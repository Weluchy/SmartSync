package repository

import (
	"database/sql"
	"fmt"
	"smartsync/internal/models"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Проверяем, есть ли у пользователя доступ к проекту
func (r *TaskRepository) CheckAccess(projectID, userID int) error {
	var ownerID int
	err := r.db.QueryRow("SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		return fmt.Errorf("проект не найден или доступ запрещен")
	}
	return nil
}

func (r *TaskRepository) GetProjectIDByTask(taskID int) (int, error) {
	var pid int
	err := r.db.QueryRow("SELECT project_id FROM tasks WHERE id = $1", taskID).Scan(&pid)
	return pid, err
}

func (r *TaskRepository) CreateTask(t *models.Task) (int, error) {
	if err := r.CheckAccess(t.ProjectID, t.UserID); err != nil {
		return 0, err
	}
	var id int
	err := r.db.QueryRow("INSERT INTO tasks (title, opt, real, pess, user_id, project_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		t.Title, t.Opt, t.Real, t.Pess, t.UserID, t.ProjectID).Scan(&id)
	return id, err
}

func (r *TaskRepository) UpdateTask(t *models.Task) error {
	_, err := r.db.Exec(`
		UPDATE tasks 
		SET title = $1, opt = $2, real = $3, pess = $4 
		WHERE id = $5 AND user_id = $6
	`, t.Title, t.Opt, t.Real, t.Pess, t.ID, t.UserID)
	return err
}

func (r *TaskRepository) DeleteTask(taskID, userID int, heal bool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var ownerID int
	err = tx.QueryRow("SELECT user_id FROM tasks WHERE id = $1", taskID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		return fmt.Errorf("доступ запрещен")
	}

	if heal {
		rowsP, _ := tx.Query("SELECT depends_on_id FROM dependencies WHERE task_id = $1", taskID)
		var parents []int
		for rowsP.Next() {
			var p int
			rowsP.Scan(&p)
			parents = append(parents, p)
		}
		rowsP.Close()

		rowsC, _ := tx.Query("SELECT task_id FROM dependencies WHERE depends_on_id = $1", taskID)
		var children []int
		for rowsC.Next() {
			var c int
			rowsC.Scan(&c)
			children = append(children, c)
		}
		rowsC.Close()

		for _, child := range children {
			for _, parent := range parents {
				tx.Exec("INSERT INTO dependencies (task_id, depends_on_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", child, parent)
			}
		}
	}

	_, err = tx.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", taskID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TaskRepository) CreateDependency(taskID, dependsOnID int) error {
	_, err := r.db.Exec("INSERT INTO dependencies (task_id, depends_on_id) VALUES ($1, $2)", taskID, dependsOnID)
	return err
}

func (r *TaskRepository) DeleteDependency(taskID, dependsOnID, userID int) error {
	_, err := r.db.Exec(`
		DELETE FROM dependencies 
		WHERE task_id = $1 AND depends_on_id = $2 
		AND task_id IN (SELECT id FROM tasks WHERE user_id = $3)
	`, taskID, dependsOnID, userID)
	return err
}

func (r *TaskRepository) ClearDependencies(projectID, userID int) error {
	if err := r.CheckAccess(projectID, userID); err != nil {
		return err
	}
	_, err := r.db.Exec(`
		DELETE FROM dependencies 
		WHERE task_id IN (SELECT id FROM tasks WHERE project_id = $1) 
		   OR depends_on_id IN (SELECT id FROM tasks WHERE project_id = $1)
	`, projectID)
	return err
}

func (r *TaskRepository) GetGraphData(projectID, userID int) (*models.GraphData, error) {
	if err := r.CheckAccess(projectID, userID); err != nil {
		return nil, err
	}
	graph := &models.GraphData{}

	rowsNodes, _ := r.db.Query("SELECT id, title, opt, real, pess, duration_hours, priority_score FROM tasks WHERE project_id = $1", projectID)
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
		WHERE t.project_id = $1`, projectID)
	defer rowsEdges.Close()
	for rowsEdges.Next() {
		var e models.GraphEdge
		rowsEdges.Scan(&e.From, &e.To)
		graph.Edges = append(graph.Edges, e)
	}

	return graph, nil
}
