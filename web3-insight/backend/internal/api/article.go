package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

type ArticleHandler struct {
	repo *repository.ArticleRepository
}

func NewArticleHandler(repo *repository.ArticleRepository) *ArticleHandler {
	return &ArticleHandler{repo: repo}
}

// ListArticles godoc
// @Summary List articles
// @Description Get paginated list of articles with optional filters
// @Tags articles
// @Accept json
// @Produce json
// @Param category_id query string false "Filter by category ID"
// @Param status query string false "Filter by status (draft, published)"
// @Param search query string false "Search in title and summary"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Success 200 {object} repository.ArticleListResult
// @Router /api/articles [get]
func (h *ArticleHandler) List(c *gin.Context) {
	params := repository.ArticleListParams{
		Status: c.Query("status"),
		Search: c.Query("search"),
	}

	if categoryID := c.Query("category_id"); categoryID != "" {
		id, err := uuid.Parse(categoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		params.CategoryID = &id
	}

	if page := c.Query("page"); page != "" {
		p, _ := strconv.Atoi(page)
		params.Page = p
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		ps, _ := strconv.Atoi(pageSize)
		params.PageSize = ps
	}

	result, err := h.repo.List(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetArticle godoc
// @Summary Get article by ID or slug
// @Description Get a single article by its ID or slug
// @Tags articles
// @Accept json
// @Produce json
// @Param id path string true "Article ID or slug"
// @Success 200 {object} model.Article
// @Router /api/articles/{id} [get]
func (h *ArticleHandler) Get(c *gin.Context) {
	idParam := c.Param("id")

	var article *model.Article
	var err error

	// Try parsing as UUID first
	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		article, err = h.repo.GetByID(id)
	} else {
		// Treat as slug
		article, err = h.repo.GetBySlug(idParam)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	// Increment view count asynchronously
	go func() {
		_ = h.repo.IncrementViewCount(article.ID)
	}()

	c.JSON(http.StatusOK, article)
}

type CreateArticleRequest struct {
	Title      string     `json:"title" binding:"required"`
	Slug       string     `json:"slug" binding:"required"`
	Content    string     `json:"content" binding:"required"`
	Summary    string     `json:"summary"`
	CategoryID *uuid.UUID `json:"categoryId"`
	Tags       []string   `json:"tags"`
	Status     string     `json:"status"`
}

// CreateArticle godoc
// @Summary Create a new article
// @Description Create a new article
// @Tags articles
// @Accept json
// @Produce json
// @Param article body CreateArticleRequest true "Article data"
// @Success 201 {object} model.Article
// @Router /api/articles [post]
func (h *ArticleHandler) Create(c *gin.Context) {
	var req CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article := &model.Article{
		Title:      req.Title,
		Slug:       req.Slug,
		Content:    req.Content,
		Summary:    req.Summary,
		CategoryID: req.CategoryID,
		Tags:       req.Tags,
		Status:     req.Status,
	}

	if article.Status == "" {
		article.Status = "draft"
	}

	if err := h.repo.Create(article); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, article)
}

type UpdateArticleRequest struct {
	Title      string     `json:"title"`
	Slug       string     `json:"slug"`
	Content    string     `json:"content"`
	Summary    string     `json:"summary"`
	CategoryID *uuid.UUID `json:"categoryId"`
	Tags       []string   `json:"tags"`
	Status     string     `json:"status"`
}

// UpdateArticle godoc
// @Summary Update an article
// @Description Update an existing article
// @Tags articles
// @Accept json
// @Produce json
// @Param id path string true "Article ID"
// @Param article body UpdateArticleRequest true "Article data"
// @Success 200 {object} model.Article
// @Router /api/articles/{id} [put]
func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	article, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Slug != "" {
		article.Slug = req.Slug
	}
	if req.Content != "" {
		article.Content = req.Content
	}
	if req.Summary != "" {
		article.Summary = req.Summary
	}
	if req.CategoryID != nil {
		article.CategoryID = req.CategoryID
	}
	if req.Tags != nil {
		article.Tags = req.Tags
	}
	if req.Status != "" {
		article.Status = req.Status
	}

	if err := h.repo.Update(article); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, article)
}

// DeleteArticle godoc
// @Summary Delete an article
// @Description Delete an article by ID
// @Tags articles
// @Accept json
// @Produce json
// @Param id path string true "Article ID"
// @Success 204 "No Content"
// @Router /api/articles/{id} [delete]
func (h *ArticleHandler) Delete(c *gin.Context) {
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

// RegenerateArticle godoc
// @Summary Regenerate article content
// @Description Trigger AI regeneration of article content
// @Tags articles
// @Accept json
// @Produce json
// @Param id path string true "Article ID"
// @Success 202 {object} map[string]string
// @Router /api/articles/{id}/regenerate [post]
func (h *ArticleHandler) Regenerate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	_, err = h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	// TODO: Queue regeneration task with Asynq
	// For now, return accepted status
	c.JSON(http.StatusAccepted, gin.H{
		"message":    "regeneration queued",
		"article_id": id,
	})
}
