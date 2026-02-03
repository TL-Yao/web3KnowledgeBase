package service

import (
	"encoding/json"
	"fmt"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/model"
)

// ConfigRepository interface defines the methods needed from repository
type ConfigRepository interface {
	Get(key string) (*model.Config, error)
	Set(key, value, description string) error
	GetAll() ([]model.Config, error)
	Delete(key string) error
	GetMap() (map[string]string, error)
	SetMultiple(configs map[string]string) error
	SetJSON(key string, jsonValue []byte, description string) error
}

// ModelSelector handles model selection with fallback logic
type ModelSelector struct {
	models     *config.ModelsConfig
	routing    *config.RoutingConfig
	configRepo ConfigRepository
}

// ModelSelectionResult contains the selected model and metadata
type ModelSelectionResult struct {
	ModelID       string `json:"modelId"`
	IsFallback    bool   `json:"isFallback"`
	PrimaryFailed bool   `json:"primaryFailed"`
	Reason        string `json:"reason,omitempty"`
}

// TaskSelection represents a user's model selection for a task
type TaskSelection struct {
	TaskID   string `json:"taskId"`
	Primary  string `json:"primary"`
	Fallback string `json:"fallback"`
}

// NewModelSelector creates a new model selector
func NewModelSelector(
	models *config.ModelsConfig,
	routing *config.RoutingConfig,
	configRepo ConfigRepository,
) *ModelSelector {
	return &ModelSelector{
		models:     models,
		routing:    routing,
		configRepo: configRepo,
	}
}

// SelectModelForTask selects appropriate model for a task with fallback logic
// Returns: model selection result with model ID, fallback status, and reason
func (ms *ModelSelector) SelectModelForTask(taskID string) (*ModelSelectionResult, error) {
	// Step 1: Get user's selections from database
	primaryModel, fallbackModel, err := ms.getUserSelection(taskID)
	if err != nil {
		// User hasn't configured, use defaults from routing.yaml
		task, err := ms.routing.GetTaskByID(taskID)
		if err != nil {
			return nil, fmt.Errorf("unknown task type: %s", taskID)
		}
		primaryModel = task.DefaultPrimary
		fallbackModel = task.DefaultFallback
	}

	// Step 2: Try primary model
	if ms.models.IsModelAvailable(primaryModel) {
		return &ModelSelectionResult{
			ModelID:    primaryModel,
			IsFallback: false,
		}, nil
	}

	// Step 3: Primary failed, try fallback
	if ms.models.IsModelAvailable(fallbackModel) {
		return &ModelSelectionResult{
			ModelID:       fallbackModel,
			IsFallback:    true,
			PrimaryFailed: true,
			Reason:        fmt.Sprintf("Primary model '%s' unavailable, using fallback", primaryModel),
		}, nil
	}

	// Step 4: Both failed, return error
	return nil, fmt.Errorf(
		"no available models for task '%s': primary '%s' and fallback '%s' both unavailable",
		taskID, primaryModel, fallbackModel,
	)
}

// getUserSelection retrieves user's model selection from database
func (ms *ModelSelector) getUserSelection(taskID string) (primary, fallback string, err error) {
	configData, err := ms.configRepo.Get("model_selections")
	if err != nil {
		return "", "", err
	}

	var selections []TaskSelection
	if err := json.Unmarshal([]byte(configData.Value), &selections); err != nil {
		return "", "", err
	}

	for _, sel := range selections {
		if sel.TaskID == taskID {
			return sel.Primary, sel.Fallback, nil
		}
	}

	return "", "", fmt.Errorf("no selection found for task: %s", taskID)
}

// GetModelStatus returns status of all models (for UI display)
func (ms *ModelSelector) GetModelStatus() map[string]bool {
	status := make(map[string]bool)
	for _, model := range ms.models.GetAllModels() {
		status[model.ID] = model.Enabled
	}
	return status
}
