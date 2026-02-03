package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Model represents a single model configuration
type Model struct {
	ID              string   `yaml:"id" json:"id"`
	Name            string   `yaml:"name" json:"name"`
	Provider        string   `yaml:"provider" json:"provider"`
	Enabled         bool     `yaml:"enabled" json:"enabled"`
	Capabilities    []string `yaml:"capabilities" json:"capabilities"`
	ContextWindow   int      `yaml:"context_window" json:"contextWindow"`
	CostPer1KTokens float64  `yaml:"cost_per_1k_tokens" json:"costPer1kTokens"`
}

// ModelsConfig holds all model configurations
type ModelsConfig struct {
	LocalModels []Model `yaml:"local_models" json:"localModels"`
	CloudModels []Model `yaml:"cloud_models" json:"cloudModels"`
}

// LoadModels reads and parses models.yaml
func LoadModels(path string) (*ModelsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read models.yaml: %w", err)
	}

	var config ModelsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse models.yaml: %w", err)
	}

	return &config, nil
}

// GetAllModels returns all models (local + cloud)
func (mc *ModelsConfig) GetAllModels() []Model {
	return append(mc.LocalModels, mc.CloudModels...)
}

// GetEnabledModels returns only enabled models
func (mc *ModelsConfig) GetEnabledModels() []Model {
	var enabled []Model
	for _, model := range mc.GetAllModels() {
		if model.Enabled {
			enabled = append(enabled, model)
		}
	}
	return enabled
}

// GetModelByID finds a model by its ID
func (mc *ModelsConfig) GetModelByID(id string) (*Model, error) {
	for _, model := range mc.GetAllModels() {
		if model.ID == id {
			return &model, nil
		}
	}
	return nil, fmt.Errorf("model not found: %s", id)
}

// IsModelAvailable checks if a model exists and is enabled
func (mc *ModelsConfig) IsModelAvailable(id string) bool {
	model, err := mc.GetModelByID(id)
	if err != nil {
		return false
	}
	return model.Enabled
}

// GetModelsByCapability returns models with specific capability
func (mc *ModelsConfig) GetModelsByCapability(capability string) []Model {
	var matching []Model
	for _, model := range mc.GetEnabledModels() {
		for _, cap := range model.Capabilities {
			if cap == capability {
				matching = append(matching, model)
				break
			}
		}
	}
	return matching
}
