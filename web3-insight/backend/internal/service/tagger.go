package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"

	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

const tagPromptTemplate = `你是一个Web3知识库标签专家。请为以下文章从标签库中选择3-7个最相关的标签。

## 标签库
### 主题标签（{{.ThemeName}}）
{{range .ThemeTags}}- {{.Name}}
{{end}}
### 通用标签
{{range .UniversalTags}}- {{.Name}}
{{end}}

## 文章信息
标题：{{.Title}}
摘要：{{.Summary}}

## 输出要求
1. 从标签库中选择3-7个最相关的标签
2. 优先选择能准确描述文章核心内容的标签
3. 如果标签库缺少关键标签，可以在 newTagSuggestions 中建议（最多2个）
4. 返回JSON格式（不要包含markdown代码块标记）：
{"tags": ["标签1", "标签2", ...], "newTagSuggestions": ["建议标签1"]}`

// Pending tag auto-activation threshold
const suggestCountThreshold = 3

// Tagger handles automatic article tagging using the tag registry
type Tagger struct {
	llmRouter   *llm.Router
	tagRepo     *repository.TagRepository
	articleRepo *repository.ArticleRepository
}

// NewTagger creates a new tagger service
func NewTagger(router *llm.Router, tagRepo *repository.TagRepository, articleRepo *repository.ArticleRepository) *Tagger {
	return &Tagger{
		llmRouter:   router,
		tagRepo:     tagRepo,
		articleRepo: articleRepo,
	}
}

// TaggingResult represents the LLM tagging response
type TaggingResult struct {
	Tags              []string `json:"tags"`
	NewTagSuggestions []string `json:"newTagSuggestions"`
}

// tagPromptData holds template variables for the tagging prompt
type tagPromptData struct {
	ThemeName    string
	ThemeTags    []model.Tag
	UniversalTags []model.Tag
	Title        string
	Summary      string
}

// TagArticle tags an article using the tag registry and LLM
func (t *Tagger) TagArticle(ctx context.Context, article *model.Article) error {
	themeID := "web3_basics" // default
	if article.ThemeID != nil && *article.ThemeID != "" {
		themeID = *article.ThemeID
	}

	// Get active tags for tagging
	universalTags, themeTags, err := t.tagRepo.FindActiveForTagging(themeID)
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	if len(universalTags)+len(themeTags) == 0 {
		log.Printf("No active tags found for theme %s, skipping tagging", themeID)
		return nil
	}

	// Build prompt
	prompt, err := t.buildTagPrompt(article, themeID, universalTags, themeTags)
	if err != nil {
		return fmt.Errorf("failed to build tag prompt: %w", err)
	}

	// Call LLM
	response, modelUsed, err := t.llmRouter.Generate(llm.TaskTagging, prompt, &llm.GenerateOptions{
		Temperature: 0.2,
		MaxTokens:   300,
	})
	if err != nil {
		return fmt.Errorf("LLM tagging failed: %w", err)
	}

	// Parse response
	result, err := parseTaggingResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse tagging response (model: %s): %w", modelUsed, err)
	}

	// Build valid tag set from registry
	validTags := make(map[string]bool, len(universalTags)+len(themeTags))
	for _, tag := range universalTags {
		validTags[tag.Name] = true
	}
	for _, tag := range themeTags {
		validTags[tag.Name] = true
	}

	// Filter: only keep tags that exist in the registry
	var validatedTags []string
	for _, tagName := range result.Tags {
		tagName = strings.TrimSpace(tagName)
		if validTags[tagName] {
			validatedTags = append(validatedTags, tagName)
		} else {
			log.Printf("Tag '%s' not in registry, skipping (article: %s)", tagName, article.Title)
		}
	}

	// Handle new tag suggestions (pending lifecycle)
	for _, suggestion := range result.NewTagSuggestions {
		suggestion = strings.TrimSpace(suggestion)
		if suggestion == "" {
			continue
		}
		t.handleNewTagSuggestion(suggestion, themeID)
	}

	// Update article tags
	if len(validatedTags) > 0 {
		article.Tags = validatedTags
		if err := t.articleRepo.Update(article); err != nil {
			return fmt.Errorf("failed to update article tags: %w", err)
		}
		log.Printf("Tagged article '%s' with %d tags: %v (model: %s)", article.Title, len(validatedTags), validatedTags, modelUsed)
	} else {
		log.Printf("No valid tags found for article '%s' (model: %s)", article.Title, modelUsed)
	}

	return nil
}

// handleNewTagSuggestion processes a new tag suggestion from the LLM
func (t *Tagger) handleNewTagSuggestion(name string, themeID string) {
	existing, err := t.tagRepo.FindByName(name)
	if err == nil {
		// Tag exists — just increment suggest count
		newCount, err := t.tagRepo.IncrementSuggestCount(existing.Name)
		if err != nil {
			log.Printf("Failed to increment suggest count for '%s': %v", name, err)
			return
		}
		// Auto-activate if threshold reached and still pending
		if existing.Status == "pending" && newCount >= suggestCountThreshold {
			if err := t.tagRepo.UpdateStatus(name, "active"); err != nil {
				log.Printf("Failed to auto-activate tag '%s': %v", name, err)
			} else {
				log.Printf("Auto-activated tag '%s' (suggest count: %d)", name, newCount)
			}
		}
		return
	}

	// Tag doesn't exist — create as pending
	tag := model.Tag{
		Name:         name,
		ThemeID:      &themeID,
		Status:       "pending",
		SuggestCount: 1,
	}
	if err := t.tagRepo.Create(&tag); err != nil {
		log.Printf("Failed to create pending tag '%s': %v", name, err)
	} else {
		log.Printf("Created pending tag suggestion: '%s' (theme: %s)", name, themeID)
	}
}

// buildTagPrompt renders the tagging prompt template
func (t *Tagger) buildTagPrompt(article *model.Article, themeID string, universalTags, themeTags []model.Tag) (string, error) {
	tmpl, err := template.New("tagging").Parse(tagPromptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse tag prompt template: %w", err)
	}

	summary := article.Summary
	if summary == "" {
		summary = truncateString(article.Content, 500)
	}

	data := tagPromptData{
		ThemeName:    themeID,
		ThemeTags:    themeTags,
		UniversalTags: universalTags,
		Title:        article.Title,
		Summary:      summary,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render tag prompt: %w", err)
	}

	return buf.String(), nil
}

// parseTaggingResponse extracts the tagging result from LLM response
func parseTaggingResponse(response string) (*TaggingResult, error) {
	response = strings.TrimSpace(response)

	// Remove markdown code blocks
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

	var result TaggingResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w, raw: %s", err, jsonStr)
	}

	return &result, nil
}

// truncateString truncates a string to maxLen runes
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
