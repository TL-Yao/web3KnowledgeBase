// bench-tagger benchmarks different tagging methods against a ground truth dataset.
//
// Usage:
//
//	go run ./cmd/bench-tagger                                    # run all methods
//	go run ./cmd/bench-tagger --method haiku-current --verbose   # run one method with detail
//	go run ./cmd/bench-tagger --dry-run                          # validate config only
//	go run ./cmd/bench-tagger --export-gt ground_truth_draft.yaml # export current tags as draft
//	go run ./cmd/bench-tagger --export results.json              # save results to JSON
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/database"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
	"gorm.io/gorm"
)

func main() {
	methodsFile := flag.String("methods", "testdata/benchmark/methods.yaml", "Methods config file")
	gtFile := flag.String("gt", "testdata/benchmark/ground_truth.yaml", "Ground truth file")
	promptsDir := flag.String("prompts", "testdata/benchmark/prompts", "Custom prompts directory")
	methodFilter := flag.String("method", "", "Run only this method ID")
	themeFilter := flag.String("theme", "", "Run only articles for this theme")
	exportFile := flag.String("export", "", "Export results to JSON file")
	exportGT := flag.String("export-gt", "", "Export current DB tags as ground truth draft")
	verbose := flag.Bool("verbose", false, "Show per-article details")
	dryRun := flag.Bool("dry-run", false, "Validate config only, don't call LLMs")
	flag.Parse()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to DB
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize key provider for dynamic API key resolution
	configRepo := repository.NewConfigRepository(db)
	keyProvider := service.NewKeyProvider(configRepo)
	claudeKeyFunc := func() string { return keyProvider.GetKey("anthropic") }
	openaiKeyFunc := func() string { return keyProvider.GetKey("openai") }

	// Handle export-gt mode
	if *exportGT != "" {
		articles, err := fetchAllArticles(db, *themeFilter)
		if err != nil {
			log.Fatalf("Failed to fetch articles: %v", err)
		}
		draft := service.ExportGroundTruthDraft(articles)
		if err := os.WriteFile(*exportGT, []byte(draft), 0644); err != nil {
			log.Fatalf("Failed to write: %v", err)
		}
		fmt.Printf("Exported %d articles to %s\n", len(articles), *exportGT)
		return
	}

	// Load tags.yaml
	tagsConfig, err := config.LoadTagsFromConfigDir()
	if err != nil {
		log.Fatalf("Failed to load tags: %v", err)
	}

	// Load ground truth
	gt, err := service.LoadGroundTruth(*gtFile)
	if err != nil {
		log.Fatalf("Failed to load ground truth: %v", err)
	}

	// Build ground truth map (title -> entry)
	gtMap := make(map[string]service.GroundTruthEntry, len(gt.Articles))
	for _, a := range gt.Articles {
		if *themeFilter != "" && a.Theme != *themeFilter {
			continue
		}
		gtMap[a.Title] = a
	}
	fmt.Printf("Ground truth: %d articles\n", len(gtMap))

	// Fetch matching articles from DB
	articles, err := fetchArticlesByTitles(db, gtMap)
	if err != nil {
		log.Fatalf("Failed to fetch articles: %v", err)
	}
	fmt.Printf("Matched articles from DB: %d\n", len(articles))

	if len(articles) == 0 {
		fmt.Println("No matching articles found. Check ground truth titles match DB.")
		return
	}

	// Load methods
	methods, err := service.LoadBenchMethods(*methodsFile)
	if err != nil {
		log.Fatalf("Failed to load methods: %v", err)
	}

	// Filter methods if specified
	var selectedMethods []service.BenchMethod
	for _, m := range methods.Methods {
		if *methodFilter == "" || m.ID == *methodFilter {
			selectedMethods = append(selectedMethods, m)
		}
	}
	if len(selectedMethods) == 0 {
		fmt.Printf("No methods matched filter '%s'. Available: ", *methodFilter)
		for _, m := range methods.Methods {
			fmt.Printf("%s ", m.ID)
		}
		fmt.Println()
		return
	}

	fmt.Printf("Methods to run: %d\n\n", len(selectedMethods))

	// Dry run: just validate config and exit
	if *dryRun {
		fmt.Println("=== DRY RUN ===")
		fmt.Printf("Ground truth entries: %d\n", len(gtMap))
		fmt.Printf("DB articles matched: %d\n", len(articles))
		fmt.Println("\nMethods:")
		for _, m := range selectedMethods {
			fmt.Printf("  - %s: model=%s, temp=%.2f, prompt=%s\n", m.ID, m.Model, m.Temperature, m.Prompt)
		}
		fmt.Println("\nGround truth articles:")
		for _, gt := range gt.Articles {
			if *themeFilter != "" && gt.Theme != *themeFilter {
				continue
			}
			fmt.Printf("  [%s] %s -> %v\n", gt.Theme, gt.Title, gt.ExpectedTags)
		}

		// Check adapter availability
		router := llm.NewRouterFromConfig(&cfg.LLM, claudeKeyFunc, openaiKeyFunc)
		fmt.Println("\nAdapter availability:")
		for _, m := range selectedMethods {
			_, ok := router.GetAdapter(m.Model)
			status := "AVAILABLE"
			if !ok {
				status = "NOT FOUND"
			}
			fmt.Printf("  %s (%s): %s\n", m.ID, m.Model, status)
		}
		return
	}

	// Create LLM router
	router := llm.NewRouterFromConfig(&cfg.LLM, claudeKeyFunc, openaiKeyFunc)

	// Create benchmark runner
	runner := service.NewBenchmarkRunner(router, tagsConfig, *promptsDir)

	// Run each method
	var allResults []*service.MethodBenchResult
	for _, method := range selectedMethods {
		fmt.Printf("Running method: %s (%s)...\n", method.ID, method.Model)
		result := runner.RunMethod(method, articles, gtMap, *verbose)
		allResults = append(allResults, result)
		fmt.Printf("  Macro-F1: %.1f%% | Errors: %d\n", result.MacroF1*100, result.ErrorCount)
	}

	// Print comparison table
	fmt.Println()
	fmt.Print(service.FormatComparisonTable(allResults))

	// Print verbose reports if requested
	if *verbose {
		for _, r := range allResults {
			fmt.Print(service.FormatVerboseReport(r))
		}
	}

	// Export JSON if requested
	if *exportFile != "" {
		data, err := json.MarshalIndent(allResults, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal results: %v", err)
		}
		if err := os.WriteFile(*exportFile, data, 0644); err != nil {
			log.Fatalf("Failed to write results: %v", err)
		}
		fmt.Printf("\nResults exported to: %s\n", *exportFile)
	}
}


