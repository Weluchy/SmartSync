package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"smartsync/internal/engine/repository"
	"smartsync/internal/engine/service"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Инфраструктура
	db, err := sql.Open("postgres", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// 2. Внедрение зависимостей
	repo := repository.NewStorage(db, rdb)
	mathEngine := service.NewCalculatorService(repo)

	// 3. Подписка на события шины
	ctx := context.Background()
	log.Println("Math Engine [Clean Architecture] запущен. Ожидание событий...")

	nc.Subscribe("graph.updated", func(m *nats.Msg) { mathEngine.RecalculateAll(ctx) })
	nc.Subscribe("task.created", func(m *nats.Msg) { mathEngine.RecalculateAll(ctx) })

	// Удержание процесса
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
