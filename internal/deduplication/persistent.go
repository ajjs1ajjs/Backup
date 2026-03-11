package deduplication

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	"novabackup/internal/database"
)

// PersistentDeduplicationManager implements DeduplicationManager with DB persistence
type PersistentDeduplicationManager struct {
	db          *database.Connection
	storagePath string
}

func NewPersistentDeduplicationManager(db *database.Connection, storagePath string) *PersistentDeduplicationManager {
	return &PersistentDeduplicationManager{
		db:          db,
		storagePath: storagePath,
	}
}

// StoreChunk splits, hashes and stores data if unique
func (p *PersistentDeduplicationManager) StoreChunk(ctx context.Context, data []byte) (string, error) {
	hash := p.CalculateHash(data)
	
	// 1. Check if chunk exists in DB
	exists, err := p.db.ChunkExists(hash)
	if err != nil {
		return "", err
	}

	if exists {
		// Increment reference count
		err = p.db.IncrementChunkRef(hash)
		return hash, err
	}

	// 2. Store new chunk file
	chunkPath := filepath.Join(p.storagePath, hash[:2], hash)
	if err := os.MkdirAll(filepath.Dir(chunkPath), 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		return "", err
	}

	// 3. Add to DB
	err = p.db.CreateChunk(hash, int64(len(data)), chunkPath)
	return hash, err
}

func (p *PersistentDeduplicationManager) CalculateHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// Chunker configuration for fixed-size chunking
type ChunkerConfig struct {
	ChunkSize int
}

func (p *PersistentDeduplicationManager) ProcessData(ctx context.Context, data []byte) ([]string, error) {
	// Simple fixed size chunker for now
	chunkSize := 1024 * 1024 // 1MB blocks
	var hashes []string

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		
		hash, err := p.StoreChunk(ctx, data[i:end])
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}

	return hashes, nil
}
