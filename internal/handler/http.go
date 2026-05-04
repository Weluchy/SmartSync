package handler

import (
	"fmt"
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
	userID, _ := c.Get("user_id") // Добавить эту строку
	taskID, _ := strconv.Atoi(c.Param("id"))
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	t.ID = taskID
	t.UserID = userID.(int) // Установить ID пользователя

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
	userID, _ := c.Get("user_id") // Получаем ID пользователя из контекста
	projectID, _ := strconv.Atoi(c.Param("project_id"))

	// Передаем userID для проверки прав доступа
	tasks, err := h.service.GetTasksByProject(projectID, userID.(int))
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
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	userID, _ := c.Get("user_id")

	// 1. Получаем задачи
	tasks, err := h.service.GetTasksByProject(projectID, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	if tasks == nil {
		tasks = []models.Task{} // Защита от null
	}

	// 2. Получаем зависимости (Проверь, что этот метод есть в твоем сервисе!)
	deps, err := h.service.GetDependenciesByProject(projectID)
	if err != nil {
		deps = []models.Dependency{} // Если метода нет или ошибка, шлем пустой список
	}
	if deps == nil {
		deps = []models.Dependency{} // Защита от null
	}

	// Отладка в консоль сервера
	fmt.Printf("DEBUG: Graph for Project %d: %d tasks, %d deps\n", projectID, len(tasks), len(deps))

	c.JSON(http.StatusOK, gin.H{
		"tasks":        tasks,
		"dependencies": deps,
	})
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
