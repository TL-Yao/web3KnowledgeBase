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

	// Warnings collected during configuration loading
	Warnings []string
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

// findConfigFile searches for a config file in standard paths
func findConfigFile(filename string) (string, error) {
	configPaths := []string{"./config", "../config", "../../config"}
	for _, basePath := range configPaths {
		testPath := basePath + "/" + filename
		if _, err := os.Stat(testPath); err == nil {
			return testPath, nil
		}
	}
	return "", fmt.Errorf("%s not found in config paths", filename)
}

// LoadAll loads all configuration files (config.yaml, models.yaml, routing.yaml)
// Returns the loaded config with any validation warnings stored in cfg.Warnings
func LoadAll() (*Config, error) {
	// Load main config.yaml
	cfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config.yaml: %w", err)
	}

	// Find and load models.yaml
	modelsPath, err := findConfigFile("models.yaml")
	if err != nil {
		return nil, err
	}

	models, err := LoadModels(modelsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load models.yaml: %w", err)
	}
	cfg.Models = models

	// Find and load routing.yaml
	routingPath, err := findConfigFile("routing.yaml")
	if err != nil {
		return nil, err
	}

	routing, err := LoadRouting(routingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load routing.yaml: %w", err)
	}
	cfg.Routing = routing

	// Validate task models against model registry and store warnings
	cfg.Warnings = cfg.Routing.ValidateTaskModels(cfg.Models)

	return cfg, nil
}
