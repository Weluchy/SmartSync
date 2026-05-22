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
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS description TEXT DEFAULT ''`)
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS milestone_id INT REFERENCES milestones(id) ON DELETE SET NULL`)
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS deadline_at BIGINT DEFAULT 0`)
	db.Exec(`CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		task_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		text TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS milestones (
		id SERIAL PRIMARY KEY,
		project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
		title VARCHAR(255) NOT NULL,
		description TEXT DEFAULT '',
		deadline TIMESTAMP NOT NULL DEFAULT NOW(),
		created_at TIMESTAMP DEFAULT NOW()
	)`)
	return &TaskRepository{db: db}
}

func (r *TaskRepository) DB() *sql.DB { return r.db }

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
	// ТОЧЕЧНЫЙ ФИКС: Добавили description
	err := r.db.QueryRow(`
		SELECT id, project_id, status, title, description, user_id, assignee_id, opt, real, pess 
		FROM tasks WHERE id = $1`, id).
		Scan(&t.ID, &t.ProjectID, &t.Status, &t.Title, &t.Description, &t.UserID, &t.AssigneeID, &t.Opt, &t.Real, &t.Pess)
	return &t, err
}

func (r *TaskRepository) CreateTask(t *models.Task) (int, error) {
	if _, err := r.CheckAccess(t.ProjectID, t.UserID, models.RoleWeights[models.RoleEditor]); err != nil {
		return 0, err
	}
	var id int
	// ФИКС: Правильная проверка указателя
	var assignee interface{} = nil
	if t.AssigneeID != nil && *t.AssigneeID != 0 {
		assignee = *t.AssigneeID
	}

	err := r.db.QueryRow(`INSERT INTO tasks (title, description, opt, real, pess, user_id, project_id, status, assignee_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'todo', $8) RETURNING id`,
		t.Title, t.Description, t.Opt, t.Real, t.Pess, t.UserID, t.ProjectID, assignee).Scan(&id)
	return id, err
}

func (r *TaskRepository) UpdateTask(t *models.Task) error {
	pid, err := r.GetProjectIDByTask(t.ID)
	if err != nil {
		return err
	}
	if _, err := r.CheckAccess(pid, t.UserID, models.RoleWeights[models.RoleEditor]); err != nil {
		return err
	}

	// ФИКС: Правильная проверка указателя
	var assignee interface{} = nil
	if t.AssigneeID != nil && *t.AssigneeID != 0 {
		assignee = *t.AssigneeID
	}

	_, err = r.db.Exec(`UPDATE tasks SET title = $1, description = $2, opt = $3, real = $4, pess = $5, assignee_id = $6 WHERE id = $7`,
		t.Title, t.Description, t.Opt, t.Real, t.Pess, assignee, t.ID)
	return err
}

func (r *TaskRepository) UpdateTaskStatus(taskID int, status string) error {
	_, err := r.db.Exec("UPDATE tasks SET status = $1 WHERE id = $2", status, taskID)
	return err
}

func (r *TaskRepository) DeleteTask(taskID, userID int, heal bool) error {
	pid, err := r.GetProjectIDByTask(taskID)
	if err != nil {
		return err
	}
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
	if _, err := r.CheckAccess(pid, userID, models.RoleWeights[models.RoleEditor]); err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM dependencies WHERE task_id = $1 AND depends_on_id = $2`, taskID, dependsOnID)
	return err
}

func (r *TaskRepository) ClearDependencies(projectID, userID int) error {
	if _, err := r.CheckAccess(projectID, userID, models.RoleWeights[models.RoleAdmin]); err != nil {
		return err
	}
	_, err := r.db.Exec(`DELETE FROM dependencies WHERE task_id IN (SELECT id FROM tasks WHERE project_id = $1) OR depends_on_id IN (SELECT id FROM tasks WHERE project_id = $1)`, projectID)
	return err
}

