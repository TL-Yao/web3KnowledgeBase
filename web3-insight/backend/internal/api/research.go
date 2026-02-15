package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
)

type ResearchHandler struct {
	researchService *service.ResearchService
	sessionRepo     *repository.ResearchSessionRepository
	articleRepo     *repository.ArticleRepository
	researchConfig  *config.ResearchConfig
}

func NewResearchHandler(
	researchService *service.ResearchService,
	sessionRepo *repository.ResearchSessionRepository,
	articleRepo *repository.ArticleRepository,
	researchConfig *config.ResearchConfig,
) *ResearchHandler {
	return &ResearchHandler{
		researchService: researchService,
		sessionRepo:     sessionRepo,
		articleRepo:     articleRepo,
		researchConfig:  researchConfig,
	}
}

// GetDomains returns available research domains from config.
func (h *ResearchHandler) GetDomains(c *gin.Context) {
	if h.researchConfig == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "research feature not configured"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"domains": h.researchConfig.Domains})
}

// StartSession creates a new research session and begins plan generation.
func (h *ResearchHandler) StartSession(c *gin.Context) {
	var req service.StartResearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if req.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}

	session, err := h.researchService.StartSession(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "concurrent session limit reached (3)" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		if strings.HasPrefix(err.Error(), "invalid domain:") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"sessionId": session.ID,
		"status":    session.Status,
	})
}

// ListSessions returns paginated research session history.
func (h *ResearchHandler) ListSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	sessions, total, err := h.sessionRepo.List(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetSession returns full session details.
func (h *ResearchHandler) GetSession(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetSessionStatus returns lightweight status info for polling.
func (h *ResearchHandler) GetSessionStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	resp := gin.H{
		"status":      session.Status,
		"stage":       session.Stage,
		"stageDetail": session.StageDetail,
	}

	if session.Status == "plan_review" {
		resp["researchPlan"] = session.ResearchPlan
	}

	if session.Status == "failed" {
		resp["error"] = session.ErrorMessage
	}

	// Include article data when completed (preloaded via GetByID)
	if session.Status == "completed" && session.Article != nil {
		resp["articleSlug"] = session.Article.Slug
		resp["article"] = session.Article
	}

	c.JSON(http.StatusOK, resp)
}

// ApprovePlan approves (or edits) the research plan and continues generation.
func (h *ResearchHandler) ApprovePlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	var req struct {
		Approved   bool   `json:"approved"`
		EditedPlan string `json:"editedPlan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.researchService.ApprovePlan(c.Request.Context(), id, req.EditedPlan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "researching"})
}

// CancelSession cancels an in-progress research session.
func (h *ResearchHandler) CancelSession(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	if err := h.researchService.CancelSession(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

// PinFinding adds a chat finding to the session's pinned list.
func (h *ResearchHandler) PinFinding(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	var req struct {
		MessageContent string `json:"messageContent" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messageContent is required"})
		return
	}

	if err := h.researchService.PinFinding(c.Request.Context(), id, req.MessageContent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated session to get latest pinned findings
	session, _ := h.sessionRepo.GetByID(id)
	c.JSON(http.StatusOK, gin.H{"pinnedFindings": session.PinnedFindings})
}

// RemovePin removes a pinned finding by index.
func (h *ResearchHandler) RemovePin(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	index, err := strconv.Atoi(c.Param("index"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
		return
	}

	if err := h.researchService.RemovePinFinding(c.Request.Context(), id, index); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, _ := h.sessionRepo.GetByID(id)
	c.JSON(http.StatusOK, gin.H{"pinnedFindings": session.PinnedFindings})
}

// SetPinPosition sets or clears the target block for a pinned finding.
func (h *ResearchHandler) SetPinPosition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	index, err := strconv.Atoi(c.Param("index"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
		return
	}

	var req struct {
		TargetBlockIndex *int   `json:"targetBlockIndex"`
		TargetPreview    string `json:"targetPreview"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.researchService.SetPinPosition(c.Request.Context(), id, index, req.TargetBlockIndex, req.TargetPreview); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, _ := h.sessionRepo.GetByID(id)
	c.JSON(http.StatusOK, gin.H{"pinnedFindings": session.PinnedFindings})
}

// IntegrateFindings triggers re-generation to merge pinned content.
func (h *ResearchHandler) IntegrateFindings(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	if err := h.researchService.IntegrateFindings(c.Request.Context(), id); err != nil {
		if err.Error() == "concurrent session limit reached (3)" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "writing"})
}

// DeleteSession deletes a research session and its linked article.
func (h *ResearchHandler) DeleteSession(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
		return
	}

	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// Delete linked article if exists
	if session.ArticleID != nil {
		if err := h.articleRepo.Delete(*session.ArticleID); err != nil {
			log.Printf("Warning: failed to delete linked article %s for session %s: %v", *session.ArticleID, id, err)
		}
	}

	if err := h.sessionRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session deleted"})
}
