package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) getMilestones(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	ms, err := h.service.GetMilestones(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ms)
}

func (h *Handler) createMilestone(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	var req struct {
		Title    string `json:"title" binding:"required"`
		Deadline string `json:"deadline" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите название и дедлайн"})
		return
	}
	m, err := h.service.CreateMilestone(projectID, req.Title, req.Deadline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) getProjectStats(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	stats, err := h.service.GetProjectStats(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
