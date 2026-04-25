package service

import (
	"context"
	"encoding/json"
	"fmt"
	"smartsync/internal/models"
	"smartsync/internal/repository"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

type TaskService struct {
	repo  *repository.TaskRepository
	nc    *nats.Conn
	redis *redis.Client
}

func NewTaskService(repo *repository.TaskRepository, nc *nats.Conn, rdb *redis.Client) *TaskService {
	return &TaskService{repo: repo, nc: nc, redis: rdb}
}

func (s *TaskService) CreateTask(t *models.Task) (int, error) {
	// Валидация
	if t.Opt == 0 {
		t.Opt = 1
	}
	if t.Real == 0 {
		t.Real = 1
	}
	if t.Pess == 0 {
		t.Pess = 1
	}

	id, err := s.repo.CreateTask(t)
	if err != nil {
		return 0, err
	}

	// Event-Driven: оповещаем математический движок
	s.nc.Publish("task.created", []byte(fmt.Sprintf("%d", id)))
	return id, nil
}

func (s *TaskService) CreateDependency(taskID, dependsOnID int) error {
	err := s.repo.CreateDependency(taskID, dependsOnID)
	if err == nil {
		s.nc.Publish("graph.updated", []byte("updated"))
	}
	return err
}

func (s *TaskService) ClearDependencies() error {
	err := s.repo.ClearDependencies()
	if err == nil {
		s.nc.Publish("graph.updated", []byte("reset"))
	}
	return err
}

// CQRS Query: Паттерн Cache-Aside
func (s *TaskService) GetGraph(ctx context.Context) (*models.GraphData, bool, error) {
	// 1. Попытка сверхбыстрого чтения из Redis
	val, err := s.redis.Get(ctx, "smartsync:graph").Result()
	if err == nil {
		var graph models.GraphData
		json.Unmarshal([]byte(val), &graph)
		return &graph, true, nil // true = взято из кэша
	}

	// 2. Если в кэше пусто — фоллбэк на Postgres
	graph, err := s.repo.GetGraphData()
	return graph, false, err
}
