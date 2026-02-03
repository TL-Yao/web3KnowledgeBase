package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/datatypes"
)

// MockConfigRepository is a mock for ConfigRepository
type MockConfigRepository struct {
	mock.Mock
}

func (m *MockConfigRepository) Get(key string) (*model.Config, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Config), args.Error(1)
}

func (m *MockConfigRepository) Set(key, value, description string) error {
	args := m.Called(key, value, description)
	return args.Error(0)
}

func (m *MockConfigRepository) GetAll() ([]model.Config, error) {
	args := m.Called()
	return args.Get(0).([]model.Config), args.Error(1)
}

func (m *MockConfigRepository) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockConfigRepository) GetMap() (map[string]string, error) {
	args := m.Called()
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockConfigRepository) SetMultiple(configs map[string]string) error {
	args := m.Called(configs)
	return args.Error(0)
}

func (m *MockConfigRepository) SetJSON(key string, jsonValue []byte, description string) error {
	args := m.Called(key, jsonValue, description)
	return args.Error(0)
}

// Test helper to create test models config
func createTestModelsConfig() *config.ModelsConfig {
	return &config.ModelsConfig{
		LocalModels: []config.Model{
			{
				ID:              "qwen2.5:7b",
				Name:            "Qwen 2.5 7B",
				Provider:        "ollama",
				Enabled:         true,
				Capabilities:    []string{"chat", "embedding"},
				ContextWindow:   32768,
				CostPer1KTokens: 0,
			},
			{
				ID:              "nomic-embed-text",
				Name:            "Nomic Embed Text",
				Provider:        "ollama",
				Enabled:         true,
				Capabilities:    []string{"embedding"},
				ContextWindow:   8192,
				CostPer1KTokens: 0,
			},
		},
		CloudModels: []config.Model{
			{
				ID:              "gpt-4o",
				Name:            "GPT-4o",
				Provider:        "openai",
				Enabled:         false, // Disabled
				Capabilities:    []string{"chat", "vision"},
				ContextWindow:   128000,
				CostPer1KTokens: 0.03,
			},
		},
	}
}

// Test helper to create test routing config
func createTestRoutingConfig() *config.RoutingConfig {
	return &config.RoutingConfig{
		TaskTypes: []config.TaskType{
			{
				ID:                 "summarize",
				Name:               "内容摘要",
				Description:        "生成文章摘要",
				DefaultPrimary:     "qwen2.5:7b",
				DefaultFallback:    "nomic-embed-text",
				RequiredCapability: "chat",
			},
			{
				ID:                 "classify",
				Name:               "内容分类",
				Description:        "智能分类",
				DefaultPrimary:     "gpt-4o",
				DefaultFallback:    "qwen2.5:7b",
				RequiredCapability: "chat",
			},
		},
	}
}

func TestModelSelector_SelectModelForTask_WithPrimaryAvailable(t *testing.T) {
	// Arrange
	mockRepo := new(MockConfigRepository)
	modelsConfig := createTestModelsConfig()
	routingConfig := createTestRoutingConfig()
	selector := NewModelSelector(modelsConfig, routingConfig, mockRepo)

	// Mock no user selections (will use defaults)
	mockRepo.On("Get", "model_selections").Return(nil, assert.AnError)

	// Act
	result, err := selector.SelectModelForTask("summarize")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "qwen2.5:7b", result.ModelID)
	assert.False(t, result.IsFallback)
	assert.False(t, result.PrimaryFailed)
	assert.Empty(t, result.Reason)
}

func TestModelSelector_SelectModelForTask_PrimaryUnavailableUseFallback(t *testing.T) {
	// Arrange
	mockRepo := new(MockConfigRepository)
	modelsConfig := createTestModelsConfig()
	routingConfig := createTestRoutingConfig()
	selector := NewModelSelector(modelsConfig, routingConfig, mockRepo)

	// Mock no user selections (will use defaults)
	mockRepo.On("Get", "model_selections").Return(nil, assert.AnError)

	// Act - task "classify" has primary "gpt-4o" which is disabled
	result, err := selector.SelectModelForTask("classify")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "qwen2.5:7b", result.ModelID)
	assert.True(t, result.IsFallback)
	assert.True(t, result.PrimaryFailed)
	assert.Contains(t, result.Reason, "Primary model 'gpt-4o' unavailable")
}

func TestModelSelector_SelectModelForTask_BothUnavailable(t *testing.T) {
	// Arrange
	mockRepo := new(MockConfigRepository)
	modelsConfig := createTestModelsConfig()
	routingConfig := &config.RoutingConfig{
		TaskTypes: []config.TaskType{
			{
				ID:                 "test-task",
				Name:               "Test Task",
				Description:        "Test",
				DefaultPrimary:     "gpt-4o",        // Disabled
				DefaultFallback:    "gpt-3.5-turbo", // Does not exist
				RequiredCapability: "chat",
			},
		},
	}
	selector := NewModelSelector(modelsConfig, routingConfig, mockRepo)

	// Mock no user selections
	mockRepo.On("Get", "model_selections").Return(nil, assert.AnError)

	// Act
	result, err := selector.SelectModelForTask("test-task")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no available models for task 'test-task'")
	assert.Contains(t, err.Error(), "primary 'gpt-4o' and fallback 'gpt-3.5-turbo' both unavailable")
}

