package handler

import (
	"net/http"
	"smartsync/internal/models"
	"smartsync/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.TaskService
}

func NewHandler(s *service.TaskService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE")
		// ОЧЕНЬ ВАЖНО: Разрешаем фронтенду отправлять токен
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Cache")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Защищенная группа роутов. Сюда нельзя попасть без токена!
	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.POST("/tasks", h.createTask)
		protected.POST("/tasks/:id/dependencies", h.createDependency)
		protected.DELETE("/dependencies", h.clearDependencies)
		protected.GET("/graph", h.getGraph)
	}

	return r
}

func (h *Handler) createTask(c *gin.Context) {
	userID, _ := c.Get("user_id") // Достаем ID пользователя из токена
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t.UserID = userID.(int) // Присваиваем владельца задаче

	id, err := h.service.CreateTask(&t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Задача создана", "id": id})
}

func (h *Handler) createDependency(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID, _ := strconv.Atoi(c.Param("id"))
	var dep models.Dependency
	c.ShouldBindJSON(&dep)

	if err := h.service.CreateDependency(taskID, dep.DependsOnID, userID.(int)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Связь создана"})
}

func (h *Handler) clearDependencies(c *gin.Context) {
	userID, _ := c.Get("user_id")
	h.service.ClearDependencies(userID.(int))
	c.JSON(http.StatusOK, gin.H{"message": "Граф сброшен"})
}

func (h *Handler) getGraph(c *gin.Context) {
	userID, _ := c.Get("user_id")
	graph, fromCache, _ := h.service.GetGraph(c.Request.Context(), userID.(int))

	if fromCache {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
	c.JSON(http.StatusOK, graph)
}
