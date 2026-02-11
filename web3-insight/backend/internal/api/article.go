package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

type ArticleHandler struct {
	repo    *repository.ArticleRepository
	db      *gorm.DB
	updater *service.ArticleUpdater
}

func NewArticleHandler(repo *repository.ArticleRepository, db *gorm.DB, updater *service.ArticleUpdater) *ArticleHandler {
	return &ArticleHandler{repo: repo, db: db, updater: updater}
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
	// Accept both "q" (frontend) and "search" (legacy) parameter names
	search := c.Query("q")
	if search == "" {
		search = c.Query("search")
	}

	params := repository.ArticleListParams{
		Status:   c.Query("status"),
		Search:   search,
		Tag:      c.Query("tag"),
		Archived: c.DefaultQuery("archived", "false"),
	}

	// Accept both "category" (frontend) and "category_id" (legacy) parameter names
	categoryIDStr := c.Query("category")
	if categoryIDStr == "" {
		categoryIDStr = c.Query("category_id")
	}
	if categoryID := categoryIDStr; categoryID != "" {
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

	// Populate SourceKeyword for each article via batch keyword lookup
	if len(result.Articles) > 0 {
		articleIDs := make([]uuid.UUID, len(result.Articles))
		for i, a := range result.Articles {
			articleIDs[i] = a.ID
		}

		var keywords []model.Keyword
		h.db.Where("article_id IN ?", articleIDs).Find(&keywords)

		kwMap := make(map[uuid.UUID]string, len(keywords))
		for _, kw := range keywords {
			if kw.ArticleID != nil {
				kwMap[*kw.ArticleID] = kw.Keyword
			}
		}

		for i := range result.Articles {
			if kw, ok := kwMap[result.Articles[i].ID]; ok {
				result.Articles[i].SourceKeyword = kw
			}
		}
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
	Archived   *bool      `json:"archived"`
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
	if req.Archived != nil {
		article.Archived = *req.Archived
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

	// Clean up keyword references before deleting article
	h.db.Model(&model.Keyword{}).Where("article_id = ?", id).Update("article_id", nil)

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ArticleHandler) ToggleArchive(c *gin.Context) {
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

	article.Archived = !article.Archived
	if err := h.repo.Update(article); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, article)
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

type UpdateTagsRequest struct {
	Tags []string `json:"tags"`
}

func (h *ArticleHandler) UpdateTags(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	var req UpdateTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Deduplicate and trim
	seen := make(map[string]bool)
	var cleanTags []string
	for _, t := range req.Tags {
		t = strings.TrimSpace(t)
		if t != "" && len(t) <= 100 && !seen[t] {
			seen[t] = true
			cleanTags = append(cleanTags, t)
		}
	}

	if err := h.repo.UpdateTags(id, cleanTags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tags"})
		return
	}

	article, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	c.JSON(http.StatusOK, article)
}

type GenerateUpdateRequest struct {
	ConversationHistory []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"conversationHistory" binding:"required"`
}

func (h *ArticleHandler) GenerateUpdate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	article, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	var req GenerateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.ConversationHistory) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation history is required"})
		return
	}

	// Convert to llm.Message
	messages := make([]llm.Message, len(req.ConversationHistory))
	for i, m := range req.ConversationHistory {
		messages[i] = llm.Message{Role: m.Role, Content: m.Content}
	}

	result, err := h.updater.GenerateUpdate(context.Background(), article, messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate update: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

type ApplyUpdateRequest struct {
	UpdatedContent string `json:"updatedContent" binding:"required"`
	ChangeSummary  string `json:"changeSummary"`
}

func (h *ArticleHandler) ApplyUpdate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article ID"})
		return
	}

	article, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	var req ApplyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ChangeSummary == "" {
		req.ChangeSummary = "文章内容已更新"
	}

	// Save current content as version snapshot
	_, err = h.repo.CreateVersion(article.ID, article.Content, "chat_refinement", req.ChangeSummary)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create version snapshot"})
		return
	}

	// Update article content
	article.Content = req.UpdatedContent
	if err := h.repo.Update(article); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"article": article,
		"message": "Article updated successfully",
	})
}
