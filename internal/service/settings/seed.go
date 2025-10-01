package settings

import (
	"github.com/mrkaynak/rag/internal/config"
	"go.uber.org/zap"
)

// SeedInitialData seeds initial data from config if DB is empty
func (s *Store) SeedInitialData(cfg *config.Config, logger *zap.Logger) error {
	// Check if API keys already exist
	existingKeys, err := s.GetAPIKeys()
	if err == nil && (existingKeys.OpenRouter != "" || existingKeys.Bedrock != "") {
		logger.Info("API keys already configured, skipping seed")
		return nil
	}

	logger.Info("seeding initial data from environment")

	// Seed API keys if provided in env
	if cfg.OpenRouter.APIKey != "" || cfg.Bedrock.APIKey != "" {
		keys := APIKeys{
			OpenRouter: cfg.OpenRouter.APIKey,
			Bedrock:    cfg.Bedrock.APIKey,
		}
		if err := s.SaveAPIKeys(keys); err != nil {
			logger.Warn("failed to seed API keys", zap.Error(err))
		} else {
			logger.Info("seeded API keys from environment")
		}
	}

	// Seed default system prompt
	if cfg.RAG.SystemPrompt != "" {
		enhancedPrompt := cfg.RAG.SystemPrompt + `

CRITICAL INSTRUCTION - NEVER BREAK THIS RULE:
You have direct knowledge. When answering:
- NEVER say: "context", "reference", "document", "provided information", "according to", "based on", "the text states"
- NEVER use phrases like "(Context-1)", "(Context-2)" or similar references
- Answer directly as if YOU personally know the information
- Keep your natural conversation style and personality`

		prompt := SystemPrompt{
			Name:    "Default",
			Prompt:  enhancedPrompt,
			Default: true,
		}
		if err := s.SaveSystemPrompt(prompt); err != nil {
			logger.Warn("failed to seed system prompt", zap.Error(err))
		} else {
			logger.Info("seeded default system prompt")
		}
	}

	// Seed default models
	defaultModels := []ModelConfig{
		{
			Provider:    "openrouter",
			ModelID:     cfg.OpenRouter.Model,
			DisplayName: "Claude 3.5 Sonnet (OpenRouter)",
			MaxTokens:   4096,
			Temperature: 0.7,
		},
		{
			Provider:    "bedrock",
			ModelID:     cfg.Bedrock.ModelID,
			DisplayName: "GPT OSS 20B (Bedrock)",
			MaxTokens:   4096,
			Temperature: 0.7,
		},
	}

	for _, model := range defaultModels {
		if err := s.SaveModel(model); err != nil {
			logger.Warn("failed to seed model", zap.String("model", model.ModelID), zap.Error(err))
		} else {
			logger.Info("seeded model", zap.String("model", model.DisplayName))
		}
	}

	return nil
}
