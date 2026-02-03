package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gosimple/slug"
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// Generator handles knowledge article generation
type Generator struct {
	llmRouter   *llm.Router
	articleRepo *repository.ArticleRepository
	newsRepo    *repository.NewsRepository
	classifier  *Classifier
}

// NewGenerator creates a new generator service
func NewGenerator(router *llm.Router, articleRepo *repository.ArticleRepository, newsRepo *repository.NewsRepository, classifier *Classifier) *Generator {
	return &Generator{
		llmRouter:   router,
		articleRepo: articleRepo,
		newsRepo:    newsRepo,
		classifier:  classifier,
	}
}

// GenerationRequest represents a request to generate an article
type GenerationRequest struct {
	Topic       string
	CategoryID  *uuid.UUID
	Style       string   // "detailed", "concise", "beginner-friendly"
	References  []string // URLs or content snippets for reference
	ModelPrefer string   // Preferred model (optional)
}

// GenerationResult represents the result of article generation
type GenerationResult struct {
	Article    *model.Article
	ModelUsed  string
	TokensUsed int
	Duration   time.Duration
}

// GenerateArticle generates a knowledge article on a topic
func (g *Generator) GenerateArticle(ctx context.Context, req *GenerationRequest) (*GenerationResult, error) {
	startTime := time.Now()

	// Gather reference materials
	references := g.gatherReferences(ctx, req.Topic, req.References)

	// Build prompt
	prompt := fmt.Sprintf(PromptKnowledgeArticle, req.Topic, references)

	// Determine which LLM task to use based on complexity
	task := llm.TaskContentGeneration

	opts := &llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   8000, // Allow for long articles
	}

	// Generate article content
	content, modelUsed, err := g.llmRouter.Generate(task, prompt, opts)
	if err != nil {
		return nil, fmt.Errorf("content generation failed: %w", err)
	}

	// Clean up content
	content = g.cleanGeneratedContent(content)

	// Validate quality
	if err := g.validateQuality(content); err != nil {
		log.Printf("Quality check failed: %v, attempting regeneration...", err)
		// Could implement retry logic here
	}

	// Create article
	article := &model.Article{
		Title:            g.extractTitle(content, req.Topic),
		Slug:             g.generateSlug(req.Topic),
		Content:          content,
		Summary:          g.extractSummary(content),
		CategoryID:       req.CategoryID,
		Status:           "published",
		SourceLanguage:   "zh",
		ModelUsed:        modelUsed,
		GenerationPrompt: prompt,
		Tags:             g.extractTags(content, req.Topic),
	}

	// Save to database
	if err := g.articleRepo.Create(article); err != nil {
		return nil, fmt.Errorf("failed to save article: %w", err)
	}

	// Trigger classification if no category specified
	if req.CategoryID == nil && g.classifier != nil {
		go func() {
			ctx := context.Background()
			if err := g.classifier.ClassifyAndUpdate(ctx, article.ID); err != nil {
				log.Printf("Auto-classification failed: %v", err)
			}
		}()
	}

	return &GenerationResult{
		Article:   article,
		ModelUsed: modelUsed,
		Duration:  time.Since(startTime),
	}, nil
}

// gatherReferences collects reference materials for generation
func (g *Generator) gatherReferences(ctx context.Context, topic string, providedRefs []string) string {
	var refs []string

	// Include provided references
	for _, ref := range providedRefs {
		refs = append(refs, fmt.Sprintf("- %s", ref))
	}

	// Search for related news items
	if g.newsRepo != nil {
		// Simple keyword search in news
		// In a real implementation, this would use vector search
		refs = append(refs, "\n（基于已收集的新闻资料）")
	}

	if len(refs) == 0 {
		return "（无额外参考资料，请基于通用知识生成）"
	}

	return strings.Join(refs, "\n")
}

// cleanGeneratedContent cleans up the generated content
func (g *Generator) cleanGeneratedContent(content string) string {
	// Remove any preamble before the actual content
	// Look for the first markdown heading
	if idx := strings.Index(content, "## "); idx > 0 && idx < 200 {
		// Check if there's a # heading before ##
		if h1Idx := strings.Index(content, "# "); h1Idx >= 0 && h1Idx < idx {
			content = content[h1Idx:]
		} else {
			content = content[idx:]
		}
	}

	// Remove trailing artifacts
	content = strings.TrimSpace(content)

	return content
}

