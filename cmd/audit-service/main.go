package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
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
	// ИСПРАВЛЕНО: для Docker используем имена сервисов nats и mongodb
	mongoURI := getEnv("MONGO_URI", "mongodb://mongodb:27017")
	natsURL := getEnv("NATS_URL", "nats://nats:4222")

	// 1. Подключение к MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Audit-service ошибка Mongo: ", err)
	}
	collection := client.Database("smartsync").Collection("audit_logs")

	// 2. Подключение к NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Audit-service ошибка NATS: ", err)
	}
	defer nc.Close()

	// 3. Асинхронная подписка на события задач
	nc.Subscribe("task.*", func(m *nats.Msg) {
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

	// Настройка CORS для локальных тестов, если это необходимо
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Маршрут 1: Логи конкретной задачи для TaskModal.jsx
	r.GET("/logs/:task_id", func(c *gin.Context) {
		taskIDStr := c.Param("task_id")
		taskID, err := strconv.Atoi(taskIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID задачи"})
			return
		}

		// ИСПРАВЛЕНО: учитываем особенности float64/int при десериализации JSON из NATS
		filter := bson.M{
			"$or": []bson.M{
				{"task_id": taskID},
				{"task_id": float64(taskID)},
			},
		}
		opts := options.Find().SetSort(bson.D{{"timestamp", -1}})

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

	// Маршрут 2: Общая лента аудита для UserProfile.jsx
	r.GET("/user/audit", func(c *gin.Context) {
		opts := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(30)
		cursor, err := collection.Find(context.TODO(), bson.M{}, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка БД"})
			return
		}
		defer cursor.Close(context.TODO())

		var logs []bson.M
		if err = cursor.All(context.TODO(), &logs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга логов"})
			return
		}

		if logs == nil {
			logs = []bson.M{}
		}
		c.JSON(http.StatusOK, logs)
	})

	log.Println("✅ Audit Service успешно запущен на порту 8083")
	r.Run(":8083")
}
