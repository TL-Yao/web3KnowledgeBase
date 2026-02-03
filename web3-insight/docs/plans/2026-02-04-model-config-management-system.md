# Model Configuration Management System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a centralized model configuration management system with YAML-based model registry and database-backed user preferences, including robust fallback handling for unavailable models.

**Architecture:** Two-layer configuration system where developers maintain model registries in YAML files (version controlled), users select preferences stored in database, and the backend validates configuration on startup and at runtime with intelligent fallback to secondary models when primary models are unavailable.

**Tech Stack:**
- Backend: Go, Viper (YAML config), GORM (database)
- Frontend: Next.js, TypeScript, React Query, shadcn/ui
- Config Files: YAML (models.yaml, routing.yaml)
- Database: PostgreSQL (configs table)

---

## Problem Summary

**Current Issues:**
1. Frontend admin panel has hardcoded model options and not connected to backend
2. Backend lacks a model registry system (no definition of available models)
3. No task routing strategy configuration
4. No handling of model availability conflicts (disabled/deleted models)

**Design Decisions:**
- **Configuration Strategy**: YAML (developer config) + Database (user selections)
- **Conflict Handling**: Validate on startup + runtime, use task-level fallback, fail task if both models unavailable
- **Fallback Strategy**: Each task has primary + fallback model, stop task if both unavailable
- **UI Behavior**: Show fallback model when primary unavailable, notify user via banner/toast

---

## Task 1: Create Model Registry Configuration Files

**Files:**
- Create: `web3-insight/backend/config/models.yaml`
- Create: `web3-insight/backend/config/routing.yaml`

**Step 1: Create models.yaml with model registry**

Create `backend/config/models.yaml`:

```yaml
# Model Registry - defines all available models
# Maintained by developers, version controlled in Git

local_models:
  - id: "llama3:70b"
    name: "Llama 3 70B"
    provider: "ollama"
    enabled: true
    capabilities: ["chat", "generation", "summarization"]
    context_window: 8192
    cost_per_1k_tokens: 0.0

  - id: "qwen2.5:32b"
    name: "Qwen 2.5 32B"
    provider: "ollama"
    enabled: true
    capabilities: ["chat", "generation", "translation", "summarization"]
    context_window: 32768
    cost_per_1k_tokens: 0.0

  - id: "mistral:7b"
    name: "Mistral 7B"
    provider: "ollama"
    enabled: false  # Example of disabled model
    capabilities: ["chat", "generation"]
    context_window: 8192
    cost_per_1k_tokens: 0.0

cloud_models:
  - id: "claude-sonnet-4-20250514"
    name: "Claude Sonnet 4"
    provider: "anthropic"
    enabled: true
    capabilities: ["chat", "generation", "summarization", "analysis"]
    context_window: 200000
    cost_per_1k_tokens: 0.003

  - id: "claude-haiku-3-20240307"
    name: "Claude Haiku 3"
    provider: "anthropic"
    enabled: true
    capabilities: ["chat", "summarization", "translation"]
    context_window: 200000
    cost_per_1k_tokens: 0.00025

  - id: "gpt-4o"
    name: "GPT-4o"
    provider: "openai"
    enabled: false  # Disabled by default
    capabilities: ["chat", "generation", "analysis"]
    context_window: 128000
    cost_per_1k_tokens: 0.005
```

**Step 2: Create routing.yaml with task definitions**

Create `backend/config/routing.yaml`:

```yaml
# Task Routing Strategy - defines task types and default model assignments
# Maintained by developers, version controlled in Git

task_types:
  - id: "simple_generation"
    name: "内容生成（简单）"
    description: "生成简单的文章内容"
    default_primary: "llama3:70b"
    default_fallback: "claude-haiku-3-20240307"
    required_capability: "generation"

  - id: "complex_generation"
    name: "内容生成（复杂）"
    description: "生成复杂的深度分析文章"
    default_primary: "claude-sonnet-4-20250514"
    default_fallback: "llama3:70b"
    required_capability: "generation"

  - id: "summarization"
    name: "摘要/分类"
    description: "生成文章摘要和自动分类"
    default_primary: "qwen2.5:32b"
    default_fallback: "claude-haiku-3-20240307"
    required_capability: "summarization"

  - id: "chat"
    name: "问答对话"
    description: "实时问答交互"
    default_primary: "llama3:70b"
    default_fallback: "claude-sonnet-4-20250514"
    required_capability: "chat"

  - id: "translation"
    name: "翻译"
    description: "内容翻译"
    default_primary: "qwen2.5:32b"
    default_fallback: "claude-haiku-3-20240307"
    required_capability: "translation"
```

