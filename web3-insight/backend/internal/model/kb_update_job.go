package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type KBUpdateJob struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Status            string         `gorm:"type:varchar(50);not null;index" json:"status"` // pending/running/completed/failed
	TriggerType       string         `gorm:"type:varchar(50)" json:"triggerType"`           // scheduled/manual
	KeywordsGenerated int            `gorm:"default:0" json:"keywordsGenerated"`
	ArticlesGenerated int            `gorm:"default:0" json:"articlesGenerated"`
	ArticlesPublished int            `gorm:"default:0" json:"articlesPublished"`
	SessionIDs        pq.StringArray `gorm:"type:text[]" json:"sessionIds"` // 记录所有 session ID
	ErrorMessage      string         `gorm:"type:text" json:"errorMessage,omitempty"`
	StartedAt         *time.Time     `json:"startedAt,omitempty"`
	CompletedAt       *time.Time     `json:"completedAt,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime;index" json:"createdAt"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (KBUpdateJob) TableName() string {
	return "kb_update_jobs"
}
