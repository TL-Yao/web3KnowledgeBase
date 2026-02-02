package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SerpAPIProvider implements SearchProvider for SerpAPI
type SerpAPIProvider struct {
	apiKey  string
	enabled bool
	client  *http.Client
}

// NewSerpAPIProvider creates a new SerpAPI provider
func NewSerpAPIProvider(apiKey string, enabled bool) *SerpAPIProvider {
	return &SerpAPIProvider{
		apiKey:  apiKey,
		enabled: enabled && apiKey != "",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *SerpAPIProvider) Name() string    { return "serpapi" }
func (s *SerpAPIProvider) IsEnabled() bool { return s.enabled }

// SerpAPIResponse represents the response from SerpAPI
type SerpAPIResponse struct {
	OrganicResults []struct {
		Position int    `json:"position"`
		Title    string `json:"title"`
		Link     string `json:"link"`
		Snippet  string `json:"snippet"`
	} `json:"organic_results"`
	AnswerBox *struct {
		Answer  string `json:"answer"`
		Snippet string `json:"snippet"`
	} `json:"answer_box"`
}

// Search performs a search using SerpAPI
func (s *SerpAPIProvider) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error) {
	if !s.enabled {
		return nil, fmt.Errorf("SerpAPI provider is not enabled")
	}

	if maxResults <= 0 {
		maxResults = 5
	}

	// Build URL
	u, _ := url.Parse("https://serpapi.com/search.json")
	q := u.Query()
	q.Set("api_key", s.apiKey)
	q.Set("q", query)
	q.Set("engine", "google")
	q.Set("num", fmt.Sprintf("%d", maxResults))
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SerpAPI returned status %d", resp.StatusCode)
	}

	var serpResp SerpAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&serpResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to SearchResponse
	response := &SearchResponse{
		Query: query,
	}

	if serpResp.AnswerBox != nil {
		response.Answer = serpResp.AnswerBox.Answer
		if response.Answer == "" {
			response.Answer = serpResp.AnswerBox.Snippet
		}
	}

	for _, r := range serpResp.OrganicResults {
		if len(response.Results) >= maxResults {
			break
		}
		response.Results = append(response.Results, SearchResult{
			Title:   r.Title,
			URL:     r.Link,
			Content: r.Snippet,
			Score:   1.0 / float64(r.Position), // Higher position = higher score
		})
	}

	return response, nil
}
