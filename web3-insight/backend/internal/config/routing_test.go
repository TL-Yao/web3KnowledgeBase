package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRouting(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		validate    func(*testing.T, *RoutingConfig)
	}{
		{
			name: "valid routing.yaml",
			yamlContent: `task_types:
  - id: "simple_generation"
    name: "内容生成（简单）"
    description: "生成简单的文章内容"
    default_primary: "llama3:70b"
    default_fallback: "claude-haiku-3-20240307"
    required_capability: "generation"

  - id: "chat"
    name: "问答对话"
    description: "实时问答交互"
    default_primary: "llama3:70b"
    default_fallback: "claude-sonnet-4-20250514"
    required_capability: "chat"
`,
			wantErr: false,
			validate: func(t *testing.T, rc *RoutingConfig) {
				if len(rc.TaskTypes) != 2 {
					t.Errorf("expected 2 task types, got %d", len(rc.TaskTypes))
				}
				if rc.TaskTypes[0].ID != "simple_generation" {
					t.Errorf("expected task ID 'simple_generation', got '%s'", rc.TaskTypes[0].ID)
				}
				if rc.TaskTypes[0].DefaultPrimary != "llama3:70b" {
					t.Errorf("expected default_primary 'llama3:70b', got '%s'", rc.TaskTypes[0].DefaultPrimary)
				}
			},
		},
		{
			name:        "file not found",
			yamlContent: "",
			wantErr:     true,
		},
		{
			name: "invalid yaml",
			yamlContent: `task_types:
  - id: "test"
    invalid yaml syntax [[[
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "file not found" {
				_, err := LoadRouting("/nonexistent/path/routing.yaml")
				if !tt.wantErr {
					t.Errorf("LoadRouting() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "routing.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			got, err := LoadRouting(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRouting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestRoutingConfig_GetTaskByID(t *testing.T) {
	rc := &RoutingConfig{
		TaskTypes: []TaskType{
			{ID: "task1", Name: "Task 1"},
			{ID: "task2", Name: "Task 2"},
		},
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		wantID  string
	}{
		{
			name:    "find existing task",
			id:      "task1",
			wantErr: false,
			wantID:  "task1",
		},
		{
			name:    "task not found",
			id:      "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rc.GetTaskByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTaskByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("GetTaskByID() got ID = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}

func TestRoutingConfig_ValidateTaskModels(t *testing.T) {
	models := &ModelsConfig{
		LocalModels: []Model{
			{ID: "llama3:70b", Name: "Llama 3", Enabled: true},
		},
		CloudModels: []Model{
			{ID: "claude-sonnet-4-20250514", Name: "Claude", Enabled: true},
			{ID: "disabled-model", Name: "Disabled", Enabled: false},
		},
	}

	tests := []struct {
		name         string
		routing      *RoutingConfig
		wantWarnings int
	}{
		{
			name: "all models available",
			routing: &RoutingConfig{
				TaskTypes: []TaskType{
					{
						ID:              "task1",
						DefaultPrimary:  "llama3:70b",
						DefaultFallback: "claude-sonnet-4-20250514",
					},
				},
			},
			wantWarnings: 0,
		},
		{
			name: "primary model unavailable",
			routing: &RoutingConfig{
				TaskTypes: []TaskType{
					{
						ID:              "task1",
						DefaultPrimary:  "nonexistent",
						DefaultFallback: "claude-sonnet-4-20250514",
					},
				},
			},
			wantWarnings: 1,
		},
		{
			name: "fallback model unavailable",
			routing: &RoutingConfig{
				TaskTypes: []TaskType{
					{
						ID:              "task1",
						DefaultPrimary:  "llama3:70b",
						DefaultFallback: "nonexistent",
					},
				},
			},
			wantWarnings: 1,
		},
		{
			name: "both models unavailable",
			routing: &RoutingConfig{
				TaskTypes: []TaskType{
					{
						ID:              "task1",
						DefaultPrimary:  "nonexistent1",
						DefaultFallback: "nonexistent2",
					},
				},
			},
			wantWarnings: 2,
		},
		{
			name: "disabled model warning",
			routing: &RoutingConfig{
				TaskTypes: []TaskType{
					{
						ID:              "task1",
						DefaultPrimary:  "disabled-model",
						DefaultFallback: "claude-sonnet-4-20250514",
					},
				},
			},
			wantWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := tt.routing.ValidateTaskModels(models)
			if len(warnings) != tt.wantWarnings {
				t.Errorf("ValidateTaskModels() got %d warnings, want %d", len(warnings), tt.wantWarnings)
				t.Logf("Warnings: %v", warnings)
			}
		})
	}
}
