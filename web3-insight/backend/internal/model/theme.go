package model

import "time"

type Theme struct {
	ID          string    `gorm:"type:varchar(50);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Category    string    `gorm:"type:varchar(50);not null;index" json:"category"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"type:varchar(20);not null;default:'paused';index" json:"status"` // active/paused
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Theme) TableName() string {
	return "themes"
}
