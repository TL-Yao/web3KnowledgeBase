package repository

import (
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

type ThemeRepository struct {
	db *gorm.DB
}

func NewThemeRepository(db *gorm.DB) *ThemeRepository {
	return &ThemeRepository{db: db}
}

// GetByID returns a theme by its ID
func (r *ThemeRepository) GetByID(id string) (*model.Theme, error) {
	var theme model.Theme
	err := r.db.Where("id = ?", id).First(&theme).Error
	if err != nil {
		return nil, err
	}
	return &theme, nil
}

// GetActive returns the currently active theme
func (r *ThemeRepository) GetActive() (*model.Theme, error) {
	var theme model.Theme
	err := r.db.Where("status = ?", "active").First(&theme).Error
	if err != nil {
		return nil, err
	}
	return &theme, nil
}

// SetActive sets a theme as active and pauses all others
func (r *ThemeRepository) SetActive(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Pause all currently active themes
		if err := tx.Model(&model.Theme{}).
			Where("status = ?", "active").
			Update("status", "paused").Error; err != nil {
			return err
		}
		// Activate the requested theme
		return tx.Model(&model.Theme{}).
			Where("id = ?", id).
			Update("status", "active").Error
	})
}

// List returns all themes ordered by sort_order
func (r *ThemeRepository) List() ([]model.Theme, error) {
	var themes []model.Theme
	err := r.db.Order("sort_order ASC").Find(&themes).Error
	return themes, err
}

// Create creates a new theme
func (r *ThemeRepository) Create(theme *model.Theme) error {
	return r.db.Create(theme).Error
}

// Update updates theme metadata (preserves status)
func (r *ThemeRepository) Update(theme *model.Theme) error {
	return r.db.Model(theme).Updates(map[string]interface{}{
		"name":        theme.Name,
		"category":    theme.Category,
		"description": theme.Description,
		"sort_order":  theme.SortOrder,
	}).Error
}
