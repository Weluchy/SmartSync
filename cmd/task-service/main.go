package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smartsync/internal/handler"
	"smartsync/internal/repository"
	"smartsync/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "smartsync/cmd/task-service/docs"
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

	// Добавляем Prometheus middleware для всех маршрутов
	router.Use(handler.PrometheusMiddleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("✅ Task Service запущен на порту 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка Task Service: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Task Service завершает работу...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	nc.Drain()
	log.Println("Task Service остановлен")
}
