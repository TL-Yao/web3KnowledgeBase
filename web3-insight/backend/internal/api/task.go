package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/repository"
)

type TaskHandler struct {
	repo *repository.TaskRepository
}

func NewTaskHandler(repo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// ListTasks godoc
// @Summary List tasks
// @Description Get paginated list of tasks with optional filters
// @Tags tasks
// @Accept json
// @Produce json
// @Param type query string false "Filter by task type"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Page size (default: 20)"
// @Success 200 {object} repository.TaskListResult
// @Router /api/tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	params := repository.TaskListParams{
		Type:   c.Query("type"),
		Status: c.Query("status"),
	}

	if page := c.Query("page"); page != "" {
		p, _ := strconv.Atoi(page)
		params.Page = p
	}

	if limit := c.Query("limit"); limit != "" {
		l, _ := strconv.Atoi(limit)
		params.Limit = l
	}

	result, err := h.repo.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTask godoc
// @Summary Get task by ID
// @Description Get a single task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} model.Task
// @Router /api/tasks/{id} [get]
func (h *TaskHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	task, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetTaskStats godoc
// @Summary Get task statistics
// @Description Get aggregated task statistics
// @Tags tasks
// @Accept json
// @Produce json
// @Success 200 {object} repository.TaskStats
// @Router /api/tasks/stats [get]
func (h *TaskHandler) GetStats(c *gin.Context) {
	stats, err := h.repo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CancelTask godoc
// @Summary Cancel a task
// @Description Cancel a pending or running task
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]string
// @Router /api/tasks/{id}/cancel [post]
func (h *TaskHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.repo.Cancel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "task cancelled",
		"task_id": id,
	})
}
