package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"smartsync/internal/models"
	"smartsync/internal/repository"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

type TaskService struct {
	repo    *repository.TaskRepository
	nc      *nats.Conn
	redis   *redis.Client
	authURL string
	cb      *gobreaker.CircuitBreaker
}

func NewTaskService(repo *repository.TaskRepository, nc *nats.Conn, rdb *redis.Client) *TaskService {
	st := gobreaker.Settings{
		Name:        "Auth-Service-CB",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     8 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.5
		},
	}
	return &TaskService{
		repo:    repo,
		nc:      nc,
		redis:   rdb,
		authURL: "http://auth-service:8081",
		cb:      gobreaker.NewCircuitBreaker(st),
	}
}

func (s *TaskService) triggerMathEngine(projectID int) {
	s.redis.Del(context.Background(), fmt.Sprintf("smartsync:graph:project:%d", projectID))
	s.nc.Publish("project.updated", []byte(fmt.Sprintf(`{"project_id": %d}`, projectID)))
}

// ТОЧЕЧНЫЙ ФИКС: Твоя логика вынесена сюда, чтобы Граф тоже её получал
func (s *TaskService) enrichTasks(tasks []models.Task) []models.Task {
	if len(tasks) == 0 {
		return tasks
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
	var names map[string]string

	_, cbErr := s.cb.Execute(func() (interface{}, error) {
		resp, err := http.Post(s.authURL+"/internal/users/bulk", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, fmt.Errorf("auth-service недоступен: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("auth-service вернул статус %d", resp.StatusCode)
		}
		json.NewDecoder(resp.Body).Decode(&names)
		return nil, nil
	})

	if cbErr != nil {
		log.Printf("⚠️ Circuit Breaker: Auth Service недоступен: %v", cbErr)
		names = make(map[string]string)
	}

	for i := range tasks {
		uid := fmt.Sprintf("%d", tasks[i].UserID)
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
	return tasks
}

func (s *TaskService) CreateTask(t *models.Task) (int, error) {
	if t.Opt <= 0 {
		t.Opt = 1
	}
	if t.Real <= 0 {
		t.Real = 2
	}
	if t.Pess <= 0 {
		t.Pess = 3
	}
	if t.Opt > t.Real {
		t.Real = t.Opt
	}
	if t.Real > t.Pess {
		t.Pess = t.Real
	}

	id, err := s.repo.CreateTask(t)
	if err == nil {
		s.triggerMathEngine(t.ProjectID)
		auditMsg, _ := json.Marshal(map[string]interface{}{
			"task_id": id,
			"action":  "created",
			"user_id": t.UserID,
			"summary": fmt.Sprintf("Создана задача «%s»", t.Title),
			"payload": map[string]interface{}{
				"title": t.Title,
				"opt":   t.Opt, "real": t.Real, "pess": t.Pess,
			},
		})
		s.nc.Publish("task.audit", auditMsg)
	}
	return id, err
}

func (s *TaskService) UpdateTask(t *models.Task) error {
	pid, _ := s.repo.GetProjectIDByTask(t.ID)
	if _, err := s.repo.CheckAccess(pid, t.UserID, models.RoleWeights[models.RoleEditor]); err != nil {
		return err
	}

	oldTask, oldErr := s.repo.GetByIDInternal(t.ID)
	err := s.repo.UpdateTask(t)
	if err == nil {
		s.triggerMathEngine(pid)

		changes := []string{}
		if oldErr == nil {
			if oldTask.Title != t.Title {
				changes = append(changes, fmt.Sprintf("название: «%s» → «%s»", oldTask.Title, t.Title))
			}
			// ТОЧЕЧНЫЙ ФИКС: логируем изменение описания
			if oldTask.Description != t.Description {
				changes = append(changes, "изменено описание задачи")
			}
			if oldTask.Opt != t.Opt {
				changes = append(changes, fmt.Sprintf("оптимистичная оценка: %dч → %dч", oldTask.Opt, t.Opt))
			}
			if oldTask.Real != t.Real {
				changes = append(changes, fmt.Sprintf("реалистичная оценка: %dч → %dч", oldTask.Real, t.Real))
			}
			if oldTask.Pess != t.Pess {
				changes = append(changes, fmt.Sprintf("пессимистичная оценка: %dч → %dч", oldTask.Pess, t.Pess))
			}
			if (oldTask.AssigneeID == nil && t.AssigneeID != nil) ||
				(oldTask.AssigneeID != nil && t.AssigneeID == nil) ||
				(oldTask.AssigneeID != nil && t.AssigneeID != nil && *oldTask.AssigneeID != *t.AssigneeID) {
				changes = append(changes, "изменён исполнитель")
			}
			if oldTask.Status != t.Status && t.Status != "" {
				changes = append(changes, fmt.Sprintf("статус: %s → %s", oldTask.Status, t.Status))
			}
		}

		summary := fmt.Sprintf("Обновлена задача «%s»", t.Title)
		if len(changes) > 0 {
			summary = fmt.Sprintf("Обновлена задача «%s»: %s", t.Title, changes[0])
			if len(changes) > 1 {
				summary += fmt.Sprintf(" (+%d изменений)", len(changes)-1)
			}
		}

		auditMsg, _ := json.Marshal(map[string]interface{}{
			"task_id": t.ID,
			"action":  "updated",
			"user_id": t.UserID,
			"summary": summary,
			"changes": changes,
			"payload": map[string]interface{}{
				"title": t.Title, "opt": t.Opt, "real": t.Real, "pess": t.Pess,
			},
		})
		s.nc.Publish("task.audit", auditMsg)
	}
	return err
}

func (s *TaskService) DeleteTask(taskID, userID int, heal bool) error {
	pid, err := s.repo.GetProjectIDByTask(taskID)
	if err != nil {
		return fmt.Errorf("задача не найдена")
	}
	if _, err = s.repo.CheckAccess(pid, userID, models.RoleWeights[models.RoleEditor]); err != nil {
		return err
	}
	err = s.repo.DeleteTask(taskID, userID, heal)
	if err == nil {
		s.triggerMathEngine(pid)
	}
	return err
}

func (s *TaskService) CreateDependency(taskID, dependsOnID, userID int) error {
	pid, err := s.repo.GetProjectIDByTask(taskID)
	if err != nil {
		return fmt.Errorf("задача не найдена")
	}
	if _, err = s.repo.CheckAccess(pid, userID, models.RoleWeights[models.RoleEditor]); err != nil {
		return err
	}
	err = s.repo.CreateDependency(taskID, dependsOnID)
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
	if _, err := s.repo.CheckAccess(projectID, userID, models.RoleWeights[models.RoleViewer]); err != nil {
		return nil, false, err
	}

	cacheKey := fmt.Sprintf("smartsync:graph:project:%d", projectID)
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var graph models.GraphData
		if err := json.Unmarshal([]byte(val), &graph); err == nil {
			graph.Nodes = s.enrichTasks(graph.Nodes) // ТОЧЕЧНЫЙ ФИКС: Обогащаем кэш
			return &graph, true, nil
		}
	}

	graph, err := s.repo.GetGraphData(projectID, userID)
	if err == nil && graph != nil {
		graph.Nodes = s.enrichTasks(graph.Nodes) // ТОЧЕЧНЫЙ ФИКС: Обогащаем БД
	}
	return graph, false, err
}

func (s *TaskService) UpdateTaskStatus(taskID, userID int, status string) error {
	task, err := s.repo.GetByIDInternal(taskID)
	if err != nil {
		return fmt.Errorf("задача не найдена")
	}
	role, err := s.repo.CheckAccess(task.ProjectID, userID, models.RoleWeights[models.RoleViewer])
	if err != nil {
		return err
	}

	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
	isManager := models.RoleWeights[role] >= models.RoleWeights[models.RoleAdmin]
	if !isAssignee && !isManager {
		return fmt.Errorf("вы не исполнитель этой задачи и не администратор проекта")
	}

	err = s.repo.UpdateTaskStatus(taskID, status)
	if err == nil {
		s.triggerMathEngine(task.ProjectID)
		auditMsg, _ := json.Marshal(map[string]interface{}{
			"task_id": taskID,
			"action":  "status_changed",
			"user_id": userID,
			"summary": fmt.Sprintf("Статус задачи «%s» изменён: %s → %s", task.Title, task.Status, status),
			"changes": []string{fmt.Sprintf("статус: %s → %s", task.Status, status)},
			"payload": map[string]interface{}{
				"old_status": task.Status, "new_status": status,
			},
		})
		s.nc.Publish("task.audit", auditMsg)
	}
	return err
}

func (s *TaskService) GetTasksByProject(projectID, userID int) ([]models.Task, error) {
	tasks, err := s.repo.GetTasksByProject(projectID, userID)
	if err != nil || len(tasks) == 0 {
		return tasks, err
	}
	return s.enrichTasks(tasks), nil // ТОЧЕЧНЫЙ ФИКС: Использована общая функция
}

func (s *TaskService) GetDependenciesByProject(projectID int) ([]models.Dependency, error) {
	return s.repo.GetDependenciesByProject(projectID)
}

func (s *TaskService) GetTaskByID(id, userID int) (*models.Task, error) {
	return s.repo.GetByID(id, userID)
}

func (s *TaskService) GetMilestones(projectID int) ([]models.Milestone, error) {
	return s.repo.GetMilestones(projectID)
}

func (s *TaskService) CreateMilestone(projectID int, title string, deadline string) (*models.Milestone, error) {
	return s.repo.CreateMilestone(projectID, title, deadline)
}

func (s *TaskService) GetProjectStats(projectID int) (*repository.ProjectStats, error) {
	return s.repo.GetProjectStats(projectID)
}

func (s *TaskService) AddComment(taskID, userID int, text string) (*models.Comment, error) {
	pid, err := s.repo.GetProjectIDByTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("задача не найдена")
	}
	if _, err := s.repo.CheckAccess(pid, userID, models.RoleWeights[models.RoleViewer]); err != nil {
		return nil, err
	}
	return s.repo.AddComment(taskID, userID, text)
}

func (s *TaskService) GetComments(taskID int) ([]models.Comment, error) {
	return s.repo.GetComments(taskID)
}
