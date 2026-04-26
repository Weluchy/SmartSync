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
	// 1. Создаем таблицу проектов
	db.Exec(`CREATE TABLE IF NOT EXISTS projects (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		owner_id INTEGER NOT NULL
	)`)

	// 2. Создаем или обновляем таблицу задач
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

	// 3. Безопасно добавляем колонку project_id в существующую таблицу задач
	// Если колонка уже есть, команда просто проигнорируется
	db.Exec(`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE`)

	// 4. Таблица зависимостей
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
	_, err := r.db.Exec(`
		DELETE FROM dependencies 
		WHERE task_id IN (SELECT id FROM tasks WHERE user_id = $1) 
		   OR depends_on_id IN (SELECT id FROM tasks WHERE user_id = $2)
	`, userID, userID)

	// Теперь мы точно увидим, если база ругнется
	if err != nil {
		fmt.Println("[БД ОШИБКА] Не удалось удалить связи:", err)
	}
	return err
}

func (r *TaskRepository) DeleteTask(taskID, userID int, heal bool) error {
	// Открываем транзакцию (выполняется либо всё, либо ничего)
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Если что-то пойдет не так, изменения откатятся

	// 1. Проверяем владельца (безопасность превыше всего)
	var ownerID int
	err = tx.QueryRow("SELECT user_id FROM tasks WHERE id = $1", taskID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		return fmt.Errorf("задача не найдена или доступ запрещен")
	}

	// 2. Если пользователь хочет "сшить" связи
	if heal {
		// Ищем родителей (от кого зависела удаляемая задача)
		rowsP, _ := tx.Query("SELECT depends_on_id FROM dependencies WHERE task_id = $1", taskID)
		var parents []int
		for rowsP.Next() {
			var p int
			rowsP.Scan(&p)
			parents = append(parents, p)
		}
		rowsP.Close()

		// Ищем детей (кто зависел от удаляемой задачи)
		rowsC, _ := tx.Query("SELECT task_id FROM dependencies WHERE depends_on_id = $1", taskID)
		var children []int
		for rowsC.Next() {
			var c int
			rowsC.Scan(&c)
			children = append(children, c)
		}
		rowsC.Close()

		// Связываем всех детей со всеми родителями напрямую
		for _, child := range children {
			for _, parent := range parents {
				// ON CONFLICT DO NOTHING защищает от ошибки, если такая связь уже существует
				tx.Exec("INSERT INTO dependencies (task_id, depends_on_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", child, parent)
			}
		}
	}

	// 3. Удаляем саму задачу (ON DELETE CASCADE удалит старые связи с ней автоматически)
	_, err = tx.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", taskID, userID)
	if err != nil {
		return err
	}

	// Фиксируем транзакцию
	return tx.Commit()
}

func (r *TaskRepository) DeleteDependency(taskID, dependsOnID, userID int) error {
	_, err := r.db.Exec(`
		DELETE FROM dependencies 
		WHERE task_id = $1 AND depends_on_id = $2 
		AND task_id IN (SELECT id FROM tasks WHERE user_id = $3)
	`, taskID, dependsOnID, userID)
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

func (r *TaskRepository) UpdateTask(t *models.Task) error {
	_, err := r.db.Exec(`
		UPDATE tasks 
		SET title = $1, opt = $2, real = $3, pess = $4 
		WHERE id = $5 AND user_id = $6
	`, t.Title, t.Opt, t.Real, t.Pess, t.ID, t.UserID)
	return err
}

// --- ЛОГИКА ПРОЕКТОВ ---

type Project struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	OwnerID int    `json:"owner_id"`
}

func (r *TaskRepository) GetUserProjects(userID int) ([]Project, error) {
	rows, err := r.db.Query("SELECT id, name, owner_id FROM projects WHERE owner_id = $1 ORDER BY id ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		rows.Scan(&p.ID, &p.Name, &p.OwnerID)
		projects = append(projects, p)
	}

	// Если проектов нет, создаем дефолтный "Мой первый проект"
	if len(projects) == 0 {
		var newID int
		err := r.db.QueryRow("INSERT INTO projects (name, owner_id) VALUES ($1, $2) RETURNING id", "Мой первый проект", userID).Scan(&newID)
		if err == nil {
			projects = append(projects, Project{ID: newID, Name: "Мой первый проект", OwnerID: userID})
		}
	}

	return projects, nil
}

func (r *TaskRepository) CreateProject(name string, ownerID int) (int, error) {
	var id int
	err := r.db.QueryRow("INSERT INTO projects (name, owner_id) VALUES ($1, $2) RETURNING id", name, ownerID).Scan(&id)
	return id, err
}
