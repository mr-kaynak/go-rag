package settings

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

// Store handles settings storage with BadgerDB
type Store struct {
	db     *badger.DB
	cipher cipher.AEAD
}

// APIKeys holds API keys for different providers
type APIKeys struct {
	OpenRouter string `json:"openrouter,omitempty"`
	Bedrock    string `json:"bedrock,omitempty"`
}

// ModelConfig represents a model configuration
type ModelConfig struct {
	ID          string  `json:"id"`
	Provider    string  `json:"provider"`
	ModelID     string  `json:"model_id"`
	DisplayName string  `json:"display_name"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// SystemPrompt represents a system prompt configuration
type SystemPrompt struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Prompt  string `json:"prompt"`
	Default bool   `json:"default"`
}

// BadgerDB key prefixes
const (
	prefixAPIKeys       = "apikeys:"
	prefixModel         = "model:"
	prefixSystemPrompt  = "prompt:"
	prefixDefaultPrompt = "default_prompt"
)

// New creates a new settings store (opens its own DB)
func New(dbPath, encryptionKey string) (*Store, error) {
	// Open BadgerDB
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable badger logs
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return newStoreWithDB(db, encryptionKey, true), nil
}

// NewWithDB creates a new settings store using an existing DB connection
func NewWithDB(db *badger.DB, encryptionKey string) *Store {
	return newStoreWithDB(db, encryptionKey, false)
}

// newStoreWithDB internal constructor
func newStoreWithDB(db *badger.DB, encryptionKey string, ownsDB bool) *Store {
	// Setup encryption for sensitive data
	var aesgcm cipher.AEAD
	if encryptionKey != "" {
		key := []byte(encryptionKey)
		// Pad or truncate to 32 bytes for AES-256
		if len(key) < 32 {
			padded := make([]byte, 32)
			copy(padded, key)
			key = padded
		} else if len(key) > 32 {
			key = key[:32]
		}

		block, err := aes.NewCipher(key)
		if err == nil {
			aesgcm, _ = cipher.NewGCM(block)
		}
	}

	return &Store{
		db:     db,
		cipher: aesgcm,
	}
}

// Close closes the database (only if this store owns it)
func (s *Store) Close() error {
	// Note: When using NewWithDB, the caller is responsible for closing the DB
	return nil
}

// === API Keys ===

// SaveAPIKeys saves encrypted API keys
func (s *Store) SaveAPIKeys(keys APIKeys) error {
	data, err := json.Marshal(keys)
	if err != nil {
		return err
	}

	// Encrypt if cipher is available
	if s.cipher != nil {
		data = s.encrypt(data)
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(prefixAPIKeys+"default"), data)
	})
}

// GetAPIKeys retrieves and decrypts API keys
func (s *Store) GetAPIKeys() (APIKeys, error) {
	var keys APIKeys

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixAPIKeys + "default"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			// Decrypt if cipher is available
			if s.cipher != nil {
				val = s.decrypt(val)
			}

			return json.Unmarshal(val, &keys)
		})
	})

	if err == badger.ErrKeyNotFound {
		return APIKeys{}, nil // Return empty keys if not found
	}

	return keys, err
}

// === Models ===

// SaveModel saves a model configuration
func (s *Store) SaveModel(model ModelConfig) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}

	data, err := json.Marshal(model)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte(prefixModel + model.ID)
		return txn.Set(key, data)
	})
}

// GetModel retrieves a model by ID
func (s *Store) GetModel(id string) (ModelConfig, error) {
	var model ModelConfig

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixModel + id))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &model)
		})
	})

	return model, err
}

// ListModels lists all models, optionally filtered by provider
func (s *Store) ListModels(provider string) ([]ModelConfig, error) {
	var models []ModelConfig

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefixModel)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var model ModelConfig
				if err := json.Unmarshal(val, &model); err != nil {
					return err
				}

				// Filter by provider if specified
				if provider == "" || model.Provider == provider {
					models = append(models, model)
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return models, err
}

// DeleteModel deletes a model
func (s *Store) DeleteModel(id string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(prefixModel + id))
	})
}

// === System Prompts ===

// SaveSystemPrompt saves a system prompt
func (s *Store) SaveSystemPrompt(prompt SystemPrompt) error {
	// If it's a default prompt, try to update existing one
	if prompt.Default {
		existingPrompts, _ := s.ListSystemPrompts()
		for _, existing := range existingPrompts {
			if existing.Default || existing.Name == prompt.Name {
				prompt.ID = existing.ID // Reuse existing ID
				break
			}
		}
	}

	if prompt.ID == "" {
		prompt.ID = uuid.New().String()
	}

	data, err := json.Marshal(prompt)
	if err != nil {
		return err
	}

	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte(prefixSystemPrompt + prompt.ID)
		if err := txn.Set(key, data); err != nil {
			return err
		}

		// If this is the default prompt, save its ID
		if prompt.Default {
			return txn.Set([]byte(prefixDefaultPrompt), []byte(prompt.ID))
		}

		return nil
	})
}

// GetSystemPrompt retrieves a system prompt by ID
func (s *Store) GetSystemPrompt(id string) (SystemPrompt, error) {
	var prompt SystemPrompt

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixSystemPrompt + id))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &prompt)
		})
	})

	return prompt, err
}

// GetDefaultSystemPrompt retrieves the default system prompt
func (s *Store) GetDefaultSystemPrompt() (SystemPrompt, error) {
	var promptID string

	// Get default prompt ID
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixDefaultPrompt))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			promptID = string(val)
			return nil
		})
	})

	if err == badger.ErrKeyNotFound {
		// Return empty prompt if no default is set
		return SystemPrompt{}, nil
	}

	if err != nil {
		return SystemPrompt{}, err
	}

	// Get the prompt
	return s.GetSystemPrompt(promptID)
}

// ListSystemPrompts lists all system prompts
func (s *Store) ListSystemPrompts() ([]SystemPrompt, error) {
	var prompts []SystemPrompt

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefixSystemPrompt)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var prompt SystemPrompt
				if err := json.Unmarshal(val, &prompt); err != nil {
					return err
				}
				prompts = append(prompts, prompt)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return prompts, err
}

// DeleteSystemPrompt deletes a system prompt
func (s *Store) DeleteSystemPrompt(id string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(prefixSystemPrompt + id))
	})
}

// === Encryption Helpers ===

func (s *Store) encrypt(data []byte) []byte {
	if s.cipher == nil {
		return data
	}

	nonce := make([]byte, s.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return data
	}

	return s.cipher.Seal(nonce, nonce, data, nil)
}

func (s *Store) decrypt(data []byte) []byte {
	if s.cipher == nil {
		return data
	}

	nonceSize := s.cipher.NonceSize()
	if len(data) < nonceSize {
		return data
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := s.cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return data
	}

	return plaintext
}
