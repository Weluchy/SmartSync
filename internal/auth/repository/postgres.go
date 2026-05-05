package repository

import (
	"database/sql"
	"fmt"
	"smartsync/internal/auth/models"

	"github.com/lib/pq"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name VARCHAR(100) DEFAULT ''`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS stack VARCHAR(255) DEFAULT ''`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(100) DEFAULT ''`)
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetUsersNames(ids []int) (map[string]string, error) {
	result := make(map[string]string)
	if len(ids) == 0 {
		return result, nil
	}

	// ИСПРАВЛЕНИЕ: NULLIF превращает пустую строку в NULL, чтобы COALESCE сработал
	query := `SELECT id, COALESCE(NULLIF(full_name, ''), username) FROM users WHERE id = ANY($1)`
	rows, err := r.db.Query(query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		result[fmt.Sprintf("%d", id)] = name
	}
	return result, nil
}

func (r *AuthRepository) GetProfileByID(userID int) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`SELECT id, username, COALESCE(full_name, ''), COALESCE(stack, ''), COALESCE(status, '') 
		FROM users WHERE id = $1`, userID).Scan(&user.ID, &user.Username, &user.FullName, &user.Stack, &user.Status)
	return user, err
}

func (r *AuthRepository) UpdateProfile(userID int, req models.ProfileUpdateRequest) error {
	_, err := r.db.Exec(`UPDATE users SET full_name = $1, stack = $2, status = $3 WHERE id = $4`,
		req.FullName, req.Stack, req.Status, userID)
	return err
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
