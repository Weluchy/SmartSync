package handler

import (
	"errors"
	"log"
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

// Убрали NATS из параметров, так как хендлеру он больше не нужен[cite: 2]
func NewHandler(ts *service.TaskService, ps *service.ProjectService) *Handler {
	return &Handler{
		service:        ts,
		projectService: ps,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	r := gin.Default()
	protected := r.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.POST("/tasks", h.createTask)
		protected.GET("/tasks/:id", h.getTask)
		protected.PUT("/tasks/:id", h.updateTask)
		protected.DELETE("/tasks/:id", h.deleteTask)
		protected.PATCH("/tasks/:id/status", h.updateTaskStatus)
		protected.POST("/tasks/:id/dependencies", h.createDependency)
		protected.DELETE("/tasks/:id/dependencies/:dep_id", h.deleteDependency)
		protected.DELETE("/projects/:project_id/dependencies", h.clearDependencies)
		protected.GET("/projects/:project_id/graph", h.getGraph)
		protected.GET("/projects/:project_id/tasks", h.getProjectTasks)
		protected.GET("/invitations/my", h.getMyInvitations)
		protected.POST("/tasks/:id/comments", h.addComment)
		protected.GET("/tasks/:id/comments", h.getComments)
		protected.GET("/projects/:project_id/milestones", h.getMilestones)
		protected.POST("/projects/:project_id/milestones", h.createMilestone)
		protected.GET("/projects/:project_id/stats", h.getProjectStats)
	}
	projectHandler := NewProjectHandler(h.projectService)
	projectHandler.RegisterRoutes(protected)
	return r
}

// @Summary Создать задачу
// @Description Создаёт новую задачу в проекте
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task body models.Task true "Данные задачи"
// @Success 201 {object} models.Task
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security bearerAuth
// @Router /tasks [post]
func (h *Handler) getTask(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	taskID, _ := strconv.Atoi(c.Param("id"))

	task, err := h.service.GetTaskByID(taskID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *Handler) createTask(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	t.UserID = userID
	id, err := h.service.CreateTask(&t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	t.ID = id
	c.JSON(http.StatusCreated, t)
}

// @Summary Обновить задачу
// @Description Изменяет название, оценки (opt/real/pess) или исполнителя задачи
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Param task body models.Task true "Новые данные"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Security bearerAuth
// @Router /tasks/{id} [put]
func (h *Handler) updateTask(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	taskID, _ := strconv.Atoi(c.Param("id"))
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	t.ID = taskID
	t.UserID = userID

	if err := h.service.UpdateTask(&t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// @Summary Удалить задачу
// @Description Удаляет задачу. Требует роль admin+
// @Tags Tasks
// @Produce json
// @Param id path int true "ID задачи"
// @Param heal query bool false "Перестроить граф"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security bearerAuth
// @Router /tasks/{id} [delete]
func (h *Handler) deleteTask(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	taskID, _ := strconv.Atoi(c.Param("id"))
	heal := c.Query("heal") == "true"

	if err := h.service.DeleteTask(taskID, userID, heal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Удалено"})
}

// @Summary Создать зависимость
// @Description Создаёт связь: task_id зависит от depends_on_id
// @Tags Dependencies
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Param dep body object true "{ \"depends_on_id\": int }"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security bearerAuth
// @Router /tasks/{id}/dependencies [post]
func (h *Handler) createDependency(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var dep struct {
		DependsOnID int `json:"depends_on_id"`
	}
	if err := c.ShouldBindJSON(&dep); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	taskID, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.CreateDependency(taskID, dep.DependsOnID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Связь создана"})
}

// @Summary Задачи проекта
// @Description Возвращает список задач с именами авторов
// @Tags Projects
// @Produce json
// @Param project_id path int true "ID проекта"
// @Success 200 {array} models.Task
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security bearerAuth
// @Router /projects/{project_id}/tasks [get]
// @Summary Изменить статус задачи
// @Description Меняет статус (todo → in_progress → done). Только исполнитель или admin+
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path int true "ID задачи"
// @Param status body object true "{ \"status\": \"todo|in_progress|done\" }"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Security bearerAuth
// @Router /tasks/{id}/status [patch]
// @Summary Граф зависимостей проекта
// @Description Возвращает все задачи и связи
// @Tags Projects
// @Produce json
// @Param project_id path int true "ID проекта"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Security bearerAuth
// @Router /projects/{project_id}/graph [get]
func (h *Handler) getProjectTasks(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	projectID, _ := strconv.Atoi(c.Param("project_id"))
	tasks, err := h.service.GetTasksByProject(projectID, userID)
	if err != nil {
		// ДОБАВЛЯЕМ ВОТ ЭТУ СТРОКУ:
		log.Printf("❌ ОШИБКА БД ПРИ ЗАГРУЗКЕ ЗАДАЧ: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tasks"})
		return
	}

	if tasks == nil {
		tasks = []models.Task{}
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) deleteDependency(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	taskID, _ := strconv.Atoi(c.Param("id"))
	depID, _ := strconv.Atoi(c.Param("dep_id"))

	if err := h.service.DeleteDependency(taskID, depID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления связи"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Связь удалена"})
}

func (h *Handler) clearDependencies(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	projectID, _ := strconv.Atoi(c.Param("project_id"))

	if err := h.service.ClearDependencies(projectID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Граф сброшен"})
}

func (h *Handler) getGraph(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	projectID, _ := strconv.Atoi(c.Param("project_id"))

	tasks, err := h.service.GetTasksByProject(projectID, userID)
	if err != nil || tasks == nil {
		tasks = []models.Task{}
	}

	deps, err := h.service.GetDependenciesByProject(projectID)
	if err != nil || deps == nil {
		deps = []models.Dependency{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":        tasks,
		"dependencies": deps,
	})
}

func (h *Handler) updateTaskStatus(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	taskID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный статус"})
		return
	}

	if err := h.service.UpdateTaskStatus(taskID, userID, req.Status); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус обновлен"})
}

func getUserID(c *gin.Context) (int, error) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, errors.New("unauthorized: user_id not found")
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, errors.New("invalid user id type")
	}
}

func (h *Handler) getMyInvitations(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	invites, err := h.projectService.GetInvitedProjects(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки приглашений"})
		return
	}
	if invites == nil {
		invites = []models.Project{}
	}
	c.JSON(http.StatusOK, invites)
}
