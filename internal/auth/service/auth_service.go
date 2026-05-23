package service

import (
	"errors"
	"os"
	"time"

	"smartsync/internal/auth/models"
	"smartsync/internal/auth/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      *repository.AuthRepository
	jwtSecret []byte
}

func (s *AuthService) Repo() *repository.AuthRepository { return s.repo }

func NewAuthService(repo *repository.AuthRepository) *AuthService {
	secret := []byte("smartsync_diploma_secret_key_2026")
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		secret = []byte(envSecret)
	}
	return &AuthService{repo: repo, jwtSecret: secret}
}

func (s *AuthService) Register(req models.AuthRequest) (int, error) {
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errors.New("неверный пароль")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}
