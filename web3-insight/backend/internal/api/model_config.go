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

	// Merge saved selections with defaults (in case new tasks were added)
	mergedSelections := h.mergeSelections(selections)

	c.JSON(http.StatusOK, mergedSelections)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Input validation
	if len(selections) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "selections cannot be empty"})
		return
	}

	if len(selections) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "too many selections (max 100)"})
		return
	}

	// Validate required fields
	for i, sel := range selections {
		if sel.TaskID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required for all selections"})
			return
		}
		if sel.Primary == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "primary model is required for all selections"})
			return
		}
		// Fallback is optional, but set empty string if not provided
		if sel.Fallback == "" {
			selections[i].Fallback = ""
		}
	}

	// Validate selections
	validatedSelections := h.validateSelections(selections)

	// Save to database as JSON (using SetJSON to avoid double-encoding)
	jsonData, err := json.Marshal(validatedSelections)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save selections"})
		return
	}

	err = h.configRepo.SetJSON("model_selections", jsonData, "User's model selection preferences")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save selections"})
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

// mergeSelections combines saved selections with defaults from routing.yaml
// This ensures that new tasks added to routing.yaml appear in the user's selections
func (h *ModelConfigHandler) mergeSelections(savedSelections []TaskSelection) []TaskSelection {
	// Create map of saved selections for fast lookup
	savedMap := make(map[string]TaskSelection)
	for _, sel := range savedSelections {
		savedMap[sel.TaskID] = sel
	}

	// Start with all current tasks from routing.yaml
	var merged []TaskSelection
	for _, task := range h.routing.TaskTypes {
		if saved, exists := savedMap[task.ID]; exists {
			// Use saved selection
			merged = append(merged, saved)
		} else {
			// Use default for new task
			merged = append(merged, TaskSelection{
				TaskID:   task.ID,
				Primary:  task.DefaultPrimary,
				Fallback: task.DefaultFallback,
			})
		}
	}

	return merged
}
