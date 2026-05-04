package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"` // Новое
	Stack        string    `json:"stack"`     // Новое
	Status       string    `json:"status"`    // Новое
	CreatedAt    time.Time `json:"created_at"`
}

// Модель для обновления профиля
type ProfileUpdateRequest struct {
	FullName string `json:"full_name"`
	Stack    string `json:"stack"`
	Status   string `json:"status"`
}

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
