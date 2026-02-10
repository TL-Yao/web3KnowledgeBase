package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

type KBUpdateHandler struct {
	orchestrator *service.KBUpdateOrchestrator
	scheduler    *service.KBScheduler
	keywordPool  *service.KeywordPoolService
	keywordRepo  *repository.KeywordRepository
	jobRepo      *repository.KBUpdateJobRepository
	themeRepo    *repository.ThemeRepository
	configRepo   *repository.ConfigRepository
	prompts      *config.PromptsConfig
}

func NewKBUpdateHandler(db *gorm.DB, prompts *config.PromptsConfig) *KBUpdateHandler {
	// Initialize repositories
	keywordRepo := repository.NewKeywordRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	jobRepo := repository.NewKBUpdateJobRepository(db)
	themeRepo := repository.NewThemeRepository(db)
	configRepo := repository.NewConfigRepository(db)

	// Initialize services
	keywordPool := service.NewKeywordPoolService(keywordRepo, prompts)
	articleGen := service.NewArticleGeneratorService(articleRepo, nil, prompts)
	orchestrator := service.NewKBUpdateOrchestrator(keywordPool, articleGen, keywordRepo, jobRepo, themeRepo, configRepo, prompts)
	scheduler := service.NewKBScheduler(orchestrator)

	return &KBUpdateHandler{
		orchestrator: orchestrator,
		scheduler:    scheduler,
		keywordPool:  keywordPool,
		keywordRepo:  keywordRepo,
		jobRepo:      jobRepo,
		themeRepo:    themeRepo,
		configRepo:   configRepo,
		prompts:      prompts,
	}
}

// TriggerUpdateRequest represents the request body for triggering an update
type TriggerUpdateRequest struct {
	TriggerType string `json:"trigger_type" binding:"omitempty,oneof=manual scheduled"`
}

// TriggerUpdate triggers a knowledge base update asynchronously
func (h *KBUpdateHandler) TriggerUpdate(c *gin.Context) {
	var req TriggerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Default to manual if not specified
	if req.TriggerType == "" {
		req.TriggerType = "manual"
	}

	log.Printf("API: Triggering KB update (type: %s)", req.TriggerType)

	// Start the update process asynchronously
	// The orchestrator will create the job and update its status as it progresses
	go func() {
		ctx := context.Background()
		job, err := h.orchestrator.RunUpdate(ctx, req.TriggerType)
		if err != nil {
			log.Printf("KB update failed: %v", err)
		} else {
			log.Printf("KB update completed successfully: job_id=%s, articles=%d/%d",
				job.ID, job.ArticlesGenerated, job.KeywordsGenerated)
		}
	}()

	// Return immediately with accepted status
	c.JSON(http.StatusAccepted, gin.H{
		"status": "accepted",
		"message": "Knowledge base update started. Use GET /api/kb/update/jobs to track progress.",
	})
}

// GetJobStatus retrieves the status of a specific job
func (h *KBUpdateHandler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	job, err := h.orchestrator.GetJobStatus(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, job)
}

// GetUpdateHistory retrieves the update history with pagination
func (h *KBUpdateHandler) GetUpdateHistory(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	jobs, total, err := h.orchestrator.GetUpdateHistory(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch update history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"items":     jobs,
	})
}

// InitKeywordPoolRequest represents the request body for initializing the keyword pool
type InitKeywordPoolRequest struct {
	Count int `json:"count" binding:"omitempty,min=1,max=500"`
}

// InitKeywordPool initializes the keyword pool
func (h *KBUpdateHandler) InitKeywordPool(c *gin.Context) {
	var req InitKeywordPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Default to 200 if not specified
	if req.Count == 0 {
		req.Count = 200
	}

	// Get active theme for initialization
	activeTheme, err := h.themeRepo.GetActive()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No active theme configured",
			"details": err.Error(),
		})
		return
	}

	themeID := activeTheme.ID
	count := req.Count
	log.Printf("API: Initializing keyword pool for theme %s with %d keywords", themeID, count)

	// Run async — keyword generation can take minutes
	go func() {
		ctx := context.Background()
		if err := h.keywordPool.InitializePool(ctx, themeID, count); err != nil {
			log.Printf("Keyword pool initialization failed for theme %s: %v", themeID, err)
		} else {
			log.Printf("Keyword pool initialization completed for theme %s", themeID)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Keyword pool initialization started",
		"count":   count,
		"themeId": themeID,
	})
}

