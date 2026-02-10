// eval-tagger evaluates the quality of article tags against the success matrix.
//
// Usage:
//
//	go run ./cmd/eval-tagger                          # evaluate with tags.yaml registry
//	go run ./cmd/eval-tagger --export review.md       # export human review table
//	go run ./cmd/eval-tagger --registry tags.txt      # use plain text registry instead
//	go run ./cmd/eval-tagger --limit 50 --theme web3  # filter articles
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/database"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

func main() {
	registryFile := flag.String("registry", "", "Path to plain text tag registry (one tag per line). If empty, loads config/tags.yaml")
	noRegistry := flag.Bool("no-registry", false, "Skip registry compliance check entirely")
	exportFile := flag.String("export", "", "Export human review table to Markdown file")
	reviewCount := flag.Int("review-count", 20, "Number of articles in human review export")
	limit := flag.Int("limit", 0, "Max articles to evaluate (0 = all)")
	themeFilter := flag.String("theme", "", "Filter articles by theme ID")
	flag.Parse()

	// Load config and connect to DB
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Fetch articles
	articles, err := fetchArticles(db, *limit, *themeFilter)
	if err != nil {
		log.Fatalf("Failed to fetch articles: %v", err)
	}

	if len(articles) == 0 {
		fmt.Println("No articles found. Nothing to evaluate.")
		return
	}

	fmt.Printf("Loaded %d articles for evaluation.\n\n", len(articles))

	// Load tag registry
	var registryTags []string
	if !*noRegistry {
		if *registryFile != "" {
			registryTags, err = loadPlainRegistry(*registryFile)
			if err != nil {
				log.Fatalf("Failed to load registry: %v", err)
			}
			fmt.Printf("Loaded %d tags from plain registry: %s\n\n", len(registryTags), *registryFile)
		} else {
			registryTags, err = loadTagsYAML()
			if err != nil {
				fmt.Printf("Warning: could not load tags.yaml: %v (skipping registry check)\n\n", err)
			} else {
				fmt.Printf("Loaded %d tags from tags.yaml registry\n\n", len(registryTags))
			}
		}
	}

	// Run evaluation
	evaluator := service.NewTaggerEvaluator(registryTags)
	result := evaluator.Evaluate(articles)

	// Print report
	fmt.Println(result.FormatReport())

	// Export human review if requested
	if *exportFile != "" {
		review := result.FormatHumanReview(*reviewCount)
		if err := os.WriteFile(*exportFile, []byte(review), 0644); err != nil {
			log.Fatalf("Failed to write review file: %v", err)
		}
		fmt.Printf("\nHuman review exported to: %s (%d articles)\n", *exportFile, min(*reviewCount, len(result.ArticleDetails)))
	}
}

// loadTagsYAML loads all tag names from config/tags.yaml (universal + all themes)
func loadTagsYAML() ([]string, error) {
	// Try relative paths from working directory
	candidates := []string{
		"config/tags.yaml",
		filepath.Join("..", "config", "tags.yaml"),
		filepath.Join("..", "..", "config", "tags.yaml"),
	}

	var tagsConfig *config.TagsConfig
	var loadErr error
	for _, path := range candidates {
		tagsConfig, loadErr = config.LoadTags(path)
		if loadErr == nil {
			break
		}
	}
	if tagsConfig == nil {
		return nil, fmt.Errorf("tags.yaml not found in any candidate path: %v", loadErr)
	}

	// Collect all tag names (universal + all themes)
	seen := make(map[string]bool)
	var tags []string
	for _, t := range tagsConfig.UniversalTags {
		name := strings.TrimSpace(t.Name)
		if !seen[name] {
			tags = append(tags, name)
			seen[name] = true
		}
	}
	for _, themeTags := range tagsConfig.ThemeTags {
		for _, t := range themeTags {
			name := strings.TrimSpace(t.Name)
			if !seen[name] {
				tags = append(tags, name)
				seen[name] = true
			}
		}
	}
	return tags, nil
}

func fetchArticles(db *gorm.DB, limit int, themeFilter string) ([]model.Article, error) {
	query := db.Model(&model.Article{}).Order("created_at DESC")

	if themeFilter != "" {
		query = query.Where("theme_id = ?", themeFilter)
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

func loadPlainRegistry(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open registry file: %w", err)
	}
	defer f.Close()

	var tags []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			tags = append(tags, line)
		}
	}
	return tags, scanner.Err()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
