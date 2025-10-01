package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mrkaynak/rag/internal/config"
	"github.com/mrkaynak/rag/internal/models"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	version string
	cfg     *config.Config
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(version string, cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		version: version,
		cfg:     cfg,
	}
}

// Health returns the health status
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(models.HealthResponse{
		Status:  "healthy",
		Version: h.version,
	})
}

// GetSystemPrompt returns the system prompt from config
func (h *HealthHandler) GetSystemPrompt(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"system_prompt": h.cfg.RAG.SystemPrompt,
	})
}
