package service

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/user/web3-insight/internal/config"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/model"
	"gopkg.in/yaml.v3"
)

// GroundTruthEntry represents one article's expected tags
type GroundTruthEntry struct {
	Title        string   `yaml:"title"`
	Theme        string   `yaml:"theme"`
	ExpectedTags []string `yaml:"expected_tags"`
}

// GroundTruthFile is the top-level ground truth YAML structure
type GroundTruthFile struct {
	Articles []GroundTruthEntry `yaml:"articles"`
}

// BenchMethod defines a tagging method to benchmark
type BenchMethod struct {
	ID          string  `yaml:"id"`
	Description string  `yaml:"description"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	Prompt      string  `yaml:"prompt"` // "default" or filename in prompts dir
}

// BenchMethodsFile is the top-level methods YAML structure
type BenchMethodsFile struct {
	Methods []BenchMethod `yaml:"methods"`
}

// ArticleBenchResult holds per-article benchmark results
type ArticleBenchResult struct {
	Title     string   `json:"title"`
	Theme     string   `json:"theme"`
	Predicted []string `json:"predicted"`
	Expected  []string `json:"expected"`
	TP        int      `json:"tp"`
	FP        int      `json:"fp"`
	FN        int      `json:"fn"`
	Precision float64  `json:"precision"`
	Recall    float64  `json:"recall"`
	F1        float64  `json:"f1"`
	Error     string   `json:"error,omitempty"`
}

// MethodBenchResult holds aggregate results for one method
type MethodBenchResult struct {
	MethodID       string               `json:"methodId"`
	Description    string               `json:"description"`
	Model          string               `json:"model"`
	Articles       []ArticleBenchResult `json:"articles"`
	MacroPrecision float64              `json:"macroPrecision"`
	MacroRecall    float64              `json:"macroRecall"`
	MacroF1        float64              `json:"macroF1"`
	MicroPrecision float64              `json:"microPrecision"`
	MicroRecall    float64              `json:"microRecall"`
	MicroF1        float64              `json:"microF1"`
	AvgTags        float64              `json:"avgTags"`
	RegistryRate   float64              `json:"registryRate"`
	ErrorCount     int                  `json:"errorCount"`
}

// BenchmarkRunner runs tagging benchmarks
type BenchmarkRunner struct {
	router    *llm.Router
	tagsConfig *config.TagsConfig
	promptsDir string
}

// NewBenchmarkRunner creates a new benchmark runner
func NewBenchmarkRunner(router *llm.Router, tagsConfig *config.TagsConfig, promptsDir string) *BenchmarkRunner {
	return &BenchmarkRunner{
		router:     router,
		tagsConfig: tagsConfig,
		promptsDir: promptsDir,
	}
}

// LoadGroundTruth loads the ground truth YAML file
func LoadGroundTruth(path string) (*GroundTruthFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ground truth: %w", err)
	}
	var gt GroundTruthFile
	if err := yaml.Unmarshal(data, &gt); err != nil {
		return nil, fmt.Errorf("parse ground truth: %w", err)
	}
	return &gt, nil
}

// LoadBenchMethods loads the methods YAML file
func LoadBenchMethods(path string) (*BenchMethodsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read methods: %w", err)
	}
	var mf BenchMethodsFile
	if err := yaml.Unmarshal(data, &mf); err != nil {
		return nil, fmt.Errorf("parse methods: %w", err)
	}
	return &mf, nil
}

// buildCanonicalTagMap builds the canonical tag lookup from the tags config for a given theme
func (br *BenchmarkRunner) buildCanonicalTagMap(themeID string) (map[string]string, []model.Tag, []model.Tag) {
	universal, theme := br.tagsConfig.GetTagsForTheme(themeID)

	// Convert config.TagEntry to model.Tag for prompt building
	var universalTags []model.Tag
	for _, t := range universal {
		universalTags = append(universalTags, model.Tag{Name: t.Name, NameEn: t.NameEn})
	}
	var themeTags []model.Tag
	for _, t := range theme {
		themeTags = append(themeTags, model.Tag{Name: t.Name, NameEn: t.NameEn})
	}

	canonical := make(map[string]string, (len(universal)+len(theme))*2)
	for _, t := range universal {
		canonical[strings.ToLower(t.Name)] = t.Name
		if t.NameEn != "" {
			canonical[strings.ToLower(t.NameEn)] = t.Name
		}
	}
	for _, t := range theme {
		canonical[strings.ToLower(t.Name)] = t.Name
		if t.NameEn != "" {
			canonical[strings.ToLower(t.NameEn)] = t.Name
		}
	}

	return canonical, universalTags, themeTags
}

// getPromptTemplate returns the template string for a method
func (br *BenchmarkRunner) getPromptTemplate(method BenchMethod) (string, error) {
	if method.Prompt == "" || method.Prompt == "default" {
		return TagPromptTemplate, nil
	}

	path := br.promptsDir + "/" + method.Prompt + ".tmpl"
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("load custom prompt %s: %w", path, err)
	}

	// Validate the template parses
	if _, err := template.New("test").Parse(string(data)); err != nil {
		return "", fmt.Errorf("invalid template %s: %w", path, err)
	}

	return string(data), nil
}

// RunMethod runs a single benchmark method against all articles
func (br *BenchmarkRunner) RunMethod(method BenchMethod, articles []model.Article, gtMap map[string]GroundTruthEntry, verbose bool) *MethodBenchResult {
	result := &MethodBenchResult{
		MethodID:    method.ID,
		Description: method.Description,
		Model:       method.Model,
	}

	promptTemplate, err := br.getPromptTemplate(method)
	if err != nil {
		log.Printf("[%s] Failed to load prompt: %v", method.ID, err)
		result.ErrorCount = len(articles)
		return result
	}

	var totalTP, totalFP, totalFN int
	var sumP, sumR, sumF1 float64
	var totalPredicted int
	var totalRegistryOK int
	var totalTagAssignments int
	scored := 0

	for _, article := range articles {
		gt, ok := gtMap[article.Title]
		if !ok {
			continue
		}

		themeID := gt.Theme
		canonical, universalTags, themeTags := br.buildCanonicalTagMap(themeID)

		// Build prompt
		prompt, err := BuildTagPromptFromTemplate(promptTemplate, &article, themeID, universalTags, themeTags)
		if err != nil {
			result.Articles = append(result.Articles, ArticleBenchResult{
				Title:    article.Title,
				Theme:    themeID,
				Expected: gt.ExpectedTags,
				Error:    fmt.Sprintf("prompt build: %v", err),
			})
			result.ErrorCount++
			continue
		}

		// Call LLM
		opts := &llm.GenerateOptions{
			Temperature: method.Temperature,
			MaxTokens:   method.MaxTokens,
		}
		response, err := br.router.GenerateWithModel(method.Model, prompt, opts)
		// Rate limit: wait 3s between cloud API calls to avoid 429
		time.Sleep(3 * time.Second)
		if err != nil {
			result.Articles = append(result.Articles, ArticleBenchResult{
				Title:    article.Title,
				Theme:    themeID,
				Expected: gt.ExpectedTags,
				Error:    fmt.Sprintf("LLM: %v", err),
			})
			result.ErrorCount++
			continue
		}

		// Parse response
		tagResult, err := ParseTaggingResponse(response)
		if err != nil {
			result.Articles = append(result.Articles, ArticleBenchResult{
				Title:    article.Title,
				Theme:    themeID,
				Expected: gt.ExpectedTags,
				Error:    fmt.Sprintf("parse: %v", err),
			})
			result.ErrorCount++
			continue
		}

		// Resolve tags through canonical map
		seen := make(map[string]bool)
		var predicted []string
		for _, tag := range tagResult.Tags {
			tag = strings.TrimSpace(tag)
			totalTagAssignments++
			c := ResolveTag(tag, canonical)
			if c != "" {
				totalRegistryOK++
				if !seen[c] {
					predicted = append(predicted, c)
					seen[c] = true
				}
			}
		}

		// Calculate precision/recall/F1
		ar := computeArticleMetrics(article.Title, themeID, predicted, gt.ExpectedTags)
		result.Articles = append(result.Articles, ar)

		totalTP += ar.TP
		totalFP += ar.FP
		totalFN += ar.FN
		sumP += ar.Precision
		sumR += ar.Recall
		sumF1 += ar.F1
		totalPredicted += len(predicted)
		scored++

		if verbose {
			log.Printf("[%s] %s: P=%.0f%% R=%.0f%% F1=%.0f%% predicted=%v",
				method.ID, truncateString(article.Title, 30),
				ar.Precision*100, ar.Recall*100, ar.F1*100, predicted)
		}
	}

	// Aggregate metrics
	if scored > 0 {
		result.MacroPrecision = sumP / float64(scored)
		result.MacroRecall = sumR / float64(scored)
		result.MacroF1 = sumF1 / float64(scored)
		result.AvgTags = float64(totalPredicted) / float64(scored)
	}

	if totalTP+totalFP > 0 {
		result.MicroPrecision = float64(totalTP) / float64(totalTP+totalFP)
	}
	if totalTP+totalFN > 0 {
		result.MicroRecall = float64(totalTP) / float64(totalTP+totalFN)
	}
	if result.MicroPrecision+result.MicroRecall > 0 {
		result.MicroF1 = 2 * result.MicroPrecision * result.MicroRecall / (result.MicroPrecision + result.MicroRecall)
	}

	if totalTagAssignments > 0 {
		result.RegistryRate = float64(totalRegistryOK) / float64(totalTagAssignments)
	}

	return result
}

// computeArticleMetrics calculates TP/FP/FN/P/R/F1 for a single article
func computeArticleMetrics(title, theme string, predicted, expected []string) ArticleBenchResult {
	expectedSet := make(map[string]bool, len(expected))
	for _, t := range expected {
		expectedSet[t] = true
	}

	var tp, fp int
	for _, t := range predicted {
		if expectedSet[t] {
			tp++
		} else {
			fp++
		}
	}
	fn := len(expected) - tp

	var precision, recall, f1 float64
	if tp+fp > 0 {
		precision = float64(tp) / float64(tp+fp)
	}
	if tp+fn > 0 {
		recall = float64(tp) / float64(tp+fn)
	}
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	return ArticleBenchResult{
		Title:     title,
		Theme:     theme,
		Predicted: predicted,
		Expected:  expected,
		TP:        tp,
		FP:        fp,
		FN:        fn,
		Precision: precision,
		Recall:    recall,
		F1:        f1,
	}
}

// FormatComparisonTable formats a comparison table across all methods
func FormatComparisonTable(results []*MethodBenchResult) string {
	var b strings.Builder

	b.WriteString("┌───────────────────┬──────────┬──────────┬──────────┬──────────┬──────────┬──────────┐\n")
	b.WriteString("│ Method            │ Macro-P  │ Macro-R  │ Macro-F1 │ Avg Tags │ Reg Rate │ Errors   │\n")
	b.WriteString("├───────────────────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┤\n")

	bestF1 := 0.0
	bestMethod := ""
	for _, r := range results {
		if r.MacroF1 > bestF1 {
			bestF1 = r.MacroF1
			bestMethod = r.MethodID
		}
		name := r.MethodID
		if len(name) > 17 {
			name = name[:17]
		}
		fmt.Fprintf(&b, "│ %-17s │  %5.1f%%  │  %5.1f%%  │  %5.1f%%  │  %5.1f   │  %5.1f%%  │  %3d     │\n",
			name,
			r.MacroPrecision*100, r.MacroRecall*100, r.MacroF1*100,
			r.AvgTags, r.RegistryRate*100, r.ErrorCount)
	}

	b.WriteString("└───────────────────┴──────────┴──────────┴──────────┴──────────┴──────────┴──────────┘\n")

	if bestMethod != "" {
		fmt.Fprintf(&b, "Best Macro-F1: %s (%.1f%%)\n", bestMethod, bestF1*100)
	}

	return b.String()
}

// FormatVerboseReport formats detailed per-article results for a method
func FormatVerboseReport(r *MethodBenchResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "\n=== %s (%s) ===\n", r.MethodID, r.Description)
	fmt.Fprintf(&b, "Model: %s | Macro-F1: %.1f%% | Errors: %d\n\n", r.Model, r.MacroF1*100, r.ErrorCount)

	for i, ar := range r.Articles {
		title := ar.Title
		if len([]rune(title)) > 40 {
			title = string([]rune(title)[:40]) + "..."
		}

		if ar.Error != "" {
			fmt.Fprintf(&b, "%2d. [ERROR] %s\n    %s\n", i+1, title, ar.Error)
			continue
		}

		fmt.Fprintf(&b, "%2d. %s\n", i+1, title)
		fmt.Fprintf(&b, "    Expected:  %s\n", strings.Join(ar.Expected, ", "))
		fmt.Fprintf(&b, "    Predicted: %s\n", strings.Join(ar.Predicted, ", "))
		fmt.Fprintf(&b, "    P=%.0f%% R=%.0f%% F1=%.0f%% (TP=%d FP=%d FN=%d)\n",
			ar.Precision*100, ar.Recall*100, ar.F1*100, ar.TP, ar.FP, ar.FN)
	}

	return b.String()
}

// ExportGroundTruthDraft exports current DB tags as a ground truth draft YAML
func ExportGroundTruthDraft(articles []model.Article) string {
	var b strings.Builder

	b.WriteString("# Ground truth draft - exported from current DB tags\n")
	b.WriteString("# Review and adjust expected_tags before using as benchmark\n\n")
	b.WriteString("articles:\n")

	lastTheme := ""
	for _, a := range articles {
		themeID := ""
		if a.ThemeID != nil {
			themeID = *a.ThemeID
		}

		if themeID != lastTheme {
			fmt.Fprintf(&b, "\n  # === %s ===\n", themeID)
			lastTheme = themeID
		}

		tags := []string(a.Tags)
		fmt.Fprintf(&b, "  - title: %q\n", a.Title)
		fmt.Fprintf(&b, "    theme: %s\n", themeID)
		fmt.Fprintf(&b, "    expected_tags: [%s]\n", formatYAMLTags(tags))
	}

	return b.String()
}

func formatYAMLTags(tags []string) string {
	quoted := make([]string, len(tags))
	for i, t := range tags {
		quoted[i] = fmt.Sprintf("%q", t)
	}
	return strings.Join(quoted, ", ")
}
