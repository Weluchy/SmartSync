package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os" // Добавили для чтения переменных окружения

	"smartsync/internal/engine/repository"
	"smartsync/internal/engine/service"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
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
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
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

	// 3. Подключение к Redis и NATS
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Priority Service: Ошибка подключения к NATS: ", err)
	}
	defer nc.Close()

	// 4. ТВОЯ ИНИЦИАЛИЗАЦИЯ (теперь с правильными переменными)
	repo := repository.NewStorage(db)
	calc := service.NewCalculator(repo, rdb)

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
	select {}
}
