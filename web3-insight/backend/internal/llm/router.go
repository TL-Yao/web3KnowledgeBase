package llm

import (
	"fmt"
	"log"
	"sync"

	"github.com/user/web3-insight/internal/config"
)

// Router manages LLM adapters and routes tasks to appropriate models
type Router struct {
	adapters map[string]LLMAdapter
	routes   map[string][]string // task -> [primary, fallback...]
	mu       sync.RWMutex
}

// NewRouter creates a new LLM router
func NewRouter() *Router {
	return &Router{
		adapters: make(map[string]LLMAdapter),
		routes:   make(map[string][]string),
	}
}

// NewRouterFromConfig creates a router initialized from config
func NewRouterFromConfig(cfg *config.LLMConfig) *Router {
	r := NewRouter()

	// Register Ollama adapters if configured
	if cfg.OllamaHost != "" && cfg.DefaultLocal != "" {
		r.RegisterAdapter(cfg.DefaultLocal, NewOllamaAdapter(cfg.OllamaHost, cfg.DefaultLocal))
	}

	// Register Claude adapter if enabled
	if cfg.Claude.Enabled && cfg.Claude.APIKey != "" {
		r.RegisterAdapter(cfg.Claude.DefaultModel, NewClaudeAdapter(cfg.Claude.APIKey, cfg.Claude.DefaultModel))
		// Also register haiku for simpler tasks
		r.RegisterAdapter("claude-haiku", NewClaudeAdapter(cfg.Claude.APIKey, "claude-3-haiku-20240307"))
	}

	// Register OpenAI adapter if enabled
	if cfg.OpenAI.Enabled && cfg.OpenAI.APIKey != "" {
		r.RegisterAdapter(cfg.OpenAI.DefaultModel, NewOpenAIAdapter(cfg.OpenAI.APIKey, cfg.OpenAI.DefaultModel))
		// Also register mini model
		r.RegisterAdapter("gpt-4o-mini", NewOpenAIAdapter(cfg.OpenAI.APIKey, "gpt-4o-mini"))
	}

	// Set up default routes based on config
	r.setupDefaultRoutes(cfg)

	return r
}

// setupDefaultRoutes configures default routing based on available adapters
func (r *Router) setupDefaultRoutes(cfg *config.LLMConfig) {
	// Content generation: prefer local, fallback to cloud
	contentRoute := []string{}
	if cfg.DefaultLocal != "" {
		contentRoute = append(contentRoute, cfg.DefaultLocal)
	}
	if cfg.Claude.Enabled {
		contentRoute = append(contentRoute, cfg.Claude.DefaultModel)
	}
	if cfg.OpenAI.Enabled {
		contentRoute = append(contentRoute, cfg.OpenAI.DefaultModel)
	}
	if len(contentRoute) > 0 {
		r.SetRoute(TaskContentGeneration, contentRoute)
	}

	// Summarization: prefer local for cost efficiency
	summaryRoute := []string{}
	if cfg.DefaultLocal != "" {
		summaryRoute = append(summaryRoute, cfg.DefaultLocal)
	}
	if cfg.Claude.Enabled {
		summaryRoute = append(summaryRoute, "claude-haiku")
	}
	if cfg.OpenAI.Enabled {
		summaryRoute = append(summaryRoute, "gpt-4o-mini")
	}
	if len(summaryRoute) > 0 {
		r.SetRoute(TaskSummarization, summaryRoute)
	}

	// Classification: similar to summarization
	if len(summaryRoute) > 0 {
		r.SetRoute(TaskClassification, summaryRoute)
	}

	// Chat: prefer local, fallback to powerful cloud model
	chatRoute := []string{}
	if cfg.DefaultLocal != "" {
		chatRoute = append(chatRoute, cfg.DefaultLocal)
	}
	if cfg.Claude.Enabled {
		chatRoute = append(chatRoute, cfg.Claude.DefaultModel)
	}
	if cfg.OpenAI.Enabled {
		chatRoute = append(chatRoute, cfg.OpenAI.DefaultModel)
	}
	if len(chatRoute) > 0 {
		r.SetRoute(TaskChat, chatRoute)
	}

	// Translation: prefer models good at Chinese
	translationRoute := []string{}
	if cfg.DefaultLocal != "" {
		translationRoute = append(translationRoute, cfg.DefaultLocal)
	}
	if cfg.Claude.Enabled {
		translationRoute = append(translationRoute, "claude-haiku")
	}
	if cfg.OpenAI.Enabled {
		translationRoute = append(translationRoute, "gpt-4o-mini")
	}
	if len(translationRoute) > 0 {
		r.SetRoute(TaskTranslation, translationRoute)
	}
}

// RegisterAdapter registers an LLM adapter
func (r *Router) RegisterAdapter(name string, adapter LLMAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[name] = adapter
	log.Printf("LLM adapter registered: %s (type: %s)", name, adapter.Type())
}

