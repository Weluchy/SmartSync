package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (s *Storage) GetAllTasks() (map[int]*models.Task, error) {
	rows, err := s.db.Query("SELECT id, title, opt, real, pess FROM tasks")
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

func (s *Storage) GetAllEdges() ([]models.GraphEdge, error) {
	rows, err := s.db.Query("SELECT depends_on_id, task_id FROM dependencies")
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

func (s *Storage) CacheGraphSnapshot(ctx context.Context, graph models.GraphData) error {
	graphJSON, _ := json.Marshal(graph)
	return s.rdb.Set(ctx, "smartsync:graph", graphJSON, 1*time.Hour).Err()
}
