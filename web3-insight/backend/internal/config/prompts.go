package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ThemeConfig defines a single theme from prompts.yaml
type ThemeConfig struct {
	ID            string `yaml:"id" json:"id"`
	Name          string `yaml:"name" json:"name"`
	Category      string `yaml:"category" json:"category"`
	Description   string `yaml:"description" json:"description"`
	SortOrder     int    `yaml:"sort_order" json:"sortOrder"`
	KeywordPrompt string `yaml:"keyword_prompt" json:"-"`
	ArticlePrompt string `yaml:"article_prompt" json:"-"`
}

// GenerationConfig holds global generation settings
type GenerationConfig struct {
	DefaultBatchSize       int `yaml:"default_batch_size" json:"defaultBatchSize"`
	MaxBatchSize           int `yaml:"max_batch_size" json:"maxBatchSize"`
	KeywordPoolTarget      int `yaml:"keyword_pool_target" json:"keywordPoolTarget"`
	KeywordRefillThreshold int `yaml:"keyword_refill_threshold" json:"keywordRefillThreshold"`
	ArticleTimeoutMinutes  int `yaml:"article_timeout_minutes" json:"articleTimeoutMinutes"`
}

// PromptsConfig holds all theme definitions and generation settings
type PromptsConfig struct {
	Themes     []ThemeConfig    `yaml:"themes" json:"themes"`
	Generation GenerationConfig `yaml:"generation" json:"generation"`
}

// LoadPrompts reads and parses prompts.yaml
func LoadPrompts(path string) (*PromptsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompts.yaml: %w", err)
	}

	var config PromptsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse prompts.yaml: %w", err)
	}

	if len(config.Themes) == 0 {
		return nil, fmt.Errorf("prompts.yaml must define at least one theme")
	}

	return &config, nil
}

// GetThemeByID finds a theme by its ID
func (pc *PromptsConfig) GetThemeByID(id string) (*ThemeConfig, error) {
	for i := range pc.Themes {
		if pc.Themes[i].ID == id {
			return &pc.Themes[i], nil
		}
	}
	return nil, fmt.Errorf("theme not found: %s", id)
}

// GetAllThemes returns all theme configs
func (pc *PromptsConfig) GetAllThemes() []ThemeConfig {
	return pc.Themes
}

// GetThemesByCategory returns themes matching the given category
func (pc *PromptsConfig) GetThemesByCategory(category string) []ThemeConfig {
	var result []ThemeConfig
	for _, t := range pc.Themes {
		if t.Category == category {
			result = append(result, t)
		}
	}
	return result
}
