package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/repository"
	"gorm.io/gorm"
)

type NewsHandler struct {
	repo *repository.NewsRepository
}

func NewNewsHandler(db *gorm.DB) *NewsHandler {
	return &NewsHandler{
		repo: repository.NewNewsRepository(db),
	}
}

// List returns a paginated list of news items
func (h *NewsHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sourceName := c.Query("source")
	processedStr := c.Query("processed")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var processed *bool
	if processedStr != "" {
		p := processedStr == "true"
		processed = &p
	}

	items, total, err := h.repo.List(page, limit, sourceName, processed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// Get returns a single news item
func (h *NewsHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// Delete deletes a news item
func (h *NewsHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// MarkProcessed marks a news item as processed
func (h *NewsHandler) MarkProcessed(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.repo.MarkProcessed(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked as processed"})
}

// GetUnprocessed returns unprocessed news items
func (h *NewsHandler) GetUnprocessed(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	items, err := h.repo.FindUnprocessed(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"count": len(items),
	})
}
