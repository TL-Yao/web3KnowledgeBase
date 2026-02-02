package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TavilyProvider implements SearchProvider for Tavily API
type TavilyProvider struct {
	apiKey  string
	enabled bool
	client  *http.Client
}

// NewTavilyProvider creates a new Tavily provider
func NewTavilyProvider(apiKey string, enabled bool) *TavilyProvider {
	return &TavilyProvider{
		apiKey:  apiKey,
		enabled: enabled && apiKey != "",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *TavilyProvider) Name() string    { return "tavily" }
func (t *TavilyProvider) IsEnabled() bool { return t.enabled }

// TavilyRequest represents the request to Tavily API
type TavilyRequest struct {
	APIKey            string   `json:"api_key"`
	Query             string   `json:"query"`
	SearchDepth       string   `json:"search_depth,omitempty"` // "basic" or "advanced"
	IncludeAnswer     bool     `json:"include_answer,omitempty"`
	IncludeRawContent bool     `json:"include_raw_content,omitempty"`
	MaxResults        int      `json:"max_results,omitempty"`
	IncludeDomains    []string `json:"include_domains,omitempty"`
	ExcludeDomains    []string `json:"exclude_domains,omitempty"`
}

// TavilyResponse represents the response from Tavily API
type TavilyResponse struct {
	Answer  string `json:"answer"`
	Results []struct {
		Title      string  `json:"title"`
		URL        string  `json:"url"`
		Content    string  `json:"content"`
		RawContent string  `json:"raw_content"`
		Score      float64 `json:"score"`
	} `json:"results"`
}

// Search performs a search using Tavily API
func (t *TavilyProvider) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error) {
	if !t.enabled {
		return nil, fmt.Errorf("Tavily provider is not enabled")
	}

	if maxResults <= 0 {
		maxResults = 5
	}

	req := TavilyRequest{
		APIKey:        t.apiKey,
		Query:         query,
		SearchDepth:   "basic",
		IncludeAnswer: true,
		MaxResults:    maxResults,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tavily API returned status %d", resp.StatusCode)
	}

	var tavilyResp TavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&tavilyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to SearchResponse
	response := &SearchResponse{
		Query:  query,
		Answer: tavilyResp.Answer,
	}

	for _, r := range tavilyResp.Results {
		response.Results = append(response.Results, SearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Content,
			Score:   r.Score,
		})
	}

	return response, nil
}
