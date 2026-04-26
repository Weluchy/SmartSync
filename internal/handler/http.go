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

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Cache")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.POST("/tasks", h.createTask)
		protected.PUT("/tasks/:id", h.updateTask)
		protected.DELETE("/tasks/:id", h.deleteTask)
		protected.POST("/tasks/:id/dependencies", h.createDependency)
		protected.DELETE("/tasks/:id/dependencies/:dep_id", h.deleteDependency)

		// Обновленные роуты (обращаются к конкретной папке)
		protected.DELETE("/projects/:project_id/dependencies", h.clearDependencies)
		protected.GET("/projects/:project_id/graph", h.getGraph)
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
	userID, _ := c.Get("user_id")
	taskID, _ := strconv.Atoi(c.Param("id"))
	var t models.Task
	c.ShouldBindJSON(&t)
	t.ID = taskID
	t.UserID = userID.(int)

	if err := h.service.UpdateTask(&t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Обновлено"})
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
	var dep models.Dependency
	c.ShouldBindJSON(&dep)
	taskID, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.CreateDependency(taskID, dep.DependsOnID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Связь создана"})
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
