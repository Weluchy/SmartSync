package service

import (
	"context"
	"encoding/json"
	"fmt"
	"smartsync/internal/engine/repository"

	"github.com/redis/go-redis/v9"
)

type Calculator struct {
	repo  *repository.Storage
	redis *redis.Client
}

func NewCalculator(repo *repository.Storage, rdb *redis.Client) *Calculator {
	return &Calculator{repo: repo, redis: rdb}
}

func (c *Calculator) RecalculateGraph(projectID int) {
	tasks, err := c.repo.GetProjectTasks(projectID)
	if err != nil || len(tasks) == 0 {
		return
	}
	edges, _ := c.repo.GetTaskDependencies(projectID)

	adj := make(map[int][]int)
	inDegree := make(map[int]int)
	durations := make(map[int]float64)
	earlyFinish := make(map[int]float64)

	// 1. Инициализация всех задач
	for _, t := range tasks {
		inDegree[t.ID] = 0
		if t.Status == "done" {
			durations[t.ID] = 0.0
		} else {
			durations[t.ID] = float64(t.Opt+4*t.Real+t.Pess) / 6.0
		}
		earlyFinish[t.ID] = durations[t.ID]
	}

	// 2. Построение связей (e.From = родитель, e.To = ребенок)
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
		inDegree[e.To]++
	}

	// 3. Топологическая сортировка (учитывает несколько входящих связей)
	var queue []int
	for _, t := range tasks {
		if inDegree[t.ID] == 0 {
			queue = append(queue, t.ID)
		}
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, child := range adj[curr] {
			// БЕРЕМ МАКСИМАЛЬНЫЙ ПУТЬ: если путь через текущего родителя дольше, обновляем вес ребенка
			if earlyFinish[curr]+durations[child] > earlyFinish[child] {
				earlyFinish[child] = earlyFinish[curr] + durations[child]
			}

			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	// 4. Сохранение метрик
	for _, t := range tasks {
		c.repo.UpdateTaskMetrics(t.ID, durations[t.ID], earlyFinish[t.ID])
	}

	// 5. Обновление кэша (чтобы фронт получил новые данные)
	graph, err := c.repo.GetFullGraph(projectID)
	if err == nil {
		cacheKey := fmt.Sprintf("smartsync:graph:project:%d", projectID)
		graphJSON, _ := json.Marshal(graph)
		c.redis.Set(context.Background(), cacheKey, graphJSON, 0)
	}
}
