package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

const (
	KeywordPoolTarget    = 200 // Target pool size
	KeywordPoolThreshold = 30  // Refill threshold
	MaxRetries           = 3   // Max retry attempts for generation
)

type KeywordPoolService struct {
	keywordRepo *repository.KeywordRepository
}

func NewKeywordPoolService(keywordRepo *repository.KeywordRepository) *KeywordPoolService {
	return &KeywordPoolService{
		keywordRepo: keywordRepo,
	}
}

// InitializePool initializes the keyword pool with the specified count
func (kp *KeywordPoolService) InitializePool(ctx context.Context, count int) error {
	// Check if pool already has keywords
	existingCount, err := kp.keywordRepo.CountPendingKeywords()
	if err != nil {
		return fmt.Errorf("failed to check existing keywords: %w", err)
	}

	if existingCount > 0 {
		log.Printf("Keyword pool already has %d pending keywords, skipping initialization", existingCount)
		return nil
	}

	log.Printf("Initializing keyword pool with %d keywords", count)

	// Generate keywords with empty exclusion list
	keywords, err := kp.generateKeywords(ctx, count, []string{})
	if err != nil {
		return fmt.Errorf("failed to generate keywords: %w", err)
	}

	// Convert to model.Keyword slice
	keywordModels := make([]model.Keyword, 0, len(keywords))
	for _, kw := range keywords {
		keywordModels = append(keywordModels, model.Keyword{
			Keyword: kw,
			Status:  "pending",
			Source:  "claude_code",
		})
	}

	// Batch insert
	if err := kp.keywordRepo.BatchCreate(keywordModels); err != nil {
		return fmt.Errorf("failed to save keywords: %w", err)
	}

	log.Printf("Successfully initialized keyword pool with %d keywords", len(keywords))
	return nil
}

// RefillPoolIfNeeded checks and refills the keyword pool if needed
func (kp *KeywordPoolService) RefillPoolIfNeeded(ctx context.Context) error {
	pendingCount, err := kp.keywordRepo.CountPendingKeywords()
	if err != nil {
		return fmt.Errorf("failed to count pending keywords: %w", err)
	}

	if pendingCount >= KeywordPoolThreshold {
		log.Printf("Keyword pool has %d pending keywords (>= %d), no refill needed", pendingCount, KeywordPoolThreshold)
		return nil
	}

	log.Printf("Keyword pool has only %d pending keywords (< %d), refilling...", pendingCount, KeywordPoolThreshold)

	// Get all used keywords for deduplication
	usedKeywords, err := kp.keywordRepo.GetAllUsedKeywords()
	if err != nil {
		return fmt.Errorf("failed to get used keywords: %w", err)
	}

	// Calculate how many keywords to generate
	neededCount := KeywordPoolTarget - int(pendingCount)
	log.Printf("Generating %d new keywords (target: %d, current: %d, exclude: %d)",
		neededCount, KeywordPoolTarget, pendingCount, len(usedKeywords))

	// Generate keywords
	keywords, err := kp.generateKeywords(ctx, neededCount, usedKeywords)
	if err != nil {
		return fmt.Errorf("failed to generate keywords: %w", err)
	}

	// Convert to model.Keyword slice
	keywordModels := make([]model.Keyword, 0, len(keywords))
	for _, kw := range keywords {
		keywordModels = append(keywordModels, model.Keyword{
			Keyword: kw,
			Status:  "pending",
			Source:  "claude_code",
		})
	}

	// Batch insert
	if err := kp.keywordRepo.BatchCreate(keywordModels); err != nil {
		return fmt.Errorf("failed to save keywords: %w", err)
	}

	log.Printf("Successfully refilled keyword pool with %d new keywords", len(keywords))
	return nil
}

// generateKeywords generates keywords using Claude Code
func (kp *KeywordPoolService) generateKeywords(ctx context.Context, count int, excludeKeywords []string) ([]string, error) {
	excludeList := ""
	if len(excludeKeywords) > 0 {
		excludeList = strings.Join(excludeKeywords, ", ")
	} else {
		excludeList = "(none)"
	}

	prompt := fmt.Sprintf(`Generate %d unique Web3/DeFi/Finance keywords for educational articles.

**Topic Categories**:
- Web3 Technology: blockchain, smart contracts, consensus mechanisms, Layer2
- DeFi Protocols: DEX, lending, stablecoins, liquidity mining
- Traditional Finance: derivatives, asset management, risk control
- Industry Knowledge: companies, products, services, standards, historical events

**Excluded Keywords** (MUST avoid):
%s

**Requirements**:
- Return ONLY a JSON array: ["keyword1", "keyword2", ...]
- Each keyword: 2-5 words (Chinese or English)
- Keywords should be specific and educational
- NO duplicates with excluded keywords

Output the JSON array directly without any additional text.`, count, excludeList)

	var keywords []string
	var lastErr error

	// Retry up to MaxRetries times
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		log.Printf("Generating keywords (attempt %d/%d)...", attempt, MaxRetries)

		// Create a new executor for each attempt to avoid session ID conflicts
		executor := NewClaudeExecutor()
		response, err := executor.Execute(ctx, prompt)
		if err != nil {
			lastErr = fmt.Errorf("execution failed: %w", err)
			log.Printf("Attempt %d failed: %v", attempt, lastErr)
			continue
		}

		// Try to parse JSON array
		if err := json.Unmarshal([]byte(response.Result), &keywords); err != nil {
			lastErr = fmt.Errorf("JSON parse failed: %w, output: %s", err, response.Result)
			log.Printf("Attempt %d: invalid JSON format: %v", attempt, lastErr)
			continue
		}

		// Validate keywords
		if len(keywords) == 0 {
			lastErr = fmt.Errorf("no keywords generated")
			log.Printf("Attempt %d: empty keyword list", attempt)
			continue
		}

		// Local deduplication check
		excludeSet := make(map[string]bool)
		for _, kw := range excludeKeywords {
			excludeSet[strings.ToLower(strings.TrimSpace(kw))] = true
		}

		// Filter out excluded keywords
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

		log.Printf("Successfully generated %d keywords (requested: %d)", len(filtered), count)
		return filtered, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", MaxRetries, lastErr)
}
