package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pgvector/pgvector-go"
)

// EmbeddingAdapter defines the interface for embedding generation
type EmbeddingAdapter interface {
	Name() string
	Dimensions() int
	GenerateEmbedding(text string) ([]float32, error)
	GenerateBatchEmbeddings(texts []string) ([][]float32, error)
	IsAvailable() bool
}

// OllamaEmbeddingAdapter generates embeddings using Ollama
type OllamaEmbeddingAdapter struct {
	host       string
	model      string
	dimensions int
	client     *http.Client
}

// NewOllamaEmbeddingAdapter creates a new Ollama embedding adapter
func NewOllamaEmbeddingAdapter(host, model string, dimensions int) *OllamaEmbeddingAdapter {
	return &OllamaEmbeddingAdapter{
		host:       host,
		model:      model,
		dimensions: dimensions,
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

// DefaultOllamaEmbeddingAdapter creates adapter with default nomic-embed-text model
func DefaultOllamaEmbeddingAdapter(host string) *OllamaEmbeddingAdapter {
	return NewOllamaEmbeddingAdapter(host, "nomic-embed-text", 768)
}

func (o *OllamaEmbeddingAdapter) Name() string       { return o.model }
func (o *OllamaEmbeddingAdapter) Dimensions() int    { return o.dimensions }

// GenerateEmbedding generates embedding for a single text
func (o *OllamaEmbeddingAdapter) GenerateEmbedding(text string) ([]float32, error) {
	payload := map[string]interface{}{
		"model":  o.model,
		"prompt": text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := o.client.Post(o.host+"/api/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Embedding, nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts
func (o *OllamaEmbeddingAdapter) GenerateBatchEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	// Ollama doesn't support batch embeddings natively, so we process sequentially
	for i, text := range texts {
		emb, err := o.GenerateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = emb
	}

	return embeddings, nil
}

// IsAvailable checks if Ollama is running and the embedding model is available
func (o *OllamaEmbeddingAdapter) IsAvailable() bool {
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

	for _, m := range result.Models {
		if m.Name == o.model || m.Name == o.model+":latest" {
			return true
		}
	}

	return false
}

// Float32ToVector converts float32 slice to pgvector.Vector
func Float32ToVector(embedding []float32) *pgvector.Vector {
	vec := pgvector.NewVector(embedding)
	return &vec
}

// VectorToFloat32 converts pgvector.Vector to float32 slice
func VectorToFloat32(vec *pgvector.Vector) []float32 {
	if vec == nil {
		return nil
	}
	return vec.Slice()
}
