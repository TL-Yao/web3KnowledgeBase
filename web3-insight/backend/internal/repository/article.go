package repository

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
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
	Tag        string // single tag filter (uses PostgreSQL array contains)
	Search     string
	Archived   string // "true", "false", or "all" (default: "false")
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
	if params.Tag != "" {
		query = query.Where("? = ANY(tags)", params.Tag)
	}

	switch params.Archived {
	case "true":
		query = query.Where("archived = ?", true)
	case "all":
		// no filter
	default:
		// "false" or empty — default: exclude archived
		query = query.Where("archived = ?", false)
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

func (r *ArticleRepository) UpdateTags(id uuid.UUID, tags []string) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Update("tags", pq.StringArray(tags)).Error
}

// RenameTag replaces oldName with newName in all articles' tag arrays
func (r *ArticleRepository) RenameTag(oldName, newName string) error {
	return r.db.Exec(
		"UPDATE articles SET tags = array_replace(tags, ?, ?) WHERE ? = ANY(tags)",
		oldName, newName, oldName,
	).Error
}

// CreateVersion saves the current article content as a version snapshot
func (r *ArticleRepository) CreateVersion(articleID uuid.UUID, content, editedBy, changeSummary string) (*model.ArticleVersion, error) {
	version := &model.ArticleVersion{
		ArticleID:     articleID,
		Content:       content,
		EditedBy:      editedBy,
		ChangeSummary: changeSummary,
	}
	return version, r.db.Create(version).Error
}

// ListVersions returns versions for an article (without content to keep response light)
func (r *ArticleRepository) ListVersions(articleID uuid.UUID) ([]model.ArticleVersion, error) {
	var versions []model.ArticleVersion
	err := r.db.Select("id, article_id, edited_by, change_summary, created_at").
		Where("article_id = ?", articleID).
		Order("created_at DESC").
		Find(&versions).Error
	return versions, err
}

// GetVersion returns a single version by ID (with content for rollback)
func (r *ArticleRepository) GetVersion(id uuid.UUID) (*model.ArticleVersion, error) {
	var version model.ArticleVersion
	err := r.db.Where("id = ?", id).First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// RemoveTag removes a tag name from all articles' tag arrays and returns affected count
func (r *ArticleRepository) RemoveTag(name string) (int64, error) {
	result := r.db.Exec(
		"UPDATE articles SET tags = array_remove(tags, ?) WHERE ? = ANY(tags)",
		name, name,
	)
	return result.RowsAffected, result.Error
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

// ListSimple returns paginated articles with optional filters
func (r *ArticleRepository) ListSimple(page, pageSize int, status string, categoryID *uuid.UUID, search string) ([]model.Article, int64, error) {
	query := r.db.Model(&model.Article{}).Preload("Category")

	if categoryID != nil {
		query = query.Where("category_id = ?", categoryID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("title ILIKE ? OR summary ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var articles []model.Article
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// UpdateEmbedding updates the embedding vector for an article
func (r *ArticleRepository) UpdateEmbedding(id uuid.UUID, embedding *pgvector.Vector) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).Update("embedding", embedding).Error
}

// FindWithoutEmbeddings returns articles that don't have embeddings
func (r *ArticleRepository) FindWithoutEmbeddings(limit int) ([]model.Article, error) {
	var articles []model.Article
	err := r.db.Where("embedding IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

// FindSimilarByEmbedding finds articles similar to the given embedding using cosine distance
func (r *ArticleRepository) FindSimilarByEmbedding(embedding *pgvector.Vector, limit int, excludeID *uuid.UUID) ([]model.Article, error) {
	var articles []model.Article

	query := r.db.Preload("Category").
		Where("embedding IS NOT NULL")

	if excludeID != nil {
		query = query.Where("id != ?", excludeID)
	}

	// Use raw SQL for vector ordering since GORM doesn't support parameterized ORDER BY
	err := query.Order(gorm.Expr("embedding <=> ?", embedding)).
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

// FindRelatedArticles finds articles related to a given article using vector similarity
func (r *ArticleRepository) FindRelatedArticles(articleID uuid.UUID, limit int) ([]model.Article, error) {
	// First get the article's embedding
	var article model.Article
	if err := r.db.Select("embedding").First(&article, "id = ?", articleID).Error; err != nil {
		return nil, err
	}

	if article.Embedding == nil {
		return nil, nil // No embedding, return empty
	}

	return r.FindSimilarByEmbedding(article.Embedding, limit, &articleID)
}

// SemanticSearch performs semantic search using vector similarity
func (r *ArticleRepository) SemanticSearch(embedding *pgvector.Vector, limit int, categoryID *uuid.UUID, status string) ([]model.Article, error) {
	var articles []model.Article

	query := r.db.Preload("Category").
		Where("embedding IS NOT NULL")

	if categoryID != nil {
		query = query.Where("category_id = ?", categoryID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Use gorm.Expr for vector ordering
	err := query.Order(gorm.Expr("embedding <=> ?", embedding)).
		Limit(limit).
		Find(&articles).Error
	return articles, err
}
