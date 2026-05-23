package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"smartsync/internal/engine/repository"
	"smartsync/internal/engine/service"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

// Функция для чтения переменных окружения из Docker
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// 1. Читаем адреса из сети Docker (или берем локальные для тестов)
	dbURL := getEnv("DATABASE_URL", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	// 2. Подключение к Postgres
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Жесткая проверка связи с БД
	if err := db.Ping(); err != nil {
		log.Fatal("Priority Service: База данных недоступна: ", err)
	}

	// 3. Подключение к NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Priority Service: Ошибка подключения к NATS: ", err)
	}
	defer nc.Close()

	// 4. Инициализация
	repo := repository.NewStorage(db)
	calc := service.NewCalculator(repo)

	// 5. ТВОЯ ЛОГИКА ПОДПИСКИ
	nc.Subscribe("project.updated", func(m *nats.Msg) {
		var payload struct {
			ProjectID int `json:"project_id"`
		}
		json.Unmarshal(m.Data, &payload)

		if payload.ProjectID != 0 {
			calc.RecalculateGraph(payload.ProjectID)
			log.Printf("Пересчитан граф для проекта %d\n", payload.ProjectID)
		}
	})

	log.Println("Математический движок запущен и слушает события...")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Priority Service завершает работу...")
	nc.Drain()
	log.Println("Priority Service остановлен")
}
