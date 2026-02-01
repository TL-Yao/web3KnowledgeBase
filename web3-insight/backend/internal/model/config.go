package model

import (
	"time"

	"gorm.io/datatypes"
)

type Config struct {
	Key         string         `gorm:"size:100;primary_key" json:"key"`
	Value       datatypes.JSON `gorm:"type:jsonb;not null" json:"value"`
	Description string         `gorm:"type:text" json:"description"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

func (Config) TableName() string {
	return "configs"
}
