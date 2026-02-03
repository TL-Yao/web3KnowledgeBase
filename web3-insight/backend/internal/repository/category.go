package repository

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

// LLMCategoryInfo contains info for creating a new category from LLM
// This is defined here to avoid circular imports with the service package
type LLMCategoryInfo struct {
	Name        string
	NameEn      string
	ParentPath  *string
	Icon        string
	Description string
}

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List() ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *CategoryRepository) GetTree() ([]model.Category, error) {
	var rootCategories []model.Category
	if err := r.db.Where("parent_id IS NULL").Order("sort_order ASC").Find(&rootCategories).Error; err != nil {
		return nil, err
	}

	for i := range rootCategories {
		if err := r.loadChildren(&rootCategories[i]); err != nil {
			return nil, err
		}
	}

	return rootCategories, nil
}

func (r *CategoryRepository) loadChildren(category *model.Category) error {
	var children []model.Category
	if err := r.db.Where("parent_id = ?", category.ID).Order("sort_order ASC").Find(&children).Error; err != nil {
		return err
	}

	category.Children = children

	for i := range children {
		if err := r.loadChildren(&children[i]); err != nil {
			return err
		}
	}

	return nil
}

func (r *CategoryRepository) GetByID(id uuid.UUID) (*model.Category, error) {
	var category model.Category
	if err := r.db.Preload("Children").First(&category, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) GetBySlug(slug string) (*model.Category, error) {
	var category model.Category
	if err := r.db.Preload("Children").First(&category, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

func (r *CategoryRepository) Update(category *model.Category) error {
	return r.db.Save(category).Error
}

func (r *CategoryRepository) Delete(id uuid.UUID) error {
	// Recursively delete all descendant categories
	if err := r.deleteDescendants(id); err != nil {
		return err
	}
	return r.db.Delete(&model.Category{}, "id = ?", id).Error
}

func (r *CategoryRepository) deleteDescendants(parentID uuid.UUID) error {
	// Find all direct children
	var children []model.Category
	if err := r.db.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}

	// Recursively delete each child's descendants first
	for _, child := range children {
		if err := r.deleteDescendants(child.ID); err != nil {
			return err
		}
	}

	// Delete direct children
	return r.db.Where("parent_id = ?", parentID).Delete(&model.Category{}).Error
}

func (r *CategoryRepository) Search(query string, limit int) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("name ILIKE ? OR name_en ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Order("sort_order ASC").
		Limit(limit).
		Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) UpdateArticleCount(id uuid.UUID) error {
	var count int64
	if err := r.db.Model(&model.Article{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	return r.db.Model(&model.Category{}).Where("id = ?", id).Update("article_count", count).Error
}

// FindAll returns all categories
func (r *CategoryRepository) FindAll() ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// FindByPath finds a category by its full path (e.g., "基础技术/区块链原理/共识机制")
func (r *CategoryRepository) FindByPath(path string) (*model.Category, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	var current *model.Category
	var parentID *uuid.UUID

	for _, name := range parts {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		var cat model.Category
		query := r.db.Where("name = ?", name)
		if parentID == nil {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", *parentID)
		}

		if err := query.First(&cat).Error; err != nil {
			return nil, fmt.Errorf("category not found: %s", name)
		}

		current = &cat
		parentID = &cat.ID
	}

	return current, nil
}

// CreateCategoryFromLLM creates a new category from LLM suggestion
// Returns the created category, whether it was newly created, and any error
func (r *CategoryRepository) CreateCategoryFromLLM(info *LLMCategoryInfo) (*model.Category, bool, error) {
	if info == nil {
		return nil, false, fmt.Errorf("category info is nil")
	}

	// Determine parent ID
	var parentID *uuid.UUID
	if info.ParentPath != nil && *info.ParentPath != "" {
		// Find or create parent category chain
		parent, _, err := r.FindOrCreateByPath(*info.ParentPath)
		if err != nil {
			return nil, false, fmt.Errorf("failed to find/create parent: %w", err)
		}
		parentID = &parent.ID
	}

	// Check if category already exists
	var existing model.Category
	query := r.db.Where("name = ?", info.Name)
	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	if err := query.First(&existing).Error; err == nil {
		// Category already exists
		return &existing, false, nil
	}

	// Generate slug
	slug := r.generateSlug(info.Name)

	// Ensure slug is unique
	slug = r.ensureUniqueSlug(slug)

	// Create the new category
	category := &model.Category{
		ID:          uuid.New(),
		Name:        info.Name,
		NameEn:      info.NameEn,
		Slug:        slug,
		ParentID:    parentID,
		Description: info.Description,
		Icon:        info.Icon,
		SortOrder:   0,
		AutoCreated: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := r.db.Create(category).Error; err != nil {
		return nil, false, fmt.Errorf("failed to create category: %w", err)
	}

	return category, true, nil
}

// FindOrCreateByPath finds a category by path, creating any missing categories along the way
// Returns the final category, whether any categories were created, and any error
func (r *CategoryRepository) FindOrCreateByPath(path string) (*model.Category, bool, error) {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil, false, fmt.Errorf("empty path")
	}

	var current *model.Category
	var parentID *uuid.UUID
	created := false

	for _, name := range parts {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		var cat model.Category
		query := r.db.Where("name = ?", name)
		if parentID == nil {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", *parentID)
		}

		if err := query.First(&cat).Error; err != nil {
			// Category doesn't exist, create it
			slug := r.generateSlug(name)
			slug = r.ensureUniqueSlug(slug)

			cat = model.Category{
				ID:          uuid.New(),
				Name:        name,
				NameEn:      name, // Default to same as name
				Slug:        slug,
				ParentID:    parentID,
				SortOrder:   0,
				AutoCreated: true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if err := r.db.Create(&cat).Error; err != nil {
				return nil, created, fmt.Errorf("failed to create category '%s': %w", name, err)
			}
			created = true
		}

		current = &cat
		parentID = &cat.ID
	}

	return current, created, nil
}

// generateSlug creates a URL-friendly slug from a name
func (r *CategoryRepository) generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace Chinese characters with pinyin-like representation or transliteration
	// For simplicity, we'll just use a hash-based approach for non-ASCII
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			result.WriteRune('-')
		} else if r > 127 {
			// For non-ASCII (like Chinese), include as-is but will be URL encoded
			result.WriteRune(r)
		}
	}

	slug = result.String()

	// Clean up multiple dashes
	re := regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")

	// Trim leading/trailing dashes
	slug = strings.Trim(slug, "-")

	// If empty, generate a timestamp-based slug
	if slug == "" {
		slug = fmt.Sprintf("category-%d", time.Now().UnixNano())
	}

	return slug
}

// ensureUniqueSlug appends a suffix if the slug already exists
func (r *CategoryRepository) ensureUniqueSlug(slug string) string {
	originalSlug := slug
	counter := 1

	for {
		var count int64
		r.db.Model(&model.Category{}).Where("slug = ?", slug).Count(&count)
		if count == 0 {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}
}
