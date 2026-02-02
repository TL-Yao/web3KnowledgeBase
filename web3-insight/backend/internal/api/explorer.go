package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ExplorerHandler struct {
	explorerRepo *repository.ExplorerRepository
	featureRepo  *repository.FeatureRepository
}

func NewExplorerHandler(db *gorm.DB) *ExplorerHandler {
	return &ExplorerHandler{
		explorerRepo: repository.NewExplorerRepository(db),
		featureRepo:  repository.NewFeatureRepository(db),
	}
}

// CreateExplorerRequest represents a request to create an explorer research entry
type CreateExplorerRequest struct {
	ChainName       string                 `json:"chainName" binding:"required"`
	ChainType       string                 `json:"chainType,omitempty"`
	ExplorerName    string                 `json:"explorerName" binding:"required"`
	ExplorerURL     string                 `json:"explorerUrl" binding:"required,url"`
	ExplorerType    string                 `json:"explorerType,omitempty"`
	Features        map[string]interface{} `json:"features,omitempty"`
	UIFeatures      map[string]interface{} `json:"uiFeatures,omitempty"`
	APIFeatures     map[string]interface{} `json:"apiFeatures,omitempty"`
	Screenshots     []string               `json:"screenshots,omitempty"`
	Analysis        string                 `json:"analysis,omitempty"`
	Strengths       []string               `json:"strengths,omitempty"`
	Weaknesses      []string               `json:"weaknesses,omitempty"`
	PopularityScore float64                `json:"popularityScore,omitempty"`
	ResearchStatus  string                 `json:"researchStatus,omitempty"`
	ResearchNotes   string                 `json:"researchNotes,omitempty"`
}

// List godoc
// @Summary List explorer research entries
// @Description Get all explorer research entries with optional filters
// @Tags explorers
// @Accept json
// @Produce json
// @Param chain query string false "Filter by chain name"
// @Param status query string false "Filter by research status"
// @Param limit query int false "Maximum results"
// @Success 200 {array} model.ExplorerResearch
// @Router /api/explorers [get]
func (h *ExplorerHandler) List(c *gin.Context) {
	chainName := c.Query("chain")
	status := c.Query("status")
	limit := 0
	if l := c.Query("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	explorers, err := h.explorerRepo.List(chainName, status, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  explorers,
		"count": len(explorers),
	})
}

// Get godoc
// @Summary Get explorer research by ID
// @Description Get a single explorer research entry
// @Tags explorers
// @Accept json
// @Produce json
// @Param id path string true "Explorer ID"
// @Success 200 {object} model.ExplorerResearch
// @Router /api/explorers/{id} [get]
func (h *ExplorerHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	explorer, err := h.explorerRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "explorer not found"})
		return
	}

	c.JSON(http.StatusOK, explorer)
}

// Create godoc
// @Summary Create explorer research entry
// @Description Create a new explorer research entry
// @Tags explorers
// @Accept json
// @Produce json
// @Param body body CreateExplorerRequest true "Explorer data"
// @Success 201 {object} model.ExplorerResearch
// @Router /api/explorers [post]
func (h *ExplorerHandler) Create(c *gin.Context) {
	var req CreateExplorerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if URL already exists
	existing, _ := h.explorerRepo.GetByURL(req.ExplorerURL)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "explorer with this URL already exists"})
		return
	}

	explorer := &model.ExplorerResearch{
		ChainName:       req.ChainName,
		ChainType:       req.ChainType,
		ExplorerName:    req.ExplorerName,
		ExplorerURL:     req.ExplorerURL,
		ExplorerType:    req.ExplorerType,
		Analysis:        req.Analysis,
		Strengths:       req.Strengths,
		Weaknesses:      req.Weaknesses,
		PopularityScore: req.PopularityScore,
		ResearchStatus:  req.ResearchStatus,
		ResearchNotes:   req.ResearchNotes,
		Screenshots:     req.Screenshots,
	}

	if req.Features != nil {
		explorer.Features = datatypes.JSON(mustMarshalJSON(req.Features))
	}
	if req.UIFeatures != nil {
		explorer.UIFeatures = datatypes.JSON(mustMarshalJSON(req.UIFeatures))
	}
	if req.APIFeatures != nil {
		explorer.APIFeatures = datatypes.JSON(mustMarshalJSON(req.APIFeatures))
	}

	if explorer.ResearchStatus == "" {
		explorer.ResearchStatus = "pending"
	}

	if err := h.explorerRepo.Create(explorer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, explorer)
}

