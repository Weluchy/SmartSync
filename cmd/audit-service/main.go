package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Модель лога аудита (соответствует тому, что шлет Task Service)
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

	// 2. Подключение к NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("❌ Ошибка NATS:", err)
	}
	defer nc.Close()

	log.Println("🚀 Audit Service слушает топик 'audit.logs'...")

	// 3. Подписка на события
	_, err = nc.Subscribe("audit.logs", func(m *nats.Msg) {
		var entry AuditLog
		if err := json.Unmarshal(m.Data, &entry); err != nil {
			log.Printf("⚠️ Ошибка парсинга: %v", err)
			return
		}

		entry.CreatedAt = time.Now()

		// Сохраняем в MongoDB
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
