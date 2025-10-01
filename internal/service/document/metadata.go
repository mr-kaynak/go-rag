package document

import (
	"encoding/json"
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

// MetadataStore handles document metadata storage with BadgerDB
type MetadataStore struct {
	db *badger.DB
}

// DocumentMetadata represents document metadata
type DocumentMetadata struct {
	ID         string    `json:"id"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	FileType   string    `json:"file_type"`
	ChunkCount int       `json:"chunk_count"`
	UploadedAt time.Time `json:"uploaded_at"`
}

const prefixDocument = "doc:"

// NewMetadataStore creates a new metadata store
func NewMetadataStore(db *badger.DB) *MetadataStore {
	return &MetadataStore{
		db: db,
	}
}

// Add adds a document metadata
func (m *MetadataStore) Add(doc DocumentMetadata) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	return m.db.Update(func(txn *badger.Txn) error {
		key := []byte(prefixDocument + doc.ID)
		return txn.Set(key, data)
	})
}

// Get retrieves a document metadata by ID
func (m *MetadataStore) Get(id string) (DocumentMetadata, error) {
	var doc DocumentMetadata

	err := m.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixDocument + id))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &doc)
		})
	})

	return doc, err
}

// List returns all document metadata
func (m *MetadataStore) List() ([]DocumentMetadata, error) {
	docs := []DocumentMetadata{} // Initialize as empty array, not nil

	err := m.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefixDocument)

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var doc DocumentMetadata
				if err := json.Unmarshal(val, &doc); err != nil {
					return err
				}
				docs = append(docs, doc)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return docs, err
}

// Delete deletes a document metadata
func (m *MetadataStore) Delete(id string) error {
	return m.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(prefixDocument + id))
	})
}
