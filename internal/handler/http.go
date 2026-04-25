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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Cache")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.POST("/tasks", h.createTask)
	r.POST("/tasks/:id/dependencies", h.createDependency)
	r.DELETE("/dependencies", h.clearDependencies)
	r.GET("/graph", h.getGraph)

	return r
}

func (h *Handler) createTask(c *gin.Context) {
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.CreateTask(&t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Задача создана", "id": id})
}

func (h *Handler) createDependency(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("id"))
	var dep models.Dependency
	c.ShouldBindJSON(&dep)

	if err := h.service.CreateDependency(taskID, dep.DependsOnID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Связь создана"})
}

func (h *Handler) clearDependencies(c *gin.Context) {
	h.service.ClearDependencies()
	c.JSON(http.StatusOK, gin.H{"message": "Граф сброшен"})
}

func (h *Handler) getGraph(c *gin.Context) {
	graph, fromCache, _ := h.service.GetGraph(c.Request.Context())

	if fromCache {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}

	c.JSON(http.StatusOK, graph)
}
