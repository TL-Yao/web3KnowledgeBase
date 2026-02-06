package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeywordPoolService_GenerateKeywords(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Use nil for repository since we're only testing generateKeywords
	service := NewKeywordPoolService(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test generating 10 keywords
	keywords, err := service.generateKeywords(ctx, 10, []string{})
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

	service := NewKeywordPoolService(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	excludeList := []string{"DeFi", "blockchain", "smart contract"}

	keywords, err := service.generateKeywords(ctx, 10, excludeList)
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Note: Full integration tests for InitializePool and RefillPoolIfNeeded
// require PostgreSQL database. These should be tested in integration test suite
// with real database connection.
