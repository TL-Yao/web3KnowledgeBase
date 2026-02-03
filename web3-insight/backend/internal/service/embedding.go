package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// EmbeddingService handles embedding generation and management
type EmbeddingService struct {
	articleRepo *repository.ArticleRepository
	adapter     llm.EmbeddingAdapter
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(articleRepo *repository.ArticleRepository, cfg *config.LLMConfig) *EmbeddingService {
	// Use Ollama for embeddings by default
	adapter := llm.DefaultOllamaEmbeddingAdapter(cfg.OllamaHost)

	return &EmbeddingService{
		articleRepo: articleRepo,
		adapter:     adapter,
	}
}

// NewEmbeddingServiceWithAdapter creates a new embedding service with custom adapter
func NewEmbeddingServiceWithAdapter(articleRepo *repository.ArticleRepository, adapter llm.EmbeddingAdapter) *EmbeddingService {
	return &EmbeddingService{
		articleRepo: articleRepo,
		adapter:     adapter,
	}
}

// GenerateForArticle generates and stores embedding for an article
func (s *EmbeddingService) GenerateForArticle(ctx context.Context, articleID uuid.UUID) error {
	article, err := s.articleRepo.GetByID(articleID)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	return s.GenerateAndStore(ctx, article)
}

// GenerateAndStore generates embedding and stores it in the article
func (s *EmbeddingService) GenerateAndStore(ctx context.Context, article *model.Article) error {
	// Prepare text for embedding: combine title, summary, and content
	text := s.prepareTextForEmbedding(article)

	// Generate embedding
	embedding, err := s.adapter.GenerateEmbedding(text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Store embedding
	vec := llm.Float32ToVector(embedding)
	if err := s.articleRepo.UpdateEmbedding(article.ID, vec); err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	log.Printf("Generated embedding for article: %s (dimensions: %d)", article.Title, len(embedding))
	return nil
}

// prepareTextForEmbedding prepares article text for embedding generation
func (s *EmbeddingService) prepareTextForEmbedding(article *model.Article) string {
	var parts []string

	// Title has high importance
	if article.Title != "" {
		parts = append(parts, "标题: "+article.Title)
	}

	// Summary provides good semantic representation
	if article.Summary != "" {
		parts = append(parts, "摘要: "+article.Summary)
	}

	// Tags add categorical context
	if len(article.Tags) > 0 {
		parts = append(parts, "标签: "+strings.Join(article.Tags, ", "))
	}

	// Content - truncate if too long (embedding models have limits)
	if article.Content != "" {
		content := article.Content
		// Truncate to ~4000 chars to keep within typical embedding model limits
		if len(content) > 4000 {
			content = content[:4000] + "..."
		}
		parts = append(parts, "内容: "+content)
	}

	return strings.Join(parts, "\n\n")
}

// GenerateForMissingArticles generates embeddings for articles without embeddings
func (s *EmbeddingService) GenerateForMissingArticles(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 10
	}

	articles, err := s.articleRepo.FindWithoutEmbeddings(batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to find articles without embeddings: %w", err)
	}

	if len(articles) == 0 {
		return 0, nil
	}

	successCount := 0
	for _, article := range articles {
		if err := s.GenerateAndStore(ctx, &article); err != nil {
			log.Printf("Failed to generate embedding for article %s: %v", article.ID, err)
			continue
		}
		successCount++
	}

	return successCount, nil
}

// GenerateForAllArticles regenerates embeddings for all articles
func (s *EmbeddingService) GenerateForAllArticles(ctx context.Context, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 10
	}

	page := 1
	for {
		articles, _, err := s.articleRepo.ListSimple(page, batchSize, "", nil, "")
		if err != nil {
			return fmt.Errorf("failed to list articles: %w", err)
		}

		if len(articles) == 0 {
			break
		}

		for _, article := range articles {
			if err := s.GenerateAndStore(ctx, &article); err != nil {
				log.Printf("Failed to generate embedding for article %s: %v", article.ID, err)
				continue
			}
		}

		page++
	}

	return nil
}

// IsAvailable checks if the embedding adapter is available
func (s *EmbeddingService) IsAvailable() bool {
	return s.adapter.IsAvailable()
}

// GetDimensions returns the embedding dimensions
func (s *EmbeddingService) GetDimensions() int {
	return s.adapter.Dimensions()
}
