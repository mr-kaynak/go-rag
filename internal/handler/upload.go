package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mrkaynak/rag/internal/config"
	"github.com/mrkaynak/rag/internal/models"
	"github.com/mrkaynak/rag/internal/service/document"
	"github.com/mrkaynak/rag/internal/service/embeddings"
	"github.com/mrkaynak/rag/internal/service/vector"
	"github.com/mrkaynak/rag/pkg/errors"
	"go.uber.org/zap"
)

const (
	// MaxFileSize is the maximum allowed file size for uploads (50MB)
	MaxFileSize = 50 * 1024 * 1024
)

var (
	// AllowedMimeTypes lists the permitted file types for upload
	AllowedMimeTypes = map[string]bool{
		"text/plain":      true,
		"text/markdown":   true,
		"text/x-markdown": true,
	}

	// AllowedExtensions lists the permitted file extensions
	AllowedExtensions = map[string]bool{
		".txt": true,
		".md":  true,
	}
)

// UploadHandler handles document upload requests
type UploadHandler struct {
	cfg           *config.Config
	logger        *zap.Logger
	docService    *document.Service
	embeddingsSvc *embeddings.Service
	vectorStore   *vector.Store
	metadataStore *document.MetadataStore
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(
	cfg *config.Config,
	logger *zap.Logger,
	docService *document.Service,
	embeddingsSvc *embeddings.Service,
	vectorStore *vector.Store,
	metadataStore *document.MetadataStore,
) *UploadHandler {
	return &UploadHandler{
		cfg:           cfg,
		logger:        logger,
		docService:    docService,
		embeddingsSvc: embeddingsSvc,
		vectorStore:   vectorStore,
		metadataStore: metadataStore,
	}
}

// detectAndValidateFileType detects the file type and validates it against allowed types
func detectAndValidateFileType(file *multipart.FileHeader) (string, error) {
	// First check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedExtensions[ext] {
		return "", fmt.Errorf("file extension '%s' is not allowed. Supported formats: .txt, .md", ext)
	}

	// Open file to detect content type
	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file for type detection: %w", err)
	}
	defer f.Close()

	// Read first 512 bytes for content type detection
	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for type detection: %w", err)
	}

	// Detect content type
	contentType := http.DetectContentType(buffer[:n])

	// Validate content type
	if !AllowedMimeTypes[contentType] {
		return "", fmt.Errorf("file type '%s' is not allowed. Supported formats: text/plain, text/markdown", contentType)
	}

	return contentType, nil
}

// Upload handles document upload and processing
func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	// Get API key from config based on provider (not needed for Ollama)
	var apiKey string
	switch h.cfg.Embeddings.Provider {
	case "ollama":
		// No API key needed for Ollama
		apiKey = ""
	case "openrouter":
		apiKey = h.cfg.OpenRouter.APIKey
	case "bedrock":
		apiKey = h.cfg.Bedrock.APIKey
	}

	if h.cfg.Embeddings.Provider != "ollama" && apiKey == "" {
		return h.sendError(c, errors.Unauthorized("API key is not configured"))
	}

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Warn("failed to parse file", zap.Error(err))
		return h.sendError(c, errors.BadRequest("file is required. Please select a file to upload."))
	}

	// Validate file size
	if file.Size > MaxFileSize {
		h.logger.Warn("file too large",
			zap.String("filename", file.Filename),
			zap.Int64("size", file.Size),
			zap.Int64("max_size", MaxFileSize),
		)
		return h.sendError(c, errors.BadRequest(
			fmt.Sprintf("file too large. Maximum file size is %d MB", MaxFileSize/(1024*1024)),
		))
	}

	// Validate file is not empty
	if file.Size == 0 {
		h.logger.Warn("empty file uploaded", zap.String("filename", file.Filename))
		return h.sendError(c, errors.BadRequest("uploaded file is empty. Please select a valid file."))
	}

	// Detect and validate file type
	fileType, err := detectAndValidateFileType(file)
	if err != nil {
		h.logger.Warn("invalid file type",
			zap.String("filename", file.Filename),
			zap.Error(err),
		)
		return h.sendError(c, errors.BadRequest(err.Error()))
	}

	h.logger.Info("processing file upload",
		zap.String("filename", file.Filename),
		zap.Int64("size", file.Size),
		zap.String("type", fileType),
	)

	// Open uploaded file
	fileContent, err := file.Open()
	if err != nil {
		h.logger.Error("failed to open uploaded file", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to open file"))
	}
	defer fileContent.Close()

	// Process document
	doc, err := h.docService.ProcessUpload(file.Filename, fileContent)
	if err != nil {
		h.logger.Error("failed to process document", zap.Error(err))
		return h.sendError(c, err)
	}

	h.logger.Info("document processed",
		zap.String("doc_id", doc.ID),
		zap.Int("chunks", len(doc.Chunks)),
	)

	// Generate embeddings
	chunks, err := h.embeddingsSvc.GenerateEmbeddings(doc.Chunks, apiKey)
	if err != nil {
		h.logger.Error("failed to generate embeddings", zap.Error(err))
		return h.sendError(c, err)
	}

	h.logger.Info("embeddings generated",
		zap.String("doc_id", doc.ID),
		zap.Int("chunks", len(chunks)),
	)

	// Store in vector store
	if err := h.vectorStore.Add(chunks); err != nil {
		h.logger.Error("failed to add to vector store", zap.Error(err))
		return h.sendError(c, err)
	}

	// Save metadata
	metadata := document.DocumentMetadata{
		ID:         doc.ID,
		FileName:   doc.FileName,
		FileSize:   file.Size,
		FileType:   fileType,
		ChunkCount: len(chunks),
		UploadedAt: doc.CreatedAt,
	}

	if err := h.metadataStore.Add(metadata); err != nil {
		h.logger.Error("failed to save metadata", zap.Error(err))
		// Non-fatal, continue
	}

	h.logger.Info("document indexed successfully",
		zap.String("doc_id", doc.ID),
		zap.String("filename", file.Filename),
	)

	return c.Status(fiber.StatusCreated).JSON(models.UploadResponse{
		DocumentID: doc.ID,
		FileName:   doc.FileName,
		ChunkCount: len(chunks),
	})
}

// ListDocuments returns all uploaded documents (GET /api/v1/documents)
func (h *UploadHandler) ListDocuments(c *fiber.Ctx) error {
	docs, err := h.metadataStore.List()
	if err != nil {
		h.logger.Error("failed to list documents", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to list documents"))
	}

	return c.Status(fiber.StatusOK).JSON(docs)
}

// DeleteDocument deletes a document and its chunks (DELETE /api/v1/documents/:id)
func (h *UploadHandler) DeleteDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.sendError(c, errors.BadRequest("document id is required"))
	}

	// Delete from metadata
	if err := h.metadataStore.Delete(id); err != nil {
		h.logger.Error("failed to delete document metadata", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to delete document"))
	}

	// Delete chunks from vector store
	if err := h.vectorStore.DeleteByDocID(id); err != nil {
		h.logger.Error("failed to delete document chunks", zap.Error(err))
		return h.sendError(c, errors.InternalWrap(err, "failed to delete document chunks"))
	}

	h.logger.Info("document deleted successfully", zap.String("doc_id", id))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "document deleted successfully",
	})
}

// sendError sends an error response
func (h *UploadHandler) sendError(c *fiber.Ctx, err error) error {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		appErr = errors.Internal("internal server error")
	}

	return c.Status(appErr.Code).JSON(models.ErrorResponse{
		Error: appErr.Message,
		Code:  appErr.Code,
	})
}
