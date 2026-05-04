package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt" // ДОБАВИТЬ
	"log"
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
	// Очищаем кэш в Redis, чтобы фронтенд получил свежие данные
	s.redis.Del(context.Background(), fmt.Sprintf("smartsync:graph:project:%d", projectID))

	// ВНИМАНИЕ: Тема должна быть project.updated, как мы договорились с движком
	payload := fmt.Sprintf(`{"project_id": %d}`, projectID)
	s.nc.Publish("project.updated", []byte(payload))
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
	pid, _ := s.repo.GetProjectIDByTask(taskID)
	err := s.repo.UpdateTaskStatus(taskID, userID, status)

	if err == nil {
		s.triggerMathEngine(pid)

		// НОВОЕ: Отправляем событие аудита в NATS
		auditMsg := fmt.Sprintf(`{"task_id": %d, "user_id": %d, "action": "STATUS_CHANGED", "new_status": "%s"}`, taskID, userID, status)
		s.nc.Publish("audit.logs", []byte(auditMsg))
	}
	return err
}

func (s *TaskService) GetTasksByProject(projectID, userID int) ([]models.Task, error) {
	// Запрашиваем "голые" задачи из базы
	tasks, err := s.repo.GetTasksByProject(projectID, userID)
	if err != nil || len(tasks) == 0 {
		return tasks, err
	}

	// 1. Собираем уникальные ID авторов, чтобы не запрашивать одно имя дважды
	userIDs := make(map[int]bool)
	for _, t := range tasks {
		userIDs[t.UserID] = true
		if t.AssigneeID != nil {
			userIDs[*t.AssigneeID] = true
		}
	}
	var ids []int
	for id := range userIDs {
		ids = append(ids, id)
	}

	// 2. Делаем СИНХРОННЫЙ HTTP ЗАПРОС в Auth Service (порт 8081)
	reqBody, _ := json.Marshal(map[string]interface{}{"ids": ids})
	resp, err := http.Post("http://localhost:8081/internal/users/bulk", "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		log.Printf("❌ ОШИБКА: Не удалось связаться с Auth Service: %v\n", err)
	} else if resp.StatusCode != http.StatusOK {
		log.Printf("❌ ОШИБКА: Auth Service вернул статус %d\n", resp.StatusCode)
	}

	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()

		var names map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&names); err != nil {
			log.Printf("❌ ОШИБКА: Не удалось распарсить ответ от Auth Service: %v\n", err)
		}

		for i := range tasks {
			// Обогащаем автора
			authorIDStr := fmt.Sprintf("%d", tasks[i].UserID)
			if name, ok := names[authorIDStr]; ok && name != "" {
				tasks[i].CreatedByName = name
			} else {
				tasks[i].CreatedByName = "Неизвестный"
			}

			// Обогащаем исполнителя
			if tasks[i].AssigneeID != nil {
				assigneeIDStr := fmt.Sprintf("%d", *tasks[i].AssigneeID)
				if name, ok := names[assigneeIDStr]; ok && name != "" {
					tasks[i].AssigneeName = name
				}
			}
		}
	} else {
		// Fallback: если Auth Service упал или вернул ошибку
		for i := range tasks {
			tasks[i].CreatedByName = "Система (Auth Service недоступен)"
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
