package model

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ArticleID      *uuid.UUID `gorm:"type:uuid" json:"articleId"`
	Article        *Article   `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	SessionID      uuid.UUID  `gorm:"type:uuid;not null" json:"sessionId"`
	Role           string     `gorm:"size:20;not null" json:"role"`
	Content        string     `gorm:"type:text;not null" json:"content"`
	ModelUsed      string     `gorm:"size:50" json:"modelUsed"`
	SavedToArticle bool       `gorm:"default:false" json:"savedToArticle"`
	CreatedAt      time.Time  `json:"createdAt"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
