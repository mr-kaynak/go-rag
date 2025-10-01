package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server     ServerConfig
	OpenRouter OpenRouterConfig
	Bedrock    BedrockConfig
	Ollama     OllamaConfig
	Embeddings EmbeddingsConfig
	Storage    StorageConfig
	Encryption EncryptionConfig
	RAG        RAGConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string
	Env  string
}

// OpenRouterConfig holds OpenRouter API configuration
type OpenRouterConfig struct {
	APIKey string
	Model  string
}

// BedrockConfig holds AWS Bedrock configuration
type BedrockConfig struct {
	APIKey  string
	Region  string
	ModelID string
}

// EmbeddingsConfig holds embeddings configuration
type EmbeddingsConfig struct {
	Provider   string
	Model      string
	Dimensions int
}

// OllamaConfig holds Ollama configuration
type OllamaConfig struct {
	BaseURL string
}

// StorageConfig holds storage paths configuration
type StorageConfig struct {
	UploadDir       string
	VectorStorePath string
	BadgerDBPath    string
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Key string
}

// RAGConfig holds RAG-specific configuration
type RAGConfig struct {
	MaxContextChunks int
	ChunkSize        int
	ChunkOverlap     int
	SystemPrompt     string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "3000"),
			Env:  getEnv("ENV", "development"),
		},
		OpenRouter: OpenRouterConfig{
			APIKey: getEnv("OPENROUTER_API_KEY", ""),
			Model:  getEnv("OPENROUTER_MODEL", "anthropic/claude-3.5-sonnet"),
		},
		Bedrock: BedrockConfig{
			APIKey:  getEnv("BEDROCK_API_KEY", ""),
			Region:  getEnv("BEDROCK_REGION", "eu-north-1"),
			ModelID: getEnv("BEDROCK_MODEL_ID", "openai.gpt-oss-20b-1:0"),
		},
		Ollama: OllamaConfig{
			BaseURL: getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		},
		Embeddings: EmbeddingsConfig{
			Provider:   getEnv("EMBEDDING_PROVIDER", "ollama"),
			Model:      getEnv("EMBEDDING_MODEL", "all-minilm:33m"),
			Dimensions: getEnvAsInt("EMBEDDING_DIMENSIONS", 384),
		},
		Storage: StorageConfig{
			UploadDir:       getEnv("UPLOAD_DIR", "./data/uploads"),
			VectorStorePath: getEnv("VECTOR_STORE_PATH", "./data/vectors"),
			BadgerDBPath:    getEnv("BADGER_DB_PATH", "./data/badger"),
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", ""),
		},
		RAG: RAGConfig{
			MaxContextChunks: getEnvAsInt("MAX_CONTEXT_CHUNKS", 5),
			ChunkSize:        getEnvAsInt("CHUNK_SIZE", 1000),
			ChunkOverlap:     getEnvAsInt("CHUNK_OVERLAP", 200),
			SystemPrompt:     getEnv("SYSTEM_PROMPT", "You are a helpful AI assistant. Answer questions based on the provided context."),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.OpenRouter.APIKey == "" && c.Bedrock.APIKey == "" {
		return fmt.Errorf("at least one LLM provider API key must be set (OPENROUTER_API_KEY or BEDROCK_API_KEY)")
	}

	if c.Embeddings.Provider != "ollama" && c.Embeddings.Provider != "openrouter" && c.Embeddings.Provider != "bedrock" {
		return fmt.Errorf("EMBEDDING_PROVIDER must be 'ollama', 'openrouter', or 'bedrock'")
	}

	if c.RAG.ChunkSize <= 0 {
		return fmt.Errorf("CHUNK_SIZE must be greater than 0")
	}

	if c.RAG.ChunkOverlap < 0 || c.RAG.ChunkOverlap >= c.RAG.ChunkSize {
		return fmt.Errorf("CHUNK_OVERLAP must be between 0 and CHUNK_SIZE")
	}

	if c.RAG.MaxContextChunks <= 0 {
		return fmt.Errorf("MAX_CONTEXT_CHUNKS must be greater than 0")
	}

	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
