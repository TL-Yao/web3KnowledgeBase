package collector

import (
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

// ContentParser handles HTML content extraction and conversion
type ContentParser struct {
	converter *md.Converter
	// Site-specific selectors for known sites
	siteSelectors map[string]SiteSelector
}

// SiteSelector defines how to extract content from a specific site
type SiteSelector struct {
	ContentSelector string   // CSS selector for main content
	TitleSelector   string   // CSS selector for title
	RemoveSelectors []string // Elements to remove before extraction
}

// NewContentParser creates a new content parser
func NewContentParser() *ContentParser {
	converter := md.NewConverter("", true, nil)

	return &ContentParser{
		converter: converter,
		siteSelectors: map[string]SiteSelector{
			"blog.ethereum.org": {
				ContentSelector: "article, .post-content, main",
				TitleSelector:   "h1",
				RemoveSelectors: []string{"nav", "footer", "aside", ".comments", ".share-buttons"},
			},
			"vitalik.eth.limo": {
				ContentSelector: "article, .post, main",
				TitleSelector:   "h1",
				RemoveSelectors: []string{"nav", "footer"},
			},
			"paradigm.xyz": {
				ContentSelector: "article, .content, main",
				TitleSelector:   "h1",
				RemoveSelectors: []string{"nav", "footer", "aside"},
			},
		},
	}
}

// ExtractedContent represents parsed content from a web page
type ExtractedContent struct {
	Title       string
	Content     string // Markdown content
	ContentHTML string // Original HTML content
	Description string
	Language    string
}

// Parse extracts content from HTML
func (p *ContentParser) Parse(html string, url string) (*ExtractedContent, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	// Get domain for site-specific handling
	domain := extractDomain(url)

	// Remove unwanted elements first
	p.removeUnwantedElements(doc, domain)

	// Extract content
	content := p.extractContent(doc, domain)

	// Extract title
	title := p.extractTitle(doc, domain)

	// Extract description
	description := p.extractDescription(doc)

	// Convert to markdown
	markdown, err := p.converter.ConvertString(content.ContentHTML)
	if err != nil {
		// If conversion fails, use raw text
		markdown = content.ContentHTML
	}

	// Clean up markdown
	markdown = p.cleanMarkdown(markdown)

	return &ExtractedContent{
		Title:       title,
		Content:     markdown,
		ContentHTML: content.ContentHTML,
		Description: description,
		Language:    detectLanguage(markdown),
	}, nil
}

type contentResult struct {
	ContentHTML string
}

// removeUnwantedElements removes navigation, footer, ads, etc.
func (p *ContentParser) removeUnwantedElements(doc *goquery.Document, domain string) {
	// Common unwanted elements
	defaultRemove := []string{
		"script", "style", "noscript", "iframe",
		"nav", "footer", "header",
		".nav", ".navigation", ".menu",
		".footer", ".header",
		".sidebar", ".aside", "aside",
		".comments", ".comment-section",
		".share", ".social", ".share-buttons",
		".advertisement", ".ad", ".ads",
		".related-posts", ".recommended",
		"[role='navigation']", "[role='banner']", "[role='contentinfo']",
	}

	for _, selector := range defaultRemove {
		doc.Find(selector).Remove()
	}

	// Site-specific removals
	if siteSelector, ok := p.siteSelectors[domain]; ok {
		for _, selector := range siteSelector.RemoveSelectors {
			doc.Find(selector).Remove()
		}
	}
}

// extractContent extracts the main content
func (p *ContentParser) extractContent(doc *goquery.Document, domain string) contentResult {
	var contentHTML string

	// Try site-specific selector first
	if siteSelector, ok := p.siteSelectors[domain]; ok {
		if siteSelector.ContentSelector != "" {
			sel := doc.Find(siteSelector.ContentSelector).First()
			if sel.Length() > 0 {
				contentHTML, _ = sel.Html()
				if len(contentHTML) > 100 {
					return contentResult{ContentHTML: contentHTML}
				}
			}
		}
	}

	// Generic extraction strategy
	// Priority order: article > main > .content > .post > body
	selectors := []string{
		"article",
		"main",
		"[role='main']",
		".content",
		".post-content",
		".article-content",
		".entry-content",
		".post",
		".article",
	}

	for _, selector := range selectors {
		sel := doc.Find(selector).First()
		if sel.Length() > 0 {
			html, _ := sel.Html()
			if len(html) > 100 {
				return contentResult{ContentHTML: html}
			}
		}
	}

	// Fallback: use body
	body := doc.Find("body")
	contentHTML, _ = body.Html()

	return contentResult{ContentHTML: contentHTML}
}

// extractTitle extracts the page title
func (p *ContentParser) extractTitle(doc *goquery.Document, domain string) string {
	// Try site-specific selector
	if siteSelector, ok := p.siteSelectors[domain]; ok {
		if siteSelector.TitleSelector != "" {
			title := doc.Find(siteSelector.TitleSelector).First().Text()
			if title != "" {
				return strings.TrimSpace(title)
			}
		}
	}

	// Try h1
	h1 := doc.Find("h1").First().Text()
	if h1 != "" {
		return strings.TrimSpace(h1)
	}

	// Try og:title
	ogTitle, _ := doc.Find("meta[property='og:title']").Attr("content")
	if ogTitle != "" {
		return ogTitle
	}

	// Fallback to title tag
	title := doc.Find("title").First().Text()
	return strings.TrimSpace(title)
}

// extractDescription extracts meta description
func (p *ContentParser) extractDescription(doc *goquery.Document) string {
	// Try og:description first
	ogDesc, _ := doc.Find("meta[property='og:description']").Attr("content")
	if ogDesc != "" {
		return ogDesc
	}

	// Try meta description
	metaDesc, _ := doc.Find("meta[name='description']").Attr("content")
	return metaDesc
}

// cleanMarkdown cleans up converted markdown
func (p *ContentParser) cleanMarkdown(mdContent string) string {
	// Remove excessive newlines
	multiNewline := regexp.MustCompile(`\n{3,}`)
	mdContent = multiNewline.ReplaceAllString(mdContent, "\n\n")

	// Remove empty links
	emptyLink := regexp.MustCompile(`\[]\([^)]*\)`)
	mdContent = emptyLink.ReplaceAllString(mdContent, "")

	// Trim whitespace
	mdContent = strings.TrimSpace(mdContent)

	return mdContent
}

// extractDomain extracts domain from URL
func extractDomain(url string) string {
	// Simple extraction - remove protocol and path
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")

	if idx := strings.Index(url, "/"); idx > 0 {
		url = url[:idx]
	}

	return url
}
