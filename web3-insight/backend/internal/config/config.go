package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Worker   WorkerConfig   `mapstructure:"worker"`
	Search   SearchConfig   `mapstructure:"search"`

	// Model registry and routing configs (loaded separately from YAML files)
	Models  *ModelsConfig
	Routing *RoutingConfig
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LLMConfig struct {
	DefaultLocal string       `mapstructure:"default_local"`
	OllamaHost   string       `mapstructure:"ollama_host"`
	Claude       ClaudeConfig `mapstructure:"claude"`
	OpenAI       OpenAIConfig `mapstructure:"openai"`
}

type ClaudeConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	APIKey       string `mapstructure:"api_key"`
	DefaultModel string `mapstructure:"default_model"`
}

type OpenAIConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	APIKey       string `mapstructure:"api_key"`
	DefaultModel string `mapstructure:"default_model"`
}

type WorkerConfig struct {
	Concurrency int            `mapstructure:"concurrency"`
	Queues      map[string]int `mapstructure:"queues"`
}

type SearchConfig struct {
	Tavily  TavilyConfig  `mapstructure:"tavily"`
	SerpAPI SerpAPIConfig `mapstructure:"serpapi"`
}

type TavilyConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key"`
}

type SerpAPIConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	// Enable environment variable substitution
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadAll loads all configuration files (config.yaml, models.yaml, routing.yaml)
func LoadAll() (*Config, error) {
	// Load main config.yaml
	cfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config.yaml: %w", err)
	}

	// Try multiple config paths to find models.yaml
	configPaths := []string{"./config", "../config", "../../config"}
	var modelsPath string
	for _, basePath := range configPaths {
		testPath := basePath + "/models.yaml"
		if _, err := os.Stat(testPath); err == nil {
			modelsPath = testPath
			break
		}
	}

	if modelsPath == "" {
		return nil, fmt.Errorf("models.yaml not found in config paths")
	}

	// Load models.yaml
	models, err := LoadModels(modelsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load models.yaml: %w", err)
	}
	cfg.Models = models

	// Try multiple config paths to find routing.yaml
	var routingPath string
	for _, basePath := range configPaths {
		testPath := basePath + "/routing.yaml"
		if _, err := os.Stat(testPath); err == nil {
			routingPath = testPath
			break
		}
	}

	if routingPath == "" {
		return nil, fmt.Errorf("routing.yaml not found in config paths")
	}

	// Load routing.yaml
	routing, err := LoadRouting(routingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load routing.yaml: %w", err)
	}
	cfg.Routing = routing

	// Validate task models against model registry
	if warnings := cfg.Routing.ValidateTaskModels(cfg.Models); len(warnings) > 0 {
		fmt.Println("⚠️  Configuration warnings detected:")
		for _, warning := range warnings {
			fmt.Printf("   - %s\n", warning)
		}
	}

	return cfg, nil
}
