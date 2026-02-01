package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const openaiAPIURL = "https://api.openai.com/v1/chat/completions"

// OpenAIAdapter implements LLMAdapter for OpenAI models
type OpenAIAdapter struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter(apiKey, model string) *OpenAIAdapter {
	return &OpenAIAdapter{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (o *OpenAIAdapter) Name() string { return o.model }
func (o *OpenAIAdapter) Type() string { return "cloud" }

// Generate performs non-streaming generation
func (o *OpenAIAdapter) Generate(prompt string, opts *GenerateOptions) (string, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return o.GenerateChat(messages, opts)
}

// GenerateStream performs streaming generation
func (o *OpenAIAdapter) GenerateStream(prompt string, opts *GenerateOptions) (<-chan StreamChunk, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return o.GenerateChatStream(messages, opts)
}

// GenerateChat performs chat completion
func (o *OpenAIAdapter) GenerateChat(messages []Message, opts *GenerateOptions) (string, error) {
	if opts == nil {
		opts = DefaultGenerateOptions()
	}

	openaiMessages := o.convertMessages(messages, opts.SystemPrompt)

	payload := map[string]interface{}{
		"model":    o.model,
		"messages": openaiMessages,
	}

	if opts.MaxTokens > 0 {
		payload["max_tokens"] = opts.MaxTokens
	}
	if opts.Temperature > 0 {
		payload["temperature"] = opts.Temperature
	}
	if opts.TopP > 0 && opts.TopP < 1 {
		payload["top_p"] = opts.TopP
	}
	if len(opts.StopWords) > 0 {
		payload["stop"] = opts.StopWords
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", openaiAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, errResp.Error.Message)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("empty response from openai")
}

// GenerateChatStream performs streaming chat completion
func (o *OpenAIAdapter) GenerateChatStream(messages []Message, opts *GenerateOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)

	if opts == nil {
		opts = DefaultGenerateOptions()
	}

	openaiMessages := o.convertMessages(messages, opts.SystemPrompt)

	payload := map[string]interface{}{
		"model":    o.model,
		"messages": openaiMessages,
		"stream":   true,
	}

	if opts.MaxTokens > 0 {
		payload["max_tokens"] = opts.MaxTokens
	}
	if opts.Temperature > 0 {
		payload["temperature"] = opts.Temperature
	}

	body, err := json.Marshal(payload)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to marshal request: %w", err)
	}

	go func() {
		defer close(ch)

		req, err := http.NewRequest("POST", openaiAPIURL, bytes.NewReader(body))
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+o.apiKey)

		resp, err := o.client.Do(req)
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ch <- StreamChunk{Error: fmt.Errorf("openai returned status %d", resp.StatusCode)}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- StreamChunk{Done: true}
				break
			}

			var event struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			if len(event.Choices) > 0 {
				if event.Choices[0].Delta.Content != "" {
					ch <- StreamChunk{Content: event.Choices[0].Delta.Content}
				}
				if event.Choices[0].FinishReason != "" {
					ch <- StreamChunk{Done: true}
					break
				}
			}
		}
	}()

	return ch, nil
}

// convertMessages converts our Message type to OpenAI's format
func (o *OpenAIAdapter) convertMessages(messages []Message, systemPrompt string) []map[string]string {
	result := make([]map[string]string, 0, len(messages)+1)

	// Add system prompt if provided
	if systemPrompt != "" {
		result = append(result, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	for _, m := range messages {
		result = append(result, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}
	return result
}

// IsAvailable checks if the OpenAI API is configured
func (o *OpenAIAdapter) IsAvailable() bool {
	return o.apiKey != ""
}

// EstimateCost estimates the cost based on OpenAI pricing
func (o *OpenAIAdapter) EstimateCost(inputTokens, outputTokens int) float64 {
	var inputPrice, outputPrice float64

	switch {
	case strings.Contains(o.model, "gpt-4o-mini"):
		inputPrice = 0.00015 // per 1K tokens
		outputPrice = 0.0006
	case strings.Contains(o.model, "gpt-4o"):
		inputPrice = 0.0025
		outputPrice = 0.01
	case strings.Contains(o.model, "gpt-4-turbo"):
		inputPrice = 0.01
		outputPrice = 0.03
	case strings.Contains(o.model, "gpt-4"):
		inputPrice = 0.03
		outputPrice = 0.06
	case strings.Contains(o.model, "gpt-3.5"):
		inputPrice = 0.0005
		outputPrice = 0.0015
	default:
		// Default to gpt-4o pricing
		inputPrice = 0.0025
		outputPrice = 0.01
	}

	return (float64(inputTokens)/1000)*inputPrice + (float64(outputTokens)/1000)*outputPrice
}
