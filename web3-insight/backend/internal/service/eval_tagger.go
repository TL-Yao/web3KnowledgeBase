package service

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/user/web3-insight/internal/model"
)

// TaggerEvaluator calculates quality metrics for article tags
type TaggerEvaluator struct {
	registry map[string]bool // known tag names (lowercased)
}

// NewTaggerEvaluator creates an evaluator with a tag registry for compliance checks.
// If registry is nil, registry compliance check is skipped.
func NewTaggerEvaluator(registryTags []string) *TaggerEvaluator {
	reg := make(map[string]bool, len(registryTags))
	for _, t := range registryTags {
		reg[strings.ToLower(strings.TrimSpace(t))] = true
	}
	return &TaggerEvaluator{registry: reg}
}

// EvalResult holds all metrics from an evaluation run
type EvalResult struct {
	TotalArticles int
	ArticlesWithTags int
	ArticlesNoTags   int

	// Tag count metrics
	AvgTagsPerArticle float64
	MinTags           int
	MaxTags           int
	InRangeCount      int // articles with 3-7 tags
	InRangeRate       float64

	// Registry compliance
	TotalTagAssignments int
	OnRegistryCount     int
	OffRegistryCount    int
	RegistryRate        float64
	OffRegistryTags     []string

	// Reuse metrics
	UniqueTags       int
	ReusedTags       int // tags used by ≥3 articles
	ReusedRate       float64
	OrphanTags       int // tags used by only 1 article
	OrphanRate       float64

	// Distribution
	TopTags          []TagCount // top 10 most used
	MaxTagCoverage   float64   // highest single tag's % of articles
	GiniCoefficient  float64

	// Per-article details (for human review export)
	ArticleDetails []ArticleEvalDetail
}

// TagCount holds a tag and its usage count
type TagCount struct {
	Tag   string
	Count int
}

// ArticleEvalDetail holds per-article evaluation info
type ArticleEvalDetail struct {
	ID      string
	Title   string
	ThemeID string
	Tags    []string
	TagCount int
	OffRegistryTags []string
}

// Evaluate runs all metrics on a set of articles
func (e *TaggerEvaluator) Evaluate(articles []model.Article) *EvalResult {
	r := &EvalResult{
		TotalArticles: len(articles),
		MinTags:       math.MaxInt32,
	}

	if len(articles) == 0 {
		r.MinTags = 0
		return r
	}

	tagUsage := make(map[string]int)    // tag -> number of articles using it
	totalTags := 0

	for _, a := range articles {
		tags := []string(a.Tags)
		n := len(tags)

		detail := ArticleEvalDetail{
			ID:       a.ID.String(),
			Title:    a.Title,
			Tags:     tags,
			TagCount: n,
		}
		if a.ThemeID != nil {
			detail.ThemeID = *a.ThemeID
		}

		if n == 0 {
			r.ArticlesNoTags++
		} else {
			r.ArticlesWithTags++
		}

		totalTags += n

		if n < r.MinTags {
			r.MinTags = n
		}
		if n > r.MaxTags {
			r.MaxTags = n
		}
		if n >= 3 && n <= 7 {
			r.InRangeCount++
		}

		// Check each tag
		for _, tag := range tags {
			norm := strings.ToLower(strings.TrimSpace(tag))
			tagUsage[norm]++
			r.TotalTagAssignments++

			if len(e.registry) > 0 {
				if e.registry[norm] {
					r.OnRegistryCount++
				} else {
					r.OffRegistryCount++
					detail.OffRegistryTags = append(detail.OffRegistryTags, tag)
				}
			}
		}

		r.ArticleDetails = append(r.ArticleDetails, detail)
	}

	// Averages
	r.AvgTagsPerArticle = float64(totalTags) / float64(len(articles))
	r.InRangeRate = float64(r.InRangeCount) / float64(len(articles))

	// Registry rate
	if r.TotalTagAssignments > 0 && len(e.registry) > 0 {
		r.RegistryRate = float64(r.OnRegistryCount) / float64(r.TotalTagAssignments)
	}

	// Reuse and orphan metrics
	r.UniqueTags = len(tagUsage)
	for _, count := range tagUsage {
		if count >= 3 {
			r.ReusedTags++
		}
		if count == 1 {
			r.OrphanTags++
		}
	}
	if r.UniqueTags > 0 {
		r.ReusedRate = float64(r.ReusedTags) / float64(r.UniqueTags)
		r.OrphanRate = float64(r.OrphanTags) / float64(r.UniqueTags)
	}

	// Top tags and distribution
	var tagCounts []TagCount
	for tag, count := range tagUsage {
		tagCounts = append(tagCounts, TagCount{Tag: tag, Count: count})
	}
	sort.Slice(tagCounts, func(i, j int) bool {
		return tagCounts[i].Count > tagCounts[j].Count
	})

	limit := 10
	if len(tagCounts) < limit {
		limit = len(tagCounts)
	}
	r.TopTags = tagCounts[:limit]

	if len(tagCounts) > 0 {
		r.MaxTagCoverage = float64(tagCounts[0].Count) / float64(len(articles))
	}

	// Gini coefficient for tag distribution evenness
	r.GiniCoefficient = calcGini(tagCounts)

	// Collect off-registry tags (deduplicated)
	if len(e.registry) > 0 {
		offSet := make(map[string]bool)
		for _, d := range r.ArticleDetails {
			for _, t := range d.OffRegistryTags {
				offSet[t] = true
			}
		}
		for t := range offSet {
			r.OffRegistryTags = append(r.OffRegistryTags, t)
		}
		sort.Strings(r.OffRegistryTags)
	}

	return r
}

