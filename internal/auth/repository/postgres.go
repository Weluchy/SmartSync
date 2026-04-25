package repository

import (
	"database/sql"
	"smartsync/internal/auth/models"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	// Автоматическая миграция таблицы пользователей
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(username, passwordHash string) (int, error) {
	var id int
	err := r.db.QueryRow("INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id",
		username, passwordHash).Scan(&id)
	return id, err
}

func (r *AuthRepository) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)
	return user, err
}
