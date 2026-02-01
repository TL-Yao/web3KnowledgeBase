package database

import (
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	// Enable pgvector extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return err
	}

	return db.AutoMigrate(
		&model.Category{},
		&model.Article{},
		&model.ArticleVersion{},
		&model.ChatMessage{},
		&model.NewsItem{},
		&model.ExplorerResearch{},
		&model.Task{},
		&model.Config{},
		&model.DataSource{},
	)
}
