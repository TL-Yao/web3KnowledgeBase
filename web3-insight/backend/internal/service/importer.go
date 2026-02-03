package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// ArticleImporter handles batch article import
type ArticleImporter struct {
	articleRepo  *repository.ArticleRepository
	categoryRepo *repository.CategoryRepository
}

// ImportArticle represents the JSON structure for importing an article
type ImportArticle struct {
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	ContentHTML  string   `json:"contentHtml,omitempty"`
	Summary      string   `json:"summary,omitempty"`
	CategoryPath string   `json:"categoryPath,omitempty"` // e.g., "基础技术/区块链原理"
	CategoryID   string   `json:"categoryId,omitempty"`   // Direct UUID if known
	Tags         []string `json:"tags,omitempty"`
	Status       string   `json:"status,omitempty"` // draft, published
	SourceURLs   []string `json:"sourceUrls,omitempty"`
	Slug         string   `json:"slug,omitempty"` // Custom slug, auto-generated if empty
}

// ImportBatch represents a batch of articles to import
type ImportBatch struct {
	Articles []ImportArticle `json:"articles"`
	Options  ImportOptions   `json:"options,omitempty"`
}

// ImportOptions configures import behavior
type ImportOptions struct {
	SkipDuplicates  bool `json:"skipDuplicates"`  // Skip articles with same slug
	UpdateExisting  bool `json:"updateExisting"`  // Update if slug exists
	GenerateSummary bool `json:"generateSummary"` // Generate summary if empty
	DefaultStatus   string `json:"defaultStatus"` // Default status for articles
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	TotalCount   int            `json:"totalCount"`
	ImportedCount int           `json:"importedCount"`
	SkippedCount int            `json:"skippedCount"`
	UpdatedCount int            `json:"updatedCount"`
	ErrorCount   int            `json:"errorCount"`
	Errors       []ImportError  `json:"errors,omitempty"`
	ImportedIDs  []uuid.UUID    `json:"importedIds,omitempty"`
}