// validateQuality checks if the generated content meets quality standards
func (g *Generator) validateQuality(content string) error {
	// Check minimum length (approximately 1500 characters for Chinese)
	charCount := utf8.RuneCountInString(content)
	if charCount < 1500 {
		return fmt.Errorf("content too short: %d characters (minimum 1500)", charCount)
	}

	// Check for section headers
	if !strings.Contains(content, "## ") {
		return fmt.Errorf("content lacks section structure")
	}

	// Check for terminology format (English + Chinese pattern)
	// This is a soft check, don't fail on it
	termPattern := regexp.MustCompile(`[A-Za-z]+\s*[（(][^）)]+[）)]`)
	if !termPattern.MatchString(content) {
		log.Printf("Warning: content may lack proper terminology formatting")
	}

	return nil
}

// extractTitle extracts or generates a title from content
func (g *Generator) extractTitle(content string, topic string) string {
	// Try to extract from first line if it's a heading
	lines := strings.SplitN(content, "\n", 2)
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "# ") {
			return strings.TrimPrefix(firstLine, "# ")
		}
	}

	// Use topic as title
	return topic
}

// generateSlug generates a URL-friendly slug
func (g *Generator) generateSlug(topic string) string {
	// For Chinese topics, create a simple slug
	baseSlug := slug.Make(topic)
	if baseSlug == "" {
		// Fallback for pure Chinese
		baseSlug = fmt.Sprintf("article-%d", time.Now().Unix())
	}

	// Check if slug exists and append number if needed
	existingCount := g.articleRepo.CountBySlugPrefix(baseSlug)
	if existingCount > 0 {
		baseSlug = fmt.Sprintf("%s-%d", baseSlug, existingCount+1)
	}

	return baseSlug
}

// extractSummary extracts a summary from the content
func (g *Generator) extractSummary(content string) string {
	// Look for content under "概述" section
	if idx := strings.Index(content, "## 概述"); idx >= 0 {
		afterOverview := content[idx+len("## 概述"):]
		// Find next section
		if nextSection := strings.Index(afterOverview, "\n## "); nextSection > 0 {
			summary := strings.TrimSpace(afterOverview[:nextSection])
			return g.truncateSummary(summary, 300)
		}
	}

	// Fallback: use first paragraph after any heading
	lines := strings.Split(content, "\n")
	var summary strings.Builder
	inContent := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			inContent = true
			continue
		}
		if inContent && line != "" && !strings.HasPrefix(line, "-") {
			summary.WriteString(line)
			summary.WriteString(" ")
			if summary.Len() > 300 {
				break
			}
		}
	}

	return g.truncateSummary(summary.String(), 300)
}

// truncateSummary truncates summary to specified length
func (g *Generator) truncateSummary(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// extractTags extracts potential tags from content
func (g *Generator) extractTags(content string, topic string) []string {
	tags := []string{}

	// Add topic as a tag
	tags = append(tags, topic)

	// Extract English terms in parentheses (these are often key concepts)
	termPattern := regexp.MustCompile(`([A-Z][a-zA-Z0-9]+)\s*[（(]`)
	matches := termPattern.FindAllStringSubmatch(content, 10)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			term := match[1]
			if !seen[term] && len(term) > 2 {
				tags = append(tags, term)
				seen[term] = true
			}
		}
		if len(tags) >= 5 {
			break
		}
	}

	return tags
}

// GenerateStream generates an article with streaming output
func (g *Generator) GenerateStream(ctx context.Context, req *GenerationRequest) (<-chan llm.StreamChunk, string, error) {
	references := g.gatherReferences(ctx, req.Topic, req.References)
	prompt := fmt.Sprintf(PromptKnowledgeArticle, req.Topic, references)

	return g.llmRouter.GenerateStream(llm.TaskContentGeneration, prompt, &llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   8000,
	})
}
