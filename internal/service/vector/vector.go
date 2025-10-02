package vector

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/mrkaynak/rag/internal/config"
	"github.com/mrkaynak/rag/internal/models"
	"github.com/mrkaynak/rag/pkg/errors"
)

// Store handles vector storage and similarity search
type Store struct {
	cfg    *config.Config
	mu     sync.RWMutex
	chunks map[string]models.Chunk // chunkID -> Chunk
}

// SimilarityResult represents a similarity search result
type SimilarityResult struct {
	Chunk      models.Chunk
	Similarity float64
}

// New creates a new vector store
func New(cfg *config.Config) (*Store, error) {
	// Ensure vector store directory exists
	if err := os.MkdirAll(cfg.Storage.VectorStorePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create vector store directory: %w", err)
	}

	store := &Store{
		cfg:    cfg,
		chunks: make(map[string]models.Chunk),
	}

	// Load existing vectors
	if err := store.load(); err != nil {
		return nil, fmt.Errorf("failed to load vector store: %w", err)
	}

	return store, nil
}

// Add adds chunks to the vector store
func (s *Store) Add(chunks []models.Chunk) error {
	// Validate first (no lock needed)
	for _, chunk := range chunks {
		if len(chunk.Embedding) == 0 {
			return errors.BadRequest(fmt.Sprintf("chunk %s has no embedding", chunk.ID))
		}
	}

	// Short lock for memory update
	s.mu.Lock()
	for _, chunk := range chunks {
		s.chunks[chunk.ID] = chunk
	}
	// Create snapshot for persistence
	snapshot := s.cloneChunks()
	s.mu.Unlock()

	// Persist outside lock to avoid blocking other operations
	return s.persistSnapshot(snapshot)
}

// Search finds similar chunks using cosine similarity
func (s *Store) Search(queryEmbedding []float64, topK int) ([]SimilarityResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(queryEmbedding) == 0 {
		return nil, errors.BadRequest("query embedding is empty")
	}

	if len(s.chunks) == 0 {
		return []SimilarityResult{}, nil
	}

	// Calculate similarities
	results := make([]SimilarityResult, 0, len(s.chunks))
	for _, chunk := range s.chunks {
		similarity := cosineSimilarity(queryEmbedding, chunk.Embedding)
		results = append(results, SimilarityResult{
			Chunk:      chunk,
			Similarity: similarity,
		})
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Return top K results
	if topK < len(results) {
		results = results[:topK]
	}

	return results, nil
}

// GetAll returns all chunks
func (s *Store) GetAll() []models.Chunk {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chunks := make([]models.Chunk, 0, len(s.chunks))
	for _, chunk := range s.chunks {
		chunks = append(chunks, chunk)
	}

	return chunks
}

// Clear removes all chunks
func (s *Store) Clear() error {
	s.mu.Lock()
	s.chunks = make(map[string]models.Chunk)
	snapshot := s.cloneChunks()
	s.mu.Unlock()

	return s.persistSnapshot(snapshot)
}

// DeleteByDocID removes all chunks belonging to a document
func (s *Store) DeleteByDocID(docID string) error {
	s.mu.Lock()
	// Find and remove chunks with matching DocID
	for id, chunk := range s.chunks {
		if chunk.DocID == docID {
			delete(s.chunks, id)
		}
	}
	snapshot := s.cloneChunks()
	s.mu.Unlock()

	return s.persistSnapshot(snapshot)
}

// cloneChunks creates a deep copy of chunks map (must be called with lock held)
func (s *Store) cloneChunks() map[string]models.Chunk {
	snapshot := make(map[string]models.Chunk, len(s.chunks))
	for id, chunk := range s.chunks {
		snapshot[id] = chunk
	}
	return snapshot
}

// persistSnapshot saves a snapshot of chunks to disk (no lock needed)
func (s *Store) persistSnapshot(snapshot map[string]models.Chunk) error {
	filePath := filepath.Join(s.cfg.Storage.VectorStorePath, "vectors.json")

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chunks: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write vector store: %w", err)
	}

	return nil
}

// persist saves the current vector store to disk (legacy method, kept for load compatibility)
func (s *Store) persist() error {
	s.mu.RLock()
	snapshot := s.cloneChunks()
	s.mu.RUnlock()

	return s.persistSnapshot(snapshot)
}

// load loads the vector store from disk
func (s *Store) load() error {
	filePath := filepath.Join(s.cfg.Storage.VectorStorePath, "vectors.json")

	// If file doesn't exist, start with empty store
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read vector store: %w", err)
	}

	if err := json.Unmarshal(data, &s.chunks); err != nil {
		return fmt.Errorf("failed to unmarshal chunks: %w", err)
	}

	return nil
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
