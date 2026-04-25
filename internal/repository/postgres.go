package repository

import (
	"database/sql"
	"smartsync/internal/models"
)

type TaskRepository struct {
	db *sql.DB
}

// Конструктор (Dependency Injection)
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CreateTask(t *models.Task) (int, error) {
	var id int
	err := r.db.QueryRow("INSERT INTO tasks (title, opt, real, pess) VALUES ($1, $2, $3, $4) RETURNING id",
		t.Title, t.Opt, t.Real, t.Pess).Scan(&id)
	return id, err
}

func (r *TaskRepository) CreateDependency(taskID, dependsOnID int) error {
	_, err := r.db.Exec("INSERT INTO dependencies (task_id, depends_on_id) VALUES ($1, $2)", taskID, dependsOnID)
	return err
}

func (r *TaskRepository) ClearDependencies() error {
	_, err := r.db.Exec("TRUNCATE dependencies")
	return err
}

// Fallback: Запрос в долгую базу, если Redis пуст
func (r *TaskRepository) GetGraphData() (*models.GraphData, error) {
	graph := &models.GraphData{}

	rowsNodes, _ := r.db.Query("SELECT id, title, opt, real, pess, duration_hours, priority_score FROM tasks")
	defer rowsNodes.Close()
	for rowsNodes.Next() {
		var t models.Task
		rowsNodes.Scan(&t.ID, &t.Title, &t.Opt, &t.Real, &t.Pess, &t.DurationHours, &t.PriorityScore)
		graph.Nodes = append(graph.Nodes, t)
	}

	rowsEdges, _ := r.db.Query("SELECT depends_on_id, task_id FROM dependencies")
	defer rowsEdges.Close()
	for rowsEdges.Next() {
		var e models.GraphEdge
		rowsEdges.Scan(&e.From, &e.To)
		graph.Edges = append(graph.Edges, e)
	}

	return graph, nil
}
