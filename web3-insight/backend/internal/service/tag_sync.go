package service

import (
	"log"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"gorm.io/gorm"
)

// SyncTagsFromConfig syncs tag definitions from tags.yaml into the database.
// Creates new tags as "active", preserves existing tag status.
func SyncTagsFromConfig(db *gorm.DB, tagsConfig *config.TagsConfig) error {
	if tagsConfig == nil {
		log.Println("Tags config not loaded, skipping tag sync")
		return nil
	}

	tagRepo := repository.NewTagRepository(db)
	created := 0

	// Sync universal tags (theme_id = nil)
	for _, entry := range tagsConfig.UniversalTags {
		tag := model.Tag{
			Name:   entry.Name,
			NameEn: entry.NameEn,
			Status: "active",
		}
		if err := tagRepo.BatchCreate([]model.Tag{tag}); err != nil {
			log.Printf("Failed to sync universal tag %s: %v", entry.Name, err)
			continue
		}
		created++
	}

	// Sync theme-specific tags
	for themeID, entries := range tagsConfig.ThemeTags {
		for _, entry := range entries {
			tid := themeID
			tag := model.Tag{
				Name:    entry.Name,
				NameEn:  entry.NameEn,
				ThemeID: &tid,
				Status:  "active",
			}
			if err := tagRepo.BatchCreate([]model.Tag{tag}); err != nil {
				log.Printf("Failed to sync theme tag %s (theme: %s): %v", entry.Name, themeID, err)
				continue
			}
			created++
		}
	}

	log.Printf("Tag sync completed: processed %d tag definitions", created)
	return nil
}
