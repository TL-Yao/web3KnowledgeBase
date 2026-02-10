package model

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	NameEn       string    `gorm:"size:100" json:"nameEn"`
	ThemeID      *string   `gorm:"type:varchar(50);index" json:"themeId,omitempty"` // nil = universal tag
	Status       string    `gorm:"size:20;not null;default:'active';index" json:"status"` // active/pending/deprecated
	SuggestCount int       `gorm:"default:0" json:"suggestCount"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Tag) TableName() string {
	return "tags"
}
