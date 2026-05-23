package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	mongoURI := getEnv("MONGO_URI", "mongodb://mongodb:27017")
	natsURL := getEnv("NATS_URL", "nats://nats:4222")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Audit-service ошибка Mongo: ", err)
	}
	collection := client.Database("smartsync").Collection("audit_logs")

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Audit-service ошибка NATS: ", err)
	}
	defer nc.Close()

	nc.Subscribe("task.audit", func(m *nats.Msg) {
		var event map[string]interface{}
		if err := json.Unmarshal(m.Data, &event); err == nil {
			event["timestamp"] = time.Now()
			_, insErr := collection.InsertOne(context.TODO(), event)
			if insErr != nil {
				log.Println("Ошибка записи лога в Mongo:", insErr)
			} else {
				log.Println("Успешно сохранен лог для задачи:", event["task_id"])
			}
		}
	})

	r := gin.Default()

	r.GET("/logs/:task_id", func(c *gin.Context) {
		taskIDStr := c.Param("task_id")
		taskID, err := strconv.Atoi(taskIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID задачи"})
			return
		}

		filter := bson.M{
			"$or": []bson.M{
				{"task_id": taskID},
				{"task_id": float64(taskID)},
			},
		}
		opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

		cursor, err := collection.Find(context.TODO(), filter, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
			return
		}
		defer cursor.Close(context.TODO())

		var logs []bson.M
		if err = cursor.All(context.TODO(), &logs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения логов"})
			return
		}

		if logs == nil {
			logs = []bson.M{}
		}
		c.JSON(http.StatusOK, logs)
	})

	r.GET("/user/audit", func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		userIDInt, _ := strconv.Atoi(userIDStr)

		// ФИКС: фильтруем историю только для текущего юзера
		filter := bson.M{"$or": []bson.M{{"user_id": userIDStr}, {"user_id": userIDInt}}}

		opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(30)
		cursor, err := collection.Find(context.TODO(), filter, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка БД"})
			return
		}
		defer cursor.Close(context.TODO())

		var logs []bson.M
		if err = cursor.All(context.TODO(), &logs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга"})
			return
		}
		if logs == nil {
			logs = []bson.M{}
		}
		c.JSON(http.StatusOK, logs)
	})

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":8083",
		Handler: r,
	}

	go func() {
		log.Println("✅ Audit Service успешно запущен на порту 8083")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка Audit Service: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Audit Service завершает работу...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	nc.Drain()
	client.Disconnect(ctx)
	log.Println("Audit Service остановлен")
}