**Step 3: Verify files are created**

Run: `ls -la /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend/config/`
Expected: Should see `models.yaml` and `routing.yaml`

**Step 4: Commit configuration files**

```bash
git add web3-insight/backend/config/models.yaml
git add web3-insight/backend/config/routing.yaml
git commit -m "feat(config): add model registry and task routing configuration files"
```

---

## Task 2: Create Backend Configuration Loader

**Files:**
- Create: `web3-insight/backend/internal/config/models.go`
- Create: `web3-insight/backend/internal/config/routing.go`
- Modify: `web3-insight/backend/internal/config/config.go`

**Step 1: Create models.go for model registry**

Create `backend/internal/config/models.go`:

```go
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
```

**Step 2: Create routing.go for task routing**

Create `backend/internal/config/routing.go`:

```go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TaskType represents a task with model routing configuration
type TaskType struct {
	ID              string `yaml:"id" json:"id"`
	Name            string `yaml:"name" json:"name"`
	Description     string `yaml:"description" json:"description"`
	DefaultPrimary  string `yaml:"default_primary" json:"defaultPrimary"`
	DefaultFallback string `yaml:"default_fallback" json:"defaultFallback"`
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
	for _, task := range rc.TaskTypes {
		if task.ID == id {
			return &task, nil
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
```

**Step 3: Update config.go to load new configurations**

Modify `backend/internal/config/config.go`, add to Config struct:

```go
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Worker   WorkerConfig   `mapstructure:"worker"`
	Search   SearchConfig   `mapstructure:"search"`

	// New fields
	Models  *ModelsConfig  // Not from YAML, loaded separately
	Routing *RoutingConfig // Not from YAML, loaded separately
}
```

Add new function at end of file:

```go
// LoadAll loads all configuration files (config.yaml, models.yaml, routing.yaml)
func LoadAll() (*Config, error) {
	// Load main config.yaml
	cfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config.yaml: %w", err)
	}

	// Load models.yaml
	modelsPath := "./config/models.yaml"
	if _, err := os.Stat(modelsPath); err == nil {
		models, err := LoadModels(modelsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load models.yaml: %w", err)
		}
		cfg.Models = models
	} else {
		return nil, fmt.Errorf("models.yaml not found at %s", modelsPath)
	}

	// Load routing.yaml
	routingPath := "./config/routing.yaml"
	if _, err := os.Stat(routingPath); err == nil {
		routing, err := LoadRouting(routingPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load routing.yaml: %w", err)
		}
		cfg.Routing = routing
	} else {
		return nil, fmt.Errorf("routing.yaml not found at %s", routingPath)
	}

	// Validate task models against model registry
	if warnings := cfg.Routing.ValidateTaskModels(cfg.Models); len(warnings) > 0 {
		fmt.Println("⚠️  Configuration warnings detected:")
		for _, warning := range warnings {
			fmt.Printf("   - %s\n", warning)
		}
	}

	return cfg, nil
}
```

**Step 4: Verify Go compiles**

Run: `/usr/local/go/bin/go build -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./...`
Expected: No compilation errors

**Step 5: Commit backend config loader**

```bash
git add web3-insight/backend/internal/config/models.go
git add web3-insight/backend/internal/config/routing.go
git add web3-insight/backend/internal/config/config.go
git commit -m "feat(config): add model and routing configuration loaders"
```

---

## Task 3: Create Model Selection API Endpoints

**Files:**
- Create: `web3-insight/backend/internal/api/model_config.go`
- Modify: `web3-insight/backend/internal/api/router.go`
- Modify: `web3-insight/backend/cmd/server/main.go`

**Step 1: Create model_config.go handler**

