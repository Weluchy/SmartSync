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

	// Защищенная группа роутов. Сюда нельзя попасть без токена!
	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.POST("/tasks", h.createTask)
		protected.POST("/tasks/:id/dependencies", h.createDependency)
		protected.DELETE("/dependencies", h.clearDependencies)
		protected.GET("/graph", h.getGraph)
		protected.DELETE("/tasks/:id", h.deleteTask)
		protected.DELETE("/tasks/:id/dependencies/:dep_id", h.deleteDependency)
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

	err := h.service.ClearDependencies(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сбросить граф"})
		return
	}

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

func (h *Handler) deleteTask(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID, _ := strconv.Atoi(c.Param("id"))

	// Ловим параметр сшивания: если в URL есть ?heal=true, то heal будет равно true
	heal := c.Query("heal") == "true"

	if err := h.service.DeleteTask(taskID, userID.(int), heal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить задачу"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Задача удалена"})
}

func (h *Handler) deleteDependency(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID, _ := strconv.Atoi(c.Param("id"))
	depID, _ := strconv.Atoi(c.Param("dep_id"))

	if err := h.service.DeleteDependency(taskID, depID, userID.(int)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить связь"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Связь удалена"})
}

func (h *Handler) updateTask(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID, _ := strconv.Atoi(c.Param("id"))

	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	t.ID = taskID
	t.UserID = userID.(int)

	if err := h.service.UpdateTask(&t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить задачу"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Задача обновлена"})
}
