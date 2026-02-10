package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/gosimple/slug"
	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

const (
	ArticleGenerationTimeout = 60 * time.Minute // 60 minutes timeout per article
	MinContentLength         = 500              // Minimum content length in characters
)

// Delimiters for structured output parsing
const (
	TitleStart   = "===TITLE_START==="
	TitleEnd     = "===TITLE_END==="
	ContentStart = "===CONTENT_START==="
	ContentEnd   = "===CONTENT_END==="
	SummaryStart = "===SUMMARY_START==="
	SummaryEnd   = "===SUMMARY_END==="
)

type ArticleGeneratorService struct {
	articleRepo *repository.ArticleRepository
	classifier  *Classifier
	tagger      *Tagger
	prompts     *config.PromptsConfig
}

func NewArticleGeneratorService(articleRepo *repository.ArticleRepository, classifier *Classifier, prompts *config.PromptsConfig) *ArticleGeneratorService {
	return &ArticleGeneratorService{
		articleRepo: articleRepo,
		classifier:  classifier,
		prompts:     prompts,
	}
}

// SetTagger sets the tagger service for post-generation tagging
func (ag *ArticleGeneratorService) SetTagger(tagger *Tagger) {
	ag.tagger = tagger
}

// ArticleData represents parsed article data
type ArticleData struct {
	Title   string
	Content string
	Summary string
}

// GenerateArticle generates a comprehensive article for the given keyword using theme-specific prompt
func (ag *ArticleGeneratorService) GenerateArticle(ctx context.Context, keyword string, themeID string) (*model.Article, string, error) {
	executor := NewClaudeExecutor()
	sessionID := executor.GetSessionID()

	log.Printf("Generating article for keyword: '%s' (theme: %s, session: %s)", keyword, themeID, sessionID)

	// Use timeout from config if available
	timeout := ArticleGenerationTimeout
	if ag.prompts != nil && ag.prompts.Generation.ArticleTimeoutMinutes > 0 {
		timeout = time.Duration(ag.prompts.Generation.ArticleTimeoutMinutes) * time.Minute
	}

	genCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	prompt, err := ag.buildThemePrompt(keyword, themeID)
	if err != nil {
		return nil, sessionID, fmt.Errorf("failed to build prompt: %w", err)
	}

	response, err := executor.Execute(genCtx, prompt)
	if err != nil {
		if genCtx.Err() == context.DeadlineExceeded {
			return nil, sessionID, fmt.Errorf("article generation timeout (%v): %w", ArticleGenerationTimeout, err)
		}
		return nil, sessionID, fmt.Errorf("failed to generate article: %w", err)
	}

	log.Printf("Article generation completed for '%s', cost: $%.4f", keyword, response.TotalCostUSD)

	// Parse delimiter-based response
	articleData, err := parseDelimitedResponse(response.Result)
	if err != nil {
		truncated := response.Result
		if len(truncated) > 500 {
			truncated = truncated[:500] + "..."
		}
		return nil, sessionID, fmt.Errorf("failed to parse response: %w\nResponse preview: %s", err, truncated)
	}

	// Validate content
	if err := ag.validateArticle(articleData); err != nil {
		return nil, sessionID, fmt.Errorf("article validation failed: %w", err)
	}

	articleSlug := ag.generateSlug(articleData.Title)

	article := &model.Article{
		Title:   articleData.Title,
		Slug:    articleSlug,
		Content: articleData.Content,
		Summary: articleData.Summary,
		Status:  "published",
		Tags:    []string{keyword},
		ThemeID: &themeID,
	}

	if err := ag.articleRepo.Create(article); err != nil {
		return nil, sessionID, fmt.Errorf("failed to save article: %w", err)
	}

	log.Printf("Article created: ID=%s, Title='%s', Slug='%s'", article.ID, article.Title, article.Slug)

	// Trigger async classification (non-blocking)
	if ag.classifier != nil {
		go func() {
			classifyCtx := context.Background()
			if err := ag.classifier.ClassifyAndUpdate(classifyCtx, article.ID); err != nil {
				log.Printf("Warning: Failed to classify article %s: %v", article.ID, err)
			} else {
				log.Printf("Successfully classified article %s", article.ID)
			}
		}()
	} else {
		log.Printf("Classifier not configured, skipping classification for article %s", article.ID)
	}

	// Trigger async tagging (non-blocking)
	if ag.tagger != nil {
		go func() {
			tagCtx := context.Background()
			// Re-fetch the article to get the latest state (after classification may have updated it)
			freshArticle, err := ag.articleRepo.GetByID(article.ID)
			if err != nil {
				log.Printf("Warning: Failed to re-fetch article %s for tagging: %v", article.ID, err)
				return
			}
			if err := ag.tagger.TagArticle(tagCtx, freshArticle); err != nil {
				log.Printf("Warning: Failed to tag article %s: %v", article.ID, err)
			} else {
				log.Printf("Successfully tagged article %s", article.ID)
			}
		}()
	}

	return article, sessionID, nil
}

