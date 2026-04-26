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

func (s *TaskService) triggerMathEngine(projectID int) {
	s.redis.Del(context.Background(), fmt.Sprintf("smartsync:graph:project:%d", projectID))
	s.nc.Publish("graph.updated", []byte(fmt.Sprintf(`{"project_id": %d}`, projectID)))
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
	if err == nil {
		s.triggerMathEngine(t.ProjectID)
	}
	return id, err
}

func (s *TaskService) UpdateTask(t *models.Task) error {
	pid, _ := s.repo.GetProjectIDByTask(t.ID)
	err := s.repo.UpdateTask(t)
	if err == nil {
		s.triggerMathEngine(pid)
	}
	return err
}

func (s *TaskService) DeleteTask(taskID, userID int, heal bool) error {
	pid, _ := s.repo.GetProjectIDByTask(taskID)
	err := s.repo.DeleteTask(taskID, userID, heal)
	if err == nil {
		s.triggerMathEngine(pid)
	}
	return err
}

func (s *TaskService) CreateDependency(taskID, dependsOnID int) error {
	pid, _ := s.repo.GetProjectIDByTask(taskID)
	err := s.repo.CreateDependency(taskID, dependsOnID)
	if err == nil {
		s.triggerMathEngine(pid)
	}
	return err
}

func (s *TaskService) DeleteDependency(taskID, dependsOnID, userID int) error {
	pid, _ := s.repo.GetProjectIDByTask(taskID)
	err := s.repo.DeleteDependency(taskID, dependsOnID, userID)
	if err == nil {
		s.triggerMathEngine(pid)
	}
	return err
}

func (s *TaskService) ClearDependencies(projectID, userID int) error {
	err := s.repo.ClearDependencies(projectID, userID)
	if err == nil {
		s.triggerMathEngine(projectID)
	}
	return err
}

func (s *TaskService) GetGraph(ctx context.Context, projectID, userID int) (*models.GraphData, bool, error) {
	cacheKey := fmt.Sprintf("smartsync:graph:project:%d", projectID)
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var graph models.GraphData
		json.Unmarshal([]byte(val), &graph)
		return &graph, true, nil
	}

	graph, err := s.repo.GetGraphData(projectID, userID)
	return graph, false, err
}

// НОВЫЙ МЕТОД ДЛЯ КАНБАНА
func (s *TaskService) UpdateTaskStatus(taskID, userID int, status string) error {
	return s.repo.UpdateTaskStatus(taskID, userID, status)
}
