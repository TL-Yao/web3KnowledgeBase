package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

const (
	MaxRetries       = 3   // Max retry attempts for generation
	KeywordBatchSize = 400 // Max keywords per LLM call for auto-batching
)

type KeywordPoolService struct {
	keywordRepo *repository.KeywordRepository
	prompts     *config.PromptsConfig
}

func NewKeywordPoolService(keywordRepo *repository.KeywordRepository, prompts *config.PromptsConfig) *KeywordPoolService {
	return &KeywordPoolService{
		keywordRepo: keywordRepo,
		prompts:     prompts,
	}
}

// InitializePool initializes the keyword pool for the given theme
func (kp *KeywordPoolService) InitializePool(ctx context.Context, themeID string, count int) error {
	existingCount, err := kp.keywordRepo.CountPendingByTheme(themeID)
	if err != nil {
		return fmt.Errorf("failed to check existing keywords: %w", err)
	}

	if existingCount > 0 {
		log.Printf("Theme %s already has %d pending keywords, skipping initialization", themeID, existingCount)
		return nil
	}

	log.Printf("Initializing keyword pool for theme %s with %d keywords", themeID, count)

	keywords, err := kp.generateKeywords(ctx, themeID, count, []string{})
	if err != nil {
		return fmt.Errorf("failed to generate keywords: %w", err)
	}

	keywordModels := make([]model.Keyword, 0, len(keywords))
	for i, kw := range keywords {
		keywordModels = append(keywordModels, model.Keyword{
			Keyword:   kw,
			Status:    "pending",
			Source:    "claude_code",
			ThemeID:   themeID,
			SortOrder: i,
		})
	}

	if err := kp.keywordRepo.BatchCreate(keywordModels); err != nil {
		return fmt.Errorf("failed to save keywords: %w", err)
	}

	log.Printf("Successfully initialized keyword pool for theme %s with %d keywords", themeID, len(keywords))
	return nil
}

// RefillPoolIfNeeded checks and refills the keyword pool for the given theme
func (kp *KeywordPoolService) RefillPoolIfNeeded(ctx context.Context, themeID string) error {
	threshold := kp.prompts.Generation.KeywordRefillThreshold
	target := kp.prompts.Generation.KeywordPoolTarget

	pendingCount, err := kp.keywordRepo.CountPendingByTheme(themeID)
	if err != nil {
		return fmt.Errorf("failed to count pending keywords: %w", err)
	}

	if int(pendingCount) >= threshold {
		log.Printf("Theme %s has %d pending keywords (>= %d), no refill needed", themeID, pendingCount, threshold)
		return nil
	}

	log.Printf("Theme %s has only %d pending keywords (< %d), refilling...", themeID, pendingCount, threshold)

	// Get all existing keywords for this theme (for dedup)
	existingKeywords, err := kp.keywordRepo.GetAllKeywordsForTheme(themeID)
	if err != nil {
		return fmt.Errorf("failed to get existing keywords: %w", err)
	}

	neededCount := target - int(pendingCount)
	log.Printf("Generating %d new keywords for theme %s (target: %d, current: %d, exclude: %d)",
		neededCount, themeID, target, pendingCount, len(existingKeywords))

	keywords, err := kp.generateKeywords(ctx, themeID, neededCount, existingKeywords)
	if err != nil {
		return fmt.Errorf("failed to generate keywords: %w", err)
	}

	// Get current max sort_order for this theme to continue ordering
	currentMax := len(existingKeywords)

	keywordModels := make([]model.Keyword, 0, len(keywords))
	for i, kw := range keywords {
		keywordModels = append(keywordModels, model.Keyword{
			Keyword:   kw,
			Status:    "pending",
			Source:    "claude_code",
			ThemeID:   themeID,
			SortOrder: currentMax + i,
		})
	}

	if err := kp.keywordRepo.BatchCreate(keywordModels); err != nil {
		return fmt.Errorf("failed to save keywords: %w", err)
	}

	log.Printf("Successfully refilled keyword pool for theme %s with %d new keywords", themeID, len(keywords))
	return nil
}

