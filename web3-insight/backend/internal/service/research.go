package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/collector"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// ResearchService handles instant research queries
type ResearchService struct {
	llmRouter    *llm.Router
	articleRepo  *repository.ArticleRepository
	searchRouter *collector.SearchRouter
	generator    *Generator
}

// NewResearchService creates a new research service
func NewResearchService(
	router *llm.Router,
	articleRepo *repository.ArticleRepository,
	searchRouter *collector.SearchRouter,
	generator *Generator,
) *ResearchService {
	return &ResearchService{
		llmRouter:    router,
		articleRepo:  articleRepo,
		searchRouter: searchRouter,
		generator:    generator,
	}
}

// ResearchRequest represents an instant research request
type ResearchRequest struct {
	Query        string
	SaveArticle  bool // Whether to save the result as an article
	UseWebSearch bool // Whether to use web search for additional context
}

// ResearchResponse represents the research result
type ResearchResponse struct {
	Content         string
	ModelUsed       string
	Sources         []string
	RelatedArticles []model.Article
	SavedArticleID  *string
	Duration        time.Duration
}

// Research performs instant research on a topic
func (s *ResearchService) Research(ctx context.Context, req *ResearchRequest) (*ResearchResponse, error) {
	startTime := time.Now()
	response := &ResearchResponse{}

	// 1. Check for existing related articles
	related, err := s.findRelatedArticles(req.Query)
	if err == nil && len(related) > 0 {
		response.RelatedArticles = related
	}

	// 2. Optionally gather web search results
	var webContext string
	if req.UseWebSearch && s.searchRouter != nil {
		searchResult, err := s.searchRouter.Search(ctx, req.Query+" Web3 blockchain", 5)
		if err == nil && searchResult != nil {
			webContext = s.formatSearchResults(searchResult)
			for _, r := range searchResult.Results {
				response.Sources = append(response.Sources, r.URL)
			}
		}
	}

	// 3. Build context from related articles
	var articleContext string
	if len(related) > 0 {
		articleContext = s.formatRelatedArticles(related)
	}

	// 4. Generate research response
	contextParts := []string{}
	if webContext != "" {
		contextParts = append(contextParts, "网络搜索结果：\n"+webContext)
	}
	if articleContext != "" {
		contextParts = append(contextParts, "相关已有文章：\n"+articleContext)
	}

	contextStr := strings.Join(contextParts, "\n\n")
	if contextStr == "" {
		contextStr = "（无额外上下文，请基于通用知识回答）"
	}

	prompt := fmt.Sprintf(PromptInstantResearch, req.Query, contextStr)

	content, modelUsed, err := s.llmRouter.Generate(llm.TaskContentGeneration, prompt, &llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   4000,
	})
	if err != nil {
		return nil, fmt.Errorf("research generation failed: %w", err)
	}

	response.Content = content
	response.ModelUsed = modelUsed
	response.Duration = time.Since(startTime)

	// 5. Optionally save as article
	if req.SaveArticle {
		savedID, err := s.saveAsArticle(ctx, req.Query, content, response.Sources)
		if err != nil {
			log.Printf("Failed to save research as article: %v", err)
		} else {
			idStr := savedID.String()
			response.SavedArticleID = &idStr
		}
	}

	return response, nil
}

// ResearchStream performs research with streaming output
func (s *ResearchService) ResearchStream(ctx context.Context, req *ResearchRequest) (<-chan llm.StreamChunk, string, error) {
	// Gather context (simplified for streaming)
	var contextStr string

	if req.UseWebSearch && s.searchRouter != nil {
		searchResult, err := s.searchRouter.Search(ctx, req.Query+" Web3 blockchain", 3)
		if err == nil && searchResult != nil {
			contextStr = s.formatSearchResults(searchResult)
		}
	}

	if contextStr == "" {
		contextStr = "（无额外上下文，请基于通用知识回答）"
	}

	prompt := fmt.Sprintf(PromptInstantResearch, req.Query, contextStr)

	return s.llmRouter.GenerateStream(llm.TaskContentGeneration, prompt, &llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   4000,
	})
}

// findRelatedArticles finds articles related to the query
func (s *ResearchService) findRelatedArticles(query string) ([]model.Article, error) {
	// Simple text search for now
	// In production, this would use vector similarity search
	return s.articleRepo.Search(query, 5)
}

// formatSearchResults formats search results for the prompt
func (s *ResearchService) formatSearchResults(results *collector.SearchResponse) string {
	var sb strings.Builder

	if results.Answer != "" {
		sb.WriteString("AI 摘要：")
		sb.WriteString(results.Answer)
		sb.WriteString("\n\n")
	}

	for i, r := range results.Results {
		if i >= 5 {
			break
		}
		sb.WriteString(fmt.Sprintf("- %s\n  %s\n", r.Title, r.Content))
	}

	return sb.String()
}

// formatRelatedArticles formats related articles for the prompt
func (s *ResearchService) formatRelatedArticles(articles []model.Article) string {
	var sb strings.Builder

	for i, a := range articles {
		if i >= 3 {
			break
		}
		sb.WriteString(fmt.Sprintf("### %s\n", a.Title))
		if a.Summary != "" {
			sb.WriteString(a.Summary)
		} else {
			// Use truncated content
			content := a.Content
			if len([]rune(content)) > 200 {
				content = string([]rune(content)[:200]) + "..."
			}
			sb.WriteString(content)
		}
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// saveAsArticle saves the research result as a knowledge article
func (s *ResearchService) saveAsArticle(ctx context.Context, query, content string, sources []string) (*uuid.UUID, error) {
	if s.generator == nil {
		return nil, fmt.Errorf("generator not available")
	}

	// Create article directly instead of using generator
	article := &model.Article{
		Title:      query,
		Slug:       s.generator.generateSlug(query),
		Content:    content,
		Summary:    s.generator.extractSummary(content),
		Status:     "published",
		SourceURLs: sources,
		Tags:       s.generator.extractTags(content, query),
	}

	if err := s.articleRepo.Create(article); err != nil {
		return nil, err
	}

	return &article.ID, nil
}
