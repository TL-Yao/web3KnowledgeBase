package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

type DataSourceRepository struct {
	db *gorm.DB
}

func NewDataSourceRepository(db *gorm.DB) *DataSourceRepository {
	return &DataSourceRepository{db: db}
}

func (r *DataSourceRepository) FindByID(id uuid.UUID) (*model.DataSource, error) {
	var source model.DataSource
	if err := r.db.First(&source, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *DataSourceRepository) FindByType(sourceType string) ([]model.DataSource, error) {
	var sources []model.DataSource
	if err := r.db.Where("type = ? AND enabled = ?", sourceType, true).Find(&sources).Error; err != nil {
		return nil, err
	}
	return sources, nil
}

func (r *DataSourceRepository) FindEnabled() ([]model.DataSource, error) {
	var sources []model.DataSource
	if err := r.db.Where("enabled = ?", true).Find(&sources).Error; err != nil {
		return nil, err
	}
	return sources, nil
}

func (r *DataSourceRepository) FindDueForFetch() ([]model.DataSource, error) {
	var sources []model.DataSource
	now := time.Now()
	// Find sources where: enabled AND (never fetched OR last_fetched + interval < now)
	if err := r.db.Where(
		"enabled = ? AND (last_fetched_at IS NULL OR last_fetched_at + (fetch_interval * interval '1 second') < ?)",
		true, now,
	).Find(&sources).Error; err != nil {
		return nil, err
	}
	return sources, nil
}

func (r *DataSourceRepository) UpdateLastFetched(id uuid.UUID, fetchedAt time.Time, lastError string) error {
	updates := map[string]interface{}{
		"last_fetched_at": fetchedAt,
		"last_error":      lastError,
	}
	return r.db.Model(&model.DataSource{}).Where("id = ?", id).Updates(updates).Error
}

func (r *DataSourceRepository) Create(source *model.DataSource) error {
	return r.db.Create(source).Error
}

func (r *DataSourceRepository) Update(source *model.DataSource) error {
	return r.db.Save(source).Error
}

func (r *DataSourceRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.DataSource{}, "id = ?", id).Error
}

func (r *DataSourceRepository) List() ([]model.DataSource, error) {
	var sources []model.DataSource
	if err := r.db.Order("created_at DESC").Find(&sources).Error; err != nil {
		return nil, err
	}
	return sources, nil
}
