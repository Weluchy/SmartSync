package main

import (
	"context"
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
	"github.com/redis/go-redis/v9"
)

func main() {
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

	repo := repository.NewStorage(db, rdb)
	mathEngine := service.NewCalculatorService(repo)

	ctx := context.Background()
	log.Println("Math Engine [Multi-Tenant Edition] запущен. Ожидание событий...")

	// Правильные подписки с парсингом user_id
	nc.Subscribe("graph.updated", func(m *nats.Msg) {
		var data struct {
			UserID int `json:"user_id"`
		}
		json.Unmarshal(m.Data, &data)
		if data.UserID > 0 {
			mathEngine.RecalculateAll(ctx, data.UserID)
		}
	})

	nc.Subscribe("task.created", func(m *nats.Msg) {
		var data struct {
			UserID int `json:"user_id"`
		}
		json.Unmarshal(m.Data, &data)
		if data.UserID > 0 {
			mathEngine.RecalculateAll(ctx, data.UserID)
		}
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
