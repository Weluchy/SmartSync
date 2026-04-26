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
	if err != nil {
		return
	}

	// Считаем PERT и обновляем БД
	for _, t := range tasks {
		duration := float64(t.Opt+4*t.Real+t.Pess) / 6.0
		priority := duration // Временно приоритет равен длительности
		c.repo.UpdateTaskMetrics(t.ID, duration, priority)
	}

	// Формируем граф и отправляем в кэш
	graph, err := c.repo.GetFullGraph(projectID)
	if err == nil {
		cacheKey := fmt.Sprintf("smartsync:graph:project:%d", projectID)
		graphJSON, _ := json.Marshal(graph)
		c.redis.Set(context.Background(), cacheKey, graphJSON, 0)
	}
}
