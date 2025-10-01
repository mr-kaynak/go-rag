package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mrkaynak/rag/internal/config"
	"github.com/mrkaynak/rag/internal/models"
	"github.com/mrkaynak/rag/internal/service/embeddings"
	"github.com/mrkaynak/rag/internal/service/llm"
	"github.com/mrkaynak/rag/internal/service/settings"
	"github.com/mrkaynak/rag/internal/service/vector"
	"github.com/mrkaynak/rag/pkg/errors"
	"go.uber.org/zap"
)

// ChatHandler handles chat requests
type ChatHandler struct {
	cfg             *config.Config
	logger          *zap.Logger
	vectorStore     *vector.Store
	embeddingsSvc   *embeddings.Service
	openRouterClient *llm.OpenRouterClient
	bedrockClient    *llm.BedrockClient
	settingsSvc     *settings.Store
}

// NewChatHandler creates a new chat handler
func NewChatHandler(
	cfg *config.Config,
	logger *zap.Logger,
	vectorStore *vector.Store,
	embeddingsSvc *embeddings.Service,
	openRouterClient *llm.OpenRouterClient,
	bedrockClient *llm.BedrockClient,
	settingsSvc *settings.Store,
) *ChatHandler {
	return &ChatHandler{
		cfg:             cfg,
		logger:          logger,
		vectorStore:     vectorStore,
		embeddingsSvc:   embeddingsSvc,
		openRouterClient: openRouterClient,
		bedrockClient:    bedrockClient,
		settingsSvc:     settingsSvc,
	}
}

