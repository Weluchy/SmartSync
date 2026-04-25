package repository

import (
	"database/sql"
	"smartsync/internal/models"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

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
