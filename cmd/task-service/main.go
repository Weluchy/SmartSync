package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"smartsync/internal/handler"
	"smartsync/internal/repository"
	"smartsync/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	var db *sql.DB
	var err error

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Ошибка БД: ", err)
	}
	defer db.Close()

	// Цикл ожидания БД (чтобы не стрелять в пустоту при запуске Docker)
	for i := 0; i < 5; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("✅ Task Service успешно подключился к Postgres!")
			break
		}
		log.Printf("⚠️ Попытка %d: БД недоступна, ждем... Ошибка: %v\n", i+1, err)
		time.Sleep(3 * time.Second)
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Ошибка NATS: ", err)
	}
	defer nc.Close()

	taskRepo := repository.NewTaskRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	taskService := service.NewTaskService(taskRepo, nc, rdb)
	projectService := service.NewProjectService(projectRepo)

	httpHandler := handler.NewHandler(taskService, projectService)
	router := httpHandler.InitRoutes()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.Println("✅ Task Service запущен на порту 8080")
	router.Run(":8080")
}
