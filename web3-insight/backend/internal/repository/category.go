package repository

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

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
