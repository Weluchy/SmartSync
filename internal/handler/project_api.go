package handler

import (
	"fmt"
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

func (h *Handler) deleteMilestone(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	projectID, _ := strconv.Atoi(c.Param("project_id"))
	milestoneID, _ := strconv.Atoi(c.Param("milestone_id"))

	if err := h.service.DeleteMilestone(projectID, milestoneID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Веха удалена"})
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

func (h *Handler) exportProjectCSV(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	projectID, _ := strconv.Atoi(c.Param("project_id"))

	tasks, err := h.service.GetTasksByProject(projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки"})
		return
	}

	// Формируем CSV
	csv := "ID,Название,Описание,Статус,Опт,Реал,Песс,Исполнитель,Приоритет,Длительность\n"
	for _, t := range tasks {
		assignee := ""
		if t.AssigneeName != "" {
			assignee = t.AssigneeName
		}
		csv += fmt.Sprintf("%d,\"%s\",\"%s\",%s,%d,%d,%d,%s,%.1f,%.1f\n",
			t.ID, t.Title, t.Description, t.Status, t.Opt, t.Real, t.Pess,
			assignee, t.PriorityScore, t.DurationHours)
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=project_%d_tasks.csv", projectID))
	c.String(200, csv)
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
