package database

import (
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	// Seed function intentionally empty - no mock data
	// Use API endpoints to create categories and content
	return nil
}