func fetchArticlesByTitles(db *gorm.DB, gtMap map[string]service.GroundTruthEntry) ([]model.Article, error) {
	titles := make([]string, 0, len(gtMap))
	for t := range gtMap {
		titles = append(titles, t)
	}

	var articles []model.Article
	if err := db.Where("title IN ?", titles).Find(&articles).Error; err != nil {
		return nil, err
	}

	// Sort by theme then title for consistent ordering
	sort.Slice(articles, func(i, j int) bool {
		ti, tj := "", ""
		if articles[i].ThemeID != nil {
			ti = *articles[i].ThemeID
		}
		if articles[j].ThemeID != nil {
			tj = *articles[j].ThemeID
		}
		if ti != tj {
			return ti < tj
		}
		return articles[i].Title < articles[j].Title
	})

	// Report mismatches
	found := make(map[string]bool)
	for _, a := range articles {
		found[a.Title] = true
	}
	var missing []string
	for t := range gtMap {
		if !found[t] {
			missing = append(missing, t)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		fmt.Printf("WARNING: %d ground truth articles not found in DB:\n", len(missing))
		for _, t := range missing {
			fmt.Printf("  - %s\n", t)
		}
	}

	return articles, nil
}

func fetchAllArticles(db *gorm.DB, themeFilter string) ([]model.Article, error) {
	query := db.Order("theme_id, title")
	if themeFilter != "" {
		query = query.Where("theme_id = ?", themeFilter)
	}
	var articles []model.Article
	if err := query.Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

