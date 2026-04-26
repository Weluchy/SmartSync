package repository

import (
	"database/sql"
	"smartsync/internal/models"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	// Enterprise-паттерн: Автоматическая миграция БД при старте сервиса

	// 1. Создаем таблицу задач (с колонкой user_id)
	db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		title VARCHAR(255) NOT NULL,
		opt INTEGER DEFAULT 1,
		real INTEGER DEFAULT 1,
		pess INTEGER DEFAULT 1,
		duration_hours FLOAT DEFAULT 0,
		priority_score FLOAT DEFAULT 0
	)`)

	// 2. Создаем таблицу зависимостей (связи графа)
	db.Exec(`CREATE TABLE IF NOT EXISTS dependencies (
		task_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
		depends_on_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
		UNIQUE(task_id, depends_on_id)
	)`)

	return &TaskRepository{db: db}
}

// ... дальше идут твои функции CreateTask, CreateDependency и т.д. (оставь их без изменений)

func (r *TaskRepository) CreateTask(t *models.Task) (int, error) {
	var id int
	err := r.db.QueryRow("INSERT INTO tasks (title, opt, real, pess, user_id) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		t.Title, t.Opt, t.Real, t.Pess, t.UserID).Scan(&id)
	return id, err
}

func (r *TaskRepository) CreateDependency(taskID, dependsOnID int) error {
	_, err := r.db.Exec("INSERT INTO dependencies (task_id, depends_on_id) VALUES ($1, $2)", taskID, dependsOnID)
	return err
}

// Удаляем связи ТОЛЬКО для задач конкретного пользователя
func (r *TaskRepository) ClearDependencies(userID int) error {
	_, err := r.db.Exec(`DELETE FROM dependencies WHERE task_id IN (SELECT id FROM tasks WHERE user_id = $1)`, userID)
	return err
}

// Достаем граф ТОЛЬКО конкретного пользователя
func (r *TaskRepository) GetGraphData(userID int) (*models.GraphData, error) {
	graph := &models.GraphData{}

	rowsNodes, _ := r.db.Query("SELECT id, title, opt, real, pess, duration_hours, priority_score FROM tasks WHERE user_id = $1", userID)
	defer rowsNodes.Close()
	for rowsNodes.Next() {
		var t models.Task
		rowsNodes.Scan(&t.ID, &t.Title, &t.Opt, &t.Real, &t.Pess, &t.DurationHours, &t.PriorityScore)
		graph.Nodes = append(graph.Nodes, t)
	}

	rowsEdges, _ := r.db.Query(`
		SELECT d.depends_on_id, d.task_id 
		FROM dependencies d 
		JOIN tasks t ON d.task_id = t.id 
		WHERE t.user_id = $1`, userID)
	defer rowsEdges.Close()
	for rowsEdges.Next() {
		var e models.GraphEdge
		rowsEdges.Scan(&e.From, &e.To)
		graph.Edges = append(graph.Edges, e)
	}

	return graph, nil
}
