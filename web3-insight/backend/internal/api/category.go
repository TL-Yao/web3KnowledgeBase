package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

type CategoryHandler struct {
	repo *repository.CategoryRepository
}

func NewCategoryHandler(repo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

// ListCategories godoc
// @Summary List all categories
// @Description Get flat list of all categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} model.Category
// @Router /api/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategoryTree godoc
// @Summary Get category tree
// @Description Get hierarchical category tree structure
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} model.Category
// @Router /api/categories/tree [get]
func (h *CategoryHandler) GetTree(c *gin.Context) {
	tree, err := h.repo.GetTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tree)
}

// GetCategory godoc
// @Summary Get category by ID or slug
// @Description Get a single category by its ID or slug
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID or slug"
// @Success 200 {object} model.Category
// @Router /api/categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
	idParam := c.Param("id")

	var category *model.Category
	var err error

	// Try parsing as UUID first
	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		category, err = h.repo.GetByID(id)
	} else {
		// Treat as slug
		category, err = h.repo.GetBySlug(idParam)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

type CreateCategoryRequest struct {
	Name        string     `json:"name" binding:"required"`
	NameEn      string     `json:"nameEn"`
	Slug        string     `json:"slug" binding:"required"`
	ParentID    *uuid.UUID `json:"parentId"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	SortOrder   int        `json:"sortOrder"`
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body CreateCategoryRequest true "Category data"
// @Success 201 {object} model.Category
// @Router /api/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := &model.Category{
		Name:        req.Name,
		NameEn:      req.NameEn,
		Slug:        req.Slug,
		ParentID:    req.ParentID,
		Description: req.Description,
		Icon:        req.Icon,
		SortOrder:   req.SortOrder,
	}

	if err := h.repo.Create(category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

type UpdateCategoryRequest struct {
	Name        string     `json:"name"`
	NameEn      string     `json:"nameEn"`
	Slug        string     `json:"slug"`
	ParentID    *uuid.UUID `json:"parentId"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	SortOrder   *int       `json:"sortOrder"`
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Update an existing category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param category body UpdateCategoryRequest true "Category data"
// @Success 200 {object} model.Category
// @Router /api/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	category, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.NameEn != "" {
		category.NameEn = req.NameEn
	}
	if req.Slug != "" {
		category.Slug = req.Slug
	}
	if req.ParentID != nil {
		category.ParentID = req.ParentID
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.Icon != "" {
		category.Icon = req.Icon
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	if err := h.repo.Update(category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a category and its children by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 204 "No Content"
// @Router /api/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
