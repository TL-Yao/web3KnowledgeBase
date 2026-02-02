package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"github.com/user/web3-insight/internal/model"
	"github.com/user/web3-insight/internal/repository"
)

// RSSCollector handles RSS feed collection
type RSSCollector struct {
	parser   *gofeed.Parser
	newsRepo *repository.NewsRepository
	dsRepo   *repository.DataSourceRepository
}

// NewRSSCollector creates a new RSS collector
func NewRSSCollector(newsRepo *repository.NewsRepository, dsRepo *repository.DataSourceRepository) *RSSCollector {
	parser := gofeed.NewParser()
	parser.UserAgent = "Web3-Insight/1.0 (RSS Reader)"

	return &RSSCollector{
		parser:   parser,
		newsRepo: newsRepo,
		dsRepo:   dsRepo,
	}
}

// RSSConfig holds RSS-specific configuration
type RSSConfig struct {
	DefaultCategory string `json:"defaultCategory,omitempty"`
	Language        string `json:"language,omitempty"`
}

// Collect fetches and stores items from an RSS feed
func (c *RSSCollector) Collect(ctx context.Context, sourceID uuid.UUID) (*CollectResult, error) {
	result := &CollectResult{SourceID: sourceID}

	// Get data source
	source, err := c.dsRepo.FindByID(sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find data source: %w", err)
	}

	if source.Type != model.DataSourceTypeRSS {
		return nil, fmt.Errorf("data source is not RSS type: %s", source.Type)
	}

	// Parse RSS config
	var config RSSConfig
	if source.Config != nil {
		if err := json.Unmarshal(source.Config, &config); err != nil {
			log.Printf("Warning: failed to parse RSS config: %v", err)
		}
	}

	// Fetch and parse feed
	feed, err := c.parser.ParseURLWithContext(source.URL, ctx)
	if err != nil {
		// Update source with error
		c.dsRepo.UpdateLastFetched(sourceID, time.Now(), err.Error())
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	result.ItemsFound = len(feed.Items)

	// Convert feed items to news items
	var newsItems []model.NewsItem
	for _, item := range feed.Items {
		newsItem := c.convertFeedItem(item, source.Name, config)
		newsItems = append(newsItems, newsItem)
	}

	// Batch insert, ignoring duplicates
	newCount, err := c.newsRepo.BatchCreateOrIgnore(newsItems)
	if err != nil {
		result.Errors = append(result.Errors, err)
		c.dsRepo.UpdateLastFetched(sourceID, time.Now(), err.Error())
		return result, err
	}

	result.ItemsNew = newCount

	// Update last fetched
	c.dsRepo.UpdateLastFetched(sourceID, time.Now(), "")

	log.Printf("RSS sync completed for %s: found=%d, new=%d", source.Name, result.ItemsFound, result.ItemsNew)

	return result, nil
}

// convertFeedItem converts a gofeed.Item to model.NewsItem
func (c *RSSCollector) convertFeedItem(item *gofeed.Item, sourceName string, config RSSConfig) model.NewsItem {
	newsItem := model.NewsItem{
		Title:          item.Title,
		OriginalTitle:  item.Title,
		SourceURL:      item.Link,
		SourceName:     sourceName,
		SourceLanguage: config.Language,
		Category:       config.DefaultCategory,
		FetchedAt:      time.Now(),
		Processed:      false,
	}

	// Extract content - prefer full content over description
	if item.Content != "" {
		newsItem.Content = item.Content
	} else if item.Description != "" {
		newsItem.Content = item.Description
	}

	// Parse published date
	if item.PublishedParsed != nil {
		newsItem.PublishedAt = item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		newsItem.PublishedAt = item.UpdatedParsed
	}

	// Extract tags from categories
	if len(item.Categories) > 0 {
		newsItem.Tags = item.Categories
	}

	// Detect language from content if not specified
	if newsItem.SourceLanguage == "" {
		newsItem.SourceLanguage = detectLanguage(newsItem.Content)
	}

	return newsItem
}

// detectLanguage is a simple heuristic to detect if content is Chinese or English
func detectLanguage(content string) string {
	if content == "" {
		return "en"
	}

	// Count Chinese characters
	chineseCount := 0
	for _, r := range content {
		if r >= 0x4E00 && r <= 0x9FFF {
			chineseCount++
		}
	}

	// If more than 10% Chinese characters, consider it Chinese
	if len(content) > 0 && float64(chineseCount)/float64(len([]rune(content))) > 0.1 {
		return "zh"
	}

	return "en"
}

// CollectAll collects from all enabled RSS sources
func (c *RSSCollector) CollectAll(ctx context.Context) ([]*CollectResult, error) {
	sources, err := c.dsRepo.FindByType(model.DataSourceTypeRSS)
	if err != nil {
		return nil, fmt.Errorf("failed to find RSS sources: %w", err)
	}

	var results []*CollectResult
	for _, source := range sources {
		result, err := c.Collect(ctx, source.ID)
		if err != nil {
			log.Printf("Failed to collect from %s: %v", source.Name, err)
			result = &CollectResult{
				SourceID: source.ID,
				Errors:   []error{err},
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// ValidateFeedURL checks if a URL is a valid RSS feed
func (c *RSSCollector) ValidateFeedURL(url string) (*gofeed.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	feed, err := c.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("invalid RSS feed: %w", err)
	}

	return feed, nil
}
