package repository

import (
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// FindAll returns all tags, optionally filtered by status
func (r *TagRepository) FindAll(status string) ([]model.Tag, error) {
	var tags []model.Tag
	query := r.db.Order("name ASC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&tags).Error
	return tags, err
}

// FindByTheme returns tags for a specific theme
func (r *TagRepository) FindByTheme(themeID string) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Where("theme_id = ?", themeID).Order("name ASC").Find(&tags).Error
	return tags, err
}

// FindByID returns a tag by its UUID
func (r *TagRepository) FindByID(id uuid.UUID) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("id = ?", id).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// FindByName returns a tag by its exact name
func (r *TagRepository) FindByName(name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// FindActiveForTagging returns active tags for a theme + all universal (theme_id IS NULL) active tags
func (r *TagRepository) FindActiveForTagging(themeID string) (universal []model.Tag, theme []model.Tag, err error) {
	err = r.db.Where("theme_id IS NULL AND status = ?", "active").Order("name ASC").Find(&universal).Error
	if err != nil {
		return nil, nil, err
	}
	err = r.db.Where("theme_id = ? AND status = ?", themeID, "active").Order("name ASC").Find(&theme).Error
	if err != nil {
		return nil, nil, err
	}
	return universal, theme, nil
}

// Create creates a new tag
func (r *TagRepository) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

// BatchCreate creates multiple tags, skipping duplicates on name conflict
func (r *TagRepository) BatchCreate(tags []model.Tag) error {
	if len(tags) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(&tags).Error
}

// UpdateStatus updates a tag's status
func (r *TagRepository) UpdateStatus(name string, status string) error {
	return r.db.Model(&model.Tag{}).Where("name = ?", name).Update("status", status).Error
}

// IncrementSuggestCount increments the suggest count for a tag and returns the new count
func (r *TagRepository) IncrementSuggestCount(name string) (int, error) {
	result := r.db.Model(&model.Tag{}).Where("name = ?", name).
		UpdateColumn("suggest_count", gorm.Expr("suggest_count + 1"))
	if result.Error != nil {
		return 0, result.Error
	}

	var tag model.Tag
	if err := r.db.Where("name = ?", name).First(&tag).Error; err != nil {
		return 0, err
	}
	return tag.SuggestCount, nil
}

// GetStats returns aggregate tag statistics
func (r *TagRepository) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	var total int64
	r.db.Model(&model.Tag{}).Count(&total)
	stats["total"] = total

	var active int64
	r.db.Model(&model.Tag{}).Where("status = ?", "active").Count(&active)
	stats["active"] = active

	var pending int64
	r.db.Model(&model.Tag{}).Where("status = ?", "pending").Count(&pending)
	stats["pending"] = pending

	var deprecated int64
	r.db.Model(&model.Tag{}).Where("status = ?", "deprecated").Count(&deprecated)
	stats["deprecated"] = deprecated

	var universal int64
	r.db.Model(&model.Tag{}).Where("theme_id IS NULL").Count(&universal)
	stats["universal"] = universal

	return stats, nil
}
