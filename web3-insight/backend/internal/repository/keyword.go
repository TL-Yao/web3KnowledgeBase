package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

type KeywordRepository struct {
	db *gorm.DB
}

func NewKeywordRepository(db *gorm.DB) *KeywordRepository {
	return &KeywordRepository{db: db}
}

// GetPendingKeywords 获取指定数量的待用关键词
func (r *KeywordRepository) GetPendingKeywords(limit int) ([]model.Keyword, error) {
	var keywords []model.Keyword
	err := r.db.
		Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(limit).
		Find(&keywords).Error
	return keywords, err
}

// CountPendingKeywords 统计待用关键词数量
func (r *KeywordRepository) CountPendingKeywords() (int64, error) {
	var count int64
	err := r.db.
		Model(&model.Keyword{}).
		Where("status = ?", "pending").
		Count(&count).Error
	return count, err
}

// MarkAsUsed 标记关键词为已使用
func (r *KeywordRepository) MarkAsUsed(id uuid.UUID, articleID uuid.UUID) error {
	now := time.Now()
	return r.db.
		Model(&model.Keyword{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "used",
			"used_at":    now,
			"article_id": articleID,
		}).Error
}

// BatchCreate 批量创建关键词
func (r *KeywordRepository) BatchCreate(keywords []model.Keyword) error {
	if len(keywords) == 0 {
		return nil
	}
	return r.db.CreateInBatches(keywords, 100).Error
}

// GetAllUsedKeywords 获取所有已使用的关键词（用于去重）
func (r *KeywordRepository) GetAllUsedKeywords() ([]string, error) {
	var keywords []string
	err := r.db.
		Model(&model.Keyword{}).
		Where("status = ?", "used").
		Pluck("keyword", &keywords).Error
	return keywords, err
}

// GetByID 根据ID获取关键词
func (r *KeywordRepository) GetByID(id uuid.UUID) (*model.Keyword, error) {
	var keyword model.Keyword
	err := r.db.
		Preload("Article").
		Where("id = ?", id).
		First(&keyword).Error
	if err != nil {
		return nil, err
	}
	return &keyword, nil
}

// List 分页获取关键词列表
func (r *KeywordRepository) List(status string, page, pageSize int) ([]model.Keyword, int64, error) {
	var keywords []model.Keyword
	var total int64

	query := r.db.Model(&model.Keyword{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("Article").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&keywords).Error

	return keywords, total, err
}
