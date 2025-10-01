package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mrkaynak/rag/internal/config"
	"github.com/mrkaynak/rag/pkg/errors"
)

// OpenRouterClient handles OpenRouter API interactions
type OpenRouterClient struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewOpenRouterClient creates a new OpenRouter client
func NewOpenRouterClient(cfg *config.Config) *OpenRouterClient {
	return &OpenRouterClient{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

// openRouterRequest represents OpenRouter chat API request
type openRouterRequest struct {
	Model    string                   `json:"model"`
	Messages []openRouterMessage      `json:"messages"`
	Stream   bool                     `json:"stream"`
}

// openRouterMessage represents a chat message
type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openRouterResponse represents OpenRouter chat API response
type openRouterResponse struct {
	Choices []struct {
		Message openRouterMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Chat sends a chat request to OpenRouter
func (c *OpenRouterClient) Chat(apiKey, model, systemPrompt, userMessage string) (string, error) {
	if apiKey == "" {
		return "", errors.Unauthorized("OpenRouter API key is required")
	}

	// Use default model if not specified
	if model == "" {
		model = c.cfg.OpenRouter.Model
	}

	messages := []openRouterMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	reqBody := openRouterRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.InternalWrap(err, "failed to marshal request")
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", errors.InternalWrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("HTTP-Referer", "https://github.com/mrkaynak/rag")
	req.Header.Set("X-Title", "Enterprise RAG System")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.InternalWrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.InternalWrap(err, "failed to read response")
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.StatusCode, fmt.Sprintf("OpenRouter API error: %s", string(body)))
	}

	var response openRouterResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", errors.InternalWrap(err, "failed to unmarshal response")
	}

	if response.Error != nil {
		return "", errors.Internal(fmt.Sprintf("OpenRouter API error: %s (code: %s)", response.Error.Message, response.Error.Code))
	}

	if len(response.Choices) == 0 {
		return "", errors.Internal("no response from OpenRouter")
	}

	return response.Choices[0].Message.Content, nil
}
