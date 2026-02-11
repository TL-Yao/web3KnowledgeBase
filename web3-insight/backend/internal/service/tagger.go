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

// TagPromptTemplate is the default tagging prompt template used by the tagger.
const TagPromptTemplate = `你是Web3知识库标签专家。请从标签列表中为文章选择最合适的标签。

## 核心规则
1. 选择恰好4-5个标签
2. 从专题标签中选至少2个（优先级最高）
3. 通用标签仅选与文章核心主题直接相关的（最多2个）
4. 必须从列表中逐字复制，禁止创造新标签
5. 仅选文章核心主题对应的标签，顺带提及的不选

## 判断标准
- "核心主题" = 文章专门讨论、解释或分析的概念
- "顺带提及" = 仅作为背景、类比或运行环境提到
- 例：闪电贷文章 → 不选"以太坊"（仅是运行平台）
- 例：ETF入门 → 不选"DeFi"（仅对比提及）

## 专题标签（至少选2个）
{{range .ThemeTags}}- {{.Name}}
{{end}}

## 通用标签（最多选2个，须与核心高度相关）
{{range .UniversalTags}}- {{.Name}}
{{end}}

## 文章
标题：{{.Title}}
摘要：{{.Summary}}
{{if .ContentExcerpt}}正文节选：{{.ContentExcerpt}}{{end}}

输出JSON（不加代码块）：
{"tags": ["标签1", "标签2", "标签3", "标签4"]}`

// Tagger handles automatic article tagging using the tag registry
type Tagger struct {
	llmRouter   *llm.Router
	tagRepo     *repository.TagRepository
	articleRepo *repository.ArticleRepository
	configRepo  *repository.ConfigRepository
}

// NewTagger creates a new tagger service
func NewTagger(router *llm.Router, tagRepo *repository.TagRepository, articleRepo *repository.ArticleRepository, configRepo *repository.ConfigRepository) *Tagger {
	return &Tagger{
		llmRouter:   router,
		tagRepo:     tagRepo,
		articleRepo: articleRepo,
		configRepo:  configRepo,
	}
}

// TaggingResult represents the LLM tagging response
type TaggingResult struct {
	Tags              []string `json:"tags"`
	NewTagSuggestions []string `json:"newTagSuggestions"`
}

// TagPromptData holds template variables for the tagging prompt
type TagPromptData struct {
	ThemeName      string
	ThemeTags      []model.Tag
	UniversalTags  []model.Tag
	Title          string
	Summary        string
	ContentExcerpt string
}

// TagArticle tags an article using the tag registry and LLM
func (t *Tagger) TagArticle(ctx context.Context, article *model.Article) error {
	// Check if auto-tagging is enabled via config
	if t.configRepo != nil {
		cfg, err := t.configRepo.Get("auto_tagging_enabled")
		if err == nil && string(cfg.Value) == `"false"` {
			log.Printf("Auto-tagging disabled via config, skipping article '%s'", article.Title)
			return nil
		}
	}

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
	result, err := ParseTaggingResponse(response)
	if err != nil {
		return fmt.Errorf("failed to parse tagging response (model: %s): %w", modelUsed, err)
	}

	// Build valid tag lookup: lowercased name/name_en -> canonical name
	canonicalTag := make(map[string]string, (len(universalTags)+len(themeTags))*2)
	for _, tag := range universalTags {
		canonicalTag[strings.ToLower(tag.Name)] = tag.Name
		if tag.NameEn != "" {
			canonicalTag[strings.ToLower(tag.NameEn)] = tag.Name
		}
	}
	for _, tag := range themeTags {
		canonicalTag[strings.ToLower(tag.Name)] = tag.Name
		if tag.NameEn != "" {
			canonicalTag[strings.ToLower(tag.NameEn)] = tag.Name
		}
	}

	// Filter: only keep tags that exist in the registry (case-insensitive, name_en also accepted)
	seen := make(map[string]bool)
	var validatedTags []string
	for _, tagName := range result.Tags {
		tagName = strings.TrimSpace(tagName)
		canonical := ResolveTag(tagName, canonicalTag)
		if canonical != "" {
			if !seen[canonical] {
				validatedTags = append(validatedTags, canonical)
				seen[canonical] = true
			}
		} else {
			log.Printf("Tag '%s' not in registry, skipping (article: %s)", tagName, article.Title)
		}
	}

	// Fallback: if too few tags after validation, supplement with keyword-matching universal tags
	if len(validatedTags) < 3 {
		titleLower := strings.ToLower(article.Title + " " + article.Summary)
		for _, tag := range universalTags {
			if seen[tag.Name] {
				continue
			}
			nameLower := strings.ToLower(tag.Name)
			nameEnLower := strings.ToLower(tag.NameEn)
			if strings.Contains(titleLower, nameLower) || (nameEnLower != "" && strings.Contains(titleLower, nameEnLower)) {
				validatedTags = append(validatedTags, tag.Name)
				seen[tag.Name] = true
				log.Printf("Auto-supplemented tag '%s' for article '%s' (fallback)", tag.Name, article.Title)
				if len(validatedTags) >= 4 {
					break
				}
			}
		}
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

// ResolveTag attempts to match a tag name against the canonical registry.
// Handles: exact match, case-insensitive, parenthetical suffix stripping (e.g. "AMM (AMM)" -> "AMM").
func ResolveTag(tagName string, canonicalTag map[string]string) string {
	lower := strings.ToLower(tagName)

	// Direct match
	if c, ok := canonicalTag[lower]; ok {
		return c
	}

	// Strip parenthetical suffix: "流动性池 (Liquidity Pool)" -> "流动性池"
	if idx := strings.Index(tagName, "("); idx > 0 {
		stripped := strings.TrimSpace(tagName[:idx])
		if c, ok := canonicalTag[strings.ToLower(stripped)]; ok {
			return c
		}
		// Also try the content inside parentheses: "(Liquidity Pool)" -> "Liquidity Pool"
		inner := strings.TrimRight(tagName[idx+1:], ") ")
		if c, ok := canonicalTag[strings.ToLower(inner)]; ok {
			return c
		}
	}

	return ""
}

// buildTagPrompt renders the tagging prompt template
func (t *Tagger) buildTagPrompt(article *model.Article, themeID string, universalTags, themeTags []model.Tag) (string, error) {
	return BuildTagPromptFromTemplate(TagPromptTemplate, article, themeID, universalTags, themeTags)
}

// BuildTagPromptFromTemplate renders a tagging prompt using a custom template string.
// This allows benchmark methods to use different prompt templates while reusing the same data logic.
func BuildTagPromptFromTemplate(templateStr string, article *model.Article, themeID string, universalTags, themeTags []model.Tag) (string, error) {
	tmpl, err := template.New("tagging").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse tag prompt template: %w", err)
	}

	summary := article.Summary
	if summary == "" {
		summary = truncateString(article.Content, 2000)
	}

	contentExcerpt := ""
	if article.Content != "" {
		contentExcerpt = truncateString(article.Content, 2000)
	}

	data := TagPromptData{
		ThemeName:      themeID,
		ThemeTags:      themeTags,
		UniversalTags:  universalTags,
		Title:          article.Title,
		Summary:        summary,
		ContentExcerpt: contentExcerpt,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render tag prompt: %w", err)
	}

	return buf.String(), nil
}

// ParseTaggingResponse extracts the tagging result from LLM response
func ParseTaggingResponse(response string) (*TaggingResult, error) {
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
