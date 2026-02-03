package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TaskType represents a task with model routing configuration
type TaskType struct {
	ID                 string `yaml:"id" json:"id"`
	Name               string `yaml:"name" json:"name"`
	Description        string `yaml:"description" json:"description"`
	DefaultPrimary     string `yaml:"default_primary" json:"defaultPrimary"`
	DefaultFallback    string `yaml:"default_fallback" json:"defaultFallback"`
	RequiredCapability string `yaml:"required_capability" json:"requiredCapability"`
}

// RoutingConfig holds all task routing configurations
type RoutingConfig struct {
	TaskTypes []TaskType `yaml:"task_types" json:"taskTypes"`
}

// LoadRouting reads and parses routing.yaml
func LoadRouting(path string) (*RoutingConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read routing.yaml: %w", err)
	}

	var config RoutingConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse routing.yaml: %w", err)
	}

	return &config, nil
}

// GetTaskByID finds a task type by its ID
func (rc *RoutingConfig) GetTaskByID(id string) (*TaskType, error) {
	for i := range rc.TaskTypes {
		if rc.TaskTypes[i].ID == id {
			return &rc.TaskTypes[i], nil
		}
	}
	return nil, fmt.Errorf("task type not found: %s", id)
}

// ValidateTaskModels checks if task's default models exist in model registry
func (rc *RoutingConfig) ValidateTaskModels(models *ModelsConfig) []string {
	var warnings []string

	for _, task := range rc.TaskTypes {
		if !models.IsModelAvailable(task.DefaultPrimary) {
			warnings = append(warnings, fmt.Sprintf(
				"Task '%s': primary model '%s' not available",
				task.ID, task.DefaultPrimary,
			))
		}

		if !models.IsModelAvailable(task.DefaultFallback) {
			warnings = append(warnings, fmt.Sprintf(
				"Task '%s': fallback model '%s' not available",
				task.ID, task.DefaultFallback,
			))
		}
	}

	return warnings
}
