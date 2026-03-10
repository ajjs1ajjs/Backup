package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
)

// Provider defines the interface for storage backends
type Provider interface {
	Store(ctx context.Context, key string, data io.Reader, size int64) error
	Retrieve(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetStats(ctx context.Context) (*Stats, error)
	Close() error
}

// Stats holds storage statistics
type Stats struct {
	TotalSize   int64  `json:"total_size"`
	UsedSize    int64  `json:"used_size"`
	ObjectCount int64  `json:"object_count"`
	Bucket      string `json:"bucket"`
	Provider    string `json:"provider"`
}

// Engine manages storage providers
type Engine struct {
	providers map[string]Provider
	mu        sync.RWMutex
	basePath  string
}

// NewEngine creates a new storage engine
func NewEngine() *Engine {
	return &Engine{
		providers: make(map[string]Provider),
	}
}

// SetBasePath sets the base path for local storage
func (e *Engine) SetBasePath(path string) {
	e.basePath = path
}

// RegisterProvider registers a storage provider
func (e *Engine) RegisterProvider(name string, provider Provider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.providers[name] = provider
}

// GetProvider gets a registered provider
func (e *Engine) GetProvider(name string) (Provider, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	provider, ok := e.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// StoreChunk stores a chunk to storage
func (e *Engine) StoreChunk(hash string, data []byte) (string, error) {
	ctx := context.Background()
	reader := bytes.NewReader(data)

	provider, err := e.GetProvider("local")
	if err != nil {
		return "", fmt.Errorf("no storage provider available: %w", err)
	}

	err = provider.Store(ctx, hash, reader, int64(len(data)))
	if err != nil {
		return "", err
	}

	return hash, nil
}

// GetChunk retrieves a chunk from storage
func (e *Engine) GetChunk(hash string) ([]byte, error) {
	ctx := context.Background()

	provider, err := e.GetProvider("local")
	if err != nil {
		return nil, fmt.Errorf("no storage provider available: %w", err)
	}

	reader, err := provider.Retrieve(ctx, hash)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk: %w", err)
	}

	return data, nil
}

// Store stores data using the specified provider
func (e *Engine) Store(ctx context.Context, provider, key string, data io.Reader, size int64) error {
	p, err := e.GetProvider(provider)
	if err != nil {
		return err
	}
	return p.Store(ctx, key, data, size)
}

// Retrieve retrieves data using the specified provider
func (e *Engine) Retrieve(ctx context.Context, provider, key string) (io.ReadCloser, error) {
	p, err := e.GetProvider(provider)
	if err != nil {
		return nil, err
	}
	return p.Retrieve(ctx, key)
}

// Delete deletes data using the specified provider
func (e *Engine) Delete(ctx context.Context, provider, key string) error {
	p, err := e.GetProvider(provider)
	if err != nil {
		return err
	}
	return p.Delete(ctx, key)
}

// GetStats gets statistics from the specified provider
func (e *Engine) GetStats(ctx context.Context, provider string) (*Stats, error) {
	p, err := e.GetProvider(provider)
	if err != nil {
		return nil, err
	}
	return p.GetStats(ctx)
}

// Close closes all providers
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	var lastErr error
	for _, provider := range e.providers {
		if err := provider.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