Create `backend/internal/api/model_config.go`:

```go
package api

import (
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
	if err := json.Unmarshal([]byte(configData.Value), &selections); err != nil {
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

	// Save to database as JSON
	jsonData, err := json.Marshal(validatedSelections)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize"})
		return
	}

	err = h.configRepo.Set("model_selections", string(jsonData), "User's model selection preferences")
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
		primaryAvailable := h.models.IsModelAvailable(sel.Primary)
		fallbackAvailable := h.models.IsModelAvailable(sel.Fallback)

		// If primary is unavailable, keep selection but mark it
		// Frontend will show fallback as current model
		if !primaryAvailable || !fallbackAvailable {
			// Keep the selection as-is, validation happens at runtime
			// Frontend will handle UI display
		}

		validated = append(validated, sel)
	}

	return validated
}
```

Add import at top:

```go
import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/repository"
)
```

**Step 2: Register routes in router.go**

Modify `backend/internal/api/router.go`, in NewServer function add:

```go
modelConfigHandler := NewModelConfigHandler(configRepo, &cfg.Models, &cfg.Routing)
```

In NewRouterWithDB function, add model config routes:

```go
// Model Configuration
models := api.Group("/models")
{
	models.GET("/registry", server.modelConfigHandler.GetModelsRegistry)
	models.GET("/tasks", server.modelConfigHandler.GetTaskTypes)
	models.GET("/selections", server.modelConfigHandler.GetUserSelections)
	models.PUT("/selections", server.modelConfigHandler.UpdateUserSelections)
}
```

Update Server struct to include the handler:

```go
type Server struct {
	config              *config.Config
	db                  *gorm.DB
	articleHandler      *ArticleHandler
	categoryHandler     *CategoryHandler
	configHandler       *ConfigHandler
	taskHandler         *TaskHandler
	searchHandler       *SearchHandler
	chatHandler         *ChatHandler
	modelConfigHandler  *ModelConfigHandler  // Add this
}
```

**Step 3: Update main.go to use LoadAll**

Modify `backend/cmd/server/main.go`, change:

```go
// Load configuration
cfg, err := config.Load()
```

To:

```go
// Load all configuration (config.yaml, models.yaml, routing.yaml)
cfg, err := config.LoadAll()
```

**Step 4: Test API endpoints**

Start backend:
```bash
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend cmd/server/main.go
```

Test endpoints:
```bash
curl http://localhost:8080/api/models/registry
curl http://localhost:8080/api/models/tasks
curl http://localhost:8080/api/models/selections
```

Expected: JSON responses with model registry, task types, and default selections

**Step 5: Commit API endpoints**

```bash
git add web3-insight/backend/internal/api/model_config.go
git add web3-insight/backend/internal/api/router.go
git add web3-insight/backend/cmd/server/main.go
git commit -m "feat(api): add model configuration API endpoints"
```

---

## Task 4: Create Model Selector Service with Fallback Logic

**Files:**
- Create: `web3-insight/backend/internal/service/model_selector.go`

**Step 1: Create model selector service**

Create `backend/internal/service/model_selector.go`:

```go
package service

import (
	"encoding/json"
	"fmt"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/repository"
)

// ModelSelector handles model selection with fallback logic
type ModelSelector struct {
	models     *config.ModelsConfig
	routing    *config.RoutingConfig
	configRepo *repository.ConfigRepository
}

// ModelSelectionResult contains the selected model and metadata
type ModelSelectionResult struct {
	ModelID       string `json:"modelId"`
	IsFallback    bool   `json:"isFallback"`
	PrimaryFailed bool   `json:"primaryFailed"`
	Reason        string `json:"reason,omitempty"`
}

func NewModelSelector(
	models *config.ModelsConfig,
	routing *config.RoutingConfig,
	configRepo *repository.ConfigRepository,
) *ModelSelector {
	return &ModelSelector{
		models:     models,
		routing:    routing,
		configRepo: configRepo,
	}
}

// SelectModelForTask selects appropriate model for a task with fallback logic
// Returns: model ID, whether fallback was used, error if no models available
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

	type TaskSelection struct {
		TaskID   string `json:"taskId"`
		Primary  string `json:"primary"`
		Fallback string `json:"fallback"`
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
```