// buildThemePrompt renders the article prompt template for the given theme and keyword.
// Falls back to a hardcoded default prompt if no theme config is available.
func (ag *ArticleGeneratorService) buildThemePrompt(keyword string, themeID string) (string, error) {
	if ag.prompts != nil {
		theme, err := ag.prompts.GetThemeByID(themeID)
		if err == nil && theme.ArticlePrompt != "" {
			tmpl, err := template.New("article").Parse(theme.ArticlePrompt)
			if err != nil {
				return "", fmt.Errorf("failed to parse article prompt template: %w", err)
			}
			data := map[string]interface{}{
				"Keyword": keyword,
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return "", fmt.Errorf("failed to render article prompt: %w", err)
			}
			return buf.String(), nil
		}
	}

	// Fallback: hardcoded default prompt
	return fmt.Sprintf(`Write a comprehensive educational article about: "%s"

**Target Audience**: Complete beginners with NO prior knowledge of this topic.

**Content Requirements**:
- Purpose: Educational - help readers LEARN and UNDERSTAND
- Be CONCRETE: Use specific examples, analogies, real-world scenarios
- NO assumptions: Explain all concepts from scratch
- Language: Chinese (中文)
- Technical terms: Use "English (中文)" format, e.g., "Smart Contract (智能合约)"
- Length: 1500-2500 words

**Research**: Use WebFetch if you need the latest information or accurate data.

**OUTPUT FORMAT** (use these exact delimiters):

===TITLE_START===
Article title in Chinese here
===TITLE_END===
===CONTENT_START===
Full Markdown content here
===CONTENT_END===
===SUMMARY_START===
80-120 word summary in Chinese
===SUMMARY_END===

IMPORTANT: Use exactly these delimiter tags. No other format.`, keyword), nil
}

// parseDelimitedResponse extracts title, content, and summary from delimiter-based format
func parseDelimitedResponse(text string) (*ArticleData, error) {
	title, err := extractBetween(text, TitleStart, TitleEnd)
	if err != nil {
		return nil, fmt.Errorf("title extraction failed: %w", err)
	}

	content, err := extractBetween(text, ContentStart, ContentEnd)
	if err != nil {
		return nil, fmt.Errorf("content extraction failed: %w", err)
	}

	summary, err := extractBetween(text, SummaryStart, SummaryEnd)
	if err != nil {
		return nil, fmt.Errorf("summary extraction failed: %w", err)
	}

	return &ArticleData{
		Title:   strings.TrimSpace(title),
		Content: strings.TrimSpace(content),
		Summary: strings.TrimSpace(summary),
	}, nil
}

// extractBetween extracts text between start and end delimiters
func extractBetween(text, startDelim, endDelim string) (string, error) {
	startIdx := strings.Index(text, startDelim)
	if startIdx == -1 {
		return "", fmt.Errorf("start delimiter %q not found", startDelim)
	}

	afterStart := startIdx + len(startDelim)
	endIdx := strings.Index(text[afterStart:], endDelim)
	if endIdx == -1 {
		return "", fmt.Errorf("end delimiter %q not found", endDelim)
	}

	return text[afterStart : afterStart+endIdx], nil
}

// validateArticle checks if the article meets quality requirements
func (ag *ArticleGeneratorService) validateArticle(data *ArticleData) error {
	if data.Title == "" {
		return fmt.Errorf("title is empty")
	}

	if len(data.Content) < MinContentLength {
		return fmt.Errorf("content too short: %d characters (minimum: %d)", len(data.Content), MinContentLength)
	}

	if data.Summary == "" {
		return fmt.Errorf("summary is empty")
	}

	return nil
}

// generateSlug generates a URL-friendly slug from the title
func (ag *ArticleGeneratorService) generateSlug(title string) string {
	baseSlug := slug.Make(title)
	if baseSlug == "" {
		baseSlug = fmt.Sprintf("article-%d", time.Now().Unix())
	}

	existingCount := ag.articleRepo.CountBySlugPrefix(baseSlug)
	if existingCount > 0 {
		baseSlug = fmt.Sprintf("%s-%d", baseSlug, existingCount+1)
	}

	return baseSlug
}
