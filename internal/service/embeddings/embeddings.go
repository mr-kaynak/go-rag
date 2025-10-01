package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mrkaynak/rag/internal/config"
	"github.com/mrkaynak/rag/internal/models"
	"github.com/mrkaynak/rag/pkg/errors"
)

// Service handles embedding generation
type Service struct {
	cfg        *config.Config
	httpClient *http.Client
}

// New creates a new embeddings service
func New(cfg *config.Config) *Service {
	return &Service{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

// openRouterRequest represents OpenRouter embeddings API request
type openRouterRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// openRouterResponse represents OpenRouter embeddings API response
type openRouterResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateEmbeddings generates embeddings for chunks
func (s *Service) GenerateEmbeddings(chunks []models.Chunk, apiKey string) ([]models.Chunk, error) {
	// API key not required for Ollama
	if s.cfg.Embeddings.Provider != "ollama" && apiKey == "" {
		return nil, errors.BadRequest("API key is required for embeddings")
	}

	for i := range chunks {
		var embedding []float64
		var err error

		switch s.cfg.Embeddings.Provider {
		case "ollama":
			embedding, err = s.generateOllamaEmbedding(chunks[i].Content)
		case "openrouter":
			embedding, err = s.generateOpenRouterEmbedding(chunks[i].Content, apiKey)
		case "bedrock":
			embedding, err = s.generateBedrockEmbedding(chunks[i].Content, apiKey)
		default:
			return nil, errors.BadRequest("unsupported embedding provider")
		}

		if err != nil {
			return nil, errors.InternalWrap(err, fmt.Sprintf("failed to generate embedding for chunk %d", i))
		}
		chunks[i].Embedding = embedding
	}

	return chunks, nil
}

// generateOpenRouterEmbedding generates embedding for a single text using OpenRouter
func (s *Service) generateOpenRouterEmbedding(text, apiKey string) ([]float64, error) {
	reqBody := openRouterRequest{
		Model: s.cfg.Embeddings.Model,
		Input: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embeddings API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response openRouterResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("embeddings API error: %s", response.Error.Message)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return response.Data[0].Embedding, nil
}

// bedrockEmbeddingRequest represents Bedrock embedding API request
type bedrockEmbeddingRequest struct {
	InputText string `json:"inputText"`
}

// bedrockEmbeddingResponse represents Bedrock embedding API response
type bedrockEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// generateBedrockEmbedding generates embedding using AWS Bedrock
func (s *Service) generateBedrockEmbedding(text, apiKey string) ([]float64, error) {
	reqBody := bedrockEmbeddingRequest{
		InputText: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build Bedrock embedding endpoint URL
	url := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke",
		s.cfg.Bedrock.Region,
		s.cfg.Embeddings.Model)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bedrock API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response bedrockEmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("Bedrock API error: %s", response.Error.Message)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return response.Embedding, nil
}

// ollamaRequest represents Ollama embeddings API request
type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaResponse represents Ollama embeddings API response
type ollamaResponse struct {
	Embedding []float64 `json:"embedding"`
}

// generateOllamaEmbedding generates embedding using Ollama
func (s *Service) generateOllamaEmbedding(text string) ([]float64, error) {
	reqBody := ollamaRequest{
		Model:  s.cfg.Embeddings.Model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embeddings", s.cfg.Ollama.BaseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response ollamaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding returned from Ollama")
	}

	return response.Embedding, nil
}
