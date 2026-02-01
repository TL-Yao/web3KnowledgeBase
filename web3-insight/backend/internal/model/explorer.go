package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type ExplorerResearch struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChainName       string         `gorm:"size:100;not null" json:"chainName"`
	ExplorerName    string         `gorm:"size:100;not null" json:"explorerName"`
	ExplorerURL     string         `gorm:"size:500;not null" json:"explorerUrl"`
	Features        datatypes.JSON `gorm:"type:jsonb" json:"features"`
	Screenshots     pq.StringArray `gorm:"type:text[]" json:"screenshots"`
	Analysis        string         `gorm:"type:text" json:"analysis"`
	PopularityScore float64        `json:"popularityScore"`
	LastUpdated     time.Time      `gorm:"default:now()" json:"lastUpdated"`
}

func (ExplorerResearch) TableName() string {
	return "explorer_research"
}
