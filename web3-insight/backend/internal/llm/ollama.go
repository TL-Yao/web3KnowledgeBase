package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaAdapter implements LLMAdapter for Ollama local models
type OllamaAdapter struct {
	host   string
	model  string
	client *http.Client
}

// NewOllamaAdapter creates a new Ollama adapter
func NewOllamaAdapter(host, model string) *OllamaAdapter {
	return &OllamaAdapter{
		host:  host,
		model: model,
		client: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for generation
		},
	}
}

func (o *OllamaAdapter) Name() string { return o.model }
func (o *OllamaAdapter) Type() string { return "local" }

// Generate performs non-streaming generation
func (o *OllamaAdapter) Generate(prompt string, opts *GenerateOptions) (string, error) {
	payload := map[string]interface{}{
		"model":  o.model,
		"prompt": prompt,
		"stream": false,
	}

	if opts != nil {
		if opts.SystemPrompt != "" {
			payload["system"] = opts.SystemPrompt
		}
		if opts.Temperature > 0 {
			payload["options"] = map[string]interface{}{
				"temperature": opts.Temperature,
			}
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := o.client.Post(o.host+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response, nil
}

// GenerateStream performs streaming generation
func (o *OllamaAdapter) GenerateStream(prompt string, opts *GenerateOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)

	payload := map[string]interface{}{
		"model":  o.model,
		"prompt": prompt,
		"stream": true,
	}

	if opts != nil {
		if opts.SystemPrompt != "" {
			payload["system"] = opts.SystemPrompt
		}
		if opts.Temperature > 0 {
			payload["options"] = map[string]interface{}{
				"temperature": opts.Temperature,
			}
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to marshal request: %w", err)
	}

	go func() {
		defer close(ch)

		resp, err := o.client.Post(o.host+"/api/generate", "application/json", bytes.NewReader(body))
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ch <- StreamChunk{Error: fmt.Errorf("ollama returned status %d", resp.StatusCode)}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var chunk struct {
				Response string `json:"response"`
				Done     bool   `json:"done"`
			}
			if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
				continue
			}
			ch <- StreamChunk{Content: chunk.Response, Done: chunk.Done}
			if chunk.Done {
				break
			}
		}
	}()

	return ch, nil
}

// GenerateChat performs chat completion
func (o *OllamaAdapter) GenerateChat(messages []Message, opts *GenerateOptions) (string, error) {
	payload := map[string]interface{}{
		"model":    o.model,
		"messages": messages,
		"stream":   false,
	}

	if opts != nil && opts.Temperature > 0 {
		payload["options"] = map[string]interface{}{
			"temperature": opts.Temperature,
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := o.client.Post(o.host+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama chat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Message.Content, nil
}

// GenerateChatStream performs streaming chat completion
func (o *OllamaAdapter) GenerateChatStream(messages []Message, opts *GenerateOptions) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk)

	payload := map[string]interface{}{
		"model":    o.model,
		"messages": messages,
		"stream":   true,
	}

	if opts != nil && opts.Temperature > 0 {
		payload["options"] = map[string]interface{}{
			"temperature": opts.Temperature,
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to marshal request: %w", err)
	}

	go func() {
		defer close(ch)

		resp, err := o.client.Post(o.host+"/api/chat", "application/json", bytes.NewReader(body))
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ch <- StreamChunk{Error: fmt.Errorf("ollama returned status %d", resp.StatusCode)}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var chunk struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				Done bool `json:"done"`
			}
			if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
				continue
			}
			ch <- StreamChunk{Content: chunk.Message.Content, Done: chunk.Done}
			if chunk.Done {
				break
			}
		}
	}()

	return ch, nil
}

// IsAvailable checks if Ollama is running and the model is available
func (o *OllamaAdapter) IsAvailable() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(o.host + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	// Check if the specific model is available
	for _, m := range result.Models {
		if m.Name == o.model || m.Name == o.model+":latest" {
			return true
		}
	}

	return false
}

// EstimateCost returns 0 for local models
func (o *OllamaAdapter) EstimateCost(inputTokens, outputTokens int) float64 {
	return 0 // Local model, no cost
}
