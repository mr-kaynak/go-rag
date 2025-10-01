package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mrkaynak/rag/internal/models"
	"github.com/mrkaynak/rag/internal/service/settings"
	"github.com/mrkaynak/rag/pkg/errors"
	"go.uber.org/zap"
)

// SettingsHandler handles settings-related requests
type SettingsHandler struct {
	logger      *zap.Logger
	settingsSvc *settings.Store
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(logger *zap.Logger, settingsSvc *settings.Store) *SettingsHandler {
	return &SettingsHandler{
		logger:      logger,
		settingsSvc: settingsSvc,
	}
}

// === API Keys ===

// SaveAPIKeys saves API keys (POST /api/v1/settings/api-keys)
func (h *SettingsHandler) SaveAPIKeys(c *fiber.Ctx) error {
	var keys settings.APIKeys
	if err := c.BodyParser(&keys); err != nil {
		return h.sendError(c, errors.BadRequest("invalid request body"))
	}

	if err := h.settingsSvc.SaveAPIKeys(keys); err != nil {
		h.logger.Error("failed to save API keys", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to save API keys"))
	}

	h.logger.Info("API keys saved successfully")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "API keys saved successfully",
	})
}

// GetAPIKeys returns API keys (masked) (GET /api/v1/settings/api-keys)
func (h *SettingsHandler) GetAPIKeys(c *fiber.Ctx) error {
	keys, err := h.settingsSvc.GetAPIKeys()
	if err != nil {
		h.logger.Error("failed to get API keys", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to get API keys"))
	}

	// Mask keys for security (show only last 4 characters)
	masked := settings.APIKeys{}
	if keys.OpenRouter != "" {
		if len(keys.OpenRouter) > 4 {
			masked.OpenRouter = "****" + keys.OpenRouter[len(keys.OpenRouter)-4:]
		} else {
			masked.OpenRouter = "****"
		}
	}
	if keys.Bedrock != "" {
		if len(keys.Bedrock) > 4 {
			masked.Bedrock = "****" + keys.Bedrock[len(keys.Bedrock)-4:]
		} else {
			masked.Bedrock = "****"
		}
	}

	return c.Status(fiber.StatusOK).JSON(masked)
}

// === Models ===

// SaveModel saves a model configuration (POST /api/v1/settings/models)
func (h *SettingsHandler) SaveModel(c *fiber.Ctx) error {
	var model settings.ModelConfig
	if err := c.BodyParser(&model); err != nil {
		return h.sendError(c, errors.BadRequest("invalid request body"))
	}

	if model.Provider == "" || model.ModelID == "" || model.DisplayName == "" {
		return h.sendError(c, errors.BadRequest("provider, model_id, and display_name are required"))
	}

	if err := h.settingsSvc.SaveModel(model); err != nil {
		h.logger.Error("failed to save model", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to save model"))
	}

	h.logger.Info("model saved", zap.String("model_id", model.ModelID))

	return c.Status(fiber.StatusCreated).JSON(model)
}

// ListModels lists all models (GET /api/v1/settings/models?provider=openrouter)
func (h *SettingsHandler) ListModels(c *fiber.Ctx) error {
	provider := c.Query("provider", "")

	models, err := h.settingsSvc.ListModels(provider)
	if err != nil {
		h.logger.Error("failed to list models", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to list models"))
	}

	return c.Status(fiber.StatusOK).JSON(models)
}

// DeleteModel deletes a model (DELETE /api/v1/settings/models/:id)
func (h *SettingsHandler) DeleteModel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.sendError(c, errors.BadRequest("model id is required"))
	}

	if err := h.settingsSvc.DeleteModel(id); err != nil {
		h.logger.Error("failed to delete model", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to delete model"))
	}

	h.logger.Info("model deleted", zap.String("model_id", id))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "model deleted successfully",
	})
}

// === System Prompts ===

// SaveSystemPrompt saves a system prompt (POST /api/v1/settings/system-prompts)
func (h *SettingsHandler) SaveSystemPrompt(c *fiber.Ctx) error {
	var prompt settings.SystemPrompt
	if err := c.BodyParser(&prompt); err != nil {
		return h.sendError(c, errors.BadRequest("invalid request body"))
	}

	if prompt.Name == "" || prompt.Prompt == "" {
		return h.sendError(c, errors.BadRequest("name and prompt are required"))
	}

	if err := h.settingsSvc.SaveSystemPrompt(prompt); err != nil {
		h.logger.Error("failed to save system prompt", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to save system prompt"))
	}

	h.logger.Info("system prompt saved", zap.String("prompt_id", prompt.ID))

	return c.Status(fiber.StatusCreated).JSON(prompt)
}

// ListSystemPrompts lists all system prompts (GET /api/v1/settings/system-prompts)
func (h *SettingsHandler) ListSystemPrompts(c *fiber.Ctx) error {
	prompts, err := h.settingsSvc.ListSystemPrompts()
	if err != nil {
		h.logger.Error("failed to list system prompts", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to list system prompts"))
	}

	return c.Status(fiber.StatusOK).JSON(prompts)
}

// GetDefaultSystemPrompt returns the default system prompt (GET /api/v1/settings/system-prompts/default)
func (h *SettingsHandler) GetDefaultSystemPrompt(c *fiber.Ctx) error {
	prompt, err := h.settingsSvc.GetDefaultSystemPrompt()
	if err != nil {
		h.logger.Error("failed to get default system prompt", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to get default system prompt"))
	}

	return c.Status(fiber.StatusOK).JSON(prompt)
}

// DeleteSystemPrompt deletes a system prompt (DELETE /api/v1/settings/system-prompts/:id)
func (h *SettingsHandler) DeleteSystemPrompt(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.sendError(c, errors.BadRequest("prompt id is required"))
	}

	if err := h.settingsSvc.DeleteSystemPrompt(id); err != nil {
		h.logger.Error("failed to delete system prompt", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to delete system prompt"))
	}

	h.logger.Info("system prompt deleted", zap.String("prompt_id", id))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "system prompt deleted successfully",
	})
}

// sendError sends an error response
func (h *SettingsHandler) sendError(c *fiber.Ctx, err error) error {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		appErr = errors.Internal("internal server error")
	}

	return c.Status(appErr.Code).JSON(models.ErrorResponse{
		Error: appErr.Message,
		Code:  appErr.Code,
	})
}
