package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/repository"
	"gorm.io/gorm"
)

// ChatService handles chat interactions with articles
type ChatService struct {
	llmRouter   *llm.Router
	articleRepo *repository.ArticleRepository
}

// NewChatService creates a new chat service
func NewChatService(db *gorm.DB, llmCfg *config.LLMConfig) *ChatService {
	return &ChatService{
		llmRouter:   llm.NewRouterFromConfig(llmCfg),
		articleRepo: repository.NewArticleRepository(db),
	}
}

// Chat handles a chat request about an article
// Returns a channel of streaming chunks, the model name used, and any error
func (s *ChatService) Chat(articleID, message, selectedText string) (<-chan llm.StreamChunk, string, error) {
	var systemPrompt string

	// If articleID is provided, fetch article context
	if articleID != "" {
		id, err := uuid.Parse(articleID)
		if err != nil {
			return nil, "", fmt.Errorf("invalid article ID: %w", err)
		}

		article, err := s.articleRepo.GetByID(id)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get article: %w", err)
		}

		systemPrompt = buildChatSystemPrompt(article.Title, article.Content)
	} else {
		systemPrompt = buildGeneralSystemPrompt()
	}

	// Build user prompt
	userPrompt := message
	if selectedText != "" {
		userPrompt = fmt.Sprintf("关于「%s」这部分内容：%s", selectedText, message)
	}

	// Configure generation options
	opts := &llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    2048,
		Temperature:  0.7,
	}

	// Use router to stream response
	stream, model, err := s.llmRouter.GenerateStream(llm.TaskChat, userPrompt, opts)
	if err != nil {
		return nil, "", fmt.Errorf("LLM generation failed: %w", err)
	}

	return stream, model, nil
}

// ChatWithMessages handles multi-turn chat with message history
func (s *ChatService) ChatWithMessages(articleID string, messages []llm.Message) (<-chan llm.StreamChunk, string, error) {
	var systemPrompt string

	// If articleID is provided, fetch article context
	if articleID != "" {
		id, err := uuid.Parse(articleID)
		if err != nil {
			return nil, "", fmt.Errorf("invalid article ID: %w", err)
		}

		article, err := s.articleRepo.GetByID(id)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get article: %w", err)
		}

		systemPrompt = buildChatSystemPrompt(article.Title, article.Content)
	} else {
		systemPrompt = buildGeneralSystemPrompt()
	}

	opts := &llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    2048,
		Temperature:  0.7,
	}

	stream, model, err := s.llmRouter.GenerateChatStream(llm.TaskChat, messages, opts)
	if err != nil {
		return nil, "", fmt.Errorf("LLM chat generation failed: %w", err)
	}

	return stream, model, nil
}

// GetAvailableModels returns the list of available LLM adapters
func (s *ChatService) GetAvailableModels() []string {
	return s.llmRouter.ListAvailableAdapters()
}

// buildChatSystemPrompt builds the system prompt for article-based chat
func buildChatSystemPrompt(title, content string) string {
	// Truncate content if too long to avoid token limits
	maxContentLen := 8000
	truncatedContent := content
	if len(content) > maxContentLen {
		truncatedContent = content[:maxContentLen] + "\n\n[内容已截断...]"
	}

	return fmt.Sprintf(`你是一个 Web3 技术助手。用户正在阅读一篇关于「%s」的文章，并对内容有疑问。

文章内容：
%s

请基于文章内容回答用户的问题。如果问题超出文章范围，可以补充相关知识。
使用中文回答，保持专业术语的一致性。回答应该准确、清晰、有帮助。`, title, truncatedContent)
}

// buildGeneralSystemPrompt builds the system prompt for general Web3 questions
func buildGeneralSystemPrompt() string {
	return `你是一个 Web3 技术助手，专门帮助用户理解区块链、加密货币、DeFi、NFT 等 Web3 相关技术。

请用中文回答用户的问题，保持专业术语的一致性。回答应该：
1. 准确、清晰、有帮助
2. 使用通俗易懂的语言解释复杂概念
3. 在适当的时候提供示例
4. 如果涉及风险，请提醒用户注意`
}
