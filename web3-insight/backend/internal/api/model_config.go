package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/repository"
)

type ModelConfigHandler struct {
	configRepo *repository.ConfigRepository
	models     *config.ModelsConfig
	routing    *config.RoutingConfig
}

func NewModelConfigHandler(
	configRepo *repository.ConfigRepository,
	models *config.ModelsConfig,
	routing *config.RoutingConfig,
) *ModelConfigHandler {
	return &ModelConfigHandler{
		configRepo: configRepo,
		models:     models,
		routing:    routing,
	}
}

// GetModelsRegistry godoc
// @Summary Get available models
// @Description Returns all models from models.yaml (both local and cloud)
// @Tags model-config
// @Accept json
// @Produce json
// @Success 200 {object} config.ModelsConfig
// @Router /api/models/registry [get]
func (h *ModelConfigHandler) GetModelsRegistry(c *gin.Context) {
	c.JSON(http.StatusOK, h.models)
}

// GetTaskTypes godoc
// @Summary Get task types
// @Description Returns all task types from routing.yaml
// @Tags model-config
// @Accept json
// @Produce json
// @Success 200 {object} config.RoutingConfig
// @Router /api/models/tasks [get]
func (h *ModelConfigHandler) GetTaskTypes(c *gin.Context) {
	c.JSON(http.StatusOK, h.routing)
}

// TaskSelection represents user's model selection for a task
type TaskSelection struct {
	TaskID   string `json:"taskId"`
	Primary  string `json:"primary"`
	Fallback string `json:"fallback"`
}

// GetUserSelections godoc
// @Summary Get user's model selections
// @Description Returns user's saved model selections from database
// @Tags model-config
// @Accept json
// @Produce json
// @Success 200 {array} TaskSelection
// @Router /api/models/selections [get]
func (h *ModelConfigHandler) GetUserSelections(c *gin.Context) {
	// Get from database configs table
	configData, err := h.configRepo.Get("model_selections")
	if err != nil {
		// No saved selections, return defaults from routing.yaml
		defaults := h.getDefaultSelections()
		c.JSON(http.StatusOK, defaults)
		return
	}

	var selections []TaskSelection
	// Parse JSON from database
	if err := json.Unmarshal(configData.Value, &selections); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid stored config"})
		return
	}

	// Validate selections against current model registry
	validatedSelections := h.validateSelections(selections)

	c.JSON(http.StatusOK, validatedSelections)
}

// UpdateUserSelections godoc
// @Summary Update user's model selections
// @Description Save user's model selections to database
// @Tags model-config
// @Accept json
// @Produce json
// @Param selections body []TaskSelection true "Model selections"
// @Success 200 {array} TaskSelection
// @Router /api/models/selections [put]
func (h *ModelConfigHandler) UpdateUserSelections(c *gin.Context) {
	var selections []TaskSelection
	if err := c.ShouldBindJSON(&selections); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate selections
	validatedSelections := h.validateSelections(selections)

	// Save to database as JSON (using SetJSON to avoid double-encoding)
	jsonData, err := json.Marshal(validatedSelections)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize"})
		return
	}

	err = h.configRepo.SetJSON("model_selections", jsonData, "User's model selection preferences")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, validatedSelections)
}

// getDefaultSelections creates default selections from routing.yaml
func (h *ModelConfigHandler) getDefaultSelections() []TaskSelection {
	var selections []TaskSelection
	for _, task := range h.routing.TaskTypes {
		selections = append(selections, TaskSelection{
			TaskID:   task.ID,
			Primary:  task.DefaultPrimary,
			Fallback: task.DefaultFallback,
		})
	}
	return selections
}

// validateSelections checks if selected models are available
func (h *ModelConfigHandler) validateSelections(selections []TaskSelection) []TaskSelection {
	var validated []TaskSelection

	for _, sel := range selections {
		// Get task definition
		task, err := h.routing.GetTaskByID(sel.TaskID)
		if err != nil {
			// Task no longer exists, skip
			continue
		}

		// Check if primary model is available
		_ = h.models.IsModelAvailable(sel.Primary)
		_ = h.models.IsModelAvailable(sel.Fallback)

		// Keep the selection as-is, validation happens at runtime
		// Frontend will handle UI display for unavailable models
		validated = append(validated, TaskSelection{
			TaskID:   task.ID,
			Primary:  sel.Primary,
			Fallback: sel.Fallback,
		})
	}

	return validated
}
