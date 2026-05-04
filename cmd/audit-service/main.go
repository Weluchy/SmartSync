package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Модель лога аудита
type AuditLog struct {
	TaskID    int    `json:"task_id"`
	UserID    int    `json:"user_id"`
	Action    string `json:"action"`
	NewStatus string `json:"new_status"`
	Timestamp string `json:"timestamp"`
}

func main() {
	// Подключаемся к брокеру NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("Не удалось подключиться к NATS:", err)
	}
	defer nc.Close()

	log.Println("Audit Service запущен и ожидает события в топике 'audit.logs'...")

	// Слушаем топик аудита
	_, err = nc.Subscribe("audit.logs", func(m *nats.Msg) {
		var logEntry AuditLog
		if err := json.Unmarshal(m.Data, &logEntry); err == nil {
			logEntry.Timestamp = time.Now().Format(time.RFC3339)

			// Для диплома: здесь можно подключить MongoDB и делать collection.InsertOne(logEntry)
			// Сейчас просто красиво выводим в консоль
			log.Printf("[AUDIT] Пользователь ID:%d изменил статус задачи ID:%d на '%s' в %s\n",
				logEntry.UserID, logEntry.TaskID, logEntry.NewStatus, logEntry.Timestamp)
		}
	})

	if err != nil {
		log.Fatal(err)
	}

	// Блокируем горутину, чтобы сервис работал вечно
	select {}
}