func (r *TaskRepository) GetGraphData(projectID, userID int) (*models.GraphData, error) {
	if _, err := r.CheckAccess(projectID, userID, models.RoleWeights[models.RoleViewer]); err != nil {
		return nil, err
	}
	graph := &models.GraphData{}

	// ТОЧЕЧНЫЙ ФИКС: Добавили description, user_id и assignee_id, чтобы Граф мог их показать
	query := `
		SELECT 
			id, title, description, opt, real, pess, user_id, assignee_id,
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
		rowsNodes.Scan(&t.ID, &t.Title, &t.Description, &t.Opt, &t.Real, &t.Pess, &t.UserID, &t.AssigneeID, &t.DurationHours, &t.PriorityScore, &t.Status)
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
	if _, err := r.CheckAccess(projectID, userID, models.RoleWeights[models.RoleViewer]); err != nil {
		return nil, err
	}
	var tasks []models.Task
	// ТОЧЕЧНЫЙ ФИКС: Добавили description
	query := `SELECT t.id, t.project_id, t.user_id, t.assignee_id, t.title, t.description, t.status, t.opt, t.real, t.pess, 
		COALESCE(t.duration_hours, 0.0), COALESCE(t.priority_score, 0.0), t.milestone_id, t.deadline_at 
		FROM tasks t WHERE t.project_id = $1`
	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t models.Task
		rows.Scan(&t.ID, &t.ProjectID, &t.UserID, &t.AssigneeID, &t.Title, &t.Description, &t.Status, &t.Opt, &t.Real, &t.Pess, &t.DurationHours, &t.PriorityScore, &t.MilestoneID, &t.DeadlineAt)
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) GetMilestones(projectID int) ([]models.Milestone, error) {
	rows, err := r.db.Query(`SELECT id, project_id, title, description, deadline, created_at FROM milestones WHERE project_id = $1 ORDER BY deadline ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ms []models.Milestone
	for rows.Next() {
		var m models.Milestone
		rows.Scan(&m.ID, &m.ProjectID, &m.Title, &m.Description, &m.Deadline, &m.CreatedAt)
		ms = append(ms, m)
	}
	return ms, nil
}

func (r *TaskRepository) CreateMilestone(projectID int, title string, deadline string) (*models.Milestone, error) {
	var m models.Milestone
	err := r.db.QueryRow(`
		INSERT INTO milestones (project_id, title, deadline) 
		VALUES ($1, $2, $3) RETURNING id, project_id, title, description, deadline, created_at`,
		projectID, title, deadline).Scan(&m.ID, &m.ProjectID, &m.Title, &m.Description, &m.Deadline, &m.CreatedAt)
	return &m, err
}

type ProjectStats struct {
	Total       int     `json:"total"`
	Todo        int     `json:"todo"`
	InProgress  int     `json:"in_progress"`
	Done        int     `json:"done"`
	TotalHours  float64 `json:"total_hours"`
	AvgPriority float64 `json:"avg_priority"`
}

func (r *TaskRepository) GetProjectStats(projectID int) (*ProjectStats, error) {
	stats := &ProjectStats{}
	err := r.db.QueryRow(`
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'todo'),
			COUNT(*) FILTER (WHERE status = 'in_progress'),
			COUNT(*) FILTER (WHERE status = 'done'),
			COALESCE(SUM(duration_hours), 0),
			COALESCE(AVG(priority_score), 0)
		FROM tasks WHERE project_id = $1`, projectID).Scan(
		&stats.Total, &stats.Todo, &stats.InProgress, &stats.Done,
		&stats.TotalHours, &stats.AvgPriority)
	return stats, err
}

func (r *TaskRepository) GetUserTasksCount(userID, projectID int) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE assignee_id = $1 AND project_id = $2`, userID, projectID).Scan(&count)
	return count, err
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
	// ТОЧЕЧНЫЙ ФИКС: Добавили description
	err := r.db.QueryRow("SELECT id, project_id, status, title, description FROM tasks WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&t.ID, &t.ProjectID, &t.Status, &t.Title, &t.Description)
	return &t, err
}

func (r *TaskRepository) AddComment(taskID, userID int, text string) (*models.Comment, error) {
	var c models.Comment
	err := r.db.QueryRow(`
		INSERT INTO comments (task_id, user_id, text) 
		VALUES ($1, $2, $3) RETURNING id, task_id, user_id, text, created_at`,
		taskID, userID, text).Scan(&c.ID, &c.TaskID, &c.UserID, &c.Text, &c.CreatedAt)
	return &c, err
}

func (r *TaskRepository) GetComments(taskID int) ([]models.Comment, error) {
	query := `
		SELECT c.id, c.task_id, c.user_id, COALESCE(NULLIF(u.username, ''), 'Пользователь'), c.text, c.created_at 
		FROM comments c 
		LEFT JOIN users u ON c.user_id = u.id 
		WHERE c.task_id = $1 ORDER BY c.created_at ASC`

	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Username, &c.Text, &c.CreatedAt)
		comments = append(comments, c)
	}
	return comments, nil
}
