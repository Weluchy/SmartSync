package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

type taskForCheck struct {
	ID         int
	Title      string
	ProjectID  int
	Status     string
	CreatedAt  time.Time
	DurHours   float64
	DeadlineAt sql.NullInt64 // Читаем наше BIGINT поле
}

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	checkInterval := 60 // секунд, можно переопределить через INTERVAL_SECONDS env

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Deadline Service: ошибка БД:", err)
	}
	defer db.Close()

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Deadline Service: ошибка NATS:", err)
	}
	defer nc.Close()

	log.Printf("Deadline Checker запущен, проверка каждые %d сек", checkInterval)

	ticker := time.NewTicker(time.Duration(checkInterval) * time.Second)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			rows, err := db.Query(`
				SELECT id, title, project_id, status, created_at, duration_hours, deadline_at 
				FROM tasks 
				WHERE status != 'done' AND (duration_hours > 0 OR deadline_at > 0)
			`)
			if err != nil {
				log.Println("Deadline: ошибка запроса:", err)
				continue
			}

			for rows.Next() {
				var t taskForCheck
				rows.Scan(&t.ID, &t.Title, &t.ProjectID, &t.Status, &t.CreatedAt, &t.DurHours, &t.DeadlineAt)

				// Выбираем: жесткий дедлайн или PERT-дедлайн
				var deadline time.Time
				if t.DeadlineAt.Valid && t.DeadlineAt.Int64 > 0 {
					deadline = time.UnixMilli(t.DeadlineAt.Int64)
				} else {
					deadline = t.CreatedAt.Add(time.Duration(t.DurHours) * time.Hour)
				}

				timeLeft := time.Until(deadline)

				// Уведомления за 24ч, 6ч, 1ч до дедлайна
				if timeLeft > 0 && timeLeft < 25*time.Hour && timeLeft > 23*time.Hour {
					publishDeadline(nc, t, "24ч", timeLeft)
				} else if timeLeft > 0 && timeLeft < 7*time.Hour && timeLeft > 5*time.Hour {
					publishDeadline(nc, t, "6ч", timeLeft)
				} else if timeLeft > 0 && timeLeft < 2*time.Hour && timeLeft > 30*time.Minute {
					publishDeadline(nc, t, "1ч", timeLeft)
				} else if timeLeft < 0 && timeLeft > -1*time.Hour {
					publishDeadline(nc, t, "просрочена", timeLeft)
				}
			}
			rows.Close()

		case <-quit:
			log.Println("Deadline Checker завершает работу...")
			nc.Drain()
			return
		}
	}
}

func publishDeadline(nc *nats.Conn, t taskForCheck, label string, timeLeft time.Duration) {
	msg, _ := json.Marshal(map[string]interface{}{
		"type":       "deadline",
		"task_id":    t.ID,
		"project_id": t.ProjectID,
		"title":      t.Title,
		"label":      label,
		"time_left":  timeLeft.String(),
		"timestamp":  time.Now(),
	})
	nc.Publish("deadline.alert", msg)
	log.Printf("🔔 Дедлайн %s: задача #%d «%s» — осталось %v", label, t.ID, t.Title, timeLeft)
}
