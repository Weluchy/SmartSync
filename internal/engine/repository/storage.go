package repository

import (
	"database/sql"
	"smartsync/internal/engine/models"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) GetProjectTasks(projectID int) ([]models.Task, error) {
	rows, err := s.db.Query("SELECT id, opt, real, pess FROM tasks WHERE project_id = $1", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		rows.Scan(&t.ID, &t.Opt, &t.Real, &t.Pess)
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *Storage) UpdateTaskMetrics(id int, duration, priority float64) error {
	_, err := s.db.Exec("UPDATE tasks SET duration_hours = $1, priority_score = $2 WHERE id = $3", duration, priority, id)
	return err
}

func (s *Storage) GetFullGraph(projectID int) (*models.GraphData, error) {
	graph := &models.GraphData{}

	rowsNodes, _ := s.db.Query("SELECT id, title, opt, real, pess, duration_hours, priority_score FROM tasks WHERE project_id = $1", projectID)
	defer rowsNodes.Close()
	for rowsNodes.Next() {
		var t models.Task
		rowsNodes.Scan(&t.ID, &t.Title, &t.Opt, &t.Real, &t.Pess, &t.DurationHours, &t.PriorityScore)
		graph.Nodes = append(graph.Nodes, t)
	}

	rowsEdges, _ := s.db.Query(`
		SELECT d.depends_on_id, d.task_id 
		FROM dependencies d 
		JOIN tasks t ON d.task_id = t.id 
		WHERE t.project_id = $1`, projectID)
	defer rowsEdges.Close()
	for rowsEdges.Next() {
		var e models.GraphEdge
		rowsEdges.Scan(&e.From, &e.To)
		graph.Edges = append(graph.Edges, e)
	}

	return graph, nil
}
