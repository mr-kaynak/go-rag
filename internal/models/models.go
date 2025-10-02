package models

import "time"

// Document represents an uploaded document
type Document struct {
	ID        string    `json:"id"`
	FileName  string    `json:"file_name"`
	Content   string    `json:"content"`
	Chunks    []Chunk   `json:"chunks,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Chunk represents a text chunk with embeddings
type Chunk struct {
	ID        string    `json:"id"`
	DocID     string    `json:"doc_id"`
	Content   string    `json:"content"`
	Embedding []float64 `json:"embedding,omitempty"`
	Index     int       `json:"index"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Message      string `json:"message" validate:"required"`
	Provider     string `json:"provider" validate:"required,oneof=openrouter bedrock"`
	Model        string `json:"model,omitempty"`
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Message      string       `json:"message"`
	Context      []string     `json:"context,omitempty"`
	TokenMetrics TokenMetrics `json:"token_metrics,omitempty"`
}

// TokenMetrics represents token usage information
type TokenMetrics struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// UploadResponse represents a document upload response
type UploadResponse struct {
	DocumentID string `json:"document_id"`
	FileName   string `json:"file_name"`
	ChunkCount int    `json:"chunk_count"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
