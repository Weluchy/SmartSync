package handler

import (
	"net/http"
	"smartsync/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) getMilestones(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	ms, err := h.service.GetMilestones(projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ms == nil {
		ms = []models.Milestone{}
	}
	c.JSON(http.StatusOK, ms)
}

func (h *Handler) createMilestone(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	var req struct {
		Title    string `json:"title" binding:"required"`
		Deadline string `json:"deadline" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите название и дедлайн"})
		return
	}
	m, err := h.service.CreateMilestone(projectID, userID, req.Title, req.Deadline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) getProjectStats(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	stats, err := h.service.GetProjectStats(projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
