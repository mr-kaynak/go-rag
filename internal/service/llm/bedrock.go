package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mrkaynak/rag/internal/config"
	"github.com/mrkaynak/rag/pkg/errors"
)

// BedrockClient handles AWS Bedrock API interactions
type BedrockClient struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewBedrockClient creates a new Bedrock client
func NewBedrockClient(cfg *config.Config) *BedrockClient {
	return &BedrockClient{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

// bedrockRequest represents Bedrock converse API request
type bedrockRequest struct {
	Messages []bedrockMessage `json:"messages"`
}

// bedrockMessage represents a chat message
type bedrockMessage struct {
	Role    string              `json:"role"`
	Content []bedrockContent    `json:"content"`
}

// bedrockContent represents message content
type bedrockContent struct {
	Text string `json:"text"`
}

// bedrockResponse represents Bedrock converse API response
type bedrockResponse struct {
	Output struct {
		Message bedrockMessage `json:"message"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Chat sends a chat request to AWS Bedrock
func (c *BedrockClient) Chat(apiKey, model, systemPrompt, userMessage string) (string, error) {
	if apiKey == "" {
		return "", errors.Unauthorized("Bedrock API key is required")
	}

	// Use default modelId if not specified
	if model == "" {
		model = c.cfg.Bedrock.ModelID
	}

	// Combine system prompt with user message (Bedrock converse format)
	fullMessage := userMessage
	if systemPrompt != "" {
		fullMessage = fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userMessage)
	}

	messages := []bedrockMessage{
		{
			Role: "user",
			Content: []bedrockContent{
				{Text: fullMessage},
			},
		},
	}

	reqBody := bedrockRequest{
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", errors.InternalWrap(err, "failed to marshal request")
	}

	// Build Bedrock endpoint URL
	url := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/converse",
		c.cfg.Bedrock.Region,
		model)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", errors.InternalWrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

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
		return "", errors.New(resp.StatusCode, fmt.Sprintf("Bedrock API error: %s", string(body)))
	}

	var response bedrockResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", errors.InternalWrap(err, "failed to unmarshal response")
	}

	if response.Error != nil {
		return "", errors.Internal(fmt.Sprintf("Bedrock API error: %s (code: %s)", response.Error.Message, response.Error.Code))
	}

	if len(response.Output.Message.Content) == 0 {
		return "", errors.Internal("no response from Bedrock")
	}

	// Find the first content item with actual text (skip reasoning content)
	for _, content := range response.Output.Message.Content {
		if content.Text != "" {
			return content.Text, nil
		}
	}

	return "", errors.Internal("no text content found in Bedrock response")
}

// bedrockStreamEvent represents a streaming event from Bedrock
type bedrockStreamEvent struct {
	ContentBlockDelta *struct {
		Delta struct {
			Text string `json:"text"`
		} `json:"delta"`
	} `json:"contentBlockDelta,omitempty"`
	MessageStop *struct{} `json:"messageStop,omitempty"`
}

// ChatStream sends a streaming chat request to AWS Bedrock
func (c *BedrockClient) ChatStream(apiKey, model, systemPrompt, userMessage string, callback func(string) error) error {
	if apiKey == "" {
		return errors.Unauthorized("Bedrock API key is required")
	}

	// Use default modelId if not specified
	if model == "" {
		model = c.cfg.Bedrock.ModelID
	}

	// Combine system prompt with user message
	fullMessage := userMessage
	if systemPrompt != "" {
		fullMessage = fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userMessage)
	}

	messages := []bedrockMessage{
		{
			Role: "user",
			Content: []bedrockContent{
				{Text: fullMessage},
			},
		},
	}

	reqBody := bedrockRequest{
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return errors.InternalWrap(err, "failed to marshal request")
	}

	// Build Bedrock streaming endpoint URL
	url := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/converse-stream",
		c.cfg.Bedrock.Region,
		model)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.InternalWrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.InternalWrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.New(resp.StatusCode, fmt.Sprintf("Bedrock API error: %s", string(body)))
	}

	// Read SSE stream
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		jsonStr := strings.TrimPrefix(line, "data:")
		jsonStr = strings.TrimSpace(jsonStr)

		if jsonStr == "" {
			continue
		}

		var event bedrockStreamEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			continue // Skip malformed events
		}

		// Handle content delta
		if event.ContentBlockDelta != nil && event.ContentBlockDelta.Delta.Text != "" {
			if err := callback(event.ContentBlockDelta.Delta.Text); err != nil {
				return err
			}
		}

		// Handle stream end
		if event.MessageStop != nil {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return errors.InternalWrap(err, "failed to read stream")
	}

	return nil
}
