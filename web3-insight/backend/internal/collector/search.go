package collector

import (
	"context"
)

// SearchResult represents a single search result
type SearchResult struct {
	Title   string
	URL     string
	Content string
	Score   float64
}

// SearchResponse represents the response from a search API
type SearchResponse struct {
	Query   string
	Results []SearchResult
	Answer  string // AI-generated answer (if available)
}

// SearchProvider defines the interface for search APIs
type SearchProvider interface {
	Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error)
	IsEnabled() bool
	Name() string
}

// SearchRouter routes search requests to available providers
type SearchRouter struct {
	providers []SearchProvider
}

// NewSearchRouter creates a new search router
func NewSearchRouter(providers ...SearchProvider) *SearchRouter {
	return &SearchRouter{providers: providers}
}

// Search performs a search using the first available provider
func (r *SearchRouter) Search(ctx context.Context, query string, maxResults int) (*SearchResponse, error) {
	for _, provider := range r.providers {
		if provider.IsEnabled() {
			return provider.Search(ctx, query, maxResults)
		}
	}
	return nil, nil // No providers enabled
}

// GetEnabledProviders returns names of enabled providers
func (r *SearchRouter) GetEnabledProviders() []string {
	var names []string
	for _, p := range r.providers {
		if p.IsEnabled() {
			names = append(names, p.Name())
		}
	}
	return names
}
