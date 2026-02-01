package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type Task struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Type        string          `gorm:"size:50;not null" json:"type"`
	Status      string          `gorm:"size:20;default:'pending'" json:"status"`
	Payload     datatypes.JSON  `gorm:"type:jsonb" json:"payload"`
	Result      datatypes.JSON  `gorm:"type:jsonb" json:"result"`
	Error       string          `gorm:"type:text" json:"error"`
	ModelUsed   string          `gorm:"size:50" json:"modelUsed"`
	TokensUsed  int             `json:"tokensUsed"`
	CostUSD     decimal.Decimal `gorm:"type:decimal(10,6)" json:"costUsd"`
	StartedAt   *time.Time      `json:"startedAt"`
	CompletedAt *time.Time      `json:"completedAt"`
	CreatedAt   time.Time       `json:"createdAt"`
}

func (Task) TableName() string {
	return "tasks"
}

// Task types
const (
	TaskTypeRSSSync         = "rss_sync"
	TaskTypeWebCrawl        = "web_crawl"
	TaskTypeContentGenerate = "content_generate"
	TaskTypeClassify        = "classify"
)

// Task statuses
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)