// Update godoc
// @Summary Update explorer research entry
// @Description Update an existing explorer research entry
// @Tags explorers
// @Accept json
// @Produce json
// @Param id path string true "Explorer ID"
// @Param body body CreateExplorerRequest true "Explorer data"
// @Success 200 {object} model.ExplorerResearch
// @Router /api/explorers/{id} [put]
func (h *ExplorerHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	explorer, err := h.explorerRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "explorer not found"})
		return
	}

	var req CreateExplorerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	explorer.ChainName = req.ChainName
	explorer.ChainType = req.ChainType
	explorer.ExplorerName = req.ExplorerName
	explorer.ExplorerURL = req.ExplorerURL
	explorer.ExplorerType = req.ExplorerType
	explorer.Analysis = req.Analysis
	explorer.Strengths = req.Strengths
	explorer.Weaknesses = req.Weaknesses
	explorer.PopularityScore = req.PopularityScore
	explorer.ResearchNotes = req.ResearchNotes
	explorer.Screenshots = req.Screenshots

	if req.ResearchStatus != "" {
		explorer.ResearchStatus = req.ResearchStatus
	}
	if req.Features != nil {
		explorer.Features = datatypes.JSON(mustMarshalJSON(req.Features))
	}
	if req.UIFeatures != nil {
		explorer.UIFeatures = datatypes.JSON(mustMarshalJSON(req.UIFeatures))
	}
	if req.APIFeatures != nil {
		explorer.APIFeatures = datatypes.JSON(mustMarshalJSON(req.APIFeatures))
	}

	if err := h.explorerRepo.Update(explorer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, explorer)
}

// Delete godoc
// @Summary Delete explorer research entry
// @Description Delete an explorer research entry
// @Tags explorers
// @Accept json
// @Produce json
// @Param id path string true "Explorer ID"
// @Success 200 {object} map[string]string
// @Router /api/explorers/{id} [delete]
func (h *ExplorerHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.explorerRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetChains godoc
// @Summary Get all chains
// @Description Get list of unique chain names
// @Tags explorers
// @Produce json
// @Success 200 {array} string
// @Router /api/explorers/chains [get]
func (h *ExplorerHandler) GetChains(c *gin.Context) {
	chains, err := h.explorerRepo.GetChains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chains": chains,
		"count":  len(chains),
	})
}

// GetStats godoc
// @Summary Get explorer research statistics
// @Description Get statistics about explorer research
// @Tags explorers
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/explorers/stats [get]
func (h *ExplorerHandler) GetStats(c *gin.Context) {
	stats, err := h.explorerRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetFeatures godoc
// @Summary Get feature checklist
// @Description Get the standard feature checklist for explorers
// @Tags explorers
// @Produce json
// @Param category query string false "Filter by category"
// @Success 200 {array} model.ExplorerFeature
// @Router /api/explorers/features [get]
func (h *ExplorerHandler) GetFeatures(c *gin.Context) {
	category := c.Query("category")

	var features []model.ExplorerFeature
	var err error

	if category != "" {
		features, err = h.featureRepo.GetByCategory(category)
	} else {
		features, err = h.featureRepo.List()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Group by category
	grouped := make(map[string][]model.ExplorerFeature)
	for _, f := range features {
		grouped[f.Category] = append(grouped[f.Category], f)
	}

	c.JSON(http.StatusOK, gin.H{
		"features":   features,
		"byCategory": grouped,
		"categories": model.StandardFeatureCategories,
	})
}

// UpdateStatus godoc
// @Summary Update explorer research status
// @Description Update the research status of an explorer
// @Tags explorers
// @Accept json
// @Produce json
// @Param id path string true "Explorer ID"
// @Param status query string true "New status (pending, in_progress, completed)"
// @Success 200 {object} map[string]string
// @Router /api/explorers/{id}/status [post]
func (h *ExplorerHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending in_progress completed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.explorerRepo.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated", "status": req.Status})
}

// Compare godoc
// @Summary Compare explorers
// @Description Compare multiple explorers side by side
// @Tags explorers
// @Produce json
// @Param ids query string true "Comma-separated explorer IDs"
// @Success 200 {object} map[string]interface{}
// @Router /api/explorers/compare [get]
func (h *ExplorerHandler) Compare(c *gin.Context) {
	idsParam := c.Query("ids")
	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids parameter required"})
		return
	}

	// Parse IDs (comma-separated)
	idStrings := splitAndTrim(idsParam, ",")
	if len(idStrings) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least 2 explorer IDs required for comparison"})
		return
	}
	if len(idStrings) > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 5 explorers can be compared"})
		return
	}

	var explorers []model.ExplorerResearch
	for _, idStr := range idStrings {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		explorer, err := h.explorerRepo.GetByID(id)
		if err != nil {
			continue
		}
		explorers = append(explorers, *explorer)
	}

	// Get features for comparison
	features, _ := h.featureRepo.List()

	c.JSON(http.StatusOK, gin.H{
		"explorers": explorers,
		"features":  features,
		"count":     len(explorers),
	})
}

// SeedFeatures godoc
// @Summary Seed standard features
// @Description Seed the database with standard explorer features
// @Tags explorers
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/explorers/features/seed [post]
func (h *ExplorerHandler) SeedFeatures(c *gin.Context) {
	if err := h.featureRepo.SeedStandardFeatures(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "features seeded successfully"})
}

// Helper functions
func splitAndTrim(s string, sep string) []string {
	parts := make([]string, 0)
	for _, part := range splitString(s, sep) {
		trimmed := trimString(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}

func mustMarshalJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
