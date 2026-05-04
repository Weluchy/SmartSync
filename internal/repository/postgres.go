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
	// Автоматическая миграция: добавляем колонку status, если её нет
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'todo'`)
	db.Exec(`UPDATE tasks SET status = 'todo' WHERE status IS NULL`)
	return &TaskRepository{db: db}
}

func (r *TaskRepository) CheckAccess(projectID, userID int, requireWrite bool) error {
	var role string
	err := r.db.QueryRow("SELECT role FROM project_members WHERE project_id = $1 AND user_id = $2", projectID, userID).Scan(&role)
	if err != nil {
		return fmt.Errorf("проект не найден или доступ запрещен")
	}
	if requireWrite && role != "owner" && role != "editor" {
		return fmt.Errorf("недостаточно прав для редактирования")
	}
	return nil
}

func (r *TaskRepository) GetProjectIDByTask(taskID int) (int, error) {
	var pid int
	err := r.db.QueryRow("SELECT project_id FROM tasks WHERE id = $1", taskID).Scan(&pid)
	return pid, err
}

func (r *TaskRepository) CreateTask(t *models.Task) (int, error) {
	if err := r.CheckAccess(t.ProjectID, t.UserID, true); err != nil {
		return 0, err
	}
	var id int
	// При создании задача автоматически получает статус 'todo'
	err := r.db.QueryRow("INSERT INTO tasks (title, opt, real, pess, user_id, project_id, status) VALUES ($1, $2, $3, $4, $5, $6, 'todo') RETURNING id",
		t.Title, t.Opt, t.Real, t.Pess, t.UserID, t.ProjectID).Scan(&id)
	return id, err
}

func (r *TaskRepository) UpdateTask(t *models.Task) error {
	pid, err := r.GetProjectIDByTask(t.ID)
	if err != nil {
		return err
	}
	if err := r.CheckAccess(pid, t.UserID, true); err != nil {
		return err
	}

	_, err = r.db.Exec("UPDATE tasks SET title = $1, opt = $2, real = $3, pess = $4 WHERE id = $5", t.Title, t.Opt, t.Real, t.Pess, t.ID)
	return err
}

// НОВЫЙ МЕТОД: Изменение только статуса (для Канбана)
func (r *TaskRepository) UpdateTaskStatus(taskID, userID int, status string) error {
	pid, err := r.GetProjectIDByTask(taskID)
	if err != nil {
		return err
	}
	if err := r.CheckAccess(pid, userID, true); err != nil {
		return err
	}

	_, err = r.db.Exec("UPDATE tasks SET status = $1 WHERE id = $2", status, taskID)
	return err
}

func (r *TaskRepository) DeleteTask(taskID, userID int, heal bool) error {
	pid, err := r.GetProjectIDByTask(taskID)
	if err != nil {
		return err
	}
	if err := r.CheckAccess(pid, userID, true); err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
	_, err = tx.Exec("DELETE FROM tasks WHERE id = $1", taskID)
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
	pid, err := r.GetProjectIDByTask(taskID)
	if err != nil {
		return err
	}
	if err := r.CheckAccess(pid, userID, true); err != nil {
		return err
	}
	_, err = r.db.Exec(`DELETE FROM dependencies WHERE task_id = $1 AND depends_on_id = $2`, taskID, dependsOnID)
	return err
}

func (r *TaskRepository) ClearDependencies(projectID, userID int) error {
	if err := r.CheckAccess(projectID, userID, true); err != nil {
		return err
	}
	_, err := r.db.Exec(`DELETE FROM dependencies WHERE task_id IN (SELECT id FROM tasks WHERE project_id = $1) OR depends_on_id IN (SELECT id FROM tasks WHERE project_id = $1)`, projectID)
	return err
}

func (r *TaskRepository) GetGraphData(projectID, userID int) (*models.GraphData, error) {
	if err := r.CheckAccess(projectID, userID, false); err != nil {
		return nil, err
	}
	graph := &models.GraphData{}

	// ТЕПЕРЬ ДОСТАЕМ И СТАТУС ТОЖЕ
	rowsNodes, _ := r.db.Query("SELECT id, title, opt, real, pess, duration_hours, priority_score, status FROM tasks WHERE project_id = $1", projectID)
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

// GetTasksByProject возвращает список всех задач конкретного проекта
func (r *TaskRepository) GetTasksByProject(projectID, userID int) ([]models.Task, error) {
	// Проверяем, есть ли у пользователя доступ к проекту (хотя бы на чтение)
	if err := r.CheckAccess(projectID, userID, false); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`
		SELECT id, title, opt, real, pess, user_id, project_id, status, duration_hours, priority_score 
		FROM tasks 
		WHERE project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		err := rows.Scan(&t.ID, &t.Title, &t.Opt, &t.Real, &t.Pess, &t.UserID, &t.ProjectID, &t.Status, &t.DurationHours, &t.PriorityScore)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
func (r *TaskRepository) GetDependenciesByProject(projectID int) ([]models.Dependency, error) {
	rows, err := r.db.Query(`
		SELECT d.task_id, d.depends_on_id 
		FROM dependencies d
		JOIN tasks t ON d.task_id = t.id
		WHERE t.project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []models.Dependency
	for rows.Next() {
		var d models.Dependency
		if err := rows.Scan(&d.TaskID, &d.DependsOnID); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, nil
}
