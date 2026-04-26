package main

import (
	"database/sql"
	"log"

	"smartsync/internal/handler"
	"smartsync/internal/repository"
	"smartsync/internal/service"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Подключение к Postgres
	db, err := sql.Open("postgres", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Подключение к Redis
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})

	// Подключение к NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
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

	log.Println("Task Service [Модульный] запущен на порту 8080")
	router.Run(":8080")
}