// ImportError describes an error during import
type ImportError struct {
	Index   int    `json:"index"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// NewArticleImporter creates a new article importer
func NewArticleImporter(articleRepo *repository.ArticleRepository, categoryRepo *repository.CategoryRepository) *ArticleImporter {
	return &ArticleImporter{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
	}
}

// Import imports a batch of articles
func (i *ArticleImporter) Import(batch ImportBatch) (*ImportResult, error) {
	result := &ImportResult{
		TotalCount: len(batch.Articles),
		Errors:     []ImportError{},
		ImportedIDs: []uuid.UUID{},
	}

	for idx, importArticle := range batch.Articles {
		if err := i.importSingle(importArticle, batch.Options, result); err != nil {
			result.Errors = append(result.Errors, ImportError{
				Index:   idx,
				Title:   importArticle.Title,
				Message: err.Error(),
			})
			result.ErrorCount++
		}
	}

	return result, nil
}

// importSingle imports a single article
func (i *ArticleImporter) importSingle(importArticle ImportArticle, opts ImportOptions, result *ImportResult) error {
	// Validate required fields
	if importArticle.Title == "" {
		return fmt.Errorf("title is required")
	}
	if importArticle.Content == "" {
		return fmt.Errorf("content is required")
	}

	// Generate or use provided slug
	articleSlug := importArticle.Slug
	if articleSlug == "" {
		articleSlug = slug.Make(importArticle.Title)
	}

	// Check for existing article by slug
	existing, err := i.articleRepo.GetBySlug(articleSlug)
	if err == nil && existing != nil {
		if opts.SkipDuplicates {
			result.SkippedCount++
			return nil
		}
		if opts.UpdateExisting {
			// Update existing article
			existing.Title = importArticle.Title
			existing.Content = importArticle.Content
			if importArticle.ContentHTML != "" {
				existing.ContentHTML = importArticle.ContentHTML
			}
			if importArticle.Summary != "" {
				existing.Summary = importArticle.Summary
			}
			if len(importArticle.Tags) > 0 {
				existing.Tags = importArticle.Tags
			}
			if len(importArticle.SourceURLs) > 0 {
				existing.SourceURLs = importArticle.SourceURLs
			}
			if importArticle.Status != "" {
				existing.Status = importArticle.Status
			}

			if err := i.articleRepo.Update(existing); err != nil {
				return fmt.Errorf("failed to update article: %w", err)
			}
			result.UpdatedCount++
			return nil
		}
		// Make slug unique by appending a number
		articleSlug = i.makeUniqueSlug(articleSlug)
	}

	// Resolve category
	var categoryID *uuid.UUID
	if importArticle.CategoryID != "" {
		if parsed, err := uuid.Parse(importArticle.CategoryID); err == nil {
			categoryID = &parsed
		}
	} else if importArticle.CategoryPath != "" {
		// Use FindOrCreateByPath to auto-create missing categories
		category, created, err := i.categoryRepo.FindOrCreateByPath(importArticle.CategoryPath)
		if err == nil && category != nil {
			categoryID = &category.ID
			if created {
				log.Printf("Auto-created category path: %s", importArticle.CategoryPath)
			}
		}
	}

	// Set default status
	status := importArticle.Status
	if status == "" {
		status = opts.DefaultStatus
		if status == "" {
			status = "draft"
		}
	}

	// Create article
	article := &model.Article{
		Title:       importArticle.Title,
		Slug:        articleSlug,
		Content:     importArticle.Content,
		ContentHTML: importArticle.ContentHTML,
		Summary:     importArticle.Summary,
		CategoryID:  categoryID,
		Tags:        importArticle.Tags,
		Status:      status,
		SourceURLs:  importArticle.SourceURLs,
	}

	if err := i.articleRepo.Create(article); err != nil {
		return fmt.Errorf("failed to create article: %w", err)
	}

	result.ImportedCount++
	result.ImportedIDs = append(result.ImportedIDs, article.ID)
	return nil
}

// makeUniqueSlug creates a unique slug by appending a number
func (i *ArticleImporter) makeUniqueSlug(baseSlug string) string {
	count := i.articleRepo.CountBySlugPrefix(baseSlug)
	return fmt.Sprintf("%s-%d", baseSlug, count+1)
}

// ParseJSON parses JSON data into ImportBatch
func (i *ArticleImporter) ParseJSON(data []byte) (*ImportBatch, error) {
	var batch ImportBatch
	if err := json.Unmarshal(data, &batch); err != nil {
		// Try parsing as array directly
		var articles []ImportArticle
		if err2 := json.Unmarshal(data, &articles); err2 != nil {
			return nil, fmt.Errorf("invalid JSON format: %w", err)
		}
		batch.Articles = articles
	}
	return &batch, nil
}

// GenerateTemplate returns a JSON template for article import
func (i *ArticleImporter) GenerateTemplate() string {
	template := ImportBatch{
		Articles: []ImportArticle{
			{
				Title:        "文章标题示例",
				Content:      "文章的 Markdown 内容...",
				Summary:      "简短的摘要描述",
				CategoryPath: "基础技术/区块链原理",
				Tags:         []string{"区块链", "技术"},
				Status:       "draft",
				SourceURLs:   []string{"https://example.com/source"},
			},
		},
		Options: ImportOptions{
			SkipDuplicates: true,
			UpdateExisting: false,
			DefaultStatus:  "draft",
		},
	}

	data, _ := json.MarshalIndent(template, "", "  ")
	return string(data)
}

// ValidateImport validates import data without actually importing
func (i *ArticleImporter) ValidateImport(batch ImportBatch) []ImportError {
	var errors []ImportError

	for idx, article := range batch.Articles {
		if article.Title == "" {
			errors = append(errors, ImportError{
				Index:   idx,
				Title:   "(empty)",
				Message: "title is required",
			})
		}
		if article.Content == "" {
			errors = append(errors, ImportError{
				Index:   idx,
				Title:   article.Title,
				Message: "content is required",
			})
		}
		if article.Status != "" && article.Status != "draft" && article.Status != "published" {
			errors = append(errors, ImportError{
				Index:   idx,
				Title:   article.Title,
				Message: fmt.Sprintf("invalid status: %s (must be 'draft' or 'published')", article.Status),
			})
		}
	}

	return errors
}

// ExportArticles exports articles to ImportBatch format
func (i *ArticleImporter) ExportArticles(articleIDs []uuid.UUID) (*ImportBatch, error) {
	batch := &ImportBatch{
		Articles: []ImportArticle{},
	}

	for _, id := range articleIDs {
		article, err := i.articleRepo.GetByID(id)
		if err != nil {
			continue
		}

		importArticle := ImportArticle{
			Title:       article.Title,
			Content:     article.Content,
			ContentHTML: article.ContentHTML,
			Summary:     article.Summary,
			Tags:        article.Tags,
			Status:      article.Status,
			SourceURLs:  article.SourceURLs,
			Slug:        article.Slug,
		}

		// Get category path if available
		if article.CategoryID != nil {
			if category, err := i.categoryRepo.GetByID(*article.CategoryID); err == nil {
				importArticle.CategoryPath = i.buildCategoryPath(category)
			}
		}

		batch.Articles = append(batch.Articles, importArticle)
	}

	return batch, nil
}

// buildCategoryPath builds the full path for a category
func (i *ArticleImporter) buildCategoryPath(category *model.Category) string {
	if category == nil {
		return ""
	}

	path := category.Name
	currentID := category.ParentID

	// Walk up the tree
	for currentID != nil {
		parent, err := i.categoryRepo.GetByID(*currentID)
		if err != nil {
			break
		}
		path = parent.Name + "/" + path
		currentID = parent.ParentID
	}

	return path
}

// ImportFromMarkdown imports a single article from markdown content
func (i *ArticleImporter) ImportFromMarkdown(title, markdown string, categoryPath string, tags []string) (*model.Article, error) {
	importArticle := ImportArticle{
		Title:        title,
		Content:      markdown,
		CategoryPath: categoryPath,
		Tags:         tags,
		Status:       "draft",
	}

	batch := ImportBatch{
		Articles: []ImportArticle{importArticle},
	}

	result, err := i.Import(batch)
	if err != nil {
		return nil, err
	}

	if result.ErrorCount > 0 {
		return nil, fmt.Errorf("import failed: %s", result.Errors[0].Message)
	}

	if len(result.ImportedIDs) == 0 {
		return nil, fmt.Errorf("article was not imported")
	}

	return i.articleRepo.GetByID(result.ImportedIDs[0])
}

// BatchExportToJSON exports multiple articles to JSON
func (i *ArticleImporter) BatchExportToJSON(categoryID *uuid.UUID, status string) ([]byte, error) {
	// Fetch articles based on filters
	params := repository.ArticleListParams{
		CategoryID: categoryID,
		Status:     status,
		PageSize:   1000, // Export limit
	}

	result, err := i.articleRepo.List(params)
	if err != nil {
		return nil, err
	}

	articleIDs := make([]uuid.UUID, len(result.Articles))
	for idx, article := range result.Articles {
		articleIDs[idx] = article.ID
	}

	batch, err := i.ExportArticles(articleIDs)
	if err != nil {
		return nil, err
	}

	batch.Options = ImportOptions{
		SkipDuplicates: true,
		DefaultStatus:  "draft",
	}

	return json.MarshalIndent(batch, "", "  ")
}
