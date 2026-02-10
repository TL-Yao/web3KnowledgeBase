package service

import (
	"testing"

	"github.com/gosimple/slug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/web3-insight/internal/config"
)

func TestArticleGeneratorService_BuildThemePrompt(t *testing.T) {
	prompts := &config.PromptsConfig{
		Themes: []config.ThemeConfig{
			{
				ID:   "web3_basics",
				Name: "Web3基础知识",
				ArticlePrompt: `Write a comprehensive educational article about: "{{.Keyword}}"

**Target Audience**: Complete beginners.
**Language**: Chinese (中文)
**Technical terms**: Use "English (中文)" format
**Length**: 1500-2500 words

**Research**: Use WebFetch if needed.

===TITLE_START===
Article title in Chinese here
===TITLE_END===`,
			},
		},
	}

	service := &ArticleGeneratorService{prompts: prompts}

	prompt, err := service.buildThemePrompt("DeFi", "web3_basics")
	require.NoError(t, err)

	assert.Contains(t, prompt, "DeFi", "Prompt should contain keyword")
	assert.Contains(t, prompt, "Chinese (中文)", "Prompt should specify Chinese output")
	assert.Contains(t, prompt, "English (中文)", "Prompt should specify term format")
	assert.Contains(t, prompt, "1500-2500 words", "Prompt should specify length")
	assert.Contains(t, prompt, "WebFetch", "Prompt should mention WebFetch")
	assert.Contains(t, prompt, "===TITLE_START===", "Prompt should specify delimiter format")
}

func TestArticleGeneratorService_BuildThemePrompt_Fallback(t *testing.T) {
	// Test fallback when no prompts config
	service := &ArticleGeneratorService{}

	prompt, err := service.buildThemePrompt("DeFi", "nonexistent")
	require.NoError(t, err)

	assert.Contains(t, prompt, "DeFi", "Fallback prompt should contain keyword")
	assert.Contains(t, prompt, "===TITLE_START===", "Fallback should use delimiter format")
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
				Content: string(make([]byte, 600)),
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
	title := "Understanding DeFi Protocols"
	expected := "understanding-defi-protocols"

	result := slug.Make(title)
	assert.Equal(t, expected, result)
}
