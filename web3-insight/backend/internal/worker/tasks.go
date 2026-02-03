package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/user/web3-insight/internal/collector"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

// Task type constants
const (
	TaskTypeContentGenerate = "content:generate"
	TaskTypeRSSSync         = "rss:sync"
	TaskTypeWebCrawl        = "web:crawl"
	TaskTypeClassify        = "content:classify"
	TaskTypeEmbedding       = "content:embedding"
)

// ContentGeneratePayload represents the payload for content generation tasks
type ContentGeneratePayload struct {
	Topic      string `json:"topic"`
	CategoryID string `json:"categoryId"`
	Style      string `json:"style,omitempty"`
}

// RSSSyncPayload represents the payload for RSS sync tasks
type RSSSyncPayload struct {
	FeedURL    string `json:"feedUrl,omitempty"`
	CategoryID string `json:"categoryId,omitempty"`
}

// WebCrawlPayload represents the payload for web crawl tasks
type WebCrawlPayload struct {
	URL        string `json:"url"`
	CategoryID string `json:"categoryId,omitempty"`
	Depth      int    `json:"depth,omitempty"`
}

// ClassifyPayload represents the payload for content classification tasks
type ClassifyPayload struct {
	ArticleID string `json:"articleId"`
}

// EmbeddingPayload represents the payload for embedding generation tasks
type EmbeddingPayload struct {
	ArticleID string `json:"articleId"`
}

// Global variables for dependency injection
var (
	rssCollector     *collector.RSSCollector
	webCrawler       *collector.WebCrawler
	embeddingService *service.EmbeddingService
	classifier       *service.Classifier
	db               *gorm.DB
	llmConfig        *config.LLMConfig
)

// InitWorkerDependencies initializes worker dependencies
func InitWorkerDependencies(database *gorm.DB, cfg *config.LLMConfig) {
	db = database
	llmConfig = cfg

	newsRepo := repository.NewNewsRepository(db)
	dsRepo := repository.NewDataSourceRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	// Initialize LLM router for services that need it
	llmRouter := llm.NewRouterFromConfig(cfg)

	rssCollector = collector.NewRSSCollector(newsRepo, dsRepo)
	webCrawler = collector.NewWebCrawler(newsRepo)
	embeddingService = service.NewEmbeddingService(articleRepo, cfg)
	classifier = service.NewClassifier(llmRouter, articleRepo, categoryRepo)
}

// NewTaskMux creates and configures the task multiplexer
func NewTaskMux() *asynq.ServeMux {
	mux := asynq.NewServeMux()

	// Register all task handlers
	mux.HandleFunc(TaskTypeContentGenerate, handleContentGenerate)
	mux.HandleFunc(TaskTypeRSSSync, handleRSSSync)
	mux.HandleFunc(TaskTypeWebCrawl, handleWebCrawl)
	mux.HandleFunc(TaskTypeClassify, handleClassify)
	mux.HandleFunc(TaskTypeEmbedding, handleEmbedding)

	return mux
}

// NewContentGenerateTask creates a new content generation task
func NewContentGenerateTask(payload ContentGeneratePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TaskTypeContentGenerate, data), nil
}

// NewRSSSyncTask creates a new RSS sync task
func NewRSSSyncTask(payload RSSSyncPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TaskTypeRSSSync, data), nil
}

// NewWebCrawlTask creates a new web crawl task
func NewWebCrawlTask(payload WebCrawlPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TaskTypeWebCrawl, data), nil
}

// NewClassifyTask creates a new classification task
func NewClassifyTask(payload ClassifyPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TaskTypeClassify, data), nil
}

// NewEmbeddingTask creates a new embedding generation task
func NewEmbeddingTask(payload EmbeddingPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TaskTypeEmbedding, data), nil
}

// handleContentGenerate handles content generation tasks
func handleContentGenerate(ctx context.Context, t *asynq.Task) error {
	var payload ContentGeneratePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing content generation task: topic=%s, categoryId=%s", payload.Topic, payload.CategoryID)

	// TODO: Implement in Phase 2
	// 1. Use LLM router to generate content
	// 2. Save to database
	// 3. Trigger embedding generation

	return nil
}

// handleRSSSync handles RSS feed synchronization tasks
func handleRSSSync(ctx context.Context, t *asynq.Task) error {
	var payload RSSSyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing RSS sync task: feedUrl=%s", payload.FeedURL)

	if rssCollector == nil {
		return fmt.Errorf("RSS collector not initialized")
	}

	// If specific feed URL provided, find the source by URL
	if payload.FeedURL != "" {
		dsRepo := repository.NewDataSourceRepository(db)
		sources, err := dsRepo.FindByType(model.DataSourceTypeRSS)
		if err != nil {
			return fmt.Errorf("failed to find RSS sources: %w", err)
		}

		for _, source := range sources {
			if source.URL == payload.FeedURL {
				_, err := rssCollector.Collect(ctx, source.ID)
				return err
			}
		}
		return fmt.Errorf("RSS source not found for URL: %s", payload.FeedURL)
	}

	// If no specific URL, sync all enabled RSS sources
	_, err := rssCollector.CollectAll(ctx)
	return err
}

// handleWebCrawl handles web crawling tasks
func handleWebCrawl(ctx context.Context, t *asynq.Task) error {
	var payload WebCrawlPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing web crawl task: url=%s, depth=%d", payload.URL, payload.Depth)

	if webCrawler == nil {
		return fmt.Errorf("web crawler not initialized")
	}

	// Crawl and save
	_, err := webCrawler.CrawlAndSave(ctx, payload.URL, "manual")
	if err != nil {
		return fmt.Errorf("crawl failed: %w", err)
	}

	return nil
}

// handleClassify handles content classification tasks
func handleClassify(ctx context.Context, t *asynq.Task) error {
	var payload ClassifyPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing classification task: articleId=%s", payload.ArticleID)

	if classifier == nil {
		return fmt.Errorf("classifier not initialized")
	}

	articleID, err := uuid.Parse(payload.ArticleID)
	if err != nil {
		return fmt.Errorf("invalid article ID: %w", err)
	}

	if err := classifier.ClassifyAndUpdate(ctx, articleID); err != nil {
		return fmt.Errorf("classification failed: %w", err)
	}

	log.Printf("Classification completed for article: %s", payload.ArticleID)
	return nil
}

// handleEmbedding handles embedding generation tasks
func handleEmbedding(ctx context.Context, t *asynq.Task) error {
	var payload EmbeddingPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing embedding task: articleId=%s", payload.ArticleID)

	if embeddingService == nil {
		return fmt.Errorf("embedding service not initialized")
	}

	articleID, err := uuid.Parse(payload.ArticleID)
	if err != nil {
		return fmt.Errorf("invalid article ID: %w", err)
	}

	if err := embeddingService.GenerateForArticle(ctx, articleID); err != nil {
		return fmt.Errorf("embedding generation failed: %w", err)
	}

	log.Printf("Embedding generated for article: %s", payload.ArticleID)
	return nil
}
