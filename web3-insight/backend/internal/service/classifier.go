package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// Classifier handles automatic content classification
type Classifier struct {
	llmRouter    *llm.Router
	articleRepo  *repository.ArticleRepository
	categoryRepo *repository.CategoryRepository
}

// NewClassifier creates a new classifier service
func NewClassifier(router *llm.Router, articleRepo *repository.ArticleRepository, categoryRepo *repository.CategoryRepository) *Classifier {
	return &Classifier{
		llmRouter:    router,
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
	}
}

// ClassificationResult represents the LLM classification response
type ClassificationResult struct {
	PrimaryCategory     string   `json:"primaryCategory"`
	SecondaryCategories []string `json:"secondaryCategories"`
	SuggestedTags       []string `json:"suggestedTags"`
	Confidence          float64  `json:"confidence"`
	Reasoning           string   `json:"reasoning"`
	NewCategorySuggestion *struct {
		Name   string `json:"name"`
		Parent string `json:"parent"`
	} `json:"newCategorySuggestion,omitempty"`
}

// ClassifyArticle classifies an article and returns the suggested category
func (c *Classifier) ClassifyArticle(ctx context.Context, article *model.Article) (*ClassificationResult, string, error) {
	// Get category tree for prompt
	categoryTree, err := c.getCategoryTreeString()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get categories: %w", err)
	}

	// Prepare content summary (truncate if too long)
	contentSummary := article.Summary
	if contentSummary == "" {
		// Use first 500 chars of content
		contentSummary = truncateString(article.Content, 500)
	}

	// Build prompt
	prompt := fmt.Sprintf(PromptClassification, categoryTree, article.Title, contentSummary)

	// Call LLM
	response, modelUsed, err := c.llmRouter.Generate(llm.TaskClassification, prompt, &llm.GenerateOptions{
		Temperature: 0.2, // Very low temperature for consistent classification
		MaxTokens:   500,
	})
	if err != nil {
		return nil, "", fmt.Errorf("LLM classification failed: %w", err)
	}

	// Parse response
	result, err := c.parseClassificationResponse(response)
	if err != nil {
		return nil, modelUsed, fmt.Errorf("failed to parse classification: %w", err)
	}

	return result, modelUsed, nil
}

// parseClassificationResponse extracts JSON from LLM response
func (c *Classifier) parseClassificationResponse(response string) (*ClassificationResult, error) {
	response = strings.TrimSpace(response)

	// Remove markdown code blocks
	response = regexp.MustCompile("(?s)```json\\s*").ReplaceAllString(response, "")
	response = regexp.MustCompile("(?s)```\\s*").ReplaceAllString(response, "")
	response = strings.TrimSpace(response)

	// Find JSON object
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object found")
	}

	jsonStr := response[start : end+1]

	var result ClassificationResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

// getCategoryTreeString returns categories formatted for the prompt
func (c *Classifier) getCategoryTreeString() (string, error) {
	categories, err := c.categoryRepo.FindAll()
	if err != nil {
		return "", err
	}

	// Build category paths
	var paths []string
	categoryMap := make(map[uuid.UUID]*model.Category)
	for i := range categories {
		categoryMap[categories[i].ID] = &categories[i]
	}

	for _, cat := range categories {
		path := c.buildCategoryPath(&cat, categoryMap)
		paths = append(paths, "- "+path)
	}

	return strings.Join(paths, "\n"), nil
}

// buildCategoryPath builds the full path for a category
func (c *Classifier) buildCategoryPath(cat *model.Category, categoryMap map[uuid.UUID]*model.Category) string {
	if cat.ParentID == nil {
		return cat.Name
	}

	parent, ok := categoryMap[*cat.ParentID]
	if !ok {
		return cat.Name
	}

	return c.buildCategoryPath(parent, categoryMap) + "/" + cat.Name
}

// ClassifyAndUpdate classifies an article and updates it in the database
func (c *Classifier) ClassifyAndUpdate(ctx context.Context, articleID uuid.UUID) error {
	article, err := c.articleRepo.GetByID(articleID)
	if err != nil {
		return fmt.Errorf("article not found: %w", err)
	}

	result, modelUsed, err := c.ClassifyArticle(ctx, article)
	if err != nil {
		return err
	}

	log.Printf("Classification result for '%s': %s (confidence: %.2f, model: %s)",
		article.Title, result.PrimaryCategory, result.Confidence, modelUsed)

	// Find the category by path
	category, err := c.categoryRepo.FindByPath(result.PrimaryCategory)
	if err != nil {
		log.Printf("Category not found: %s, keeping current category", result.PrimaryCategory)
		// Don't fail, just log and continue
	} else {
		article.CategoryID = &category.ID
	}

	// Update tags
	if len(result.SuggestedTags) > 0 {
		article.Tags = result.SuggestedTags
	}

	return c.articleRepo.Update(article)
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
