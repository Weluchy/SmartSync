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

	// МГНОВЕННАЯ ИНВАЛИДАЦИЯ КЭША
	s.redis.Del(context.Background(), fmt.Sprintf("smartsync:graph:user:%d", t.UserID))

	payload := fmt.Sprintf(`{"task_id": %d, "user_id": %d}`, id, t.UserID)
	s.nc.Publish("task.created", []byte(payload))
	return id, nil
}

func (s *TaskService) CreateDependency(taskID, dependsOnID, userID int) error {
	err := s.repo.CreateDependency(taskID, dependsOnID)
	if err == nil {
		// МГНОВЕННАЯ ИНВАЛИДАЦИЯ КЭША
		s.redis.Del(context.Background(), fmt.Sprintf("smartsync:graph:user:%d", userID))
		s.nc.Publish("graph.updated", []byte(fmt.Sprintf(`{"user_id": %d}`, userID)))
	}
	return err
}

func (s *TaskService) ClearDependencies(userID int) error {
	err := s.repo.ClearDependencies(userID)
	if err == nil {
		// МГНОВЕННАЯ ИНВАЛИДАЦИЯ КЭША
		s.redis.Del(context.Background(), fmt.Sprintf("smartsync:graph:user:%d", userID))
		s.nc.Publish("graph.updated", []byte(fmt.Sprintf(`{"user_id": %d}`, userID)))
	}
	return err
}

func (s *TaskService) GetGraph(ctx context.Context, userID int) (*models.GraphData, bool, error) {
	cacheKey := fmt.Sprintf("smartsync:graph:user:%d", userID)
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var graph models.GraphData
		json.Unmarshal([]byte(val), &graph)
		return &graph, true, nil
	}

	graph, err := s.repo.GetGraphData(userID)
	return graph, false, err
}

func (s *TaskService) DeleteTask(taskID, userID int, heal bool) error {
	err := s.repo.DeleteTask(taskID, userID, heal)
	if err == nil {
		s.redis.Del(context.Background(), fmt.Sprintf("smartsync:graph:user:%d", userID))
		s.nc.Publish("graph.updated", []byte(fmt.Sprintf(`{"user_id": %d}`, userID)))
	}
	return err
}

func (s *TaskService) DeleteDependency(taskID, dependsOnID, userID int) error {
	err := s.repo.DeleteDependency(taskID, dependsOnID, userID)
	if err == nil {
		s.redis.Del(context.Background(), fmt.Sprintf("smartsync:graph:user:%d", userID))
		s.nc.Publish("graph.updated", []byte(fmt.Sprintf(`{"user_id": %d}`, userID)))
	}
	return err
}

func (s *TaskService) UpdateTask(t *models.Task) error {
	err := s.repo.UpdateTask(t)
	if err == nil {
		s.redis.Del(context.Background(), fmt.Sprintf("smartsync:graph:user:%d", t.UserID))
		s.nc.Publish("graph.updated", []byte(fmt.Sprintf(`{"user_id": %d}`, t.UserID)))
	}
	return err
}