// calcGini computes Gini coefficient for tag usage distribution.
// 0 = perfectly even, 1 = maximally uneven.
func calcGini(tags []TagCount) float64 {
	n := len(tags)
	if n <= 1 {
		return 0
	}

	// Sort ascending by count for Gini calculation
	sorted := make([]int, n)
	for i, tc := range tags {
		sorted[i] = tc.Count
	}
	sort.Ints(sorted)

	var sumOfDiffs float64
	var total float64
	for i := 0; i < n; i++ {
		total += float64(sorted[i])
		for j := 0; j < n; j++ {
			sumOfDiffs += math.Abs(float64(sorted[i]) - float64(sorted[j]))
		}
	}

	if total == 0 {
		return 0
	}
	return sumOfDiffs / (2 * float64(n) * total)
}

// FormatReport generates a formatted text report with pass/fail indicators
func (r *EvalResult) FormatReport() string {
	var b strings.Builder

	b.WriteString("╔══════════════════════════════════════════════════════╗\n")
	b.WriteString("║          TAGGER EVALUATION REPORT                   ║\n")
	b.WriteString("╚══════════════════════════════════════════════════════╝\n\n")

	fmt.Fprintf(&b, "Articles evaluated: %d (with tags: %d, without: %d)\n\n", r.TotalArticles, r.ArticlesWithTags, r.ArticlesNoTags)

	b.WriteString("── Tag Count ──────────────────────────────────────────\n")
	fmt.Fprintf(&b, "  Avg tags/article:   %.1f\n", r.AvgTagsPerArticle)
	fmt.Fprintf(&b, "  Range:              %d - %d\n", r.MinTags, r.MaxTags)
	fmt.Fprintf(&b, "  In range (3-7):     %d/%d (%.0f%%)  %s\n",
		r.InRangeCount, r.TotalArticles, r.InRangeRate*100, passFail(r.InRangeRate >= 0.95, ">95%"))

	b.WriteString("\n── Registry Compliance ─────────────────────────────────\n")
	if len(r.OffRegistryTags) > 0 || r.OnRegistryCount > 0 {
		fmt.Fprintf(&b, "  On-registry:        %d/%d (%.0f%%)  %s\n",
			r.OnRegistryCount, r.TotalTagAssignments, r.RegistryRate*100, passFail(r.RegistryRate >= 0.95, ">95%"))
		if len(r.OffRegistryTags) > 0 {
			fmt.Fprintf(&b, "  Off-registry tags:  %s\n", strings.Join(r.OffRegistryTags, ", "))
		}
	} else {
		b.WriteString("  (no registry loaded — skipped)\n")
	}

	b.WriteString("\n── Reuse & Orphans ─────────────────────────────────────\n")
	fmt.Fprintf(&b, "  Unique tags:        %d\n", r.UniqueTags)
	fmt.Fprintf(&b, "  Reused (≥3 arts):   %d/%d (%.0f%%)  %s\n",
		r.ReusedTags, r.UniqueTags, r.ReusedRate*100, passFail(r.ReusedRate >= 0.80, ">80%"))
	fmt.Fprintf(&b, "  Orphans (1 art):    %d/%d (%.0f%%)  %s\n",
		r.OrphanTags, r.UniqueTags, r.OrphanRate*100, passFail(r.OrphanRate <= 0.15, "<15%"))

	b.WriteString("\n── Distribution ────────────────────────────────────────\n")
	fmt.Fprintf(&b, "  Max tag coverage:   %.0f%%  %s\n",
		r.MaxTagCoverage*100, passFail(r.MaxTagCoverage <= 0.40, "≤40%"))
	fmt.Fprintf(&b, "  Gini coefficient:   %.2f  %s\n",
		r.GiniCoefficient, passFail(r.GiniCoefficient <= 0.60, "≤0.60"))
	if len(r.TopTags) > 0 {
		b.WriteString("  Top tags:\n")
		for i, tc := range r.TopTags {
			fmt.Fprintf(&b, "    %2d. %-30s  %d articles\n", i+1, tc.Tag, tc.Count)
		}
	}

	b.WriteString("\n── Empty/Error Rate ────────────────────────────────────\n")
	emptyRate := float64(r.ArticlesNoTags) / float64(r.TotalArticles)
	fmt.Fprintf(&b, "  Articles with no tags: %d/%d (%.0f%%)  %s\n",
		r.ArticlesNoTags, r.TotalArticles, emptyRate*100, passFail(r.ArticlesNoTags == 0, "0%"))

	return b.String()
}

