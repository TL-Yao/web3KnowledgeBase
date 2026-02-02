package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// Summarizer handles news summarization and translation
type Summarizer struct {
	llmRouter *llm.Router
	newsRepo  *repository.NewsRepository
}

// NewSummarizer creates a new summarizer service
func NewSummarizer(router *llm.Router, newsRepo *repository.NewsRepository) *Summarizer {
	return &Summarizer{
		llmRouter: router,
		newsRepo:  newsRepo,
	}
}

// SummaryResult represents the parsed LLM response
type SummaryResult struct {
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

// SummarizeNews generates a Chinese summary for a news item
func (s *Summarizer) SummarizeNews(ctx context.Context, item *model.NewsItem) (*SummaryResult, string, error) {
	// Build prompt
	prompt := fmt.Sprintf(PromptNewsSummary, item.OriginalTitle, item.Content)

	// Call LLM
	response, modelUsed, err := s.llmRouter.Generate(llm.TaskSummarization, prompt, &llm.GenerateOptions{
		Temperature: 0.3, // Lower temperature for more consistent output
		MaxTokens:   1000,
	})
	if err != nil {
		return nil, "", fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse JSON response
	result, err := s.parseSummaryResponse(response)
	if err != nil {
		log.Printf("Failed to parse LLM response, using raw output: %v", err)
		// Fallback: use the raw response as summary
		result = &SummaryResult{
			Title:    item.OriginalTitle,
			Summary:  response,
			Category: "tech",
			Tags:     []string{},
		}
	}

	return result, modelUsed, nil
}

// parseSummaryResponse extracts JSON from LLM response
func (s *Summarizer) parseSummaryResponse(response string) (*SummaryResult, error) {
	// Try to find JSON in the response
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	response = regexp.MustCompile("(?s)```json\\s*").ReplaceAllString(response, "")
	response = regexp.MustCompile("(?s)```\\s*").ReplaceAllString(response, "")
	response = strings.TrimSpace(response)

	// Find JSON object
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	jsonStr := response[start : end+1]

	var result SummaryResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// ProcessUnprocessedNews processes all unprocessed news items
func (s *Summarizer) ProcessUnprocessedNews(ctx context.Context, batchSize int) (int, error) {
	items, err := s.newsRepo.FindUnprocessed(batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to find unprocessed items: %w", err)
	}

	processed := 0
	for _, item := range items {
		select {
		case <-ctx.Done():
			return processed, ctx.Err()
		default:
		}

		result, modelUsed, err := s.SummarizeNews(ctx, &item)
		if err != nil {
			log.Printf("Failed to summarize news %s: %v", item.ID, err)
			continue
		}

		// Update the news item
		err = s.newsRepo.UpdateSummary(item.ID, result.Summary, result.Category, result.Tags)
		if err != nil {
			log.Printf("Failed to update news %s: %v", item.ID, err)
			continue
		}

		// Also update the title if we got a translated one
		if result.Title != "" && result.Title != item.OriginalTitle {
			item.Title = result.Title
		}

		log.Printf("Summarized news: %s (model: %s)", item.Title, modelUsed)
		processed++
	}

	return processed, nil
}

// SummarizeByID summarizes a specific news item by ID
func (s *Summarizer) SummarizeByID(ctx context.Context, id uuid.UUID) error {
	item, err := s.newsRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("news item not found: %w", err)
	}

	result, _, err := s.SummarizeNews(ctx, item)
	if err != nil {
		return err
	}

	return s.newsRepo.UpdateSummary(id, result.Summary, result.Category, result.Tags)
}
