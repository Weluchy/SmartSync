package handler

import (
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
		protected.POST("/tasks", h.createTask)
		protected.PUT("/tasks/:id", h.updateTask)
		protected.DELETE("/tasks/:id", h.deleteTask)

		protected.POST("/tasks/:id/dependencies", h.createDependency)
		protected.DELETE("/tasks/:id/dependencies/:dep_id", h.deleteDependency)

		protected.DELETE("/projects/:project_id/dependencies", h.clearDependencies)
		protected.GET("/projects/:project_id/graph", h.getGraph)

		// МАРШРУТ ДЛЯ КАНБАН-ДОСКИ
		protected.PATCH("/tasks/:id/status", h.updateTaskStatus)
		api := r.Group("/api")
		{
			api.GET("/projects/:project_id/tasks", h.getProjectTasks) // Добавь эту строку
			api.PATCH("/tasks/:id/status", h.updateTask)
		}
	}

	projectHandler := NewProjectHandler(h.projectService)
	projectHandler.RegisterRoutes(protected)

	return r
}

func (h *Handler) createTask(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t.UserID = userID.(int)

	id, err := h.service.CreateTask(&t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Задача создана", "id": id})
}

func (h *Handler) updateTask(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("id"))
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	t.ID = taskID
	if err := h.service.UpdateTask(&t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) deleteTask(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID, _ := strconv.Atoi(c.Param("id"))
	heal := c.Query("heal") == "true"

	if err := h.service.DeleteTask(taskID, userID.(int), heal); err != nil {
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
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	tasks, err := h.service.GetTasksByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки задач"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}
func (h *Handler) deleteDependency(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID, _ := strconv.Atoi(c.Param("id"))
	depID, _ := strconv.Atoi(c.Param("dep_id"))

	if err := h.service.DeleteDependency(taskID, depID, userID.(int)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления связи"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Связь удалена"})
}

func (h *Handler) clearDependencies(c *gin.Context) {
	userID, _ := c.Get("user_id")
	projectID, _ := strconv.Atoi(c.Param("project_id"))

	h.service.ClearDependencies(projectID, userID.(int))
	c.JSON(http.StatusOK, gin.H{"message": "Граф сброшен"})
}

func (h *Handler) getGraph(c *gin.Context) {
	userID, _ := c.Get("user_id")
	projectID, _ := strconv.Atoi(c.Param("project_id"))

	graph, fromCache, _ := h.service.GetGraph(c.Request.Context(), projectID, userID.(int))
	if fromCache {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
	c.JSON(http.StatusOK, graph)
}

func (h *Handler) updateTaskStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный статус"})
		return
	}

	if err := h.service.UpdateTaskStatus(taskID, userID.(int), req.Status); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Статус обновлен"})
}
