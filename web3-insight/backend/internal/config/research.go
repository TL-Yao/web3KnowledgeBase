package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ResearchConfig struct {
	Domains    []ResearchDomain   `yaml:"domains" json:"domains"`
	Generation ResearchGeneration `yaml:"generation" json:"generation"`
}

type ResearchDomain struct {
	ID            string `yaml:"id" json:"id"`
	Name          string `yaml:"name" json:"name"`
	NameEn        string `yaml:"nameEn" json:"nameEn"`
	Description   string `yaml:"description" json:"description"`
	Icon          string `yaml:"icon" json:"icon"`
	SortOrder     int    `yaml:"sort_order" json:"sortOrder"`
	SystemContext string `yaml:"system_context" json:"-"` // Never expose to frontend
}

type ResearchGeneration struct {
	DefaultModel       string `yaml:"default_model" json:"defaultModel"`
	TimeoutMinutes     int    `yaml:"timeout_minutes" json:"timeoutMinutes"`
	PlanTimeoutMinutes int    `yaml:"plan_timeout_minutes" json:"planTimeoutMinutes"`
	MinReportLength    int    `yaml:"min_report_length" json:"minReportLength"`
	MaxReportLength    int    `yaml:"max_report_length" json:"maxReportLength"`
}

// LoadResearch reads and parses research.yaml
func LoadResearch(path string) (*ResearchConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read research.yaml: %w", err)
	}

	var config ResearchConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse research.yaml: %w", err)
	}

	if len(config.Domains) == 0 {
		return nil, fmt.Errorf("research.yaml must define at least one domain")
	}

	return &config, nil
}

// GetDomainByID finds a domain by its ID
func (rc *ResearchConfig) GetDomainByID(id string) (*ResearchDomain, error) {
	for i := range rc.Domains {
		if rc.Domains[i].ID == id {
			return &rc.Domains[i], nil
		}
	}
	return nil, fmt.Errorf("research domain not found: %s", id)
}
