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
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'todo'`)
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS assignee_id INT REFERENCES users(id) ON DELETE SET NULL`)
	return &TaskRepository{db: db}
}

func (r *TaskRepository) DB() *sql.DB { return r.db }

// CheckAccess теперь универсален и работает на системе весов
func (r *TaskRepository) CheckAccess(projectID, userID int, requiredWeight int) (string, error) {
	var role string
	err := r.db.QueryRow("SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2", projectID, userID).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("у вас нет доступа к этому проекту")
	}

	userWeight := models.RoleWeights[role]
	if userWeight < requiredWeight {
		return role, fmt.Errorf("ваша роль (%s) недостаточно высока для этого действия", role)
	}

	return role, nil
}

func (r *TaskRepository) GetProjectIDByTask(taskID int) (int, error) {
	var pid int
	err := r.db.QueryRow("SELECT project_id FROM tasks WHERE id = $1", taskID).Scan(&pid)
	return pid, err
}

func (r *TaskRepository) GetByIDInternal(id int) (*models.Task, error) {
	var t models.Task
	err := r.db.QueryRow(`
		SELECT id, project_id, status, title, user_id, assignee_id 
		FROM tasks WHERE id = $1`, id).
		Scan(&t.ID, &t.ProjectID, &t.Status, &t.Title, &t.UserID, &t.AssigneeID)
	return &t, err
}

func (r *TaskRepository) CreateTask(t *models.Task) (int, error) {
	// Для создания задачи нужен вес Editor (40)
	if _, err := r.CheckAccess(t.ProjectID, t.UserID, models.RoleWeights[models.RoleEditor]); err != nil {
		return 0, err
	}
	var id int
	err := r.db.QueryRow(`INSERT INTO tasks (title, opt, real, pess, user_id, project_id, status, assignee_id) 
		VALUES ($1, $2, $3, $4, $5, $6, 'todo', $7) RETURNING id`,
		t.Title, t.Opt, t.Real, t.Pess, t.UserID, t.ProjectID, t.AssigneeID).Scan(&id)
	return id, err
}

func (r *TaskRepository) UpdateTask(t *models.Task) error {
	pid, err := r.GetProjectIDByTask(t.ID)
	if err != nil {
		return err
	}
	// Для изменения параметров задачи (имя, оценки) нужен вес Editor (40)
	if _, err := r.CheckAccess(pid, t.UserID, models.RoleWeights[models.RoleEditor]); err != nil {
		return err
	}
	_, err = r.db.Exec(`UPDATE tasks SET title = $1, opt = $2, real = $3, pess = $4, assignee_id = $5 WHERE id = $6`,
		t.Title, t.Opt, t.Real, t.Pess, t.AssigneeID, t.ID)
	return err
}

func (r *TaskRepository) UpdateTaskStatus(taskID int, status string) error {
	// Сама смена статуса в БД не проверяет права, это делает Service
	_, err := r.db.Exec("UPDATE tasks SET status = $1 WHERE id = $2", status, taskID)
	return err
}

func (r *TaskRepository) DeleteTask(taskID, userID int, heal bool) error {
	pid, err := r.GetProjectIDByTask(taskID)
	if err != nil {
		return err
	}
	// Удаление — серьезное действие, требующее веса Admin (80)
	if _, err := r.CheckAccess(pid, userID, models.RoleWeights[models.RoleAdmin]); err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM tasks WHERE id = $1", taskID)
	return err
}

func (r *TaskRepository) CreateDependency(taskID, dependsOnID int) error {
	_, err := r.db.Exec("INSERT INTO dependencies (task_id, depends_on_id) VALUES ($1, $2)", taskID, dependsOnID)
	return err
}

func (r *TaskRepository) DeleteDependency(taskID, dependsOnID, userID int) error {
	pid, err := r.GetProjectIDByTask(taskID)
	if err != nil {
		return err
	}
	// Для изменения структуры графа нужен вес Editor (40)
	if _, err := r.CheckAccess(pid, userID, models.RoleWeights[models.RoleEditor]); err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM dependencies WHERE task_id = $1 AND depends_on_id = $2`, taskID, dependsOnID)
	return err
}

func (r *TaskRepository) ClearDependencies(projectID, userID int) error {
	// Сброс всех связей — деструктивное действие, нужен вес Admin (80)
	if _, err := r.CheckAccess(projectID, userID, models.RoleWeights[models.RoleAdmin]); err != nil {
		return err
	}
	_, err := r.db.Exec(`DELETE FROM dependencies WHERE task_id IN (SELECT id FROM tasks WHERE project_id = $1) OR depends_on_id IN (SELECT id FROM tasks WHERE project_id = $1)`, projectID)
	return err
}

func (r *TaskRepository) GetGraphData(projectID, userID int) (*models.GraphData, error) {
	// Для просмотра достаточно веса Viewer (10)
	if _, err := r.CheckAccess(projectID, userID, models.RoleWeights[models.RoleViewer]); err != nil {
		return nil, err
	}
	graph := &models.GraphData{}

	query := `
		SELECT 
			id, title, opt, real, pess, 
			COALESCE(duration_hours, 0.0), 
			COALESCE(priority_score, 0.0), 
			status 
		FROM tasks 
		WHERE project_id = $1
	`
	rowsNodes, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rowsNodes.Close()

	for rowsNodes.Next() {
		var t models.Task
		rowsNodes.Scan(&t.ID, &t.Title, &t.Opt, &t.Real, &t.Pess, &t.DurationHours, &t.PriorityScore, &t.Status)
		graph.Nodes = append(graph.Nodes, t)
	}

	rowsEdges, _ := r.db.Query(`SELECT d.depends_on_id, d.task_id FROM dependencies d JOIN tasks t ON d.task_id = t.id WHERE t.project_id = $1`, projectID)
	defer rowsEdges.Close()
	for rowsEdges.Next() {
		var e models.GraphEdge
		rowsEdges.Scan(&e.From, &e.To)
		graph.Edges = append(graph.Edges, e)
	}

	return graph, nil
}

func (r *TaskRepository) GetTasksByProject(projectID, userID int) ([]models.Task, error) {
	// Для просмотра списка достаточно веса Viewer (10)
	if _, err := r.CheckAccess(projectID, userID, models.RoleWeights[models.RoleViewer]); err != nil {
		return nil, err
	}
	var tasks []models.Task
	query := `SELECT id, project_id, user_id, assignee_id, title, status, opt, real, pess, 
		COALESCE(duration_hours, 0.0), COALESCE(priority_score, 0.0) FROM tasks WHERE project_id = $1`
	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t models.Task
		rows.Scan(&t.ID, &t.ProjectID, &t.UserID, &t.AssigneeID, &t.Title, &t.Status, &t.Opt, &t.Real, &t.Pess, &t.DurationHours, &t.PriorityScore)
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) GetDependenciesByProject(projectID int) ([]models.Dependency, error) {
	rows, err := r.db.Query(`SELECT d.task_id, d.depends_on_id FROM dependencies d JOIN tasks t ON d.task_id = t.id WHERE t.project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var deps []models.Dependency
	for rows.Next() {
		var d models.Dependency
		rows.Scan(&d.TaskID, &d.DependsOnID)
		deps = append(deps, d)
	}
	return deps, nil
}

func (r *TaskRepository) GetByID(id, userID int) (*models.Task, error) {
	var t models.Task
	err := r.db.QueryRow("SELECT id, project_id, status, title FROM tasks WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&t.ID, &t.ProjectID, &t.Status, &t.Title)
	return &t, err
}
