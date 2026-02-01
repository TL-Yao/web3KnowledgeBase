package llm

// LLMAdapter defines the interface for all LLM providers
type LLMAdapter interface {
	Name() string
	Type() string // "local" or "cloud"
	Generate(prompt string, opts *GenerateOptions) (string, error)
	GenerateStream(prompt string, opts *GenerateOptions) (<-chan StreamChunk, error)
	GenerateChat(messages []Message, opts *GenerateOptions) (string, error)
	GenerateChatStream(messages []Message, opts *GenerateOptions) (<-chan StreamChunk, error)
	IsAvailable() bool
	EstimateCost(inputTokens, outputTokens int) float64
}

// GenerateOptions holds configuration for LLM generation
type GenerateOptions struct {
	SystemPrompt string
	MaxTokens    int
	Temperature  float64
	TopP         float64
	StopWords    []string
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// StreamChunk represents a chunk in streaming response
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}

// GenerateResult holds the result of a generation
type GenerateResult struct {
	Content   string
	Model     string
	Usage     Usage
	CostUSD   float64
	Cached    bool
	FinishReason string
}

// Task types for routing
const (
	TaskContentGeneration = "content_generation"
	TaskSummarization     = "summarization"
	TaskClassification    = "classification"
	TaskChat              = "chat"
	TaskTranslation       = "translation"
	TaskEmbedding         = "embedding"
)

// Default options
func DefaultGenerateOptions() *GenerateOptions {
	return &GenerateOptions{
		MaxTokens:   4096,
		Temperature: 0.7,
		TopP:        1.0,
	}
}
