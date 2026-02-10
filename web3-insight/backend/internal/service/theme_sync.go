package service

import (
	"log"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"gorm.io/gorm"
)

// SyncThemesFromConfig syncs theme definitions from prompts.yaml into the database.
// Creates new themes as "paused", updates metadata for existing ones (preserving status).
// If no theme is active, activates the first theme. Also migrates orphaned keywords/articles.
func SyncThemesFromConfig(db *gorm.DB, prompts *config.PromptsConfig) error {
	themeRepo := repository.NewThemeRepository(db)

	for _, tc := range prompts.Themes {
		existing, err := themeRepo.GetByID(tc.ID)
		if err != nil {
			// Theme doesn't exist — create with status "paused"
			theme := &model.Theme{
				ID:          tc.ID,
				Name:        tc.Name,
				Category:    tc.Category,
				Description: tc.Description,
				Status:      "paused",
				SortOrder:   tc.SortOrder,
			}
			if err := themeRepo.Create(theme); err != nil {
				log.Printf("Failed to create theme %s: %v", tc.ID, err)
				continue
			}
			log.Printf("Created theme: %s (%s)", tc.ID, tc.Name)
		} else {
			// Theme exists — update metadata, preserve status
			existing.Name = tc.Name
			existing.Category = tc.Category
			existing.Description = tc.Description
			existing.SortOrder = tc.SortOrder
			if err := themeRepo.Update(existing); err != nil {
				log.Printf("Failed to update theme %s: %v", tc.ID, err)
			}
		}
	}

	// If no theme is active, activate the first one
	_, err := themeRepo.GetActive()
	if err != nil && len(prompts.Themes) > 0 {
		firstThemeID := prompts.Themes[0].ID
		if err := themeRepo.SetActive(firstThemeID); err != nil {
			log.Printf("Failed to set default active theme: %v", err)
		} else {
			log.Printf("Set default active theme: %s", firstThemeID)
		}
	}

	// Migrate orphaned keywords (no theme_id) to web3_basics
	result := db.Model(&model.Keyword{}).
		Where("theme_id IS NULL OR theme_id = ''").
		Updates(map[string]interface{}{"theme_id": "web3_basics"})
	if result.RowsAffected > 0 {
		log.Printf("Migrated %d orphaned keywords to web3_basics", result.RowsAffected)
	}

	// Migrate orphaned articles that have associated keywords to web3_basics
	result = db.Model(&model.Article{}).
		Where("theme_id IS NULL").
		Where("id IN (SELECT article_id FROM keywords WHERE article_id IS NOT NULL)").
		Updates(map[string]interface{}{"theme_id": "web3_basics"})
	if result.RowsAffected > 0 {
		log.Printf("Migrated %d orphaned articles to web3_basics", result.RowsAffected)
	}

	return nil
}