**Step 2: Verify Go compiles**

Run: `/usr/local/go/bin/go build -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./...`
Expected: No compilation errors

**Step 3: Commit model selector service**

```bash
git add web3-insight/backend/internal/service/model_selector.go
git commit -m "feat(service): add model selector with fallback logic"
```

---

## Task 5: Update Frontend API Types

**Files:**
- Modify: `web3-insight/frontend/lib/api.ts`

**Step 1: Add model configuration types**

Add to `frontend/lib/api.ts`:

```typescript
// Model Configuration Types
export interface Model {
  id: string
  name: string
  provider: string
  enabled: boolean
  capabilities: string[]
  contextWindow: number
  costPer1kTokens: number
}

export interface ModelsConfig {
  localModels: Model[]
  cloudModels: Model[]
}

export interface TaskType {
  id: string
  name: string
  description: string
  defaultPrimary: string
  defaultFallback: string
  requiredCapability: string
}

export interface RoutingConfig {
  taskTypes: TaskType[]
}

export interface TaskSelection {
  taskId: string
  primary: string
  fallback: string
}

export interface ModelSelectionResult {
  modelId: string
  isFallback: boolean
  primaryFailed: boolean
  reason?: string
}
```

**Step 2: Add model configuration API methods**

Add to `frontend/lib/api.ts`:

```typescript
// Model Configuration API
export const modelConfigAPI = {
  // Get available models from registry
  getModelsRegistry: () =>
    fetchAPI<ModelsConfig>('/api/models/registry'),

  // Get task types
  getTaskTypes: () =>
    fetchAPI<RoutingConfig>('/api/models/tasks'),

  // Get user's model selections
  getUserSelections: () =>
    fetchAPI<TaskSelection[]>('/api/models/selections'),

  // Update user's model selections
  updateUserSelections: (selections: TaskSelection[]) =>
    fetchAPI<TaskSelection[]>('/api/models/selections', {
      method: 'PUT',
      body: JSON.stringify(selections),
    }),
}
```

**Step 3: Verify TypeScript compiles**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm run build`
Expected: No type errors

**Step 4: Commit frontend types**

```bash
git add web3-insight/frontend/lib/api.ts
git commit -m "feat(frontend): add model configuration API types"
```

---

## Task 6: Rebuild Frontend Model Config Component

**Files:**
- Modify: `web3-insight/frontend/components/admin/model-config.tsx`

**Step 1: Replace hardcoded component with API-driven version**

Replace entire content of `frontend/components/admin/model-config.tsx`:

```typescript
'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'
import { Check, AlertCircle, RefreshCw } from 'lucide-react'
import { modelConfigAPI, type TaskSelection, type Model } from '@/lib/api'
import { useToast } from '@/hooks/use-toast'

