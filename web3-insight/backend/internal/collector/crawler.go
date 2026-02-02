package collector

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// WebCrawler handles web page crawling
type WebCrawler struct {
	newsRepo      *repository.NewsRepository
	contentParser *ContentParser
	rateLimiter   *RateLimiter
	userAgents    []string
	uaIndex       int
	uaMutex       sync.Mutex
}

// RateLimiter manages per-domain rate limiting
type RateLimiter struct {
	lastRequest map[string]time.Time
	minDelay    time.Duration
	maxDelay    time.Duration
	mutex       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(minDelay, maxDelay time.Duration) *RateLimiter {
	return &RateLimiter{
		lastRequest: make(map[string]time.Time),
		minDelay:    minDelay,
		maxDelay:    maxDelay,
	}
}

// Wait waits if necessary before making a request to the domain
func (r *RateLimiter) Wait(domain string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if lastTime, ok := r.lastRequest[domain]; ok {
		elapsed := time.Since(lastTime)
		if elapsed < r.minDelay {
			time.Sleep(r.minDelay - elapsed)
		}
	}

	r.lastRequest[domain] = time.Now()
}

// NewWebCrawler creates a new web crawler
func NewWebCrawler(newsRepo *repository.NewsRepository) *WebCrawler {
	return &WebCrawler{
		newsRepo:      newsRepo,
		contentParser: NewContentParser(),
		rateLimiter:   NewRateLimiter(30*time.Second, 60*time.Second),
		userAgents: []string{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
		},
	}
}

// getNextUserAgent returns the next user agent in rotation
func (c *WebCrawler) getNextUserAgent() string {
	c.uaMutex.Lock()
	defer c.uaMutex.Unlock()
	ua := c.userAgents[c.uaIndex]
	c.uaIndex = (c.uaIndex + 1) % len(c.userAgents)
	return ua
}

// CrawlResult represents the result of crawling a URL
type CrawlResult struct {
	URL         string
	Title       string
	Content     string
	ContentHTML string
	Description string
	Language    string
	Error       error
}

// Crawl fetches and parses a single URL
func (c *WebCrawler) Crawl(ctx context.Context, targetURL string) (*CrawlResult, error) {
	result := &CrawlResult{URL: targetURL}

	// Parse URL to get domain for rate limiting
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Wait for rate limit
	c.rateLimiter.Wait(parsedURL.Host)

	// Create collector
	collector := colly.NewCollector(
		colly.UserAgent(c.getNextUserAgent()),
		colly.AllowURLRevisit(),
	)

	// Set timeout
	collector.SetRequestTimeout(30 * time.Second)

	var htmlContent string

	// Handle response
	collector.OnResponse(func(r *colly.Response) {
		htmlContent = string(r.Body)
	})

	// Handle errors
	collector.OnError(func(r *colly.Response, err error) {
		result.Error = fmt.Errorf("crawl failed: %w (status: %d)", err, r.StatusCode)
	})

	// Visit URL
	if err := collector.Visit(targetURL); err != nil {
		return nil, fmt.Errorf("failed to visit URL: %w", err)
	}

	// Wait for collector to finish
	collector.Wait()

	if result.Error != nil {
		return result, result.Error
	}

	if htmlContent == "" {
		return nil, fmt.Errorf("no content received from URL")
	}

	// Parse content
	extracted, err := c.contentParser.Parse(htmlContent, targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	result.Title = extracted.Title
	result.Content = extracted.Content
	result.ContentHTML = extracted.ContentHTML
	result.Description = extracted.Description
	result.Language = extracted.Language

	return result, nil
}

// CrawlAndSave crawls a URL and saves it to the database
func (c *WebCrawler) CrawlAndSave(ctx context.Context, targetURL string, sourceName string) (*model.NewsItem, error) {
	// Check if already exists
	existing, err := c.newsRepo.FindBySourceURL(targetURL)
	if err == nil && existing != nil {
		log.Printf("URL already exists in database: %s", targetURL)
		return existing, nil
	}

	// Crawl the URL
	result, err := c.Crawl(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	// Create news item
	newsItem := &model.NewsItem{
		Title:          result.Title,
		OriginalTitle:  result.Title,
		Content:        result.Content,
		Summary:        result.Description,
		SourceURL:      targetURL,
		SourceName:     sourceName,
		SourceLanguage: result.Language,
		FetchedAt:      time.Now(),
		Processed:      false,
	}

	// Save to database
	created, err := c.newsRepo.CreateOrIgnore(newsItem)
	if err != nil {
		return nil, fmt.Errorf("failed to save news item: %w", err)
	}

	if !created {
		log.Printf("URL already exists (race condition): %s", targetURL)
	}

	log.Printf("Crawled and saved: %s", result.Title)

	return newsItem, nil
}

// FillMissingContent crawls URLs that have no content (e.g., from RSS feeds with only summaries)
func (c *WebCrawler) FillMissingContent(ctx context.Context, item *model.NewsItem) error {
	if item.Content != "" && len(item.Content) > 500 {
		// Already has sufficient content
		return nil
	}

	result, err := c.Crawl(ctx, item.SourceURL)
	if err != nil {
		return err
	}

	// Update the item
	item.Content = result.Content
	if item.Title == "" {
		item.Title = result.Title
	}

	return nil
}

// CrawlMultiple crawls multiple URLs concurrently with rate limiting
func (c *WebCrawler) CrawlMultiple(ctx context.Context, urls []string, sourceName string, concurrency int) []*CrawlResult {
	if concurrency < 1 {
		concurrency = 2
	}

	results := make([]*CrawlResult, len(urls))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, u := range urls {
		wg.Add(1)
		go func(idx int, targetURL string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			result, err := c.Crawl(ctx, targetURL)
			if err != nil {
				results[idx] = &CrawlResult{URL: targetURL, Error: err}
			} else {
				results[idx] = result
			}
		}(i, u)
	}

	wg.Wait()
	return results
}