// generateKeywords generates keywords using Claude Code with auto-batching for large counts
func (kp *KeywordPoolService) generateKeywords(ctx context.Context, themeID string, count int, excludeKeywords []string) ([]string, error) {
	theme, err := kp.prompts.GetThemeByID(themeID)
	if err != nil {
		return nil, fmt.Errorf("unknown theme: %w", err)
	}

	// Auto-batching: if count > KeywordBatchSize, generate in batches
	var allKeywords []string
	batchNum := 1

	for len(allKeywords) < count {
		remaining := count - len(allKeywords)
		batchCount := remaining
		if batchCount > KeywordBatchSize {
			batchCount = KeywordBatchSize
		}

		// Combine exclude list with already-generated keywords to avoid self-duplicates
		currentExclude := make([]string, 0, len(excludeKeywords)+len(allKeywords))
		currentExclude = append(currentExclude, excludeKeywords...)
		currentExclude = append(currentExclude, allKeywords...)

		batch, err := kp.generateKeywordBatch(ctx, theme, batchCount, currentExclude, batchNum)
		if err != nil {
			if len(allKeywords) > 0 {
				// Return what we have so far
				log.Printf("Batch %d failed, returning %d keywords generated so far: %v", batchNum, len(allKeywords), err)
				break
			}
			return nil, err
		}

		allKeywords = append(allKeywords, batch...)
		batchNum++

		// Diminishing returns check
		if len(batch) < batchCount/2 {
			log.Printf("Diminishing returns: got %d/%d keywords in batch %d, stopping", len(batch), batchCount, batchNum-1)
			break
		}
	}

	return allKeywords, nil
}

// generateKeywordBatch generates a single batch of keywords
func (kp *KeywordPoolService) generateKeywordBatch(ctx context.Context, theme *config.ThemeConfig, count int, excludeKeywords []string, batchNumber int) ([]string, error) {
	prompt, err := kp.renderKeywordPrompt(theme, count, excludeKeywords)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	var keywords []string
	var lastErr error

	for attempt := 1; attempt <= MaxRetries; attempt++ {
		log.Printf("Generating keywords for theme %s batch %d (attempt %d/%d)...", theme.ID, batchNumber, attempt, MaxRetries)

		executor := NewClaudeExecutor()
		response, err := executor.Execute(ctx, prompt)
		if err != nil {
			lastErr = fmt.Errorf("execution failed: %w", err)
			log.Printf("Attempt %d failed: %v", attempt, lastErr)
			continue
		}

		cleaned := extractJSON(response.Result)
		if err := json.Unmarshal([]byte(cleaned), &keywords); err != nil {
			lastErr = fmt.Errorf("JSON parse failed: %w, output: %s", err, response.Result)
			log.Printf("Attempt %d: invalid JSON format: %v", attempt, lastErr)
			continue
		}

		if len(keywords) == 0 {
			lastErr = fmt.Errorf("no keywords generated")
			log.Printf("Attempt %d: empty keyword list", attempt)
			continue
		}

		// Dedup filter
		excludeSet := make(map[string]bool, len(excludeKeywords))
		for _, kw := range excludeKeywords {
			excludeSet[strings.ToLower(strings.TrimSpace(kw))] = true
		}

		filtered := make([]string, 0, len(keywords))
		for _, kw := range keywords {
			normalized := strings.ToLower(strings.TrimSpace(kw))
			if !excludeSet[normalized] && kw != "" {
				filtered = append(filtered, kw)
			}
		}

		if len(filtered) < count/2 {
			lastErr = fmt.Errorf("too many duplicates: got %d valid keywords out of %d requested", len(filtered), count)
			log.Printf("Attempt %d: %v", attempt, lastErr)
			continue
		}

		log.Printf("Successfully generated %d keywords for theme %s (requested: %d)", len(filtered), theme.ID, count)
		return filtered, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", MaxRetries, lastErr)
}

// extractJSON strips markdown code block wrappers and leading/trailing text from LLM output
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// Strip ```json ... ``` or ``` ... ``` wrappers
	if idx := strings.Index(s, "```"); idx != -1 {
		s = s[idx+3:]
		// Remove optional language tag (e.g., "json")
		if nl := strings.Index(s, "\n"); nl != -1 && nl < 20 {
			s = s[nl+1:]
		}
		if end := strings.LastIndex(s, "```"); end != -1 {
			s = s[:end]
		}
	}
	s = strings.TrimSpace(s)
	// Find the first [ and last ] for array extraction
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start != -1 && end != -1 && end > start {
		s = s[start : end+1]
	}
	return s
}

// renderKeywordPrompt renders the theme's keyword_prompt template with variables
func (kp *KeywordPoolService) renderKeywordPrompt(theme *config.ThemeConfig, count int, excludeKeywords []string) (string, error) {
	tmpl, err := template.New("keyword").Parse(theme.KeywordPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse keyword prompt template: %w", err)
	}

	excludeList := "(none)"
	if len(excludeKeywords) > 0 {
		excludeList = strings.Join(excludeKeywords, ", ")
	}

	data := map[string]interface{}{
		"Count":            count,
		"ExistingKeywords": excludeList,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render keyword prompt: %w", err)
	}

	return buf.String(), nil
}