// GetKeywordStats retrieves statistics about the keyword pool for the active theme
func (h *KBUpdateHandler) GetKeywordStats(c *gin.Context) {
	activeTheme, err := h.themeRepo.GetActive()
	if err != nil {
		// Fall back to global stats if no active theme
		pendingCount, err := h.keywordRepo.CountPendingKeywords()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch keyword stats"})
			return
		}
		usedKeywords, err := h.keywordRepo.GetAllUsedKeywords()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch used keywords"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"pending": pendingCount,
			"used":    len(usedKeywords),
			"total":   int(pendingCount) + len(usedKeywords),
		})
		return
	}

	pendingCount, err := h.keywordRepo.CountPendingByTheme(activeTheme.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch keyword stats"})
		return
	}

	totalCount, err := h.keywordRepo.CountByTheme(activeTheme.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch theme keywords"})
		return
	}

	usedCount := totalCount - pendingCount
	c.JSON(http.StatusOK, gin.H{
		"pending": pendingCount,
		"used":    usedCount,
		"total":   totalCount,
		"themeId": activeTheme.ID,
	})
}

// GetSchedulerStatus retrieves the current scheduler status
func (h *KBUpdateHandler) GetSchedulerStatus(c *gin.Context) {
	status := h.scheduler.GetSchedulerStatus()
	c.JSON(http.StatusOK, status)
}

// StartScheduler starts the automatic scheduler
func (h *KBUpdateHandler) StartScheduler(c *gin.Context) {
	if err := h.scheduler.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start scheduler",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Scheduler started successfully",
		"status":  h.scheduler.GetSchedulerStatus(),
	})
}

// StopScheduler stops the automatic scheduler
func (h *KBUpdateHandler) StopScheduler(c *gin.Context) {
	h.scheduler.Stop()
	c.JSON(http.StatusOK, gin.H{
		"message": "Scheduler stopped successfully",
	})
}

// === Theme Management Endpoints ===

// GetThemes returns all themes with keyword stats
func (h *KBUpdateHandler) GetThemes(c *gin.Context) {
	themes, err := h.themeRepo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch themes"})
		return
	}

	type ThemeWithStats struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Category        string `json:"category"`
		Description     string `json:"description"`
		Status          string `json:"status"`
		SortOrder       int    `json:"sortOrder"`
		KeywordsPending int64  `json:"keywordsPending"`
		KeywordsUsed    int64  `json:"keywordsUsed"`
		KeywordsTotal   int64  `json:"keywordsTotal"`
	}

	var result []ThemeWithStats
	var activeThemeID string

	for _, t := range themes {
		pending, _ := h.keywordRepo.CountPendingByTheme(t.ID)
		total, _ := h.keywordRepo.CountByTheme(t.ID)
		used := total - pending

		if t.Status == "active" {
			activeThemeID = t.ID
		}

		result = append(result, ThemeWithStats{
			ID:              t.ID,
			Name:            t.Name,
			Category:        t.Category,
			Description:     t.Description,
			Status:          t.Status,
			SortOrder:       t.SortOrder,
			KeywordsPending: pending,
			KeywordsUsed:    used,
			KeywordsTotal:   total,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"themes":        result,
		"activeThemeId": activeThemeID,
	})
}

// GetActiveTheme returns the currently active theme
func (h *KBUpdateHandler) GetActiveTheme(c *gin.Context) {
	theme, err := h.themeRepo.GetActive()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active theme"})
		return
	}
	c.JSON(http.StatusOK, theme)
}

// SetActiveTheme activates a theme by ID (pauses all others)
func (h *KBUpdateHandler) SetActiveTheme(c *gin.Context) {
	themeID := c.Param("id")

	// Verify theme exists in config
	if _, err := h.prompts.GetThemeByID(themeID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Theme not found: %s", themeID)})
		return
	}

	if err := h.themeRepo.SetActive(themeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate theme"})
		return
	}

	log.Printf("API: Activated theme %s", themeID)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Theme '%s' activated", themeID),
		"themeId": themeID,
	})
}

// === KB Config Endpoints ===

// GetKBConfig returns KB configuration (batch size, etc.)
func (h *KBUpdateHandler) GetKBConfig(c *gin.Context) {
	batchSize := h.orchestrator.GetBatchSize()

	c.JSON(http.StatusOK, gin.H{
		"batchSize":    batchSize,
		"maxBatchSize": service.MaxKeywordBatchSize,
	})
}

// UpdateBatchSizeRequest is the request body for updating batch size
type UpdateBatchSizeRequest struct {
	BatchSize int `json:"batchSize" binding:"required,min=1,max=10"`
}

// UpdateBatchSize updates the KB batch size in DB config
func (h *KBUpdateHandler) UpdateBatchSize(c *gin.Context) {
	var req UpdateBatchSizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid batch size (must be 1-10)"})
		return
	}

	val, _ := json.Marshal(strconv.Itoa(req.BatchSize))
	if err := h.configRepo.SetJSON("kb.batch_size", val, "Number of articles per KB update cycle"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save batch size"})
		return
	}

	log.Printf("API: Updated batch size to %d", req.BatchSize)
	c.JSON(http.StatusOK, gin.H{
		"message":   "Batch size updated",
		"batchSize": req.BatchSize,
	})
}
