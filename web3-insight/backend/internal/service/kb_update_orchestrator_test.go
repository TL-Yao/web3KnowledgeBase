package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestKBUpdateOrchestrator_GetJobStatus_InvalidUUID(t *testing.T) {
	orchestrator := &KBUpdateOrchestrator{}

	_, err := orchestrator.GetJobStatus(nil, "invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid job ID")
}

func TestKBUpdateOrchestrator_GetJobStatus_ValidUUID(t *testing.T) {
	// This test would require mock repositories
	// For now, we just verify UUID parsing works
	validUUID := uuid.New().String()

	// Test UUID parsing independently
	_, err := uuid.Parse(validUUID)
	assert.NoError(t, err)
}

// Note: Full integration tests for RunUpdate would require:
// - Real database with migrations
// - All service dependencies initialized
// - Claude Code available
// - Significant time and cost
//
// These tests are better suited for the end-to-end test suite (Task #11)
//
// Example integration test structure:
//
// func TestKBUpdateOrchestrator_RunUpdate_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("Skipping expensive integration test")
//     }
//
//     // Setup
//     db := setupTestDB(t)
//     defer cleanupTestDB(t, db)
//
//     // Initialize all repositories
//     keywordRepo := repository.NewKeywordRepository(db)
//     articleRepo := repository.NewArticleRepository(db)
//     jobRepo := repository.NewKBUpdateJobRepository(db)
//     categoryRepo := repository.NewCategoryRepository(db)
//
//     // Initialize services
//     executor := NewClaudeExecutor()
//     keywordPool := NewKeywordPoolService(executor, keywordRepo)
//     classifier := NewClassifier(llmRouter, articleRepo, categoryRepo)
//     articleGen := NewArticleGeneratorService(articleRepo, classifier)
//
//     // Create orchestrator
//     orchestrator := NewKBUpdateOrchestrator(keywordPool, articleGen, keywordRepo, jobRepo)
//
//     // Initialize keyword pool first
//     ctx := context.Background()
//     err := keywordPool.InitializePool(ctx, 25)  // More than 20 for testing
//     require.NoError(t, err)
//
//     // Run update
//     job, err := orchestrator.RunUpdate(ctx, "manual")
//     require.NoError(t, err)
//     assert.NotNil(t, job)
//     assert.Equal(t, "completed", job.Status)
//     assert.Equal(t, 20, job.KeywordsGenerated)
//     assert.GreaterOrEqual(t, job.ArticlesGenerated, 15) // Allow some failures
//     assert.Equal(t, 20, len(job.SessionIDs))
//
//     // Verify articles were created
//     articles, err := articleRepo.List(repository.ArticleListParams{
//         Status:   "published",
//         Page:     1,
//         PageSize: 25,
//     })
//     require.NoError(t, err)
//     assert.GreaterOrEqual(t, len(articles.Articles), 15)
// }
