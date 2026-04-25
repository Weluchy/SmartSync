package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

func main() {
	connStr := "postgres://user:password@127.0.0.1:5433/smartsync?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	fmt.Println("Math Engine [PERT + DFS] запущен. Ожидание...")

	recalculateGraph := func() {
		fmt.Println("\n[Математика] Вычисление PERT и перестроение графа...")

		// 1. Вычисляем PERT для каждой задачи
		rows, _ := db.Query("SELECT id, opt, real, pess FROM tasks")
		durationsTE := make(map[int]float64) // Ожидаемое время TE

		for rows.Next() {
			var id, opt, real, pess int
			rows.Scan(&id, &opt, &real, &pess)

			// Математика: Формула PERT
			te := float64(opt+4*real+pess) / 6.0
			durationsTE[id] = te

			// Сохраняем вычисленное время обратно в БД для визуала
			db.Exec("UPDATE tasks SET duration_hours = $1 WHERE id = $2", te, id)
		}
		rows.Close()

		// 2. Строим связи
		deps, _ := db.Query("SELECT depends_on_id, task_id FROM dependencies")
		graph := make(map[int][]int)
		for deps.Next() {
			var parent, child int
			deps.Scan(&parent, &child)
			graph[parent] = append(graph[parent], child)
		}
		deps.Close()

		// 3. Расчет критического пути (теперь на основе вероятностного TE)
		memo := make(map[int]float64)
		visiting := make(map[int]bool)

		var dfs func(node int) (float64, error)
		dfs = func(node int) (float64, error) {
			if visiting[node] {
				return 0, fmt.Errorf("цикл")
			}
			if val, exists := memo[node]; exists {
				return val, nil
			}

			visiting[node] = true
			maxChildPath := 0.0
			for _, child := range graph[node] {
				childWeight, err := dfs(child)
				if err != nil {
					return 0, err
				}
				if childWeight > maxChildPath {
					maxChildPath = childWeight
				}
			}
			visiting[node] = false

			weight := durationsTE[node] + maxChildPath
			memo[node] = weight
			return weight, nil
		}

		// 4. Обновление приоритетов
		for node := range durationsTE {
			score, err := dfs(node)
			if err == nil {
				db.Exec("UPDATE tasks SET priority_score = $1 WHERE id = $2", score, node)
			}
		}

		fmt.Println("[Успех] Статистика и Граф пересчитаны.")
	}

	nc.Subscribe("graph.updated", func(m *nats.Msg) { recalculateGraph() })
	nc.Subscribe("task.created", func(m *nats.Msg) { recalculateGraph() })

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
