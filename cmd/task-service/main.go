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
	// 1. Инициализация инфраструктуры
	db, err := sql.Open("postgres", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")
	if err != nil {
		log.Fatal("Postgres error:", err)
	}
	defer db.Close()

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("NATS error:", err)
	}
	defer nc.Close()

	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})

	// 2. Сборка слоев приложения (Dependency Injection)
	repo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(repo, nc, rdb)
	httpHandler := handler.NewHandler(taskService)

	// 3. Запуск сервера
	router := httpHandler.InitRoutes()
	log.Println("Task Service [Clean Architecture] запущен на порту 8080")
	router.Run(":8080")
}
