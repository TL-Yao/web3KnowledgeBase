package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadModels(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		validate    func(*testing.T, *ModelsConfig)
	}{
		{
			name: "valid models.yaml",
			yamlContent: `local_models:
  - id: "llama3:70b"
    name: "Llama 3 70B"
    provider: "ollama"
    enabled: true
    capabilities: ["chat", "generation"]
    context_window: 8192
    cost_per_1k_tokens: 0.0

cloud_models:
  - id: "claude-sonnet-4-20250514"
    name: "Claude Sonnet 4"
    provider: "anthropic"
    enabled: true
    capabilities: ["chat", "analysis"]
    context_window: 200000
    cost_per_1k_tokens: 0.003
`,
			wantErr: false,
			validate: func(t *testing.T, mc *ModelsConfig) {
				if len(mc.LocalModels) != 1 {
					t.Errorf("expected 1 local model, got %d", len(mc.LocalModels))
				}
				if len(mc.CloudModels) != 1 {
					t.Errorf("expected 1 cloud model, got %d", len(mc.CloudModels))
				}
				if mc.LocalModels[0].ID != "llama3:70b" {
					t.Errorf("expected local model ID 'llama3:70b', got '%s'", mc.LocalModels[0].ID)
				}
				if mc.CloudModels[0].ID != "claude-sonnet-4-20250514" {
					t.Errorf("expected cloud model ID 'claude-sonnet-4-20250514', got '%s'", mc.CloudModels[0].ID)
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
			yamlContent: `local_models:
  - id: "test"
    invalid yaml syntax [[[
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "file not found" {
				_, err := LoadModels("/nonexistent/path/models.yaml")
				if !tt.wantErr {
					t.Errorf("LoadModels() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "models.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			got, err := LoadModels(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadModels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestModelsConfig_GetAllModels(t *testing.T) {
	mc := &ModelsConfig{
		LocalModels: []Model{
			{ID: "local1", Name: "Local 1", Enabled: true},
			{ID: "local2", Name: "Local 2", Enabled: false},
		},
		CloudModels: []Model{
			{ID: "cloud1", Name: "Cloud 1", Enabled: true},
		},
	}

	all := mc.GetAllModels()
	if len(all) != 3 {
		t.Errorf("expected 3 models, got %d", len(all))
	}
}

func TestModelsConfig_GetEnabledModels(t *testing.T) {
	mc := &ModelsConfig{
		LocalModels: []Model{
			{ID: "local1", Name: "Local 1", Enabled: true},
			{ID: "local2", Name: "Local 2", Enabled: false},
		},
		CloudModels: []Model{
			{ID: "cloud1", Name: "Cloud 1", Enabled: true},
			{ID: "cloud2", Name: "Cloud 2", Enabled: false},
		},
	}

	enabled := mc.GetEnabledModels()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled models, got %d", len(enabled))
	}

	for _, model := range enabled {
		if !model.Enabled {
			t.Errorf("model %s should be enabled", model.ID)
		}
	}
}

func TestModelsConfig_GetModelByID(t *testing.T) {
	mc := &ModelsConfig{
		LocalModels: []Model{
			{ID: "local1", Name: "Local 1", Enabled: true},
		},
		CloudModels: []Model{
			{ID: "cloud1", Name: "Cloud 1", Enabled: true},
		},
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		wantID  string
	}{
		{
			name:    "find local model",
			id:      "local1",
			wantErr: false,
			wantID:  "local1",
		},
		{
			name:    "find cloud model",
			id:      "cloud1",
			wantErr: false,
			wantID:  "cloud1",
		},
		{
			name:    "model not found",
			id:      "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mc.GetModelByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetModelByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("GetModelByID() got ID = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}

func TestModelsConfig_IsModelAvailable(t *testing.T) {
	mc := &ModelsConfig{
		LocalModels: []Model{
			{ID: "enabled", Name: "Enabled", Enabled: true},
			{ID: "disabled", Name: "Disabled", Enabled: false},
		},
	}

	tests := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "enabled model is available",
			id:   "enabled",
			want: true,
		},
		{
			name: "disabled model is not available",
			id:   "disabled",
			want: false,
		},
		{
			name: "nonexistent model is not available",
			id:   "nonexistent",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mc.IsModelAvailable(tt.id); got != tt.want {
				t.Errorf("IsModelAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelsConfig_GetModelsByCapability(t *testing.T) {
	mc := &ModelsConfig{
		LocalModels: []Model{
			{ID: "local1", Name: "Local 1", Enabled: true, Capabilities: []string{"chat", "generation"}},
			{ID: "local2", Name: "Local 2", Enabled: true, Capabilities: []string{"translation"}},
			{ID: "local3", Name: "Local 3", Enabled: false, Capabilities: []string{"chat"}},
		},
		CloudModels: []Model{
			{ID: "cloud1", Name: "Cloud 1", Enabled: true, Capabilities: []string{"chat", "analysis"}},
		},
	}

	tests := []struct {
		name       string
		capability string
		wantCount  int
		wantIDs    []string
	}{
		{
			name:       "chat capability",
			capability: "chat",
			wantCount:  2,
			wantIDs:    []string{"local1", "cloud1"},
		},
		{
			name:       "translation capability",
			capability: "translation",
			wantCount:  1,
			wantIDs:    []string{"local2"},
		},
		{
			name:       "nonexistent capability",
			capability: "nonexistent",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mc.GetModelsByCapability(tt.capability)
			if len(got) != tt.wantCount {
				t.Errorf("GetModelsByCapability() count = %v, want %v", len(got), tt.wantCount)
			}

			if tt.wantIDs != nil {
				gotIDs := make(map[string]bool)
				for _, model := range got {
					gotIDs[model.ID] = true
				}
				for _, wantID := range tt.wantIDs {
					if !gotIDs[wantID] {
						t.Errorf("GetModelsByCapability() missing expected ID: %s", wantID)
					}
				}
			}
		})
	}
}