func TestModelSelector_SelectModelForTask_UnknownTask(t *testing.T) {
	// Arrange
	mockRepo := new(MockConfigRepository)
	modelsConfig := createTestModelsConfig()
	routingConfig := createTestRoutingConfig()
	selector := NewModelSelector(modelsConfig, routingConfig, mockRepo)

	// Mock no user selections
	mockRepo.On("Get", "model_selections").Return(nil, assert.AnError)

	// Act
	result, err := selector.SelectModelForTask("nonexistent-task")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown task type: nonexistent-task")
}

func TestModelSelector_SelectModelForTask_WithUserSelections(t *testing.T) {
	// Arrange
	mockRepo := new(MockConfigRepository)
	modelsConfig := createTestModelsConfig()
	routingConfig := createTestRoutingConfig()
	selector := NewModelSelector(modelsConfig, routingConfig, mockRepo)

	// Create user selections
	selections := []TaskSelection{
		{
			TaskID:   "summarize",
			Primary:  "nomic-embed-text", // User selected different model
			Fallback: "qwen2.5:7b",
		},
	}
	selectionsJSON, _ := json.Marshal(selections)

	// Mock user has saved selections
	mockRepo.On("Get", "model_selections").Return(&model.Config{
		Key:   "model_selections",
		Value: datatypes.JSON(selectionsJSON),
	}, nil)

	// Act
	result, err := selector.SelectModelForTask("summarize")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "nomic-embed-text", result.ModelID) // User's choice, not default
	assert.False(t, result.IsFallback)
}

func TestModelSelector_SelectModelForTask_UserSelectionPrimaryUnavailable(t *testing.T) {
	// Arrange
	mockRepo := new(MockConfigRepository)
	modelsConfig := createTestModelsConfig()
	routingConfig := createTestRoutingConfig()
	selector := NewModelSelector(modelsConfig, routingConfig, mockRepo)

	// Create user selections with unavailable primary
	selections := []TaskSelection{
		{
			TaskID:   "summarize",
			Primary:  "gpt-4o",       // Disabled
			Fallback: "qwen2.5:7b",   // Available
		},
	}
	selectionsJSON, _ := json.Marshal(selections)

	mockRepo.On("Get", "model_selections").Return(&model.Config{
		Key:   "model_selections",
		Value: datatypes.JSON(selectionsJSON),
	}, nil)

	// Act
	result, err := selector.SelectModelForTask("summarize")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "qwen2.5:7b", result.ModelID)
	assert.True(t, result.IsFallback)
	assert.True(t, result.PrimaryFailed)
	assert.Contains(t, result.Reason, "Primary model 'gpt-4o' unavailable")
}

func TestModelSelector_GetModelStatus(t *testing.T) {
	// Arrange
	mockRepo := new(MockConfigRepository)
	modelsConfig := createTestModelsConfig()
	routingConfig := createTestRoutingConfig()
	selector := NewModelSelector(modelsConfig, routingConfig, mockRepo)

	// Act
	status := selector.GetModelStatus()

	// Assert
	assert.NotNil(t, status)
	assert.Equal(t, true, status["qwen2.5:7b"])
	assert.Equal(t, true, status["nomic-embed-text"])
	assert.Equal(t, false, status["gpt-4o"]) // Disabled
}

func TestModelSelector_getUserSelection_NoSelections(t *testing.T) {
	// Arrange
	mockRepo := new(MockConfigRepository)
	modelsConfig := createTestModelsConfig()
	routingConfig := createTestRoutingConfig()
	selector := NewModelSelector(modelsConfig, routingConfig, mockRepo)

	// Mock no selections in database
	mockRepo.On("Get", "model_selections").Return(nil, assert.AnError)

	// Act
	primary, fallback, err := selector.getUserSelection("summarize")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, primary)
	assert.Empty(t, fallback)
}

func TestModelSelector_getUserSelection_TaskNotInSelections(t *testing.T) {
	// Arrange
	mockRepo := new(MockConfigRepository)
	modelsConfig := createTestModelsConfig()
	routingConfig := createTestRoutingConfig()
	selector := NewModelSelector(modelsConfig, routingConfig, mockRepo)

	// Create selections that don't include the task we're looking for
	selections := []TaskSelection{
		{
			TaskID:   "other-task",
			Primary:  "qwen2.5:7b",
			Fallback: "nomic-embed-text",
		},
	}
	selectionsJSON, _ := json.Marshal(selections)

	mockRepo.On("Get", "model_selections").Return(&model.Config{
		Key:   "model_selections",
		Value: datatypes.JSON(selectionsJSON),
	}, nil)

	// Act
	primary, fallback, err := selector.getUserSelection("summarize")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no selection found for task: summarize")
	assert.Empty(t, primary)
	assert.Empty(t, fallback)
}
