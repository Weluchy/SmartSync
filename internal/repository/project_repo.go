package repository

import (
	"database/sql"
	"smartsync/internal/models"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	// Создаем таблицу проектов (если её нет)
	db.Exec(`CREATE TABLE IF NOT EXISTS projects (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		owner_id INTEGER NOT NULL
	)`)

	// Создаем таблицу участников для совместного доступа (RBAC)
	db.Exec(`CREATE TABLE IF NOT EXISTS project_members (
		project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'viewer',
		PRIMARY KEY (project_id, user_id)
	)`)

	return &ProjectRepository{db: db}
}

// Получаем все проекты: и свои, и те, куда нас пригласили
func (r *ProjectRepository) GetUserProjects(userID int) ([]models.Project, error) {
	query := `
		SELECT p.id, p.name, p.owner_id, pm.role 
		FROM projects p
		JOIN project_members pm ON p.id = pm.project_id
		WHERE pm.user_id = $1
		ORDER BY p.id ASC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		rows.Scan(&p.ID, &p.Name, &p.OwnerID, &p.Role)
		projects = append(projects, p)
	}

	// Дефолтный проект для новых пользователей
	if len(projects) == 0 {
		return r.createDefaultProject(userID)
	}

	return projects, nil
}

func (r *ProjectRepository) createDefaultProject(userID int) ([]models.Project, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var newID int
	err = tx.QueryRow("INSERT INTO projects (name, owner_id) VALUES ($1, $2) RETURNING id", "Мой первый проект", userID).Scan(&newID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, 'owner')", newID, userID)
	if err != nil {
		return nil, err
	}

	tx.Commit()
	return []models.Project{{ID: newID, Name: "Мой первый проект", OwnerID: userID, Role: "owner"}}, nil
}

// Создание нового проекта с транзакцией
func (r *ProjectRepository) CreateProject(name string, ownerID int) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRow("INSERT INTO projects (name, owner_id) VALUES ($1, $2) RETURNING id", name, ownerID).Scan(&id)
	if err != nil {
		return 0, err
	}

	// Сразу выдаем создателю права 'owner'
	_, err = tx.Exec("INSERT INTO project_members (project_id, user_id, role) VALUES ($1, $2, 'owner')", id, ownerID)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit()
}

// Безопасное удаление проекта (только для владельцев)
func (r *ProjectRepository) DeleteProject(projectID, userID int) error {
	_, err := r.db.Exec("DELETE FROM projects WHERE id = $1 AND owner_id = $2", projectID, userID)
	return err
}
