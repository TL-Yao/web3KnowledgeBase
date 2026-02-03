package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
)

type SearchHandler struct {
	articleRepo    *repository.ArticleRepository
	categoryRepo   *repository.CategoryRepository
	semanticSearch *service.SemanticSearchService
}

func NewSearchHandler(articleRepo *repository.ArticleRepository, categoryRepo *repository.CategoryRepository) *SearchHandler {
	return &SearchHandler{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
	}
}

// NewSearchHandlerWithSemantic creates a search handler with semantic search support
func NewSearchHandlerWithSemantic(articleRepo *repository.ArticleRepository, categoryRepo *repository.CategoryRepository, semanticSearch *service.SemanticSearchService) *SearchHandler {
	return &SearchHandler{
		articleRepo:    articleRepo,
		categoryRepo:   categoryRepo,
		semanticSearch: semanticSearch,
	}
}

type SearchResult struct {
	Articles   []model.Article  `json:"articles"`
	Categories []model.Category `json:"categories"`
	TotalHits  int              `json:"totalHits"`
}

// Search godoc
// @Summary Search across articles and categories
// @Description Full-text search across articles and categories
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Maximum results (default: 20)"
// @Param type query string false "Filter by type (articles, categories)"
// @Success 200 {object} SearchResult
// @Router /api/search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Cap maximum limit to prevent DoS
	if limit > 100 {
		limit = 100
	}

	searchType := c.Query("type")

	result := SearchResult{
		Articles:   []model.Article{},
		Categories: []model.Category{},
	}

	// Search articles using database ILIKE
	if searchType == "" || searchType == "articles" {
		articles, err := h.articleRepo.Search(query, limit)
		if err == nil {
			result.Articles = articles
		}
	}

	// Search categories using database ILIKE
	if searchType == "" || searchType == "categories" {
		categories, err := h.categoryRepo.Search(query, limit)
		if err == nil {
			result.Categories = categories
		}
	}

	result.TotalHits = len(result.Articles) + len(result.Categories)

	c.JSON(http.StatusOK, result)
}

// SemanticSearch godoc
// @Summary Semantic search for articles
// @Description Search articles using semantic similarity (vector search)
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Maximum results (default: 10)"
// @Param categoryId query string false "Filter by category ID"
// @Param mode query string false "Search mode: semantic, keyword, or hybrid (default: hybrid)"
// @Success 200 {array} model.Article
// @Router /api/search/semantic [get]
func (h *SearchHandler) SemanticSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if limit > 50 {
		limit = 50
	}

	var categoryID *uuid.UUID
	if catID := c.Query("categoryId"); catID != "" {
		if parsed, err := uuid.Parse(catID); err == nil {
			categoryID = &parsed
		}
	}

	mode := c.DefaultQuery("mode", "hybrid")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var articles []model.Article
	var err error

	// Check if semantic search is available
	if h.semanticSearch == nil || !h.semanticSearch.IsAvailable() {
		// Fall back to keyword search
		articles, err = h.articleRepo.Search(query, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"articles": articles,
			"mode":     "keyword",
			"fallback": true,
		})
		return
	}

	switch mode {
	case "semantic":
		articles, err = h.semanticSearch.Search(ctx, service.SearchRequest{
			Query:      query,
			CategoryID: categoryID,
			Limit:      limit,
		})
	case "keyword":
		articles, err = h.articleRepo.Search(query, limit)
	default: // hybrid
		articles, err = h.semanticSearch.HybridSearch(ctx, query, limit, categoryID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": articles,
		"mode":     mode,
		"count":    len(articles),
	})
}

// RelatedArticles godoc
// @Summary Get related articles
// @Description Find articles related to a given article using vector similarity
// @Tags search
// @Accept json
// @Produce json
// @Param id path string true "Article ID"
// @Param limit query int false "Maximum results (default: 5)"
// @Success 200 {array} model.Article
// @Router /api/articles/{id}/related [get]
func (h *SearchHandler) RelatedArticles(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	limit := 5
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if limit > 20 {
		limit = 20
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Check if semantic search is available
	if h.semanticSearch == nil || !h.semanticSearch.IsAvailable() {
		c.JSON(http.StatusOK, gin.H{
			"articles":  []model.Article{},
			"available": false,
		})
		return
	}

	articles, err := h.semanticSearch.GetRelatedArticles(ctx, id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"articles":  articles,
		"count":     len(articles),
		"available": true,
	})
}
