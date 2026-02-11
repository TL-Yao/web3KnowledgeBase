package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
)

const articleUpdateSystemPrompt = `你是Web3知识库的文章编辑专家。你的任务是根据用户与AI助手的对话，对现有文章进行定向补充和优化。

## 核心原则

1. **定向补充，不重写**：只修改或添加与对话讨论相关的内容。保留文章的原始结构、措辞和段落。
2. **无缝融入**：新增内容应自然融入现有文章，就像原作者自己补充的一样。保持一致的写作风格、术语和格式。
3. **读者友好**：最终文章是给读者看的，不是给对话者看的。不要出现"根据讨论"、"如前所述"等元叙述。
4. **保持完整性**：输出必须是完整的文章（不是片段或补丁）。包含所有原始内容加上新增/修改的部分。
5. **格式一致**：使用与原文相同的Markdown格式级别（标题层级、列表风格、代码块格式等）。
6. **不输出元数据**：输出的文章正文不要包含标题行。<article-title> 标签中的标题仅供参考，不要在正文中重复。直接从文章正文内容开始。

## 输出格式

使用以下定界符格式输出（不要使用JSON）：

===UPDATED_CONTENT_START===
（完整的更新后文章正文内容，Markdown格式。不要包含标题行。）
===UPDATED_CONTENT_END===

===CHANGE_SUMMARY_START===
（简要描述做了哪些修改，一句话概括）
===CHANGE_SUMMARY_END===`

// UpdateResult holds the result of an article update generation
type UpdateResult struct {
	UpdatedContent string `json:"updatedContent"`
	ChangeSummary  string `json:"changeSummary"`
	Model          string `json:"model"`
}

// ArticleUpdater generates targeted article updates from conversation context
type ArticleUpdater struct {
	llmRouter *llm.Router
}

// NewArticleUpdater creates a new ArticleUpdater
func NewArticleUpdater(llmRouter *llm.Router) *ArticleUpdater {
	return &ArticleUpdater{llmRouter: llmRouter}
}

// GenerateUpdate generates a targeted article update based on conversation history
func (u *ArticleUpdater) GenerateUpdate(
	ctx context.Context,
	article *model.Article,
	conversationHistory []llm.Message,
) (*UpdateResult, error) {
	prompt := buildUpdatePrompt(article, conversationHistory)

	opts := &llm.GenerateOptions{
		SystemPrompt: articleUpdateSystemPrompt,
		MaxTokens:    8192,
		Temperature:  0.3,
	}

	result, modelUsed, err := u.llmRouter.Generate(llm.TaskArticleUpdate, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	updatedContent, changeSummary, err := parseUpdateResponse(result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse update response: %w", err)
	}

	return &UpdateResult{
		UpdatedContent: updatedContent,
		ChangeSummary:  changeSummary,
		Model:          modelUsed,
	}, nil
}

func buildUpdatePrompt(article *model.Article, history []llm.Message) string {
	var sb strings.Builder

	sb.WriteString("## 原始文章\n\n")
	sb.WriteString("<article-title>" + article.Title + "</article-title>\n\n")
	sb.WriteString("<article-content>\n")
	sb.WriteString(article.Content)
	sb.WriteString("\n</article-content>")
	sb.WriteString("\n\n## 用户与AI助手的对话\n\n")

	for _, msg := range history {
		switch msg.Role {
		case "user":
			sb.WriteString("用户：" + msg.Content + "\n\n")
		case "assistant":
			sb.WriteString("助手：" + msg.Content + "\n\n")
		}
	}

	sb.WriteString("## 任务\n\n")
	sb.WriteString("根据以上对话内容，对原始文章进行定向补充和优化。按照系统提示中的原则和格式输出。")

	return sb.String()
}

func parseUpdateResponse(response string) (updatedContent, changeSummary string, err error) {
	// Reuses extractBetween from article_generator.go
	updatedContent, err = extractBetween(response, "===UPDATED_CONTENT_START===", "===UPDATED_CONTENT_END===")
	if err != nil {
		return "", "", fmt.Errorf("no updated content found in response")
	}

	changeSummary, err = extractBetween(response, "===CHANGE_SUMMARY_START===", "===CHANGE_SUMMARY_END===")
	if err != nil {
		changeSummary = "文章内容已更新"
	}

	updatedContent = stripLeadingTitle(strings.TrimSpace(updatedContent))
	return updatedContent, strings.TrimSpace(changeSummary), nil
}

// stripLeadingTitle removes a leading "标题：...\n" or "# Title\n" line that the LLM sometimes echoes back
var leadingTitleRe = regexp.MustCompile(`^(标题[：:].+|#\s+.+)\n{1,2}`)

func stripLeadingTitle(content string) string {
	return leadingTitleRe.ReplaceAllString(content, "")
}
