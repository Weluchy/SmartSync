package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"smartsync/internal/engine/models"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	db  *sql.DB
	rdb *redis.Client
}

func NewStorage(db *sql.DB, rdb *redis.Client) *Storage {
	return &Storage{db: db, rdb: rdb}
}

func (s *Storage) GetAllTasks(userID int) (map[int]*models.Task, error) {
	rows, err := s.db.Query("SELECT id, title, opt, real, pess FROM tasks WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make(map[int]*models.Task)
	for rows.Next() {
		t := &models.Task{}
		rows.Scan(&t.ID, &t.Title, &t.Opt, &t.Real, &t.Pess)
		tasks[t.ID] = t
	}
	return tasks, nil
}

func (s *Storage) GetAllEdges(userID int) ([]models.GraphEdge, error) {
	rows, err := s.db.Query(`
		SELECT d.depends_on_id, d.task_id 
		FROM dependencies d 
		JOIN tasks t ON d.task_id = t.id 
		WHERE t.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []models.GraphEdge
	for rows.Next() {
		var e models.GraphEdge
		rows.Scan(&e.From, &e.To)
		edges = append(edges, e)
	}
	return edges, nil
}

func (s *Storage) UpdateTaskScores(t *models.Task) error {
	_, err := s.db.Exec("UPDATE tasks SET duration_hours = $1, priority_score = $2 WHERE id = $3",
		t.DurationHours, t.PriorityScore, t.ID)
	return err
}

func (s *Storage) CacheGraphSnapshot(ctx context.Context, userID int, graph models.GraphData) error {
	graphJSON, _ := json.Marshal(graph)
	// Сохраняем изолированно для каждого пользователя!
	cacheKey := fmt.Sprintf("smartsync:graph:user:%d", userID)
	return s.rdb.Set(ctx, cacheKey, graphJSON, 1*time.Hour).Err()
}
