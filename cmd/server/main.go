package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/mrkaynak/rag/internal/config"
	"github.com/mrkaynak/rag/internal/handler"
	"github.com/mrkaynak/rag/internal/middleware"
	"github.com/mrkaynak/rag/internal/service/document"
	"github.com/mrkaynak/rag/internal/service/embeddings"
	"github.com/mrkaynak/rag/internal/service/llm"
	"github.com/mrkaynak/rag/internal/service/settings"
	"github.com/mrkaynak/rag/internal/service/vector"
	"go.uber.org/zap"
)

const version = "1.0.0"

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.Server.Env)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("starting RAG server",
		zap.String("version", version),
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize BadgerDB (single instance)
	opts := badger.DefaultOptions(cfg.Storage.BadgerDBPath)
	opts.Logger = nil // Disable badger logs
	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to open badger db: %w", err)
	}
	defer db.Close()

	logger.Info("badger db initialized", zap.String("path", cfg.Storage.BadgerDBPath))

	// Initialize settings service (uses existing db)
	settingsSvc := settings.NewWithDB(db, cfg.Encryption.Key)

	// Seed initial data from env if DB is empty
	if err := settingsSvc.SeedInitialData(cfg, logger); err != nil {
		logger.Warn("failed to seed initial data", zap.Error(err))
	}

	// Initialize services
	docService, err := document.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize document service: %w", err)
	}

	embeddingsSvc := embeddings.New(cfg)

	vectorStore, err := vector.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize vector store: %w", err)
	}

	// Initialize metadata store
	metadataStore := document.NewMetadataStore(db)

	openRouterClient := llm.NewOpenRouterClient(cfg)
	bedrockClient := llm.NewBedrockClient(cfg)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(version, cfg)
	uploadHandler := handler.NewUploadHandler(cfg, logger, docService, embeddingsSvc, vectorStore, metadataStore)
	chatHandler := handler.NewChatHandler(cfg, logger, vectorStore, embeddingsSvc, openRouterClient, bedrockClient, settingsSvc)
	settingsHandler := handler.NewSettingsHandler(logger, settingsSvc)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler:          customErrorHandler(logger),
		DisableStartupMessage: true,
		AppName:               "Enterprise RAG System",
	})

	// Global middleware
	app.Use(middleware.Recovery(logger))
	app.Use(middleware.Logger(logger))
	app.Use(middleware.CORS())

	// Routes
	api := app.Group("/api/v1")

	// Health & Info
	api.Get("/health", healthHandler.Health)
	api.Get("/system-prompt", healthHandler.GetSystemPrompt)

	// Documents
	api.Post("/upload", uploadHandler.Upload)
	api.Get("/documents", uploadHandler.ListDocuments)
	api.Delete("/documents/:id", uploadHandler.DeleteDocument)

	// Chat
	api.Post("/chat", chatHandler.Chat)
	api.Post("/chat/stream", chatHandler.ChatStream)

	// Settings - API Keys
	api.Post("/settings/api-keys", settingsHandler.SaveAPIKeys)
	api.Get("/settings/api-keys", settingsHandler.GetAPIKeys)

	// Settings - Models
	api.Post("/settings/models", settingsHandler.SaveModel)
	api.Get("/settings/models", settingsHandler.ListModels)
	api.Delete("/settings/models/:id", settingsHandler.DeleteModel)

	// Settings - System Prompts
	api.Post("/settings/system-prompts", settingsHandler.SaveSystemPrompt)
	api.Get("/settings/system-prompts", settingsHandler.ListSystemPrompts)
	api.Get("/settings/system-prompts/default", settingsHandler.GetDefaultSystemPrompt)
	api.Delete("/settings/system-prompts/:id", settingsHandler.DeleteSystemPrompt)

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		logger.Info("server listening", zap.String("address", addr))

		if err := app.Listen(addr); err != nil {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	if err := app.Shutdown(); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}

// initLogger initializes the logger based on environment
func initLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// customErrorHandler handles Fiber errors
func customErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		logger.Error("request error",
			zap.Error(err),
			zap.Int("status", code),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)

		return c.Status(code).JSON(fiber.Map{
			"error": err.Error(),
			"code":  code,
		})
	}
}
