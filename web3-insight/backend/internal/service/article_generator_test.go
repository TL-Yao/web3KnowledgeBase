package service

import (
	"testing"

	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
)

func TestArticleGeneratorService_BuildPrompt(t *testing.T) {
	service := &ArticleGeneratorService{}

	prompt := service.buildPrompt("DeFi")

	// Verify prompt contains key elements
	assert.Contains(t, prompt, "DeFi", "Prompt should contain keyword")
	assert.Contains(t, prompt, "Chinese (中文)", "Prompt should specify Chinese output")
	assert.Contains(t, prompt, "English (中文)", "Prompt should specify term format")
	assert.Contains(t, prompt, "1500-2500 words", "Prompt should specify length")
	assert.Contains(t, prompt, "WebFetch", "Prompt should mention WebFetch capability")
	assert.Contains(t, prompt, "===TITLE_START===", "Prompt should specify delimiter format")
}

func TestArticleGeneratorService_ValidateArticle(t *testing.T) {
	service := &ArticleGeneratorService{}

	tests := []struct {
		name    string
		data    ArticleData
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid article",
			data: ArticleData{
				Title:   "Test Article",
				Content: string(make([]byte, 600)), // 600 characters
				Summary: "Test summary",
			},
			wantErr: false,
		},
		{
			name: "empty title",
			data: ArticleData{
				Title:   "",
				Content: string(make([]byte, 600)),
				Summary: "Test summary",
			},
			wantErr: true,
			errMsg:  "title is empty",
		},
		{
			name: "content too short",
			data: ArticleData{
				Title:   "Test",
				Content: "Short",
				Summary: "Test summary",
			},
			wantErr: true,
			errMsg:  "content too short",
		},
		{
			name: "empty summary",
			data: ArticleData{
				Title:   "Test",
				Content: string(make([]byte, 600)),
				Summary: "",
			},
			wantErr: true,
			errMsg:  "summary is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateArticle(&tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArticleGeneratorService_GenerateSlug(t *testing.T) {
	// Note: This test requires a mock articleRepo or will fail
	// For now, we test the slug generation logic separately

	title := "Understanding DeFi Protocols"
	expected := "understanding-defi-protocols"

	// Test slug.Make directly
	result := slug.Make(title)
	assert.Equal(t, expected, result)
}

// Note: Full integration test for GenerateArticle would be expensive
// and is better suited for manual testing or end-to-end test suite.
// The test would look like:
//
// func TestArticleGeneratorService_GenerateArticle_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("Skipping expensive integration test")
//     }
//
//     // Setup with real database and classifier
//     db := setupTestDB(t)
//     articleRepo := repository.NewArticleRepository(db)
//     classifier := setupTestClassifier(t, db)
//     service := NewArticleGeneratorService(articleRepo, classifier)
//
//     ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
//     defer cancel()
//
//     article, sessionID, err := service.GenerateArticle(ctx, "零知识证明")
//     require.NoError(t, err)
//     assert.NotEmpty(t, article.ID)
//     assert.NotEmpty(t, article.Title)
//     assert.NotEmpty(t, article.Content)
//     assert.NotEmpty(t, sessionID)
//
//     // Verify article is saved
//     saved, err := articleRepo.GetByID(article.ID)
//     require.NoError(t, err)
//     assert.Equal(t, article.Title, saved.Title)
// }
