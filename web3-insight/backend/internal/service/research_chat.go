package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/repository"
	"gorm.io/gorm"
)

// ResearchChatService handles chat interactions alongside research sessions.
type ResearchChatService struct {
	llmRouter   *llm.Router
	sessionRepo *repository.ResearchSessionRepository
	articleRepo *repository.ArticleRepository
	researchCfg *config.ResearchConfig
}

// NewResearchChatService creates a new research chat service.
func NewResearchChatService(db *gorm.DB, llmCfg *config.LLMConfig, claudeKeyFunc, openaiKeyFunc func() string, researchCfg *config.ResearchConfig) *ResearchChatService {
	return &ResearchChatService{
		llmRouter:   llm.NewRouterFromConfig(llmCfg, claudeKeyFunc, openaiKeyFunc),
		sessionRepo: repository.NewResearchSessionRepository(db),
		articleRepo: repository.NewArticleRepository(db),
		researchCfg: researchCfg,
	}
}

// ChatWithModel handles a single-message chat with optional model override.
func (s *ResearchChatService) ChatWithModel(sessionID, message, modelOverride string) (<-chan llm.StreamChunk, string, error) {
	systemPrompt, err := s.buildSystemPrompt(sessionID)
	if err != nil {
		return nil, "", err
	}

	opts := &llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    2048,
		Temperature:  0.7,
	}

	if modelOverride != "" {
		adapterName := resolveModelName(modelOverride)
		stream, err := s.llmRouter.GenerateStreamWithModel(adapterName, message, opts)
		if err != nil {
			return nil, "", fmt.Errorf("LLM generation failed with model %s: %w", adapterName, err)
		}
		return stream, adapterName, nil
	}

	stream, model, err := s.llmRouter.GenerateStream(llm.TaskChat, message, opts)
	if err != nil {
		return nil, "", fmt.Errorf("LLM generation failed: %w", err)
	}

	return stream, model, nil
}

// ChatWithMessagesAndModel handles multi-turn chat with message history and optional model override.
func (s *ResearchChatService) ChatWithMessagesAndModel(sessionID string, messages []llm.Message, modelOverride string) (<-chan llm.StreamChunk, string, error) {
	systemPrompt, err := s.buildSystemPrompt(sessionID)
	if err != nil {
		return nil, "", err
	}

	opts := &llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    2048,
		Temperature:  0.7,
	}

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

// buildSystemPrompt creates context from the research session and linked article.
func (s *ResearchChatService) buildSystemPrompt(sessionID string) (string, error) {
	if sessionID == "" {
		return buildResearchGeneralPrompt(), nil
	}

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return "", fmt.Errorf("invalid session ID: %w", err)
	}

	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	// Build domain context
	domainContext := ""
	if s.researchCfg != nil {
		if domain, err := s.researchCfg.GetDomainByID(session.Domain); err == nil {
			domainContext = domain.SystemContext
		}
	}

	// If session has a linked article, include its content
	if session.ArticleID != nil {
		article, err := s.articleRepo.GetByID(*session.ArticleID)
		if err == nil {
			return buildResearchArticlePrompt(session.Question, article.Content, domainContext), nil
		}
	}

	// No article yet — just the research question + domain guidance
	return buildResearchQuestionPrompt(session.Question, domainContext), nil
}

func buildResearchArticlePrompt(question, content, domainContext string) string {
	// Truncate content to avoid token limits
	maxContentLen := 8000
	truncatedContent := content
	if len(content) > maxContentLen {
		truncatedContent = content[:maxContentLen] + "\n\n[内容已截断...]"
	}

	prompt := fmt.Sprintf(`**重要：无论用户使用什么语言提问，你必须始终使用中文回答。**

你是一个研究助手。用户正在进行关于「%s」的深度研究，已生成了研究报告。

`, question)

	if domainContext != "" {
		prompt += fmt.Sprintf("领域上下文：\n%s\n\n", domainContext)
	}

	prompt += fmt.Sprintf(`研究报告内容：
%s

请基于研究报告内容回答用户的问题。你可以：
1. 解释报告中的复杂概念
2. 提供报告未覆盖的补充信息
3. 分析报告中的数据和观点
4. 回答用户的后续问题

用中文回答，保持专业准确。如果用户发现了有价值的信息，提醒他们可以"固定"消息以整合到报告中。`, truncatedContent)

	return prompt
}

func buildResearchQuestionPrompt(question, domainContext string) string {
	prompt := fmt.Sprintf(`**重要：无论用户使用什么语言提问，你必须始终使用中文回答。**

你是一个研究助手。用户正在进行关于「%s」的研究，报告尚在生成中。

`, question)

	if domainContext != "" {
		prompt += fmt.Sprintf("领域上下文：\n%s\n\n", domainContext)
	}

	prompt += `请帮助用户：
1. 回答关于研究主题的初步问题
2. 提供相关背景知识
3. 建议可能的研究方向
4. 解释相关术语和概念

用中文回答，保持专业准确。`

	return prompt
}

func buildResearchGeneralPrompt() string {
	return `**重要：无论用户使用什么语言提问，你必须始终使用中文回答。**

你是一个通用研究助手，可以帮助用户进行各种主题的深度研究。

请用中文回答用户的问题，保持专业准确。你可以：
1. 帮助用户明确研究方向
2. 提供相关背景知识
3. 分析不同观点和数据
4. 建议深入研究的切入点`
}
