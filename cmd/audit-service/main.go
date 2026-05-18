package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AuditLog struct {
	TaskID    int       `json:"task_id" bson:"task_id"`
	UserID    int       `json:"user_id" bson:"user_id"`
	Action    string    `json:"action" bson:"action"`
	NewStatus string    `json:"new_status" bson:"new_status"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

func main() {
	// 1. Подключение к MongoDB
	mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("❌ Ошибка MongoDB:", err)
	}
	defer client.Disconnect(mongoCtx)

	collection := client.Database("smartsync_audit").Collection("logs")
	log.Println("✅ Подключено к MongoDB")

	// 2. HTTP-сервер для чтения логов профилем (Запуск в отдельной горутине)
	go func() {
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()

		r.GET("/audit", func(c *gin.Context) {
			// Достаем ID пользователя, который пробросил API Gateway
			userIDStr := c.GetHeader("X-User-ID")
			if userIDStr == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не идентифицирован"})
				return
			}
			userID, _ := strconv.Atoi(userIDStr)

			// Фильтр по UserID и сортировка по дате (сначала новые)
			filter := bson.M{"user_id": userID}
			findOptions := options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(50)

			cursor, err := collection.Find(context.Background(), filter, findOptions)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных логов"})
				return
			}
			defer cursor.Close(context.Background())

			var logs []AuditLog
			if err := cursor.All(context.Background(), &logs); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки логов"})
				return
			}

			if logs == nil {
				logs = []AuditLog{}
			}
			c.JSON(http.StatusOK, logs)
		})

		log.Println("🌐 HTTP API Audit Service запущен на порту 8083")
		r.Run(":8083")
	}()

	// 3. Подключение к NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("❌ Ошибка NATS:", err)
	}
	defer nc.Close()

	log.Println("🚀 Audit Service слушает топик 'audit.logs'...")

	// Подписка на события изменений
	_, err = nc.Subscribe("audit.logs", func(m *nats.Msg) {
		var entry AuditLog
		if err := json.Unmarshal(m.Data, &entry); err != nil {
			log.Printf("⚠️ Ошибка парсинга: %v", err)
			return
		}

		entry.CreatedAt = time.Now()

		_, err := collection.InsertOne(context.Background(), entry)
		if err != nil {
			log.Printf("❌ Ошибка записи в Mongo: %v", err)
		} else {
			log.Printf("📝 Записан лог: Задача %d -> %s (Юзер %d)", entry.TaskID, entry.NewStatus, entry.UserID)
		}
	})

	if err != nil {
		log.Fatal(err)
	}

	select {} // Работаем вечно
}
