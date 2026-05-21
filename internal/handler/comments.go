package handler

import (
	"net/http"
	"smartsync/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary Добавить комментарий
// @Description Добавляет комментарий к задаче
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Param comment body object true "{ \"text\": \"string\" }"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security bearerAuth
// @Router /tasks/{id}/comments [post]
func (h *Handler) addComment(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	taskID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Текст комментария обязателен"})
		return
	}

	comment, err := h.service.AddComment(taskID, userID, req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// @Summary Получить комментарии задачи
// @Description Возвращает все комментарии к задаче
// @Tags Comments
// @Produce json
// @Param id path int true "ID задачи"
// @Success 200 {array} models.Comment
// @Failure 401 {object} map[string]string
// @Security bearerAuth
// @Router /tasks/{id}/comments [get]
func (h *Handler) getComments(c *gin.Context) {
	taskID, _ := strconv.Atoi(c.Param("id"))
	comments, err := h.service.GetComments(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if comments == nil {
		comments = []models.Comment{}
	}
	c.JSON(http.StatusOK, comments)
}