export function ModelConfig() {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [selections, setSelections] = useState<TaskSelection[]>([])

  // Fetch model registry
  const { data: modelsRegistry, isLoading: loadingModels } = useQuery({
    queryKey: ['models-registry'],
    queryFn: modelConfigAPI.getModelsRegistry,
  })

  // Fetch task types
  const { data: routingConfig, isLoading: loadingTasks } = useQuery({
    queryKey: ['task-types'],
    queryFn: modelConfigAPI.getTaskTypes,
  })

  // Fetch user selections
  const { data: userSelections, isLoading: loadingSelections } = useQuery({
    queryKey: ['model-selections'],
    queryFn: modelConfigAPI.getUserSelections,
    onSuccess: (data) => setSelections(data),
  })

  // Save mutation
  const saveMutation = useMutation({
    mutationFn: modelConfigAPI.updateUserSelections,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model-selections'] })
      toast({
        title: '保存成功',
        description: '模型配置已更新',
      })
    },
    onError: () => {
      toast({
        title: '保存失败',
        description: '请稍后重试',
        variant: 'destructive',
      })
    },
  })

  const handleSave = () => {
    saveMutation.mutate(selections)
  }

  const updateSelection = (taskId: string, field: 'primary' | 'fallback', value: string) => {
    setSelections(prev =>
      prev.map(sel =>
        sel.taskId === taskId
          ? { ...sel, [field]: value }
          : sel
      )
    )
  }

  const isModelAvailable = (modelId: string): boolean => {
    if (!modelsRegistry) return false
    const allModels = [...modelsRegistry.localModels, ...modelsRegistry.cloudModels]
    const model = allModels.find(m => m.id === modelId)
    return model ? model.enabled : false
  }

  const getModelDisplayName = (modelId: string): string => {
    if (!modelsRegistry) return modelId
    const allModels = [...modelsRegistry.localModels, ...modelsRegistry.cloudModels]
    const model = allModels.find(m => m.id === modelId)
    return model ? model.name : modelId
  }

  const renderModelSelect = (
    taskId: string,
    currentValue: string,
    field: 'primary' | 'fallback',
    capability: string
  ) => {
    if (!modelsRegistry) return null

    const allModels = [...modelsRegistry.localModels, ...modelsRegistry.cloudModels]
    const availableModels = allModels.filter(m =>
      m.capabilities.includes(capability)
    )

    const isCurrentAvailable = isModelAvailable(currentValue)

    return (
      <div className="flex items-center gap-2">
        <Select
          value={currentValue}
          onValueChange={(value) => updateSelection(taskId, field, value)}
        >
          <SelectTrigger className="h-8 w-[200px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {availableModels.map(model => (
              <SelectItem
                key={model.id}
                value={model.id}
                disabled={!model.enabled}
              >
                {model.name} {!model.enabled && '(已禁用)'}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {!isCurrentAvailable && (
          <Badge variant="outline" className="text-yellow-600 border-yellow-600">
            <AlertCircle className="w-3 h-3 mr-1" />
            不可用
          </Badge>
        )}
      </div>
    )
  }

  if (loadingModels || loadingTasks || loadingSelections) {
    return <div className="flex items-center justify-center py-12">加载中...</div>
  }

  const hasUnavailableModels = selections.some(sel =>
    !isModelAvailable(sel.primary) || !isModelAvailable(sel.fallback)
  )

  return (
    <div className="space-y-6">
      {/* Warning Banner */}
      {hasUnavailableModels && (
        <Card className="border-yellow-500 bg-yellow-50">
          <CardContent className="pt-6">
            <div className="flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-yellow-600 mt-0.5" />
              <div>
                <p className="font-medium text-yellow-900">检测到模型配置问题</p>
                <p className="text-sm text-yellow-700 mt-1">
                  部分任务的首选模型不可用。系统将自动使用备用模型。建议更新配置。
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Local Models */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">本地模型 (Ollama)</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg divide-y">
            {modelsRegistry?.localModels.map((model) => (
              <div key={model.id} className="flex items-center justify-between p-3">
                <div className="flex items-center gap-3">
                  <span className="font-mono text-sm">{model.name}</span>
                  <Badge variant={model.enabled ? 'default' : 'secondary'}>
                    {model.enabled ? '已启用' : '已禁用'}
                  </Badge>
                </div>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span className="text-xs">{model.capabilities.join(', ')}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Cloud Models */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">云端模型</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg divide-y">
            {modelsRegistry?.cloudModels.map((model) => (
              <div key={model.id} className="flex items-center justify-between p-3">
                <div className="flex items-center gap-3">
                  <span className="font-mono text-sm">{model.name}</span>
                  <Badge variant={model.enabled ? 'default' : 'secondary'}>
                    {model.enabled ? '已启用' : '已禁用'}
                  </Badge>
                </div>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span className="text-xs">${model.costPer1kTokens}/1K tokens</span>
                  <span className="text-xs">{model.capabilities.join(', ')}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Task Routing */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">任务模型路由</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg">
            <div className="grid grid-cols-4 gap-4 p-3 bg-muted text-sm font-medium">
              <span>任务类型</span>
              <span>首选模型</span>
              <span>备用模型</span>
              <span>状态</span>
            </div>
            {routingConfig?.taskTypes.map((task) => {
              const selection = selections.find(s => s.taskId === task.id)
              if (!selection) return null

              const primaryAvailable = isModelAvailable(selection.primary)
              const fallbackAvailable = isModelAvailable(selection.fallback)
              const bothUnavailable = !primaryAvailable && !fallbackAvailable

              return (
                <div key={task.id} className="grid grid-cols-4 gap-4 p-3 border-t text-sm items-center">
                  <div>
                    <div className="font-medium">{task.name}</div>
                    <div className="text-xs text-muted-foreground">{task.description}</div>
                  </div>

                  {renderModelSelect(task.id, selection.primary, 'primary', task.requiredCapability)}
                  {renderModelSelect(task.id, selection.fallback, 'fallback', task.requiredCapability)}

                  <div>
                    {bothUnavailable ? (
                      <Badge variant="destructive">
                        <AlertCircle className="w-3 h-3 mr-1" />
                        无可用模型
                      </Badge>
                    ) : !primaryAvailable ? (
                      <Badge variant="outline" className="text-yellow-600 border-yellow-600">
                        使用备用模型
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="text-green-600 border-green-600">
                        ✓ 正常
                      </Badge>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end gap-2">
        <Button
          variant="outline"
          onClick={() => setSelections(userSelections || [])}
        >
          <RefreshCw className="w-4 h-4 mr-2" />
          重置
        </Button>
        <Button
          onClick={handleSave}
          disabled={saveMutation.isLoading}
        >
          <Check className="w-4 h-4 mr-2" />
          保存配置
        </Button>
      </div>
    </div>
  )
}
```

**Step 2: Verify frontend builds**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm run build`
Expected: No errors

**Step 3: Commit updated component**

```bash
git add web3-insight/frontend/components/admin/model-config.tsx
git commit -m "feat(frontend): rebuild model config component with API integration"
```

---

## Task 7: Update CLAUDE.md with Model Fallback Pattern

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Add Model Fallback Pattern section**

Add new section to CLAUDE.md after "Chrome Automation Capabilities" section:

```markdown
## Model Fallback Pattern

**CRITICAL:** All features that use AI models MUST implement the model fallback pattern to handle unavailable models gracefully.

### Core Pattern

When any feature needs to use an AI model:

1. **Use ModelSelector service** to get the appropriate model
2. **Handle three possible outcomes**:
   - ✅ Primary model available → Use it
   - ⚠️ Primary unavailable, fallback available → Use fallback, log warning, notify user
   - ❌ Both unavailable → Stop task, return error to user

### Implementation Requirements

**Backend (Go):**

```go
// In any service that uses AI models
modelSelector := service.NewModelSelector(cfg.Models, cfg.Routing, configRepo)

// Select model for task
result, err := modelSelector.SelectModelForTask("simple_generation")
if err != nil {
    // Both primary and fallback unavailable
    return nil, fmt.Errorf("cannot execute task: %w", err)
}

// Log if using fallback
if result.IsFallback {
    log.Printf("⚠️  Using fallback model: %s", result.Reason)
}

// Use the selected model
modelID := result.ModelID
// ... proceed with LLM call
```

**Frontend (TypeScript):**

```typescript
// When displaying task status or executing AI operations
const { data: selections } = useQuery({
  queryKey: ['model-selections'],
  queryFn: modelConfigAPI.getUserSelections,
})

const { data: modelsRegistry } = useQuery({
  queryKey: ['models-registry'],
  queryFn: modelConfigAPI.getModelsRegistry,
})

// Check model availability
const isModelAvailable = (modelId: string) => {
  const allModels = [
    ...(modelsRegistry?.localModels ?? []),
    ...(modelsRegistry?.cloudModels ?? [])
  ]
  const model = allModels.find(m => m.id === modelId)
  return model?.enabled ?? false
}

// Show warning if primary model unavailable
{!isModelAvailable(selection.primary) && (
  <Alert variant="warning">
    首选模型不可用，当前使用备用模型: {selection.fallback}
  </Alert>
)}
```

### Error Messages

**User-facing error messages:**
- ❌ Primary unavailable: "首选模型不可用，当前使用备用模型"
- ❌ Both unavailable: "无可用模型，请在管理面板中配置模型"

**Log messages (backend):**
- `⚠️  Using fallback model for task '%s': primary '%s' unavailable`
- `❌ No available models for task '%s': both primary and fallback unavailable`

### Features That MUST Use This Pattern

All AI-dependent features must implement this pattern:

1. **Content Generation** (`generator.go`)
   - Simple generation tasks
   - Complex generation tasks

2. **Summarization** (`summarizer.go`)
   - Article summarization
   - Chat history summarization

3. **Classification** (`classifier.go`)
   - Automatic categorization
   - Tag generation

4. **Chat/Q&A** (`chat.go`)
   - Real-time chat
   - Article-based Q&A

5. **Translation** (future feature)
   - Content translation

6. **Research** (`research.go`)
   - Instant research queries

### Configuration Files

Model availability is defined in:
- `backend/config/models.yaml` - Model registry (what models exist and their status)
- `backend/config/routing.yaml` - Task routing (default model assignments)
- Database `configs` table - User selections (which models user chose)

### Testing Checklist

When implementing AI features, verify:
- [ ] Uses `ModelSelector` service
- [ ] Handles primary model available case
- [ ] Handles fallback model case (logs warning)
- [ ] Handles both unavailable case (returns error)
- [ ] Frontend shows appropriate status/warning
- [ ] User receives clear error message if task fails
```

**Step 2: Commit documentation**

```bash
git add CLAUDE.md
git commit -m "docs: add model fallback pattern requirements to CLAUDE.md"
```

---

## Task 8: Integration Testing

**Files:**
- Test: End-to-end testing of the configuration system

**Step 1: Start all services**

```bash
/Users/tongleyao/claudeProjects/explorerResearch/web3-insight/scripts/start-all.sh
```

Expected: Services start, see validation warnings in logs if any

**Step 2: Test model registry API**

```bash
curl http://localhost:8080/api/models/registry | jq
```

Expected: JSON with localModels and cloudModels arrays

**Step 3: Test task types API**

```bash
curl http://localhost:8080/api/models/tasks | jq
```

Expected: JSON with taskTypes array

**Step 4: Test user selections API (GET)**

```bash
curl http://localhost:8080/api/models/selections | jq
```

Expected: JSON with default selections from routing.yaml

**Step 5: Test user selections API (PUT)**

```bash
curl -X PUT http://localhost:8080/api/models/selections \
  -H "Content-Type: application/json" \
  -d '[{"taskId":"chat","primary":"qwen2.5:32b","fallback":"claude-haiku-3-20240307"}]' | jq
```

Expected: JSON with updated selections

**Step 6: Open frontend model config page**

Navigate to: http://localhost:3000/admin/config

Expected:
- See local and cloud models from backend
- See task routing table with model selects
- Model status badges (available/unavailable) correct
- Warning banner if any models unavailable

**Step 7: Test model selection update in UI**

1. Change a task's primary model
2. Click "保存配置"
3. Refresh page
4. Verify selection persisted

**Step 8: Test unavailable model handling**

1. Edit `backend/config/models.yaml`
2. Set `llama3:70b` to `enabled: false`
3. Restart backend
4. Refresh frontend
5. Verify:
   - Tasks using llama3:70b show warning
   - Status badge shows "使用备用模型"
   - Warning banner appears at top

**Step 9: Stop services**

```bash
/Users/tongleyao/claudeProjects/explorerResearch/web3-insight/scripts/stop-all.sh
```

**Step 10: Document test results**

No commit needed, manual testing verification.

---

## Success Criteria

- ✅ YAML configuration files (models.yaml, routing.yaml) created and loaded
- ✅ Backend API endpoints return model registry and task types
- ✅ User selections stored in database
- ✅ Frontend displays real backend data (not hardcoded)
- ✅ Model availability validation on startup (warnings logged)
- ✅ ModelSelector service implements fallback logic
- ✅ Frontend shows warnings for unavailable models
- ✅ CLAUDE.md documents model fallback pattern
- ✅ Integration tests pass

## Notes

- Keep YAML files simple (YAGNI principle)
- Focus on core functionality first
- Model fallback pattern is CRITICAL for all AI features
- Frontend only displays backend state (no hardcoded assumptions)
- Configuration changes require backend restart
- User selections survive backend restarts (stored in database)
