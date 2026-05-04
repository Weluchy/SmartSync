package handler

import (
	"net/http"
	"smartsync/internal/auth/models"
	"smartsync/internal/auth/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}
func (h *AuthHandler) InitRoutes() *gin.Engine {
	r := gin.Default()

	// Настройка CORS

	r.POST("/register", h.register)
	r.POST("/login", h.login)

	// Добавляем маршрут, который ждет фронтенд
	r.GET("/user/profile", h.getProfile)
	r.PUT("/user/profile", h.updateProfile)

	return r
}

func (h *AuthHandler) updateProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *AuthHandler) register(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	id, err := h.service.Register(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Имя пользователя уже занято"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Пользователь успешно зарегистрирован", "user_id": id})
}

func (h *AuthHandler) login(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	token, err := h.service.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{Token: token})
}

func (h *AuthHandler) getProfile(c *gin.Context) {
	// Возвращаем JSON, чтобы App.jsx не падал с SyntaxError
	c.JSON(http.StatusOK, gin.H{
		"username": "Неизвестно",
		"stack":    "",
		"status":   "",
	})
}
