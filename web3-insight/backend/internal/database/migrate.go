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

	// Rename articles.category → articles.article_type (one-time migration)
	var colExists bool
	db.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='articles' AND column_name='category')").Scan(&colExists)
	if colExists {
		db.Exec("ALTER TABLE articles RENAME COLUMN category TO article_type")
	}

	return db.AutoMigrate(
		&model.Theme{},
		&model.Category{},
		&model.Article{},
		&model.ArticleVersion{},
		&model.ChatMessage{},
		&model.NewsItem{},
		&model.ExplorerResearch{},
		&model.Task{},
		&model.Config{},
		&model.DataSource{},
		&model.Keyword{},
		&model.KBUpdateJob{},
		&model.Tag{},
		&model.ResearchSession{},
	)
}
