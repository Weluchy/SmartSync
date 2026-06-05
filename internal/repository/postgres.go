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
	var assignee interface{} = nil
	if t.AssigneeID != nil && *t.AssigneeID != 0 {
		assignee = *t.AssigneeID
	}

	// ПАРСИМ ВЕХУ
	var milestone interface{} = nil
	if t.MilestoneID != nil && *t.MilestoneID != 0 {
		milestone = *t.MilestoneID
	}

	// ПАРСИМ ДЕДЛАЙН
	var deadline interface{} = nil
	if t.DeadlineAt != nil && *t.DeadlineAt > 0 {
		deadline = *t.DeadlineAt
	}

	// ИСПРАВЛЕНИЕ: Добавили milestone_id в SQL запрос
	err := r.db.QueryRow(`INSERT INTO tasks (title, description, opt, real, pess, user_id, project_id, status, assignee_id, milestone_id, deadline_at, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'todo', $8, $9, $10, NOW()) RETURNING id`,
		t.Title, t.Description, t.Opt, t.Real, t.Pess, t.UserID, t.ProjectID, assignee, milestone, deadline).Scan(&id)
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

	var assignee interface{} = nil
	if t.AssigneeID != nil && *t.AssigneeID != 0 {
		assignee = *t.AssigneeID
	}

	// ПАРСИМ ВЕХУ
	var milestone interface{} = nil
	if t.MilestoneID != nil && *t.MilestoneID != 0 {
		milestone = *t.MilestoneID
	}

	// ПАРСИМ ДЕДЛАЙН
	var deadline interface{} = nil
	if t.DeadlineAt != nil && *t.DeadlineAt > 0 {
		deadline = *t.DeadlineAt
	}

	// ИСПРАВЛЕНИЕ: Добавили milestone_id = $7 и deadline_at = $8
	_, err = r.db.Exec(`UPDATE tasks SET title = $1, description = $2, opt = $3, real = $4, pess = $5, assignee_id = $6, milestone_id = $7, deadline_at = $8 WHERE id = $9`,
		t.Title, t.Description, t.Opt, t.Real, t.Pess, assignee, milestone, deadline, t.ID)
	return err
}

func (r *TaskRepository) UpdateTaskStatus(taskID int, status string) error {
	_, err := r.db.Exec("UPDATE tasks SET status = $1 WHERE id = $2", status, taskID)
	return err
}

// DeleteTask с поддержкой "сшивания" (heal) графа.
// Если heal = true, то при удалении задачи:
// 1. Находим все зависимости, где эта задача была depends_on (т.е. кто ссылался на неё)
// 2. Находим все зависимости, где эта задача была task_id (т.е. на кого ссылалась она)
// 3. Перенаправляем: все кто зависел от удаляемой — переключаем на те задачи, от которых зависела удаляемая
// 4. Удаляем старые связи и саму задачу
func (r *TaskRepository) DeleteTask(taskID, userID int, heal bool) error {
	pid, err := r.GetProjectIDByTask(taskID)
	if err != nil {
		return err
	}
	if _, err := r.CheckAccess(pid, userID, models.RoleWeights[models.RoleEditor]); err != nil {
		return err
	}

	if heal {
		// Сшивание графа:
		// 1. Кто зависел от удаляемой задачи (depends_on_id = taskID)
		rowsDependents, err := r.db.Query(`SELECT task_id FROM dependencies WHERE depends_on_id = $1`, taskID)
		if err == nil {
			var dependents []int
			for rowsDependents.Next() {
				var tid int
				rowsDependents.Scan(&tid)
				dependents = append(dependents, tid)
			}
			rowsDependents.Close()

			// 2. На кого ссылалась удаляемая задача (task_id = taskID)
			rowsParents, err := r.db.Query(`SELECT depends_on_id FROM dependencies WHERE task_id = $1`, taskID)
			var parents []int
			if err == nil {
				for rowsParents.Next() {
					var pid int
					rowsParents.Scan(&pid)
					parents = append(parents, pid)
				}
				rowsParents.Close()
			}

			// 3. Сшиваем: все dependents перенаправляем на parents
			// DELETE старых связей (где depends_on_id = taskID), INSERT новых
			if len(parents) > 0 {
				// Удаляем старые связи, где эта задача была depends_on
				r.db.Exec(`DELETE FROM dependencies WHERE depends_on_id = $1`, taskID)
				// Создаём новые: каждый dependent теперь зависит от каждого parent
				for _, dep := range dependents {
					for _, par := range parents {
						// Проверяем, не создаст ли это цикл
						r.db.Exec(`INSERT INTO dependencies (task_id, depends_on_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, dep, par)
					}
				}
			} else {
				// Нет родителей — просто удаляем связи, задачи становятся независимыми
				r.db.Exec(`DELETE FROM dependencies WHERE depends_on_id = $1`, taskID)
			}
		}

		// Удаляем все связи, где task_id = taskID (на кого ссылалась удаляемая)
		r.db.Exec(`DELETE FROM dependencies WHERE task_id = $1`, taskID)
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

	// ДОБАВЛЕНО: milestone_id и deadline_at в конце SELECT
	query := `
		SELECT 
			id, title, description, opt, real, pess, user_id, assignee_id,
			COALESCE(duration_hours, 0.0), 
			COALESCE(priority_score, 0.0), 
			status, milestone_id, deadline_at
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
		// ДОБАВЛЕНО: &t.MilestoneID, &t.DeadlineAt в конце Scan
		rowsNodes.Scan(&t.ID, &t.Title, &t.Description, &t.Opt, &t.Real, &t.Pess, &t.UserID, &t.AssigneeID, &t.DurationHours, &t.PriorityScore, &t.Status, &t.MilestoneID, &t.DeadlineAt)
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

	// ФИКС: Добавили t.created_at в конец SELECT
	query := `SELECT t.id, t.project_id, t.user_id, t.assignee_id, t.title, t.description, t.status, t.opt, t.real, t.pess, 
		COALESCE(t.duration_hours, 0.0), COALESCE(t.priority_score, 0.0), t.milestone_id, t.deadline_at, t.created_at 
		FROM tasks t WHERE t.project_id = $1`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Task
		// ФИКС: Добавили &t.CreatedAt в самый конец списка Scan
		rows.Scan(&t.ID, &t.ProjectID, &t.UserID, &t.AssigneeID, &t.Title, &t.Description, &t.Status, &t.Opt, &t.Real, &t.Pess, &t.DurationHours, &t.PriorityScore, &t.MilestoneID, &t.DeadlineAt, &t.CreatedAt)
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

func (r *TaskRepository) AddComment(taskID, userID int, text string) (*models.Comment, error) {
	var c models.Comment
	err := r.db.QueryRow(`
		INSERT INTO comments (task_id, user_id, text) 
		VALUES ($1, $2, $3) RETURNING id, task_id, user_id, text, created_at`,
		taskID, userID, text).Scan(&c.ID, &c.TaskID, &c.UserID, &c.Text, &c.CreatedAt)
	return &c, err
}

func (r *TaskRepository) DeleteMilestone(projectID, milestoneID int) error {
	_, err := r.db.Exec("DELETE FROM milestones WHERE id = $1 AND project_id = $2", milestoneID, projectID)
	return err
}

func (r *TaskRepository) GetComments(taskID int) ([]models.Comment, error) {
	query := `
		SELECT c.id, c.task_id, c.user_id, COALESCE(NULLIF(u.full_name, ''), u.username, 'Пользователь'), c.text, c.created_at 
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
