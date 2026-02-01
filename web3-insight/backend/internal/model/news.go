package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type NewsItem struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title          string          `gorm:"size:500;not null" json:"title"`
	OriginalTitle  string          `gorm:"size:500" json:"originalTitle"`
	Content        string          `gorm:"type:text" json:"content"`
	Summary        string          `gorm:"type:text" json:"summary"`
	SourceURL      string          `gorm:"size:1000;uniqueIndex;not null" json:"sourceUrl"`
	SourceName     string          `gorm:"size:100" json:"sourceName"`
	SourceLanguage string          `gorm:"size:10" json:"sourceLanguage"`
	Category       string          `gorm:"size:50" json:"category"`
	Tags           pq.StringArray  `gorm:"type:text[]" json:"tags"`
	PublishedAt    *time.Time      `json:"publishedAt"`
	FetchedAt      time.Time       `gorm:"default:now()" json:"fetchedAt"`
	Processed      bool            `gorm:"default:false" json:"processed"`
	Embedding      pgvector.Vector `gorm:"type:vector(1536)" json:"-"`
}

func (NewsItem) TableName() string {
	return "news_items"
}
