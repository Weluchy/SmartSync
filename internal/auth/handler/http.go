package handler

import (
	"fmt"
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

	r.POST("/internal/users/bulk", h.getUsersBulk)

	return r
}

func (h *AuthHandler) getUsersBulk(c *gin.Context) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format"})
		return
	}

	// В идеале вызывать через сервис, но для простоты стучимся в репо
	names, err := h.service.Repo().GetUsersNames(req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, names)
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

// Получение профиля
func (h *AuthHandler) getProfile(c *gin.Context) {
	// Gateway прокидывает X-User-ID, но мы его парсим
	userIDStr := c.GetHeader("X-User-ID")

	// Пока обращаемся напрямую к репозиторию для скорости (в идеале через сервис)
	user, err := h.service.Repo().GetProfileByID(parseID(userIDStr))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Профиль не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"full_name": user.FullName,
		"stack":     user.Stack,
		"status":    user.Status,
	})
}

// Обновление профиля
func (h *AuthHandler) updateProfile(c *gin.Context) {
	userIDStr := c.GetHeader("X-User-ID")

	var req models.ProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат"})
		return
	}

	if err := h.service.Repo().UpdateProfile(parseID(userIDStr), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Профиль обновлен"})
}

// Вспомогательная функция для парсинга ID
func parseID(idStr string) int {
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	return id
}
