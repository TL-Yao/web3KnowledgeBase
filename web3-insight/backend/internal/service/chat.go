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
func NewChatService(db *gorm.DB, llmCfg *config.LLMConfig, claudeKeyFunc, openaiKeyFunc func() string) *ChatService {
	return &ChatService{
		llmRouter:   llm.NewRouterFromConfig(llmCfg, claudeKeyFunc, openaiKeyFunc),
		articleRepo: repository.NewArticleRepository(db),
	}
}

// Chat handles a chat request about an article
// Returns a channel of streaming chunks, the model name used, and any error
func (s *ChatService) Chat(articleID, message, selectedText string) (<-chan llm.StreamChunk, string, error) {
	return s.ChatWithModel(articleID, message, selectedText, "")
}

// ChatWithModel handles a chat request with optional model override
func (s *ChatService) ChatWithModel(articleID, message, selectedText, modelOverride string) (<-chan llm.StreamChunk, string, error) {
	systemPrompt, err := s.buildSystemPrompt(articleID)
	if err != nil {
		return nil, "", err
	}

	// Build user prompt
	userPrompt := message
	if selectedText != "" {
		userPrompt = fmt.Sprintf("关于「%s」这部分内容：%s", selectedText, message)
	}

	opts := &llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    2048,
		Temperature:  0.7,
	}

	// Use specific model if requested, otherwise use router
	if modelOverride != "" {
		adapterName := resolveModelName(modelOverride)
		stream, err := s.llmRouter.GenerateStreamWithModel(adapterName, userPrompt, opts)
		if err != nil {
			return nil, "", fmt.Errorf("LLM generation failed with model %s: %w", adapterName, err)
		}
		return stream, adapterName, nil
	}

	stream, model, err := s.llmRouter.GenerateStream(llm.TaskChat, userPrompt, opts)
	if err != nil {
		return nil, "", fmt.Errorf("LLM generation failed: %w", err)
	}

	return stream, model, nil
}

// ChatWithMessages handles multi-turn chat with message history
func (s *ChatService) ChatWithMessages(articleID string, messages []llm.Message) (<-chan llm.StreamChunk, string, error) {
	return s.ChatWithMessagesAndModel(articleID, messages, "")
}

// ChatWithMessagesAndModel handles multi-turn chat with optional model override
func (s *ChatService) ChatWithMessagesAndModel(articleID string, messages []llm.Message, modelOverride string) (<-chan llm.StreamChunk, string, error) {
	systemPrompt, err := s.buildSystemPrompt(articleID)
	if err != nil {
		return nil, "", err
	}

	opts := &llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    2048,
		Temperature:  0.7,
	}

	// Use specific model if requested, otherwise use router
	if modelOverride != "" {
		adapterName := resolveModelName(modelOverride)
		stream, err := s.llmRouter.GenerateChatStreamWithModel(adapterName, messages, opts)
		if err != nil {
			return nil, "", fmt.Errorf("LLM chat generation failed with model %s: %w", adapterName, err)
		}
		return stream, adapterName, nil
	}

	stream, model, err := s.llmRouter.GenerateChatStream(llm.TaskChat, messages, opts)
	if err != nil {
		return nil, "", fmt.Errorf("LLM chat generation failed: %w", err)
	}

	return stream, model, nil
}

// buildSystemPrompt resolves article context and returns the appropriate system prompt
func (s *ChatService) buildSystemPrompt(articleID string) (string, error) {
	if articleID != "" {
		id, err := uuid.Parse(articleID)
		if err != nil {
			return "", fmt.Errorf("invalid article ID: %w", err)
		}
		article, err := s.articleRepo.GetByID(id)
		if err != nil {
			return "", fmt.Errorf("failed to get article: %w", err)
		}
		return buildChatSystemPrompt(article.Title, article.Content), nil
	}
	return buildGeneralSystemPrompt(), nil
}

// resolveModelName maps short model names to registered adapter names
func resolveModelName(shortName string) string {
	switch shortName {
	case "haiku":
		return "claude-haiku-4-5"
	case "sonnet":
		return "claude-sonnet-4-5"
	case "opus":
		return "claude-opus-4-6"
	default:
		return shortName
	}
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

	return fmt.Sprintf(`**重要：无论用户使用什么语言提问，你必须始终使用中文回答。**

你是一个 Web3 技术助手。用户正在阅读一篇关于「%s」的文章，并对内容有疑问。

文章内容：
%s

请基于文章内容回答用户的问题。如果问题超出文章范围，可以补充相关知识。
使用中文回答，保持专业术语的一致性。回答应该准确、清晰、有帮助。`, title, truncatedContent)
}

// buildGeneralSystemPrompt builds the system prompt for general Web3 questions
func buildGeneralSystemPrompt() string {
	return `**重要：无论用户使用什么语言提问，你必须始终使用中文回答。**

你是一个 Web3 技术助手，专门帮助用户理解区块链、加密货币、DeFi、NFT 等 Web3 相关技术。

请用中文回答用户的问题，保持专业术语的一致性。回答应该：
1. 准确、清晰、有帮助
2. 使用通俗易懂的语言解释复杂概念
3. 在适当的时候提供示例
4. 如果涉及风险，请提醒用户注意`
}
