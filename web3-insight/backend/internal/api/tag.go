package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

type TagHandler struct {
	db          *gorm.DB
	tagRepo     *repository.TagRepository
	articleRepo *repository.ArticleRepository
	llmConfig   *config.LLMConfig
	keyProvider *service.KeyProvider
}

func NewTagHandler(db *gorm.DB, llmCfg *config.LLMConfig, kp *service.KeyProvider) *TagHandler {
	return &TagHandler{
		db:          db,
		tagRepo:     repository.NewTagRepository(db),
		articleRepo: repository.NewArticleRepository(db),
		llmConfig:   llmCfg,
		keyProvider: kp,
	}
}

// ListTags returns tags with optional filtering
func (h *TagHandler) List(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("q")
	themeID := c.Query("themeId")
	if themeID == "" {
		themeID = c.Query("theme")
	}

	if themeID != "" {
		tags, err := h.tagRepo.FindByTheme(themeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tags": tags})
		return
	}

	tags, err := h.tagRepo.FindAll(status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// GetStats returns tag usage statistics
func (h *TagHandler) GetStats(c *gin.Context) {
	stats, err := h.tagRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tag stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"totalTags":   stats["total"],
		"activeTags":  stats["active"],
		"pendingTags": stats["pending"],
		"universalTags": stats["universal"],
	})
}

// UpdateStatusRequest represents the request to change tag status
type UpdateTagStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active deprecated"`
}

// UpdateStatus changes a tag's status
func (h *TagHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	var req UpdateTagStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.tagRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	if err := h.tagRepo.UpdateStatus(tag.Name, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag status updated",
		"name":    tag.Name,
		"status":  req.Status,
	})
}

// ApprovePending approves a pending tag (sets status to active)
func (h *TagHandler) ApprovePending(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	tag, err := h.tagRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	if tag.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending tags can be approved"})
		return
	}

	if err := h.tagRepo.UpdateStatus(tag.Name, "active"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve tag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag approved",
		"name":    tag.Name,
	})
}

// GetInUse returns tags currently in use by articles with article counts
func (h *TagHandler) GetInUse(c *gin.Context) {
	tags, err := h.tagRepo.FindInUseWithCounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tags in use"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// Search returns tags matching a query string for autocomplete
func (h *TagHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q parameter required"})
		return
	}
	status := c.DefaultQuery("status", "active")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit > 50 {
		limit = 50
	}

	tags, err := h.tagRepo.Search(q, status, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	// Return lightweight response for autocomplete
	type tagResult struct {
		Name    string  `json:"name"`
		NameEn  string  `json:"nameEn"`
		ThemeID *string `json:"themeId"`
	}
	results := make([]tagResult, len(tags))
	for i, t := range tags {
		results[i] = tagResult{Name: t.Name, NameEn: t.NameEn, ThemeID: t.ThemeID}
	}

	c.JSON(http.StatusOK, gin.H{"tags": results})
}

type CreateTagRequest struct {
	Name    string  `json:"name" binding:"required,max=100"`
	NameEn  string  `json:"nameEn"`
	ThemeID *string `json:"themeId"`
}

func (h *TagHandler) Create(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag name is required"})
		return
	}

	existing, _ := h.tagRepo.FindByName(req.Name)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "tag already exists", "tag": existing})
		return
	}

	tag := &model.Tag{
		Name:   req.Name,
		NameEn: strings.TrimSpace(req.NameEn),
		ThemeID: req.ThemeID,
		Status: "active",
	}

	if err := h.tagRepo.Create(tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

type UpdateTagRequest struct {
	Name    *string `json:"name"`
	NameEn  *string `json:"nameEn"`
	ThemeID *string `json:"themeId"`
}

func (h *TagHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	tag, err := h.tagRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldName := tag.Name

	if req.Name != nil {
		newName := strings.TrimSpace(*req.Name)
		if newName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be empty"})
			return
		}
		if len(newName) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name too long (max 100 chars)"})
			return
		}
		if newName != tag.Name {
			existing, _ := h.tagRepo.FindByName(newName)
			if existing != nil {
				c.JSON(http.StatusConflict, gin.H{"error": "tag name already exists"})
				return
			}
			tag.Name = newName
		}
	}
	if req.NameEn != nil {
		tag.NameEn = strings.TrimSpace(*req.NameEn)
	}
	if req.ThemeID != nil {
		if *req.ThemeID == "" {
			tag.ThemeID = nil
		} else {
			tag.ThemeID = req.ThemeID
		}
	}

	if err := h.tagRepo.Update(tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tag"})
		return
	}

	if tag.Name != oldName {
		if err := h.articleRepo.RenameTag(oldName, tag.Name); err != nil {
			log.Printf("WARNING: tag renamed but article cascade failed: %v", err)
		}
	}

	c.JSON(http.StatusOK, tag)
}

func (h *TagHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag id"})
		return
	}

	tag, err := h.tagRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	affected, err := h.articleRepo.RemoveTag(tag.Name)
	if err != nil {
		log.Printf("WARNING: failed to remove tag from articles: %v", err)
	}

	if err := h.tagRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete tag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Tag deleted",
		"name":             tag.Name,
		"articlesAffected": affected,
	})
}

// BulkTag tags all untagged articles (or all with force=true query param)
func (h *TagHandler) BulkTag(c *gin.Context) {
	force := c.Query("force") == "true"

	// Get articles to tag
	var articles []model.Article
	query := h.db.Model(&model.Article{}).Order("created_at ASC")
	if !force {
		query = query.Where("tags IS NULL OR array_length(tags, 1) IS NULL OR array_length(tags, 1) = 0")
	}
	if err := query.Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}

	if len(articles) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No articles to tag", "tagged": 0})
		return
	}

	// Create tagger
	claudeKeyFunc := func() string { return h.keyProvider.GetKey("anthropic") }
	openaiKeyFunc := func() string { return h.keyProvider.GetKey("openai") }
	llmRouter := llm.NewRouterFromConfig(h.llmConfig, claudeKeyFunc, openaiKeyFunc)
	configRepo := repository.NewConfigRepository(h.db)
	tagger := service.NewTagger(llmRouter, h.tagRepo, h.articleRepo, configRepo)

	// Tag in background, return immediately
	total := len(articles)
	c.JSON(http.StatusAccepted, gin.H{
		"message": fmt.Sprintf("Tagging %d articles in background", total),
		"total":   total,
	})

	go func() {
		ctx := context.Background()
		success, failed := 0, 0
		for i, article := range articles {
			if err := tagger.TagArticle(ctx, &article); err != nil {
				log.Printf("[bulk-tag %d/%d] FAILED '%s': %v", i+1, total, article.Title, err)
				failed++
			} else {
				success++
				log.Printf("[bulk-tag %d/%d] OK '%s'", i+1, total, article.Title)
			}
		}
		log.Printf("[bulk-tag] Done: %d/%d success, %d failed", success, total, failed)
	}()
}
