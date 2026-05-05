package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	s.nc.Publish("project.updated", []byte(fmt.Sprintf(`{"project_id": %d}`, projectID)))
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

func (s *TaskService) UpdateTaskStatus(taskID, userID int, status string) error {
	task, err := s.repo.GetByIDInternal(taskID) // ИСПРАВЛЕНИЕ: вызываем внутренний метод
	if err != nil {
		return fmt.Errorf("задача не найдена")
	}

	var role string
	s.repo.DB().QueryRow("SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2", task.ProjectID, userID).Scan(&role)

	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
	isPrivileged := role == "owner" || role == "admin"

	if !isAssignee && !isPrivileged {
		return fmt.Errorf("менять статус может только исполнитель или администратор")
	}

	err = s.repo.UpdateTaskStatus(taskID, status)
	if err == nil {
		s.triggerMathEngine(task.ProjectID)
		auditMsg := fmt.Sprintf(`{"task_id": %d, "user_id": %d, "action": "STATUS_CHANGED", "new_status": "%s"}`, taskID, userID, status)
		s.nc.Publish("audit.logs", []byte(auditMsg))
	}
	return err
}

func (s *TaskService) GetTasksByProject(projectID, userID int) ([]models.Task, error) {
	tasks, err := s.repo.GetTasksByProject(projectID, userID)
	if err != nil || len(tasks) == 0 {
		return tasks, err
	}

	userIDsMap := make(map[int]bool)
	for _, t := range tasks {
		userIDsMap[t.UserID] = true
		if t.AssigneeID != nil {
			userIDsMap[*t.AssigneeID] = true
		}
	}
	var ids []int
	for id := range userIDsMap {
		ids = append(ids, id)
	}

	reqBody, _ := json.Marshal(map[string]interface{}{"ids": ids})
	resp, err := http.Post("http://localhost:8081/internal/users/bulk", "application/json", bytes.NewBuffer(reqBody))

	names := make(map[string]string)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&names)
	}

	for i := range tasks {
		uid := fmt.Sprintf("%d", tasks[i].UserID)
		// ИСПРАВЛЕНИЕ: Проверяем val != ""
		if val, ok := names[uid]; ok && val != "" {
			tasks[i].CreatedByName = val
		} else {
			tasks[i].CreatedByName = "Неизвестный автор"
		}

		if tasks[i].AssigneeID != nil {
			aid := fmt.Sprintf("%d", *tasks[i].AssigneeID)
			if val, ok := names[aid]; ok && val != "" {
				tasks[i].AssigneeName = val
			} else {
				tasks[i].AssigneeName = "Не назначен"
			}
		}
	}
	return tasks, nil
}

func (s *TaskService) GetDependenciesByProject(projectID int) ([]models.Dependency, error) {
	return s.repo.GetDependenciesByProject(projectID)
}

func (s *TaskService) GetTaskByID(id, userID int) (*models.Task, error) {
	return s.repo.GetByID(id, userID)
}
