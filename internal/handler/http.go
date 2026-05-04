package handler

import (
	"errors"
	"net/http"
	"smartsync/internal/models"
	"smartsync/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service        *service.TaskService
	projectService *service.ProjectService
}

func NewHandler(ts *service.TaskService, ps *service.ProjectService) *Handler {
	return &Handler{
		service:        ts,
		projectService: ps,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.Default()

	// Внимание: Блок CORS удален, так как мы настроили его в Gateway!
	// Если оставить его здесь, браузер выдаст ошибку "multiple values '*, *'"

	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		// Управление задачами
		protected.POST("/tasks", h.createTask)
		protected.PUT("/tasks/:id", h.updateTask)
		protected.DELETE("/tasks/:id", h.deleteTask)
		protected.PATCH("/tasks/:id/status", h.updateTaskStatus) // Для Канбан-доски

		// Зависимости
		protected.POST("/tasks/:id/dependencies", h.createDependency)
		protected.DELETE("/tasks/:id/dependencies/:dep_id", h.deleteDependency)
		protected.DELETE("/projects/:project_id/dependencies", h.clearDependencies)

		// Проекты
		protected.GET("/projects/:project_id/graph", h.getGraph)
		protected.GET("/projects/:project_id/tasks", h.getProjectTasks)
	}

	projectHandler := NewProjectHandler(h.projectService)
	projectHandler.RegisterRoutes(protected)

	return r
}

func (h *Handler) createTask(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	t.UserID = userID

	// CreateTask должен возвращать ID созданной записи
	id, err := h.service.CreateTask(&t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	t.ID = id                     // Присваиваем полученный ID объекту
	c.JSON(http.StatusCreated, t) // Возвращаем ВЕСЬ объект
}

func (h *Handler) updateTask(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	taskID, _ := strconv.Atoi(c.Param("id"))
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	t.ID = taskID
	t.UserID = userID

	if err := h.service.UpdateTask(&t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) deleteTask(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	taskID, _ := strconv.Atoi(c.Param("id"))
	heal := c.Query("heal") == "true"

	if err := h.service.DeleteTask(taskID, userID, heal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Удалено"})
}

func (h *Handler) createDependency(c *gin.Context) {
	// ИСПРАВЛЕНИЕ: используем анонимную структуру, чтобы не зависеть от models.Dependency
	var dep struct {
		DependsOnID int `json:"depends_on_id"`
	}
	c.ShouldBindJSON(&dep)
	taskID, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.CreateDependency(taskID, dep.DependsOnID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Связь создана"})
}

func (h *Handler) getProjectTasks(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	projectID, _ := strconv.Atoi(c.Param("project_id"))
	tasks, err := h.service.GetTasksByProject(projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tasks"})
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) deleteDependency(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	taskID, _ := strconv.Atoi(c.Param("id"))
	depID, _ := strconv.Atoi(c.Param("dep_id"))

	if err := h.service.DeleteDependency(taskID, depID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления связи"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Связь удалена"})
}

func (h *Handler) clearDependencies(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	projectID, _ := strconv.Atoi(c.Param("project_id"))

	h.service.ClearDependencies(projectID, userID)
	c.JSON(http.StatusOK, gin.H{"message": "Граф сброшен"})
}

func (h *Handler) getGraph(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	projectID, _ := strconv.Atoi(c.Param("project_id"))

	tasks, err := h.service.GetTasksByProject(projectID, userID)
	if err != nil || tasks == nil {
		tasks = []models.Task{}
	}

	deps, err := h.service.GetDependenciesByProject(projectID)
	if err != nil || deps == nil {
		deps = []models.Dependency{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":        tasks,
		"dependencies": deps,
	})
}

func (h *Handler) updateTaskStatus(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	taskID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный статус"})
		return
	}

	if err := h.service.UpdateTaskStatus(taskID, userID, req.Status); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Статус обновлен"})
}

// Безопасное извлечение ID пользователя
// getUserID безопасно извлекает ID пользователя из контекста Gin,
// куда его положил AuthMiddleware.
// getUserID безопасно достает ID пользователя из контекста Gin
func getUserID(c *gin.Context) (int, error) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("unauthorized: user_id not found")
	}

	// Проверяем тип, так как JWT может вернуть float64
	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, errors.New("invalid user id type")
	}
}
