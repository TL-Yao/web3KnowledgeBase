package repository

import (
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NewsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) *NewsRepository {
	return &NewsRepository{db: db}
}

func (r *NewsRepository) Create(item *model.NewsItem) error {
	return r.db.Create(item).Error
}

// CreateOrIgnore creates a news item or ignores if source_url already exists
func (r *NewsRepository) CreateOrIgnore(item *model.NewsItem) (bool, error) {
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source_url"}},
		DoNothing: true,
	}).Create(item)

	if result.Error != nil {
		return false, result.Error
	}
	// RowsAffected == 0 means it was a duplicate
	return result.RowsAffected > 0, nil
}

// BatchCreateOrIgnore creates multiple items, ignoring duplicates
func (r *NewsRepository) BatchCreateOrIgnore(items []model.NewsItem) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}

	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source_url"}},
		DoNothing: true,
	}).Create(&items)

	if result.Error != nil {
		return 0, result.Error
	}
	return int(result.RowsAffected), nil
}

func (r *NewsRepository) FindByID(id uuid.UUID) (*model.NewsItem, error) {
	var item model.NewsItem
	if err := r.db.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *NewsRepository) FindBySourceURL(url string) (*model.NewsItem, error) {
	var item model.NewsItem
	if err := r.db.First(&item, "source_url = ?", url).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *NewsRepository) FindUnprocessed(limit int) ([]model.NewsItem, error) {
	var items []model.NewsItem
	if err := r.db.Where("processed = ?", false).
		Order("fetched_at ASC").
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *NewsRepository) MarkProcessed(id uuid.UUID) error {
	return r.db.Model(&model.NewsItem{}).
		Where("id = ?", id).
		Update("processed", true).Error
}

func (r *NewsRepository) UpdateSummary(id uuid.UUID, summary string, category string, tags []string) error {
	return r.db.Model(&model.NewsItem{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"summary":   summary,
			"category":  category,
			"tags":      tags,
			"processed": true,
		}).Error
}

func (r *NewsRepository) List(page, limit int, sourceName string, processed *bool) ([]model.NewsItem, int64, error) {
	var items []model.NewsItem
	var total int64

	query := r.db.Model(&model.NewsItem{})

	if sourceName != "" {
		query = query.Where("source_name = ?", sourceName)
	}
	if processed != nil {
		query = query.Where("processed = ?", *processed)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("fetched_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *NewsRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.NewsItem{}, "id = ?", id).Error
}
