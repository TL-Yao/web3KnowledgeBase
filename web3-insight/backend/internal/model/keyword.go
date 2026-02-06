package model

import (
	"time"

	"github.com/google/uuid"
)

type Keyword struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Keyword   string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"keyword"`
	Status    string     `gorm:"type:varchar(50);not null;index" json:"status"` // pending/used
	UsedAt    *time.Time `gorm:"index" json:"usedAt,omitempty"`
	ArticleID *uuid.UUID `gorm:"type:uuid" json:"articleId,omitempty"`
	Article   *Article   `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	Category  string     `gorm:"type:varchar(100)" json:"category"`
	Source    string     `gorm:"type:varchar(50);default:'claude_code'" json:"source"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Keyword) TableName() string {
	return "keywords"
}
