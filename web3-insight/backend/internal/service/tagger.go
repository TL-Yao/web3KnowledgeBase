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

const tagPromptTemplate = `你是一个Web3知识库标签专家。请为以下文章选择最合适的标签。

## 选标规则（重要）
1. 仅当文章的**核心主题**是关于某标签时才选择，**不要**仅因文章提及就选择
2. **必须**从专题标签中选择至少2个（这些是最精准的标签）
3. 通用标签最多选3个，仅选择与文章核心高度相关的
4. 每篇文章必须选择至少4个标签，确保充分描述文章内容
5. **禁止自创标签**，必须从下方标签列表中逐字复制标签名。不在列表中的词（如PoW、宏观经济、算法稳定币等）不可使用
6. 如果专题标签不够4个，请用通用标签补齐

## 注意区分（以下情况不应选择该标签）
- 文章讲"闪电贷原理" → 不应选"以太坊"（以太坊只是运行平台，不是文章主题）
- 文章讲"ETF入门" → 不应选"DeFi"（DeFi只是顺带提及的对比）
- 文章讲"Terra崩盘" → 不应选"以太坊"（Terra不是以太坊生态）

## 正确示例
- 文章讨论"以太坊2.0的PoS共识机制升级" → 应选"以太坊"、"共识机制"（都是核心主题）
- 文章介绍"Uniswap V3的集中流动性机制" → 应选"AMM"、"DEX"、"集中流动性"、"流动性池"

## 专题标签 —— 优先从这里选择
{{range .ThemeTags}}- {{.Name}}
{{end}}

## 通用标签 —— 仅当与文章核心高度相关时选择（最多3个）
{{range .UniversalTags}}- {{.Name}}
{{end}}

## 文章信息
标题：{{.Title}}
摘要：{{.Summary}}
{{if .ContentExcerpt}}
正文节选：
{{.ContentExcerpt}}
{{end}}

## 输出要求
选择4-6个标签，返回JSON格式（不要包含markdown代码块标记）：
{"tags": ["标签1", "标签2", "标签3", "标签4"]}`

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
	ThemeName      string
	ThemeTags      []model.Tag
	UniversalTags  []model.Tag
	Title          string
	Summary        string
	ContentExcerpt string
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
		canonical := resolveTag(tagName, canonicalTag)
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

// resolveTag attempts to match a tag name against the canonical registry.
// Handles: exact match, case-insensitive, parenthetical suffix stripping (e.g. "AMM (AMM)" -> "AMM").
func resolveTag(tagName string, canonicalTag map[string]string) string {
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
	tmpl, err := template.New("tagging").Parse(tagPromptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse tag prompt template: %w", err)
	}

	summary := article.Summary
	if summary == "" {
		summary = truncateString(article.Content, 2000)
	}

	// Include content excerpt for better context (first 2000 chars of content)
	contentExcerpt := ""
	if article.Content != "" {
		contentExcerpt = truncateString(article.Content, 2000)
	}

	data := tagPromptData{
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
