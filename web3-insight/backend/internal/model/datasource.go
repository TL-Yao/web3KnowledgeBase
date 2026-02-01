package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type DataSource struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Type          string         `gorm:"size:50;not null" json:"type"`
	URL           string         `gorm:"size:1000" json:"url"`
	Config        datatypes.JSON `gorm:"type:jsonb" json:"config"`
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	FetchInterval int            `gorm:"default:3600" json:"fetchInterval"`
	LastFetchedAt *time.Time     `json:"lastFetchedAt"`
	LastError     string         `gorm:"type:text" json:"lastError"`
	CreatedAt     time.Time      `json:"createdAt"`
}

func (DataSource) TableName() string {
	return "data_sources"
}

// DataSource types
const (
	DataSourceTypeRSS   = "rss"
	DataSourceTypeAPI   = "api"
	DataSourceTypeCrawl = "crawl"
)
