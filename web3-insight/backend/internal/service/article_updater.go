package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
)

const articleUpdateSystemPrompt = `你是Web3知识库的文章编辑专家。你的任务是根据用户与AI助手的对话，判断是否需要对现有文章进行定向补充和优化。

## 核心原则

1. **有实质内容才更新**：只有当对话中包含文章没有覆盖的新知识、纠正、或新视角时才进行更新。以下情况不需要更新：
   - 用户只是在提问文章已有的内容（问答澄清）
   - 对话偏离了文章主题
   - 对话内容是泛泛而谈，没有具体可融入的知识点
   - 不要为了"更新而更新"——没有实质新内容就不要改
2. **定向补充，不重写**：只修改或添加与对话讨论相关的内容。保留文章的原始结构、措辞和段落。
3. **无缝融入**：新增内容应自然融入现有文章，就像原作者自己补充的一样。保持一致的写作风格、术语和格式。
4. **读者友好**：最终文章是给读者看的，不是给对话者看的。不要出现"根据讨论"、"如前所述"等元叙述。
5. **保持完整性**：输出必须是完整的文章（不是片段或补丁）。包含所有原始内容加上新增/修改的部分。
6. **格式一致**：使用与原文相同的Markdown格式级别（标题层级、列表风格、代码块格式等）。
7. **不输出元数据**：输出的文章正文不要包含标题行。<article-title> 标签中的标题仅供参考，不要在正文中重复。直接从文章正文内容开始。

## 输出格式

**如果对话包含实质新内容，需要更新文章**，使用以下定界符格式输出：

===UPDATED_CONTENT_START===
（完整的更新后文章正文内容，Markdown格式。不要包含标题行。）
===UPDATED_CONTENT_END===

===CHANGE_SUMMARY_START===
（简要描述做了哪些修改，一句话概括）
===CHANGE_SUMMARY_END===

**如果对话不包含实质新内容，不需要更新**，使用以下格式：

===NO_CHANGE===
（简要说明为什么不需要更新，一句话）
===NO_CHANGE_END===`

// UpdateResult holds the result of an article update generation
type UpdateResult struct {
	UpdatedContent string `json:"updatedContent"`
	ChangeSummary  string `json:"changeSummary"`
	Model          string `json:"model"`
	NoChange       bool   `json:"noChange,omitempty"`
	NoChangeReason string `json:"noChangeReason,omitempty"`
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

	return parseUpdateResult(result, modelUsed)
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

// parseUpdateResult parses the LLM response and returns an UpdateResult,
// handling both the update case and the no-change case.
func parseUpdateResult(response, modelUsed string) (*UpdateResult, error) {
	// Check for no-change output first
	noChangeReason, err := extractBetween(response, "===NO_CHANGE===", "===NO_CHANGE_END===")
	if err == nil {
		return &UpdateResult{
			Model:          modelUsed,
			NoChange:       true,
			NoChangeReason: strings.TrimSpace(noChangeReason),
		}, nil
	}

	// Parse normal update
	updatedContent, changeSummary, err := parseUpdateResponse(response)
	if err != nil {
		return nil, err
	}

	return &UpdateResult{
		UpdatedContent: updatedContent,
		ChangeSummary:  changeSummary,
		Model:          modelUsed,
	}, nil
}

func parseUpdateResponse(response string) (updatedContent, changeSummary string, err error) {
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
