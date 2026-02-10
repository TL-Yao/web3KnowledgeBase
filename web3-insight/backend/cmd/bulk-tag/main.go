// bulk-tag runs the tagger on all untagged articles (or all articles with --force).
//
// Usage:
//
//	go run ./cmd/bulk-tag              # tag articles with empty tags
//	go run ./cmd/bulk-tag --force      # re-tag all articles
//	go run ./cmd/bulk-tag --limit 10   # tag at most 10 articles
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/database"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

func main() {
	force := flag.Bool("force", false, "Re-tag all articles (not just untagged)")
	limit := flag.Int("limit", 0, "Max articles to tag (0 = all)")
	flag.Parse()

	// Load config and connect to DB
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Expand env vars in API keys (config.yaml uses ${ANTHROPIC_API_KEY} syntax)
	cfg.LLM.Claude.APIKey = os.ExpandEnv(cfg.LLM.Claude.APIKey)
	cfg.LLM.OpenAI.APIKey = os.ExpandEnv(cfg.LLM.OpenAI.APIKey)

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Fetch articles to tag
	articles, err := fetchArticlesToTag(db, *force, *limit)
	if err != nil {
		log.Fatalf("Failed to fetch articles: %v", err)
	}

	if len(articles) == 0 {
		fmt.Println("No articles to tag.")
		return
	}

	fmt.Printf("Found %d articles to tag.\n\n", len(articles))

	// Initialize LLM router
	llmRouter := llm.NewRouterFromConfig(&cfg.LLM)

	// Create tagger
	tagRepo := repository.NewTagRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	tagger := service.NewTagger(llmRouter, tagRepo, articleRepo)

	// Tag each article
	ctx := context.Background()
	success := 0
	failed := 0
	for i, article := range articles {
		fmt.Printf("[%d/%d] Tagging: %s ... ", i+1, len(articles), article.Title)
		start := time.Now()

		if err := tagger.TagArticle(ctx, &article); err != nil {
			fmt.Printf("FAILED (%v)\n", err)
			failed++
			continue
		}

		// Reload to see updated tags
		var updated model.Article
		db.First(&updated, "id = ?", article.ID)

		elapsed := time.Since(start)
		fmt.Printf("OK (%d tags, %.1fs)\n", len(updated.Tags), elapsed.Seconds())
		for _, tag := range updated.Tags {
			fmt.Printf("  - %s\n", tag)
		}
		success++

		// Brief pause to avoid rate limiting
		if i < len(articles)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	fmt.Printf("\nDone! Tagged %d/%d articles (%d failed)\n", success, len(articles), failed)
}

func fetchArticlesToTag(db *gorm.DB, force bool, limit int) ([]model.Article, error) {
	query := db.Model(&model.Article{}).Order("created_at ASC")

	if !force {
		// Only untagged articles (empty array or null)
		query = query.Where("tags IS NULL OR array_length(tags, 1) IS NULL OR array_length(tags, 1) = 0")
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	var articles []model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}
