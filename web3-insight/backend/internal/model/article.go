package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type Article struct {
	ID               uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title            string          `gorm:"size:500;not null" json:"title"`
	Slug             string          `gorm:"size:500;uniqueIndex;not null" json:"slug"`
	Content          string          `gorm:"type:text;not null" json:"content"`
	ContentHTML      string          `gorm:"type:text" json:"contentHtml"`
	Summary          string          `gorm:"type:text" json:"summary"`
	CategoryID       *uuid.UUID      `gorm:"type:uuid" json:"categoryId"`
	Category         *Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags             pq.StringArray  `gorm:"type:text[]" json:"tags"`
	Status           string          `gorm:"size:20;default:'published'" json:"status"`
	SourceURLs       pq.StringArray  `gorm:"type:text[]" json:"sourceUrls"`
	SourceLanguage   string          `gorm:"size:10" json:"sourceLanguage"`
	ModelUsed        string          `gorm:"size:50" json:"modelUsed"`
	GenerationPrompt string          `gorm:"type:text" json:"generationPrompt"`
	ViewCount        int             `gorm:"default:0" json:"viewCount"`
	Embedding        pgvector.Vector `gorm:"type:vector(1536)" json:"-"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

func (Article) TableName() string {
	return "articles"
}

type ArticleVersion struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ArticleID     uuid.UUID `gorm:"type:uuid;not null" json:"articleId"`
	Article       *Article  `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"article,omitempty"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	EditedBy      string    `gorm:"size:20;default:'ai'" json:"editedBy"`
	ChangeSummary string    `gorm:"type:text" json:"changeSummary"`
	CreatedAt     time.Time `json:"createdAt"`
}

func (ArticleVersion) TableName() string {
	return "article_versions"
}
