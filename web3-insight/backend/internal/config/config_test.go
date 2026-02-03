package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAll(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	if err := os.Mkdir(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Create config.yaml
	configYAML := `server:
  host: localhost
  port: 8080

database:
  host: localhost
  port: 5432
  user: test
  password: test
  dbname: test
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

llm:
  default_local: "llama3:70b"
  ollama_host: "http://localhost:11434"
  claude:
    enabled: false
    api_key: ""
    default_model: ""
  openai:
    enabled: false
    api_key: ""
    default_model: ""

worker:
  concurrency: 10
  queues:
    default: 5
    high: 3
    low: 2

search:
  tavily:
    enabled: false
    api_key: ""
  serpapi:
    enabled: false
    api_key: ""
`

	// Create models.yaml
	modelsYAML := `local_models:
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
`

	// Create routing.yaml
	routingYAML := `task_types:
  - id: "chat"
    name: "问答对话"
    description: "实时问答交互"
    default_primary: "llama3:70b"
    default_fallback: "claude-sonnet-4-20250514"
    required_capability: "chat"
`

	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(configDir, "models.yaml"), []byte(modelsYAML), 0644); err != nil {
		t.Fatalf("failed to write models.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(configDir, "routing.yaml"), []byte(routingYAML), 0644); err != nil {
		t.Fatalf("failed to write routing.yaml: %v", err)
	}

	// Change to temp directory so viper can find the config
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Test LoadAll
	cfg, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	// Verify main config loaded
	if cfg.Server.Port != 8080 {
		t.Errorf("expected server port 8080, got %d", cfg.Server.Port)
	}

	// Verify models loaded
	if cfg.Models == nil {
		t.Fatal("Models config is nil")
	}
	if len(cfg.Models.LocalModels) != 1 {
		t.Errorf("expected 1 local model, got %d", len(cfg.Models.LocalModels))
	}
	if len(cfg.Models.CloudModels) != 1 {
		t.Errorf("expected 1 cloud model, got %d", len(cfg.Models.CloudModels))
	}

	// Verify routing loaded
	if cfg.Routing == nil {
		t.Fatal("Routing config is nil")
	}
	if len(cfg.Routing.TaskTypes) != 1 {
		t.Errorf("expected 1 task type, got %d", len(cfg.Routing.TaskTypes))
	}

	// Verify cross-validation works (no warnings since models are valid)
	warnings := cfg.Routing.ValidateTaskModels(cfg.Models)
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got: %v", warnings)
	}
}

func TestLoadAll_MissingModelsYAML(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	if err := os.Mkdir(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Create minimal config.yaml
	configYAML := `server:
  host: localhost
  port: 8080
database:
  host: localhost
  port: 5432
  user: test
  password: test
  dbname: test
  sslmode: disable
redis:
  host: localhost
  port: 6379
llm:
  default_local: "llama3:70b"
  ollama_host: "http://localhost:11434"
worker:
  concurrency: 10
search:
  tavily:
    enabled: false
  serpapi:
    enabled: false
`

	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config.yaml: %v", err)
	}

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Test LoadAll should fail without models.yaml
	_, err = LoadAll()
	if err == nil {
		t.Error("LoadAll() should fail when models.yaml is missing")
	}
}
