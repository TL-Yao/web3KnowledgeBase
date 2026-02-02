package repository

import (
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

type ArticleListParams struct {
	CategoryID *uuid.UUID
	Status     string
	Tags       []string
	Search     string
	Page       int
	PageSize   int
}

type ArticleListResult struct {
	Articles []model.Article `json:"articles"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

func (r *ArticleRepository) List(params ArticleListParams) (*ArticleListResult, error) {
	query := r.db.Model(&model.Article{}).Preload("Category")

	if params.CategoryID != nil {
		query = query.Where("category_id = ?", params.CategoryID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Search != "" {
		query = query.Where("title ILIKE ? OR summary ILIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	offset := (params.Page - 1) * params.PageSize

	var articles []model.Article
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&articles).Error; err != nil {
		return nil, err
	}

	return &ArticleListResult{
		Articles: articles,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (r *ArticleRepository) GetByID(id uuid.UUID) (*model.Article, error) {
	var article model.Article
	if err := r.db.Preload("Category").First(&article, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) GetBySlug(slug string) (*model.Article, error) {
	var article model.Article
	if err := r.db.Preload("Category").First(&article, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) Create(article *model.Article) error {
	// Omit embedding field if nil to avoid pgvector empty dimension error
	if article.Embedding == nil {
		return r.db.Omit("Embedding").Create(article).Error
	}
	return r.db.Create(article).Error
}

func (r *ArticleRepository) Update(article *model.Article) error {
	// Omit embedding field if nil to avoid pgvector empty dimension error
	if article.Embedding == nil {
		return r.db.Omit("Embedding").Save(article).Error
	}
	return r.db.Save(article).Error
}

func (r *ArticleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Article{}, "id = ?", id).Error
}

func (r *ArticleRepository) IncrementViewCount(id uuid.UUID) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *ArticleRepository) Search(query string, limit int) ([]model.Article, error) {
	var articles []model.Article
	err := r.db.Preload("Category").
		Where("title ILIKE ? OR content ILIKE ? OR summary ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Order("view_count DESC, created_at DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

// CountBySlugPrefix counts articles with slugs starting with prefix
func (r *ArticleRepository) CountBySlugPrefix(prefix string) int64 {
	var count int64
	r.db.Model(&model.Article{}).Where("slug LIKE ?", prefix+"%").Count(&count)
	return count
}
