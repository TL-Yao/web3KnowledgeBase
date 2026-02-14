package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ResearchSession tracks an instant research generation session.
// Links to Article once the report is saved.
type ResearchSession struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ArticleID      *uuid.UUID     `gorm:"type:uuid" json:"articleId,omitempty"`
	Article        *Article       `gorm:"foreignKey:ArticleID;constraint:OnDelete:SET NULL" json:"article,omitempty"`
	Question       string         `gorm:"type:text;not null" json:"question"`
	Domain         string         `gorm:"type:varchar(50);not null;default:'auto'" json:"domain"`
	Status         string         `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	ResearchPlan   string         `gorm:"type:text" json:"researchPlan,omitempty"`
	PlanApproved   bool           `gorm:"default:false" json:"planApproved"`
	Citations      datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"citations"`
	PinnedFindings datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"pinnedFindings"`
	Stage          string         `gorm:"type:varchar(50)" json:"stage,omitempty"`
	StageDetail    string         `gorm:"type:text" json:"stageDetail,omitempty"`
	ErrorMessage   string         `gorm:"type:text" json:"errorMessage,omitempty"`
	ModelUsed      string         `gorm:"type:varchar(50)" json:"modelUsed,omitempty"`
	GenerationCost float64        `gorm:"type:decimal(10,6);default:0" json:"generationCost"`
	CLISessionID   string         `gorm:"type:varchar(100)" json:"cliSessionId,omitempty"`
	StartedAt      *time.Time     `json:"startedAt,omitempty"`
	CompletedAt    *time.Time     `json:"completedAt,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;index:idx_research_created,sort:desc" json:"createdAt"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ResearchSession) TableName() string { return "research_sessions" }
