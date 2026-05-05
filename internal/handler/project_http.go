package handler

import (
	"net/http"
	"smartsync/internal/models"
	"smartsync/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(s *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: s}
}

func (h *ProjectHandler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.GET("/projects", h.getProjects)
	protected.POST("/projects", h.createProject)
	protected.DELETE("/projects/:project_id", h.deleteProject)
	protected.PUT("/projects/:project_id", h.renameProject)
	protected.POST("/projects/:project_id/members", h.addMember)
	protected.GET("/projects/:project_id/members", h.getMembers)
	protected.DELETE("/projects/:project_id/members/:user_id", h.removeMember)
	// РЕГИСТРАЦИЯ PATCH
	protected.PATCH("/projects/:project_id/members/:user_id", h.updateMemberRole)
}

func (h *ProjectHandler) getMembers(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	members, err := h.service.GetProjectMembers(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить участников"})
		return
	}
	if members == nil {
		members = []models.ProjectMember{}
	}
	c.JSON(http.StatusOK, members)
}

func (h *ProjectHandler) removeMember(c *gin.Context) {
	ownerID, _ := c.Get("user_id")
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	targetUserID, _ := strconv.Atoi(c.Param("user_id"))
	err := h.service.RemoveMember(projectID, ownerID.(int), targetUserID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Участник удален"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Имя не должно быть пустым"})
		return
	}
	id, err := h.service.CreateProject(req.Name, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Создано", "id": id})
}

func (h *ProjectHandler) deleteProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	if err := h.service.DeleteProject(projectID, userID.(int)); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Только создатель может удалить проект"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Проект удален"})
}

func (h *ProjectHandler) renameProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	c.ShouldBindJSON(&req)
	if err := h.service.RenameProject(projectID, userID.(int), req.Name); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Только создатель может переименовать"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Проект переименован"})
}

func (h *ProjectHandler) addMember(c *gin.Context) {
	userID, _ := c.Get("user_id")
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	var req struct {
		Username string `json:"username" binding:"required"`
		Role     string `json:"role" binding:"required"` // Добавили поле
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите логин и роль"})
		return
	}

	if err := h.service.AddMember(projectID, userID.(int), req.Username, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Участник приглашен"})
}

func (h *ProjectHandler) updateMemberRole(c *gin.Context) {
	ownerID, _ := c.Get("user_id")
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	targetUserID, _ := strconv.Atoi(c.Param("user_id"))
	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверная роль"})
		return
	}

	if err := h.service.UpdateMemberRole(projectID, ownerID.(int), targetUserID, req.Role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Роль обновлена"})
}
