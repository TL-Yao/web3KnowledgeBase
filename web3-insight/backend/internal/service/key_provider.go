package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/user/web3-insight/internal/repository"
)

// KeyProvider resolves API keys at runtime from the database with caching.
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
	if err != nil {
		return "" // not configured
	}

	var value string
	if err := json.Unmarshal(cfg.Value, &value); err != nil {
		return ""
	}

	kp.mu.Lock()
	kp.cache[provider] = cachedKey{value: value, fetchedAt: time.Now()}
	kp.mu.Unlock()

	return value
}

// InvalidateCache forces the next GetKey call to hit the DB.
func (kp *KeyProvider) InvalidateCache() {
	kp.mu.Lock()
	kp.cache = make(map[string]cachedKey)
	kp.mu.Unlock()
}

// SeedAPIKeysFromEnv writes env var values to DB for providers that have no key stored yet.
// Does NOT overwrite existing DB values.
func SeedAPIKeysFromEnv(configRepo *repository.ConfigRepository, envMap map[string]string) {
	for provider, envVar := range envMap {
		dbKey := "api_key." + provider
		if _, err := configRepo.Get(dbKey); err == nil {
			continue // already has a value in DB
		}
		if value := os.Getenv(envVar); value != "" {
			if err := configRepo.Set(dbKey, value, fmt.Sprintf("API key for %s (seeded from env)", provider)); err != nil {
				log.Printf("Warning: failed to seed API key for %s: %v", provider, err)
			} else {
				log.Printf("Seeded API key for %s from $%s", provider, envVar)
			}
		}
	}
}
