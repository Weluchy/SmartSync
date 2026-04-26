package handler

import (
	"net/http"
	"smartsync/internal/service"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(s *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: s}
}

// Эта функция "прикрепляет" маршруты к существующему роутеру
func (h *ProjectHandler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.GET("/projects", h.getProjects)
	protected.POST("/projects", h.createProject)
}

func (h *ProjectHandler) getProjects(c *gin.Context) {
	userID, _ := c.Get("user_id")

	projects, err := h.service.GetUserProjects(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения проектов"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) createProject(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Имя проекта не должно быть пустым"})
		return
	}

	id, err := h.service.CreateProject(req.Name, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать проект"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Проект успешно создан", "id": id})
}
