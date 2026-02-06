package api

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
}

func NewKBUpdateHandler(db *gorm.DB) *KBUpdateHandler {
	// Initialize repositories
	keywordRepo := repository.NewKeywordRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	jobRepo := repository.NewKBUpdateJobRepository(db)

	// Initialize services
	keywordPool := service.NewKeywordPoolService(keywordRepo)

	// Note: Classifier disabled for now (would need actual LLM router from config)
	// Pass nil classifier to skip auto-classification
	articleGen := service.NewArticleGeneratorService(articleRepo, nil)

	orchestrator := service.NewKBUpdateOrchestrator(keywordPool, articleGen, keywordRepo, jobRepo)
	scheduler := service.NewKBScheduler(orchestrator)

	return &KBUpdateHandler{
		orchestrator: orchestrator,
		scheduler:    scheduler,
		keywordPool:  keywordPool,
		keywordRepo:  keywordRepo,
		jobRepo:      jobRepo,
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

	log.Printf("API: Initializing keyword pool with %d keywords", req.Count)

	err := h.keywordPool.InitializePool(c.Request.Context(), req.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize keyword pool",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Keyword pool initialized successfully",
		"count":   req.Count,
	})
}

// GetKeywordStats retrieves statistics about the keyword pool
func (h *KBUpdateHandler) GetKeywordStats(c *gin.Context) {
	pendingCount, err := h.keywordRepo.CountPendingKeywords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch keyword stats",
			"details": err.Error(),
		})
		return
	}

	// Count used keywords
	usedKeywords, err := h.keywordRepo.GetAllUsedKeywords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch used keywords",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pending": pendingCount,
		"used":    len(usedKeywords),
		"total":   int(pendingCount) + len(usedKeywords),
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
