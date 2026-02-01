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

const claudeAPIURL = "https://api.anthropic.com/v1/messages"
const claudeAPIVersion = "2023-06-01"

// ClaudeAdapter implements LLMAdapter for Anthropic Claude models
type ClaudeAdapter struct {
	apiKey string
	model  string
	client *http.Client
}

// NewClaudeAdapter creates a new Claude adapter
func NewClaudeAdapter(apiKey, model string) *ClaudeAdapter {
	return &ClaudeAdapter{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (c *ClaudeAdapter) Name() string { return c.model }
func (c *ClaudeAdapter) Type() string { return "cloud" }

// Generate performs non-streaming generation
func (c *ClaudeAdapter) Generate(prompt string, opts *GenerateOptions) (string, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return c.GenerateChat(messages, opts)
}

// GenerateStream performs streaming generation
func (c *ClaudeAdapter) GenerateStream(prompt string, opts *GenerateOptions) (<-chan StreamChunk, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return c.GenerateChatStream(messages, opts)
}

// GenerateChat performs chat completion
func (c *ClaudeAdapter) GenerateChat(messages []Message, opts *GenerateOptions) (string, error) {
	if opts == nil {
		opts = DefaultGenerateOptions()
	}

	payload := map[string]interface{}{
		"model":      c.model,
		"max_tokens": opts.MaxTokens,
		"messages":   c.convertMessages(messages),
	}

	if opts.SystemPrompt != "" {
		payload["system"] = opts.SystemPrompt
	}
	if opts.Temperature > 0 {
		payload["temperature"] = opts.Temperature
	}
	if opts.TopP > 0 && opts.TopP < 1 {
		payload["top_p"] = opts.TopP
	}
	if len(opts.StopWords) > 0 {
		payload["stop_sequences"] = opts.StopWords
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", claudeAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", claudeAPIVersion)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return "", fmt.Errorf("claude returned status %d: %s", resp.StatusCode, errResp.Error.Message)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Content) > 0 && result.Content[0].Type == "text" {
		return result.Content[0].Text, nil
	}

	return "", fmt.Errorf("empty response from claude")
}

// GenerateChatStream performs streaming chat completion
func (c *ClaudeAdapter) GenerateChatStream(messages []Message, opts *GenerateOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)

	if opts == nil {
		opts = DefaultGenerateOptions()
	}

	payload := map[string]interface{}{
		"model":      c.model,
		"max_tokens": opts.MaxTokens,
		"messages":   c.convertMessages(messages),
		"stream":     true,
	}

	if opts.SystemPrompt != "" {
		payload["system"] = opts.SystemPrompt
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

		req, err := http.NewRequest("POST", claudeAPIURL, bytes.NewReader(body))
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("anthropic-version", claudeAPIVersion)

		resp, err := c.client.Do(req)
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ch <- StreamChunk{Error: fmt.Errorf("claude returned status %d", resp.StatusCode)}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// SSE format: "event: xxx" followed by "data: xxx"
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- StreamChunk{Done: true}
				break
			}

			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
				ch <- StreamChunk{Content: event.Delta.Text}
			} else if event.Type == "message_stop" {
				ch <- StreamChunk{Done: true}
				break
			}
		}
	}()

	return ch, nil
}

// convertMessages converts our Message type to Claude's format
func (c *ClaudeAdapter) convertMessages(messages []Message) []map[string]string {
	result := make([]map[string]string, 0, len(messages))
	for _, m := range messages {
		// Claude doesn't support "system" role in messages array
		if m.Role == "system" {
			continue
		}
		result = append(result, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}
	return result
}

// IsAvailable checks if the Claude API is configured
func (c *ClaudeAdapter) IsAvailable() bool {
	return c.apiKey != ""
}

// EstimateCost estimates the cost based on Claude pricing
func (c *ClaudeAdapter) EstimateCost(inputTokens, outputTokens int) float64 {
	var inputPrice, outputPrice float64

	switch {
	case strings.Contains(c.model, "opus"):
		inputPrice = 0.015  // per 1K tokens
		outputPrice = 0.075 // per 1K tokens
	case strings.Contains(c.model, "sonnet"):
		inputPrice = 0.003
		outputPrice = 0.015
	case strings.Contains(c.model, "haiku"):
		inputPrice = 0.00025
		outputPrice = 0.00125
	default:
		// Default to sonnet pricing
		inputPrice = 0.003
		outputPrice = 0.015
	}

	return (float64(inputTokens)/1000)*inputPrice + (float64(outputTokens)/1000)*outputPrice
}
