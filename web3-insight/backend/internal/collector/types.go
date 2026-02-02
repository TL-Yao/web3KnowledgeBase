package collector

import (
	"time"

	"github.com/google/uuid"
)

// CollectedItem represents a single item collected from any source
type CollectedItem struct {
	Title          string
	OriginalTitle  string
	Content        string
	Summary        string
	SourceURL      string
	SourceName     string
	SourceLanguage string
	PublishedAt    *time.Time
	Tags           []string
}

// CollectResult represents the result of a collection operation
type CollectResult struct {
	SourceID    uuid.UUID
	ItemsFound  int
	ItemsNew    int
	ItemsFailed int
	Errors      []error
}

// Collector interface for all data collectors
type Collector interface {
	Collect(sourceID uuid.UUID) (*CollectResult, error)
}
