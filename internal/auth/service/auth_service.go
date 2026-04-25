package service

import (
	"errors"
	"time"

	"smartsync/internal/auth/models"
	"smartsync/internal/auth/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// В реальных проектах ключ лежит в .env, для диплома оставим здесь
var jwtSecret = []byte("smartsync_diploma_secret_key_2026")

type AuthService struct {
	repo *repository.AuthRepository
}

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Register(req models.AuthRequest) (int, error) {
	// Хешируем пароль (cost=10 - баланс между скоростью и безопасностью)
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return 0, err
	}
	return s.repo.CreateUser(req.Username, string(hash))
}

func (s *AuthService) Login(req models.AuthRequest) (string, error) {
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return "", errors.New("пользователь не найден")
	}

	// Сравниваем хеши
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errors.New("неверный пароль")
	}

	// Генерируем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Токен живет 24 часа
	})

	return token.SignedString(jwtSecret)
}
