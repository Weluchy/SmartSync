package main

import (
	"database/sql"
	"encoding/json"
	"log"

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

	repo := repository.NewStorage(db)
	calc := service.NewCalculator(repo, rdb)

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
