package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

type SearchHandler struct {
	articleRepo  *repository.ArticleRepository
	categoryRepo *repository.CategoryRepository
}

func NewSearchHandler(articleRepo *repository.ArticleRepository, categoryRepo *repository.CategoryRepository) *SearchHandler {
	return &SearchHandler{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
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