// Chat handles chat requests with RAG
func (h *ChatHandler) Chat(c *fiber.Ctx) error {
	var req models.ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return h.sendError(c, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if req.Message == "" {
		return h.sendError(c, errors.BadRequest("message is required"))
	}

	if req.Provider != "openrouter" && req.Provider != "bedrock" {
		return h.sendError(c, errors.BadRequest("provider must be 'openrouter' or 'bedrock'"))
	}

	// Get API key from config based on provider
	var apiKey string
	switch req.Provider {
	case "openrouter":
		apiKey = h.cfg.OpenRouter.APIKey
	case "bedrock":
		apiKey = h.cfg.Bedrock.APIKey
	}

	if apiKey == "" {
		return h.sendError(c, errors.Unauthorized("API key is not configured for provider: "+req.Provider))
	}

	h.logger.Info("processing chat request",
		zap.String("provider", req.Provider),
		zap.String("message", req.Message),
	)

	// Generate embedding for the query
	queryChunk := models.Chunk{Content: req.Message}
	chunks, err := h.embeddingsSvc.GenerateEmbeddings([]models.Chunk{queryChunk}, apiKey)
	if err != nil {
		h.logger.Error("failed to generate query embedding", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to generate query embedding"))
	}

	queryEmbedding := chunks[0].Embedding

	// Search for similar chunks
	results, err := h.vectorStore.Search(queryEmbedding, h.cfg.RAG.MaxContextChunks)
	if err != nil {
		h.logger.Error("failed to search vector store", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to search context"))
	}

	// Build context from results
	var contextParts []string
	var contextTexts []string

	for _, result := range results {
		// Just append the content without "Context X" labels
		contextParts = append(contextParts, result.Chunk.Content)
		contextTexts = append(contextTexts, result.Chunk.Content)
	}

	context := strings.Join(contextParts, "\n\n---\n\n")

	// Build system prompt (use custom if provided, otherwise try DB, then config default)
	basePrompt := req.SystemPrompt
	if basePrompt == "" {
		// Try to get from DB first
		if dbPrompt, err := h.settingsSvc.GetDefaultSystemPrompt(); err == nil && dbPrompt.Prompt != "" {
			basePrompt = dbPrompt.Prompt
			h.logger.Debug("using system prompt from DB")
		} else {
			// Fallback to config
			basePrompt = h.cfg.RAG.SystemPrompt
			h.logger.Debug("using system prompt from config")
		}
	}
	systemPrompt := h.buildSystemPrompt(basePrompt, context)

	// Call LLM
	var response string
	switch req.Provider {
	case "openrouter":
		response, err = h.openRouterClient.Chat(apiKey, req.Model, systemPrompt, req.Message)
	case "bedrock":
		response, err = h.bedrockClient.Chat(apiKey, req.Model, systemPrompt, req.Message)
	default:
		return h.sendError(c, errors.BadRequest("unsupported provider"))
	}

	if err != nil {
		h.logger.Error("LLM request failed", zap.Error(err), zap.String("provider", req.Provider))
		return h.sendError(c, err)
	}

	h.logger.Info("chat request completed",
		zap.String("provider", req.Provider),
		zap.Int("context_chunks", len(results)),
	)

	return c.Status(fiber.StatusOK).JSON(models.ChatResponse{
		Message: response,
		Context: contextTexts,
	})
}

// ChatStream handles streaming chat requests with RAG
func (h *ChatHandler) ChatStream(c *fiber.Ctx) error {
	var req models.ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return h.sendError(c, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if req.Message == "" {
		return h.sendError(c, errors.BadRequest("message is required"))
	}

	if req.Provider != "openrouter" && req.Provider != "bedrock" {
		return h.sendError(c, errors.BadRequest("provider must be 'openrouter' or 'bedrock'"))
	}

	// Get API key from config based on provider
	var apiKey string
	switch req.Provider {
	case "openrouter":
		apiKey = h.cfg.OpenRouter.APIKey
	case "bedrock":
		apiKey = h.cfg.Bedrock.APIKey
	}

	if apiKey == "" {
		return h.sendError(c, errors.Unauthorized("API key is not configured for provider: "+req.Provider))
	}

	h.logger.Info("processing streaming chat request",
		zap.String("provider", req.Provider),
		zap.String("message", req.Message),
	)

	// Generate embedding for the query
	queryChunk := models.Chunk{Content: req.Message}
	chunks, err := h.embeddingsSvc.GenerateEmbeddings([]models.Chunk{queryChunk}, apiKey)
	if err != nil {
		h.logger.Error("failed to generate query embedding", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to generate query embedding"))
	}

	queryEmbedding := chunks[0].Embedding

	// Search for similar chunks
	results, err := h.vectorStore.Search(queryEmbedding, h.cfg.RAG.MaxContextChunks)
	if err != nil {
		h.logger.Error("failed to search vector store", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to search context"))
	}

	// Build context from results
	var contextParts []string
	var contextTexts []string

	for _, result := range results {
		// Just append the content without "Context X" labels
		contextParts = append(contextParts, result.Chunk.Content)
		contextTexts = append(contextTexts, result.Chunk.Content)
	}

	context := strings.Join(contextParts, "\n\n---\n\n")

	// Build system prompt (use custom if provided, otherwise try DB, then config default)
	basePrompt := req.SystemPrompt
	if basePrompt == "" {
		// Try to get from DB first
		if dbPrompt, err := h.settingsSvc.GetDefaultSystemPrompt(); err == nil && dbPrompt.Prompt != "" {
			basePrompt = dbPrompt.Prompt
			h.logger.Debug("using system prompt from DB")
		} else {
			// Fallback to config
			basePrompt = h.cfg.RAG.SystemPrompt
			h.logger.Debug("using system prompt from config")
		}
	}
	systemPrompt := h.buildSystemPrompt(basePrompt, context)

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Send context first
		contextJSON, _ := json.Marshal(map[string]interface{}{
			"type":    "context",
			"context": contextTexts,
		})
		fmt.Fprintf(w, "data: %s\n\n", contextJSON)
		w.Flush()

		// Stream LLM response
		switch req.Provider {
		case "bedrock":
			err = h.bedrockClient.ChatStream(apiKey, req.Model, systemPrompt, req.Message, func(chunk string) error {
				eventData, _ := json.Marshal(map[string]interface{}{
					"type": "chunk",
					"text": chunk,
				})
				fmt.Fprintf(w, "data: %s\n\n", eventData)
				return w.Flush()
			})
		default:
			// OpenRouter streaming not implemented yet
			eventData, _ := json.Marshal(map[string]interface{}{
				"type":  "error",
				"error": "streaming not supported for this provider",
			})
			fmt.Fprintf(w, "data: %s\n\n", eventData)
			w.Flush()
			return
		}

		if err != nil {
			h.logger.Error("streaming failed", zap.Error(err))
			eventData, _ := json.Marshal(map[string]interface{}{
				"type":  "error",
				"error": err.Error(),
			})
			fmt.Fprintf(w, "data: %s\n\n", eventData)
			w.Flush()
			return
		}

		// Send done event
		doneData, _ := json.Marshal(map[string]interface{}{
			"type": "done",
		})
		fmt.Fprintf(w, "data: %s\n\n", doneData)
		w.Flush()

		h.logger.Info("streaming chat request completed",
			zap.String("provider", req.Provider),
			zap.Int("context_chunks", len(results)),
		)
	})

	return nil
}

// buildSystemPrompt builds the system prompt with context
func (h *ChatHandler) buildSystemPrompt(basePrompt, context string) string {
	if context == "" {
		return basePrompt
	}

	return fmt.Sprintf(`%s

KNOWLEDGE BASE:
%s

Use this knowledge to answer questions naturally.`, basePrompt, context)
}

// sendError sends an error response
func (h *ChatHandler) sendError(c *fiber.Ctx, err error) error {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		appErr = errors.Internal("internal server error")
	}

	return c.Status(appErr.Code).JSON(models.ErrorResponse{
		Error: appErr.Message,
		Code:  appErr.Code,
	})
}
