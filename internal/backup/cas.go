package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ChunkStore handles persistent storage of unique data blocks
type ChunkStore struct {
	BaseDir string
}

// NewChunkStore creates a new chunk store
func NewChunkStore(baseDir string) *ChunkStore {
	return &ChunkStore{
		BaseDir: baseDir,
	}
}

// Put saves a block if it doesn't exist
func (s *ChunkStore) Put(hash string, data []byte) error {
	if len(hash) < 4 {
		return fmt.Errorf("invalid hash length")
	}

	// Two-level directory structure: chunks/ab/cd/hash
	dir := filepath.Join(s.BaseDir, hash[:2], hash[2:4])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, hash)
	if _, err := os.Stat(path); err == nil {
		return nil // Already exists
	}

	return os.WriteFile(path, data, 0644)
}

// Get retrieves a block by hash
func (s *ChunkStore) Get(hash string) ([]byte, error) {
	if len(hash) < 4 {
		return nil, fmt.Errorf("invalid hash length")
	}

	path := filepath.Join(s.BaseDir, hash[:2], hash[2:4], hash)
	return os.ReadFile(path)
}

// Exists checks if a block is already stored
func (s *ChunkStore) Exists(hash string) bool {
	if len(hash) < 4 {
		return false
	}
	path := filepath.Join(s.BaseDir, hash[:2], hash[2:4], hash)
	_, err := os.Stat(path)
	return err == nil
}

// GetPath returns the physical path of a chunk
func (s *ChunkStore) GetPath(hash string) string {
	return filepath.Join(s.BaseDir, hash[:2], hash[2:4], hash)
}

// CopyToWriter copies chunk data to a writer
func (s *ChunkStore) CopyToWriter(hash string, w io.Writer) error {
	path := s.GetPath(hash)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}
