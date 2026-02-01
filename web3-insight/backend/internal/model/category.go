package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	NameEn       string     `gorm:"size:100" json:"nameEn"`
	Slug         string     `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	ParentID     *uuid.UUID `gorm:"type:uuid" json:"parentId"`
	Parent       *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children     []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Description  string     `gorm:"type:text" json:"description"`
	Icon         string     `gorm:"size:50" json:"icon"`
	SortOrder    int        `gorm:"default:0" json:"sortOrder"`
	AutoCreated  bool       `gorm:"default:false" json:"autoCreated"`
	ArticleCount int        `gorm:"default:0" json:"articleCount"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (Category) TableName() string {
	return "categories"
}
