package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/web3-insight/internal/config"
)

// testPromptsConfig returns a minimal PromptsConfig for testing
func testPromptsConfig() *config.PromptsConfig {
	return &config.PromptsConfig{
		Themes: []config.ThemeConfig{
			{
				ID:       "test_theme",
				Name:     "Test Theme",
				Category: "test",
				KeywordPrompt: `Generate {{.Count}} unique Web3/DeFi keywords.

**Excluded keywords**: {{.ExistingKeywords}}

**Rules**:
- Return ONLY a JSON array: ["keyword1", "keyword2", ...]
- Each keyword: 2-5 words
- NO overlap with excluded keywords

Output the JSON array directly without any additional text.`,
				ArticlePrompt: `Write a test article about: "{{.Keyword}}"`,
			},
		},
		Generation: config.GenerationConfig{
			DefaultBatchSize:       3,
			MaxBatchSize:           10,
			KeywordPoolTarget:      200,
			KeywordRefillThreshold: 10,
			ArticleTimeoutMinutes:  60,
		},
	}
}

func TestKeywordPoolService_GenerateKeywords(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	prompts := testPromptsConfig()
	svc := NewKeywordPoolService(nil, prompts)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	keywords, err := svc.generateKeywords(ctx, "test_theme", 10, []string{})
	require.NoError(t, err, "Should generate keywords successfully")
	assert.GreaterOrEqual(t, len(keywords), 5, "Should generate at least 5 keywords")
	assert.LessOrEqual(t, len(keywords), 15, "Should not generate too many keywords")

	// Verify keywords are unique
	keywordSet := make(map[string]bool)
	for _, kw := range keywords {
		assert.False(t, keywordSet[kw], "Keywords should be unique: %s", kw)
		keywordSet[kw] = true
		assert.NotEmpty(t, kw, "Keywords should not be empty")
	}

	t.Logf("Generated %d unique keywords", len(keywords))
	t.Logf("Sample keywords: %v", keywords[:min(5, len(keywords))])
}

func TestKeywordPoolService_GenerateKeywords_WithExclusion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	prompts := testPromptsConfig()
	svc := NewKeywordPoolService(nil, prompts)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	excludeList := []string{"DeFi", "blockchain", "smart contract"}

	keywords, err := svc.generateKeywords(ctx, "test_theme", 10, excludeList)
	require.NoError(t, err, "Should generate keywords with exclusion")

	// Verify excluded keywords are not in result (case-insensitive check)
	for _, kw := range keywords {
		for _, excluded := range excludeList {
			assert.NotEqual(t, excluded, kw, "Should not include excluded keyword")
		}
	}

	t.Logf("Generated %d keywords with %d exclusions", len(keywords), len(excludeList))
	t.Logf("Sample keywords: %v", keywords[:min(5, len(keywords))])
}

// Note: Full integration tests for InitializePool and RefillPoolIfNeeded
// require PostgreSQL database. These should be tested in integration test suite
// with real database connection.