// FormatHumanReview generates a Markdown table for spot-checking.
// If maxArticles > 0, it stratifies by theme and picks up to that many.
func (r *EvalResult) FormatHumanReview(maxArticles int) string {
	details := r.ArticleDetails
	if maxArticles > 0 && len(details) > maxArticles {
		details = stratifyByTheme(details, maxArticles)
	}

	var b strings.Builder
	b.WriteString("# Tagger Human Review\n\n")
	b.WriteString("Rating: ✓ = correct, △ = mostly correct (1-2 questionable), ✗ = wrong/misleading\n\n")
	b.WriteString("| # | Theme | Title | Tags | Rating |\n")
	b.WriteString("|---|-------|-------|------|--------|\n")

	for i, d := range details {
		title := truncateRunes(d.Title, 25)
		tags := truncateRunes(strings.Join(d.Tags, ", "), 40)
		fmt.Fprintf(&b, "| %d | %s | %s | %s | |\n", i+1, d.ThemeID, title, tags)
	}

	return b.String()
}

// stratifyByTheme picks articles evenly across themes
func stratifyByTheme(details []ArticleEvalDetail, max int) []ArticleEvalDetail {
	byTheme := make(map[string][]ArticleEvalDetail)
	var themes []string
	for _, d := range details {
		theme := d.ThemeID
		if theme == "" {
			theme = "(no theme)"
		}
		if _, exists := byTheme[theme]; !exists {
			themes = append(themes, theme)
		}
		byTheme[theme] = append(byTheme[theme], d)
	}
	sort.Strings(themes)

	// Round-robin pick from each theme
	var result []ArticleEvalDetail
	perTheme := max / len(themes)
	if perTheme < 2 {
		perTheme = 2
	}

	for _, theme := range themes {
		arts := byTheme[theme]
		limit := perTheme
		if limit > len(arts) {
			limit = len(arts)
		}
		result = append(result, arts[:limit]...)
		if len(result) >= max {
			break
		}
	}

	if len(result) > max {
		result = result[:max]
	}
	return result
}

func truncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}

func passFail(ok bool, target string) string {
	if ok {
		return fmt.Sprintf("[PASS] (target: %s)", target)
	}
	return fmt.Sprintf("[FAIL] (target: %s)", target)
}
