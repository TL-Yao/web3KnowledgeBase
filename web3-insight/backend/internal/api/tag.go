package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
}

func NewTagHandler(db *gorm.DB, llmCfg *config.LLMConfig) *TagHandler {
	return &TagHandler{
		db:          db,
		tagRepo:     repository.NewTagRepository(db),
		articleRepo: repository.NewArticleRepository(db),
		llmConfig:   llmCfg,
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
	llmRouter := llm.NewRouterFromConfig(h.llmConfig)
	tagger := service.NewTagger(llmRouter, h.tagRepo, h.articleRepo)

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
