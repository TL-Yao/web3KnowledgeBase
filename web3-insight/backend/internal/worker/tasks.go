package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
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

	// TODO: Implement in Phase 2
	// 1. Fetch RSS feed
	// 2. Parse entries
	// 3. Create news items in database
	// 4. Trigger classification for new items

	return nil
}

// handleWebCrawl handles web crawling tasks
func handleWebCrawl(ctx context.Context, t *asynq.Task) error {
	var payload WebCrawlPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing web crawl task: url=%s, depth=%d", payload.URL, payload.Depth)

	// TODO: Implement in Phase 2
	// 1. Fetch web page
	// 2. Extract content
	// 3. Save to database
	// 4. Trigger classification and embedding

	return nil
}

// handleClassify handles content classification tasks
func handleClassify(ctx context.Context, t *asynq.Task) error {
	var payload ClassifyPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing classification task: articleId=%s", payload.ArticleID)

	// TODO: Implement in Phase 2
	// 1. Fetch article content
	// 2. Use LLM to classify
	// 3. Update article with category

	return nil
}

// handleEmbedding handles embedding generation tasks
func handleEmbedding(ctx context.Context, t *asynq.Task) error {
	var payload EmbeddingPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Processing embedding task: articleId=%s", payload.ArticleID)

	// TODO: Implement in Phase 2
	// 1. Fetch article content
	// 2. Generate embedding using LLM
	// 3. Update article with embedding vector

	return nil
}
