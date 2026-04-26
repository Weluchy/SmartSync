package repository

import (
	"database/sql"
	"smartsync/internal/models"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) GetUserProjects(userID int) ([]models.Project, error) {
	rows, err := r.db.Query("SELECT id, name, owner_id FROM projects WHERE owner_id = $1 ORDER BY id ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		rows.Scan(&p.ID, &p.Name, &p.OwnerID)
		projects = append(projects, p)
	}

	// Если проектов еще нет, сразу создаем дефолтный
	if len(projects) == 0 {
		var newID int
		err := r.db.QueryRow("INSERT INTO projects (name, owner_id) VALUES ($1, $2) RETURNING id", "Мой первый проект", userID).Scan(&newID)
		if err == nil {
			projects = append(projects, models.Project{ID: newID, Name: "Мой первый проект", OwnerID: userID})
		}
	}

	return projects, nil
}

func (r *ProjectRepository) CreateProject(name string, ownerID int) (int, error) {
	var id int
	err := r.db.QueryRow("INSERT INTO projects (name, owner_id) VALUES ($1, $2) RETURNING id", name, ownerID).Scan(&id)
	return id, err
}
