package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/collector"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DataSourceHandler struct {
	repo         *repository.DataSourceRepository
	rssCollector *collector.RSSCollector
}

func NewDataSourceHandler(db *gorm.DB) *DataSourceHandler {
	repo := repository.NewDataSourceRepository(db)
	newsRepo := repository.NewNewsRepository(db)
	return &DataSourceHandler{
		repo:         repo,
		rssCollector: collector.NewRSSCollector(newsRepo, repo),
	}
}

// List returns all data sources
func (h *DataSourceHandler) List(c *gin.Context) {
	sources, err := h.repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sources)
}

// Get returns a single data source
func (h *DataSourceHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	source, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "data source not found"})
		return
	}
	c.JSON(http.StatusOK, source)
}

// CreateDataSourceRequest represents the request body for creating a data source
type CreateDataSourceRequest struct {
	Name          string         `json:"name" binding:"required"`
	Type          string         `json:"type" binding:"required,oneof=rss api crawl"`
	URL           string         `json:"url" binding:"required,url"`
	Config        datatypes.JSON `json:"config,omitempty"`
	Enabled       *bool          `json:"enabled,omitempty"`
	FetchInterval int            `json:"fetchInterval,omitempty"`
}

// Create creates a new data source
func (h *DataSourceHandler) Create(c *gin.Context) {
	var req CreateDataSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate RSS URL if type is RSS
	if req.Type == model.DataSourceTypeRSS {
		_, err := h.rssCollector.ValidateFeedURL(req.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid RSS feed URL: " + err.Error()})
			return
		}
	}

	source := &model.DataSource{
		Name:    req.Name,
		Type:    req.Type,
		URL:     req.URL,
		Enabled: true,
	}

	if req.Config != nil {
		source.Config = req.Config
	}
	if req.Enabled != nil {
		source.Enabled = *req.Enabled
	}
	if req.FetchInterval > 0 {
		source.FetchInterval = req.FetchInterval
	} else {
		source.FetchInterval = 3600 // default 1 hour
	}

	if err := h.repo.Create(source); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, source)
}

// Update updates a data source
func (h *DataSourceHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	source, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "data source not found"})
		return
	}

	var req CreateDataSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate RSS URL if type is RSS and URL changed
	if req.Type == model.DataSourceTypeRSS && req.URL != source.URL {
		_, err := h.rssCollector.ValidateFeedURL(req.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid RSS feed URL: " + err.Error()})
			return
		}
	}

	source.Name = req.Name
	source.Type = req.Type
	source.URL = req.URL
	if req.Config != nil {
		source.Config = req.Config
	}
	if req.Enabled != nil {
		source.Enabled = *req.Enabled
	}
	if req.FetchInterval > 0 {
		source.FetchInterval = req.FetchInterval
	}

	if err := h.repo.Update(source); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, source)
}

// Delete deletes a data source
func (h *DataSourceHandler) Delete(c *gin.Context) {
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

// TriggerSync manually triggers a sync for a data source
func (h *DataSourceHandler) TriggerSync(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	source, err := h.repo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "data source not found"})
		return
	}

	// For RSS sources, trigger immediate sync
	if source.Type == model.DataSourceTypeRSS {
		result, err := h.rssCollector.Collect(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":    "sync completed",
			"itemsFound": result.ItemsFound,
			"itemsNew":   result.ItemsNew,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "sync not supported for this source type yet"})
}

// ValidateURL validates a feed URL without creating a source
func (h *DataSourceHandler) ValidateURL(c *gin.Context) {
	var req struct {
		URL  string `json:"url" binding:"required,url"`
		Type string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type == model.DataSourceTypeRSS {
		feed, err := h.rssCollector.ValidateFeedURL(req.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid": false,
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"valid":       true,
			"title":       feed.Title,
			"description": feed.Description,
			"itemCount":   len(feed.Items),
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "validation not supported for this type"})
}
