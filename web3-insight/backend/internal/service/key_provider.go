package service

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/user/web3-insight/internal/repository"
)

// providerEnvVars maps provider names to their corresponding environment variable names.
var providerEnvVars = map[string]string{
	"anthropic": "ANTHROPIC_API_KEY",
	"openai":    "OPENAI_API_KEY",
	"tavily":    "TAVILY_API_KEY",
	"serpapi":   "SERPAPI_API_KEY",
}

// KeyProvider resolves API keys at runtime from the database with caching.
// Falls back to environment variables when no DB value is set.
type KeyProvider struct {
	configRepo *repository.ConfigRepository
	cache      map[string]cachedKey
	mu         sync.RWMutex
	cacheTTL   time.Duration
}

type cachedKey struct {
	value     string
	fetchedAt time.Time
}

// NewKeyProvider creates a new KeyProvider that reads keys from the configs DB table.
func NewKeyProvider(configRepo *repository.ConfigRepository) *KeyProvider {
	return &KeyProvider{
		configRepo: configRepo,
		cache:      make(map[string]cachedKey),
		cacheTTL:   30 * time.Second,
	}
}

// GetKey returns the API key for a provider. Checks cache first, then DB.
func (kp *KeyProvider) GetKey(provider string) string {
	kp.mu.RLock()
	if cached, ok := kp.cache[provider]; ok {
		if time.Since(cached.fetchedAt) < kp.cacheTTL {
			kp.mu.RUnlock()
			return cached.value
		}
	}
	kp.mu.RUnlock()

	// Cache miss or expired — fetch from DB
	dbKey := "api_key." + provider
	cfg, err := kp.configRepo.Get(dbKey)
	if err == nil {
		var value string
		if err := json.Unmarshal(cfg.Value, &value); err == nil && value != "" {
			kp.mu.Lock()
			kp.cache[provider] = cachedKey{value: value, fetchedAt: time.Now()}
			kp.mu.Unlock()
			return value
		}
	}

	// Fallback to environment variable (not written to DB)
	if envVar, ok := providerEnvVars[provider]; ok {
		if value := os.Getenv(envVar); value != "" {
			kp.mu.Lock()
			kp.cache[provider] = cachedKey{value: value, fetchedAt: time.Now()}
			kp.mu.Unlock()
			return value
		}
	}

	return ""
}

// InvalidateCache forces the next GetKey call to hit the DB.
func (kp *KeyProvider) InvalidateCache() {
	kp.mu.Lock()
	kp.cache = make(map[string]cachedKey)
	kp.mu.Unlock()
}

