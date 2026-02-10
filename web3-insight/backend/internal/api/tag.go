package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/repository"
	"gorm.io/gorm"
)

type TagHandler struct {
	tagRepo     *repository.TagRepository
	articleRepo *repository.ArticleRepository
}

func NewTagHandler(db *gorm.DB) *TagHandler {
	return &TagHandler{
		tagRepo:     repository.NewTagRepository(db),
		articleRepo: repository.NewArticleRepository(db),
	}
}

// ListTags returns tags with optional filtering
func (h *TagHandler) List(c *gin.Context) {
	status := c.Query("status")
	themeID := c.Query("themeId")

	if themeID != "" {
		tags, err := h.tagRepo.FindByTheme(themeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tags": tags})
		return
	}

	tags, err := h.tagRepo.FindAll(status)
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

	// Find tag by ID to get its name
	tags, err := h.tagRepo.FindAll("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}

	var tagName string
	for _, t := range tags {
		if t.ID == id {
			tagName = t.Name
			break
		}
	}
	if tagName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	if err := h.tagRepo.UpdateStatus(tagName, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag status updated",
		"name":    tagName,
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

	tags, err := h.tagRepo.FindAll("pending")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}

	var tagName string
	for _, t := range tags {
		if t.ID == id {
			tagName = t.Name
			break
		}
	}
	if tagName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending tag not found"})
		return
	}

	if err := h.tagRepo.UpdateStatus(tagName, "active"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve tag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag approved",
		"name":    tagName,
	})
}
