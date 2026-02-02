package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// SemanticSearchService handles semantic search operations
type SemanticSearchService struct {
	articleRepo *repository.ArticleRepository
	adapter     llm.EmbeddingAdapter
}

// SearchResult represents a semantic search result
type SearchResult struct {
	Article model.Article `json:"article"`
	Score   float64       `json:"score,omitempty"` // Similarity score if available
}

// SearchRequest represents a semantic search request
type SearchRequest struct {
	Query      string     `json:"query"`
	CategoryID *uuid.UUID `json:"categoryId,omitempty"`
	Status     string     `json:"status,omitempty"`
	Limit      int        `json:"limit,omitempty"`
}

// NewSemanticSearchService creates a new semantic search service
func NewSemanticSearchService(articleRepo *repository.ArticleRepository, cfg *config.LLMConfig) *SemanticSearchService {
	adapter := llm.DefaultOllamaEmbeddingAdapter(cfg.OllamaHost)

	return &SemanticSearchService{
		articleRepo: articleRepo,
		adapter:     adapter,
	}
}

// Search performs semantic search using the query text
func (s *SemanticSearchService) Search(ctx context.Context, req SearchRequest) ([]model.Article, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Generate embedding for the query
	embedding, err := s.adapter.GenerateEmbedding(req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	vec := llm.Float32ToVector(embedding)

	// Perform semantic search
	articles, err := s.articleRepo.SemanticSearch(vec, req.Limit, req.CategoryID, req.Status)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	return articles, nil
}

// GetRelatedArticles finds articles related to a given article
func (s *SemanticSearchService) GetRelatedArticles(ctx context.Context, articleID uuid.UUID, limit int) ([]model.Article, error) {
	if limit <= 0 {
		limit = 5
	}

	articles, err := s.articleRepo.FindRelatedArticles(articleID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find related articles: %w", err)
	}

	return articles, nil
}

// HybridSearch combines semantic search with keyword search
func (s *SemanticSearchService) HybridSearch(ctx context.Context, query string, limit int, categoryID *uuid.UUID) ([]model.Article, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if limit <= 0 {
		limit = 10
	}

	// Perform semantic search
	semanticResults, err := s.Search(ctx, SearchRequest{
		Query:      query,
		CategoryID: categoryID,
		Limit:      limit,
	})
	if err != nil {
		// Fall back to keyword search if semantic search fails
		return s.articleRepo.Search(query, limit)
	}

	// If semantic search returns few results, supplement with keyword search
	if len(semanticResults) < limit {
		keywordResults, _ := s.articleRepo.Search(query, limit-len(semanticResults))

		// Deduplicate results
		resultMap := make(map[uuid.UUID]model.Article)
		for _, a := range semanticResults {
			resultMap[a.ID] = a
		}
		for _, a := range keywordResults {
			if _, exists := resultMap[a.ID]; !exists {
				semanticResults = append(semanticResults, a)
			}
		}
	}

	return semanticResults, nil
}

// IsAvailable checks if the semantic search service is available
func (s *SemanticSearchService) IsAvailable() bool {
	return s.adapter.IsAvailable()
}
