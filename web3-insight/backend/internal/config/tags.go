package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TagEntry represents a single tag in the config
type TagEntry struct {
	Name   string `yaml:"name" json:"name"`
	NameEn string `yaml:"name_en" json:"nameEn"`
}

// TagsConfig holds the full tag registry from tags.yaml
type TagsConfig struct {
	UniversalTags []TagEntry              `yaml:"universal_tags" json:"universalTags"`
	ThemeTags     map[string][]TagEntry   `yaml:"theme_tags" json:"themeTags"`
}

// LoadTags reads and parses tags.yaml
func LoadTags(path string) (*TagsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags.yaml: %w", err)
	}

	var config TagsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse tags.yaml: %w", err)
	}

	return &config, nil
}

// GetTagsForTheme returns universal tags + theme-specific tags
func (tc *TagsConfig) GetTagsForTheme(themeID string) (universal []TagEntry, theme []TagEntry) {
	universal = tc.UniversalTags
	theme = tc.ThemeTags[themeID]
	return
}
