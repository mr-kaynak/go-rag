package document

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mrkaynak/rag/internal/config"
	"github.com/mrkaynak/rag/internal/models"
	"github.com/mrkaynak/rag/pkg/errors"
)

// Service handles document operations
type Service struct {
	cfg *config.Config
}

// New creates a new document service
func New(cfg *config.Config) (*Service, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.Storage.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return &Service{
		cfg: cfg,
	}, nil
}

// ProcessUpload processes an uploaded file
func (s *Service) ProcessUpload(filename string, reader io.Reader) (*models.Document, error) {
	docID := uuid.New().String()

	// Read file content
	content, err := s.readContent(reader)
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to read file content")
	}

	// Save original file
	if err := s.saveFile(docID, filename, content); err != nil {
		return nil, errors.InternalWrap(err, "failed to save file")
	}

	// Create document
	doc := &models.Document{
		ID:        docID,
		FileName:  filename,
		Content:   content,
		CreatedAt: time.Now(),
	}

	// Split into chunks
	chunks := s.chunkText(doc.ID, content)
	doc.Chunks = chunks

	return doc, nil
}

// readContent reads content from reader based on file type
func (s *Service) readContent(reader io.Reader) (string, error) {
	var builder strings.Builder
	scanner := bufio.NewScanner(reader)

	// Set a larger buffer for long lines
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		builder.WriteString(scanner.Text())
		builder.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading content: %w", err)
	}

	content := builder.String()
	if content == "" {
		return "", fmt.Errorf("file is empty")
	}

	return content, nil
}

// saveFile saves file to disk
func (s *Service) saveFile(docID, filename, content string) error {
	filePath := filepath.Join(s.cfg.Storage.UploadDir, fmt.Sprintf("%s_%s", docID, filename))

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// chunkText splits text into overlapping chunks
func (s *Service) chunkText(docID, text string) []models.Chunk {
	chunkSize := s.cfg.RAG.ChunkSize
	overlap := s.cfg.RAG.ChunkOverlap

	var chunks []models.Chunk
	runes := []rune(text)
	index := 0

	for i := 0; i < len(runes); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunkContent := strings.TrimSpace(string(runes[i:end]))
		if chunkContent == "" {
			continue
		}

		chunk := models.Chunk{
			ID:      uuid.New().String(),
			DocID:   docID,
			Content: chunkContent,
			Index:   index,
		}

		chunks = append(chunks, chunk)
		index++

		if end >= len(runes) {
			break
		}
	}

	return chunks
}

// GetDocument retrieves a document by ID
func (s *Service) GetDocument(docID string) (*models.Document, error) {
	// In a production system, this would query a database
	// For now, this is a placeholder
	return nil, errors.NotFound("document not found")
}
