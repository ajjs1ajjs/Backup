package s3

import (
	"context"
	"fmt"
	"time"

	"novabackup/pkg/models"
)

// StorageInterface defines the interface for storage backends
type StorageInterface interface {
	StoreChunk(ctx context.Context, hash string, data []byte) (string, error)
	GetChunk(ctx context.Context, hash string) ([]byte, error)
	ChunkExists(ctx context.Context, hash string) (bool, error)
	DeleteChunk(ctx context.Context, hash string) error
	ListChunks(ctx context.Context, prefix string) ([]interface{}, error)
	GetStorageInfo(ctx context.Context) (*models.StorageInfo, error)
	TestConnection(ctx context.Context) error
	Cleanup(ctx context.Context, olderThan time.Time) error
	Close() error
}

// Ensure S3Engine implements StorageInterface
var _ StorageInterface = (*S3Engine)(nil)

// StorageManager manages multiple storage backends
type StorageManager struct {
	backends map[string]StorageInterface
	primary  string
}

// NewStorageManager creates a new storage manager
func NewStorageManager() *StorageManager {
	return &StorageManager{
		backends: make(map[string]StorageInterface),
	}
}

// AddBackend adds a storage backend
func (sm *StorageManager) AddBackend(name string, backend StorageInterface, isPrimary bool) error {
	if backend == nil {
		return fmt.Errorf("backend cannot be nil")
	}
	
	sm.backends[name] = backend
	if isPrimary || sm.primary == "" {
		sm.primary = name
	}
	
	return nil
}

// GetBackend returns a storage backend by name
func (sm *StorageManager) GetBackend(name string) (StorageInterface, error) {
	backend, exists := sm.backends[name]
	if !exists {
		return nil, fmt.Errorf("backend '%s' not found", name)
	}
	return backend, nil
}

// GetPrimary returns the primary storage backend
func (sm *StorageManager) GetPrimary() (StorageInterface, error) {
	if sm.primary == "" {
		return nil, fmt.Errorf("no primary backend set")
	}
	return sm.GetBackend(sm.primary)
}

// ListBackends returns all backend names
func (sm *StorageManager) ListBackends() []string {
	var names []string
	for name := range sm.backends {
		names = append(names, name)
	}
	return names
}

// SetPrimary sets the primary storage backend
func (sm *StorageManager) SetPrimary(name string) error {
	if _, exists := sm.backends[name]; !exists {
		return fmt.Errorf("backend '%s' not found", name)
	}
	sm.primary = name
	return nil
}

// StoreChunk stores a chunk using the primary backend
func (sm *StorageManager) StoreChunk(ctx context.Context, hash string, data []byte) (string, error) {
	backend, err := sm.GetPrimary()
	if err != nil {
		return "", err
	}
	return backend.StoreChunk(ctx, hash, data)
}

// GetChunk retrieves a chunk using the primary backend
func (sm *StorageManager) GetChunk(ctx context.Context, hash string) ([]byte, error) {
	backend, err := sm.GetPrimary()
	if err != nil {
		return nil, err
	}
	return backend.GetChunk(ctx, hash)
}

// ChunkExists checks if a chunk exists in the primary backend
func (sm *StorageManager) ChunkExists(ctx context.Context, hash string) (bool, error) {
	backend, err := sm.GetPrimary()
	if err != nil {
		return false, err
	}
	return backend.ChunkExists(ctx, hash)
}

// DeleteChunk removes a chunk from the primary backend
func (sm *StorageManager) DeleteChunk(ctx context.Context, hash string) error {
	backend, err := sm.GetPrimary()
	if err != nil {
		return err
	}
	return backend.DeleteChunk(ctx, hash)
}

// TestAllConnections tests all backend connections
func (sm *StorageManager) TestAllConnections(ctx context.Context) map[string]error {
	results := make(map[string]error)
	
	for name, backend := range sm.backends {
		err := backend.TestConnection(ctx)
		results[name] = err
	}
	
	return results
}

// Close closes all storage backends
func (sm *StorageManager) Close() error {
	var lastErr error
	
	for name, backend := range sm.backends {
		if err := backend.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close backend '%s': %w", name, err)
		}
	}
	
	return lastErr
}
