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
	mongoURI := getEnv("MONGO_URI", "mongodb://127.0.0.1:27017")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	// 1. Подключение к MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database("smartsync").Collection("audit_logs")

	// 2. Подключение к NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// 3. Подписка на события (Асинхронная запись)
	nc.Subscribe("task.*", func(m *nats.Msg) {
		var event map[string]interface{}
		if err := json.Unmarshal(m.Data, &event); err == nil {
			event["timestamp"] = time.Now()
			collection.InsertOne(context.TODO(), bson.M{
				"action":    event["action"],
				"task_id":   event["task_id"],
				"payload":   event["payload"],
				"timestamp": event["timestamp"],
			})
			log.Println("Записан лог аудита для задачи:", event["task_id"])
		}
	})

	// 4. HTTP-сервер для выдачи логов фронтенду
	r := gin.Default()

	r.GET("/logs/:task_id", func(c *gin.Context) {
		taskIDStr := c.Param("task_id")
		taskID, err := strconv.Atoi(taskIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID задачи"})
			return
		}

		// ИССЛЕДОВАНИЕ: Защита от несовпадения типов int и float64 в MongoDB
		filter := bson.M{"task_id": bson.M{"$in": []interface{}{taskID, float64(taskID)}}}
		opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}) // -1 означает новые сверху

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

	// ИССЛЕДОВАНИЕ: Маршрут для страницы профиля (Лента активности)
	r.GET("/user/audit", func(c *gin.Context) {
		// Просто берем 20 самых свежих логов из базы
		opts := options.Find().SetSort(bson.D{{"timestamp", -1}}).SetLimit(20)

		cursor, err := collection.Find(context.TODO(), bson.M{}, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка БД"})
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

	log.Println("✅ Audit Service запущен на порту 8083")
	r.Run(":8083") // Теперь сервер держит процесс, select{} больше не нужен
}