// SetRoute sets the routing for a task type
func (r *Router) SetRoute(task string, models []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[task] = models
}

// GetAdapter returns a specific adapter by name
func (r *Router) GetAdapter(name string) (LLMAdapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[name]
	return adapter, ok
}

// ListAdapters returns all registered adapters
func (r *Router) ListAdapters() map[string]LLMAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]LLMAdapter)
	for k, v := range r.adapters {
		result[k] = v
	}
	return result
}

// ListAvailableAdapters returns only adapters that are currently available
func (r *Router) ListAvailableAdapters() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []string
	for name, adapter := range r.adapters {
		if adapter.IsAvailable() {
			result = append(result, name)
		}
	}
	return result
}

// Generate routes a generation request to the appropriate model
// Returns the result, the model used, and any error
func (r *Router) Generate(task, prompt string, opts *GenerateOptions) (string, string, error) {
	r.mu.RLock()
	models := r.routes[task]
	r.mu.RUnlock()

	if len(models) == 0 {
		return "", "", fmt.Errorf("no models configured for task: %s", task)
	}

	for _, modelName := range models {
		adapter, ok := r.adapters[modelName]
		if !ok {
			log.Printf("adapter not found: %s", modelName)
			continue
		}

		if !adapter.IsAvailable() {
			log.Printf("adapter not available: %s", modelName)
			continue
		}

		result, err := adapter.Generate(prompt, opts)
		if err != nil {
			log.Printf("generation failed with %s: %v", modelName, err)
			continue
		}

		return result, modelName, nil
	}

	return "", "", fmt.Errorf("all models failed for task: %s", task)
}

// GenerateStream routes a streaming generation request
func (r *Router) GenerateStream(task, prompt string, opts *GenerateOptions) (<-chan StreamChunk, string, error) {
	r.mu.RLock()
	models := r.routes[task]
	r.mu.RUnlock()

	if len(models) == 0 {
		return nil, "", fmt.Errorf("no models configured for task: %s", task)
	}

	for _, modelName := range models {
		adapter, ok := r.adapters[modelName]
		if !ok {
			continue
		}

		if !adapter.IsAvailable() {
			continue
		}

		stream, err := adapter.GenerateStream(prompt, opts)
		if err != nil {
			log.Printf("stream generation failed with %s: %v", modelName, err)
			continue
		}

		return stream, modelName, nil
	}

	return nil, "", fmt.Errorf("all models failed for task: %s", task)
}

// GenerateChat routes a chat request to the appropriate model
func (r *Router) GenerateChat(task string, messages []Message, opts *GenerateOptions) (string, string, error) {
	r.mu.RLock()
	models := r.routes[task]
	r.mu.RUnlock()

	if len(models) == 0 {
		return "", "", fmt.Errorf("no models configured for task: %s", task)
	}

	for _, modelName := range models {
		adapter, ok := r.adapters[modelName]
		if !ok {
			continue
		}

		if !adapter.IsAvailable() {
			continue
		}

		result, err := adapter.GenerateChat(messages, opts)
		if err != nil {
			log.Printf("chat generation failed with %s: %v", modelName, err)
			continue
		}

		return result, modelName, nil
	}

	return "", "", fmt.Errorf("all models failed for task: %s", task)
}

// GenerateChatStream routes a streaming chat request
func (r *Router) GenerateChatStream(task string, messages []Message, opts *GenerateOptions) (<-chan StreamChunk, string, error) {
	r.mu.RLock()
	models := r.routes[task]
	r.mu.RUnlock()

	if len(models) == 0 {
		return nil, "", fmt.Errorf("no models configured for task: %s", task)
	}

	for _, modelName := range models {
		adapter, ok := r.adapters[modelName]
		if !ok {
			continue
		}

		if !adapter.IsAvailable() {
			continue
		}

		stream, err := adapter.GenerateChatStream(messages, opts)
		if err != nil {
			log.Printf("stream chat generation failed with %s: %v", modelName, err)
			continue
		}

		return stream, modelName, nil
	}

	return nil, "", fmt.Errorf("all models failed for task: %s", task)
}

// GenerateWithModel generates using a specific model (bypasses routing)
func (r *Router) GenerateWithModel(modelName, prompt string, opts *GenerateOptions) (string, error) {
	adapter, ok := r.adapters[modelName]
	if !ok {
		return "", fmt.Errorf("model not found: %s", modelName)
	}

	if !adapter.IsAvailable() {
		return "", fmt.Errorf("model not available: %s", modelName)
	}

	return adapter.Generate(prompt, opts)
}

// EstimateCost estimates the cost for a specific model
func (r *Router) EstimateCost(modelName string, inputTokens, outputTokens int) float64 {
	adapter, ok := r.adapters[modelName]
	if !ok {
		return 0
	}
	return adapter.EstimateCost(inputTokens, outputTokens)
}
