package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/web3-insight/internal/repository"
	"github.com/user/web3-insight/internal/service"
)

// providerInfo defines a supported API key provider
type providerInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var supportedProviders = []providerInfo{
	{ID: "anthropic", Name: "Anthropic Claude"},
	{ID: "openai", Name: "OpenAI"},
	{ID: "tavily", Name: "Tavily"},
	{ID: "serpapi", Name: "SerpAPI"},
}

type ApiKeyHandler struct {
	configRepo  *repository.ConfigRepository
	keyProvider *service.KeyProvider
}

func NewApiKeyHandler(configRepo *repository.ConfigRepository, keyProvider *service.KeyProvider) *ApiKeyHandler {
	return &ApiKeyHandler{
		configRepo:  configRepo,
		keyProvider: keyProvider,
	}
}

// ListKeys returns configured providers with masked key values
func (h *ApiKeyHandler) ListKeys(c *gin.Context) {
	type providerStatus struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Configured bool   `json:"configured"`
		Masked     string `json:"masked"`
	}

	var providers []providerStatus
	for _, p := range supportedProviders {
		key := h.keyProvider.GetKey(p.ID)
		providers = append(providers, providerStatus{
			ID:         p.ID,
			Name:       p.Name,
			Configured: key != "",
			Masked:     maskKey(key),
		})
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// SaveKeys saves or updates API keys for one or more providers
func (h *ApiKeyHandler) SaveKeys(c *gin.Context) {
	var req struct {
		Keys map[string]string `json:"keys" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: keys map required"})
		return
	}

	// Validate provider IDs
	validProviders := make(map[string]bool)
	for _, p := range supportedProviders {
		validProviders[p.ID] = true
	}

	for provider, key := range req.Keys {
		if !validProviders[provider] {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unknown provider: %s", provider)})
			return
		}

		dbKey := "api_key." + provider
		if key == "" {
			// Empty string removes the key
			if err := h.configRepo.Delete(dbKey); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove key"})
				return
			}
		} else {
			desc := fmt.Sprintf("API key for %s", provider)
			if err := h.configRepo.Set(dbKey, key, desc); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save key"})
				return
			}
		}
	}

	// Invalidate cache so changes take effect immediately
	h.keyProvider.InvalidateCache()

	// Return updated list
	h.ListKeys(c)
}

// TestKey tests connectivity for a specific provider
func (h *ApiKeyHandler) TestKey(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
		Key      string `json:"key"` // optional: if empty, tests stored key
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	key := req.Key
	if key == "" {
		key = h.keyProvider.GetKey(req.Provider)
	}

	if key == "" {
		c.JSON(http.StatusOK, gin.H{
			"provider": req.Provider,
			"success":  false,
			"message":  "No API key configured",
		})
		return
	}

	success, message := testProviderKey(req.Provider, key)
	c.JSON(http.StatusOK, gin.H{
		"provider": req.Provider,
		"success":  success,
		"message":  message,
	})
}

// maskKey masks an API key for display: show first 6 + last 4 chars
func maskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) < 12 {
		if len(key) <= 4 {
			return "****"
		}
		return key[:2] + "..." + key[len(key)-2:]
	}
	return key[:6] + "..." + key[len(key)-4:]
}

// testProviderKey tests if an API key is valid by making a minimal request
func testProviderKey(provider, key string) (bool, string) {
	client := &http.Client{Timeout: 15 * time.Second}

	switch provider {
	case "anthropic":
		return testAnthropicKey(client, key)
	case "openai":
		return testOpenAIKey(client, key)
	case "tavily":
		return testTavilyKey(client, key)
	case "serpapi":
		return testSerpAPIKey(client, key)
	default:
		return false, fmt.Sprintf("Unknown provider: %s", provider)
	}
}

func testAnthropicKey(client *http.Client, key string) (bool, string) {
	payload := map[string]interface{}{
		"model":      "claude-haiku-4-5-20251001",
		"max_tokens": 1,
		"messages":   []map[string]string{{"role": "user", "content": "hi"}},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", key)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, "Connected successfully"
	}

	return false, fmt.Sprintf("API returned status %d", resp.StatusCode)
}

func testOpenAIKey(client *http.Client, key string) (bool, string) {
	// Use the models endpoint — lightweight, no tokens consumed
	req, _ := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, "Connected successfully"
	}
	return false, fmt.Sprintf("API returned status %d", resp.StatusCode)
}

func testTavilyKey(client *http.Client, key string) (bool, string) {
	payload := map[string]interface{}{
		"api_key":     key,
		"query":       "test",
		"max_results": 1,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://api.tavily.com/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, "Connected successfully"
	}
	return false, fmt.Sprintf("API returned status %d", resp.StatusCode)
}

func testSerpAPIKey(client *http.Client, key string) (bool, string) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://serpapi.com/search.json?api_key=%s&q=test&num=1", key), nil)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, "Connected successfully"
	}
	return false, fmt.Sprintf("API returned status %d", resp.StatusCode)
}
