package main

import (
	"database/sql"
	"log"
	"os" // Добавили для чтения переменных окружения

	"smartsync/internal/handler"
	"smartsync/internal/repository"
	"smartsync/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// Вспомогательная функция для чтения переменных окружения
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Читаем настройки из окружения Docker, либо используем локальные по умолчанию
	dbURL := getEnv("DATABASE_URL", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	// Подключение к Postgres
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Ошибка БД: ", err)
	}
	defer db.Close()

	// Подключение к Redis
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	// Подключение к NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Ошибка NATS: ", err)
	}
	defer nc.Close()

	// 1. Инициализация слоев БД
	taskRepo := repository.NewTaskRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	// 2. Инициализация бизнес-логики
	taskService := service.NewTaskService(taskRepo, nc, rdb)
	projectService := service.NewProjectService(projectRepo)

	// 3. Сборка контроллера и запуск
	httpHandler := handler.NewHandler(taskService, projectService)
	router := httpHandler.InitRoutes()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.Println("Task Service запущен на порту 8080")
	router.Run(":8080")
}
