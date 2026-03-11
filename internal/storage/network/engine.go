package network

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"novabackup/pkg/models"
)

// NetworkStorageType defines the type of network storage
type NetworkStorageType string

const (
	StorageNFS NetworkStorageType = "nfs"
	StorageSMB NetworkStorageType = "smb"
)

// NetworkStorageConfig contains network storage configuration
type NetworkStorageConfig struct {
	Type        NetworkStorageType `json:"type"`
	Host        string             `json:"host"`
	Share       string             `json:"share"`
	Path        string             `json:"path"`
	Username    string             `json:"username,omitempty"`
	Password    string             `json:"password,omitempty"`
	Domain      string             `json:"domain,omitempty"`
	MountPoint  string             `json:"mount_point"`
	AutoMount   bool               `json:"auto_mount"`
	AutoUnmount bool               `json:"auto_unmount"`
	Timeout     time.Duration      `json:"timeout"`
	RetryCount  int                `json:"retry_count"`
	BufferSize  int                `json:"buffer_size"`
}

// NetworkStorageEngine handles network storage operations
type NetworkStorageEngine struct {
	config     *NetworkStorageConfig
	mounted    bool
	engineType NetworkStorageType
}

// NetworkStorageInterface defines the interface for network storage operations
type NetworkStorageInterface interface {
	Mount(ctx context.Context) error
	Unmount(ctx context.Context) error
	IsMounted() bool
	TestConnection(ctx context.Context) error
	StoreChunk(ctx context.Context, hash string, data []byte) (string, error)
	GetChunk(ctx context.Context, hash string) ([]byte, error)
	ChunkExists(ctx context.Context, hash string) (bool, error)
	DeleteChunk(ctx context.Context, hash string) error
	ListChunks(ctx context.Context, prefix string) ([]interface{}, error)
	GetStorageInfo(ctx context.Context) (*models.StorageInfo, error)
	Cleanup(ctx context.Context, olderThan time.Time) error
	Close() error
}

// ValidateConfig validates the network storage configuration
func (e *NetworkStorageEngine) ValidateConfig() error {
	if e.config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if e.config.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if e.config.Share == "" {
		return fmt.Errorf("share cannot be empty")
	}

	if e.config.MountPoint == "" {
		return fmt.Errorf("mount point cannot be empty")
	}

	if e.engineType == StorageSMB && e.config.Username == "" {
		return fmt.Errorf("username required for SMB shares")
	}

	return nil
}

// NewNetworkStorageEngine creates a new network storage engine
func NewNetworkStorageEngine(cfg *NetworkStorageConfig) (*NetworkStorageEngine, error) {
	if cfg == nil {
		return nil, fmt.Errorf("network storage config cannot be nil")
	}

	// Set defaults
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.RetryCount == 0 {
		cfg.RetryCount = 3
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 64 * 1024 // 64KB
	}

	engine := &NetworkStorageEngine{
		config:     cfg,
		mounted:    false,
		engineType: cfg.Type,
	}

	// Auto-mount if configured
	if cfg.AutoMount {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		if err := engine.Mount(ctx); err != nil {
			return nil, fmt.Errorf("auto-mount failed: %w", err)
		}
	}

	return engine, nil
}

// Mount mounts the network share
func (e *NetworkStorageEngine) Mount(ctx context.Context) error {
	if e.mounted {
		return nil
	}

	switch e.engineType {
	case StorageNFS:
		return e.mountNFS(ctx)
	case StorageSMB:
		return e.mountSMB(ctx)
	default:
		return fmt.Errorf("unsupported storage type: %s", e.engineType)
	}
}

// Unmount unmounts the network share
func (e *NetworkStorageEngine) Unmount(ctx context.Context) error {
	if !e.mounted {
		return nil
	}

	switch e.engineType {
	case StorageNFS:
		return e.unmountNFS(ctx)
	case StorageSMB:
		return e.unmountSMB(ctx)
	default:
		return fmt.Errorf("unsupported storage type: %s", e.engineType)
	}
}

// IsMounted returns whether the network share is mounted
func (e *NetworkStorageEngine) IsMounted() bool {
	return e.mounted
}

// TestConnection tests the network storage connection
func (e *NetworkStorageEngine) TestConnection(ctx context.Context) error {
	// Ensure mounted
	if !e.mounted {
		if err := e.Mount(ctx); err != nil {
			return fmt.Errorf("failed to mount for connection test: %w", err)
		}
		defer e.Unmount(ctx)
	}

	// Test write access
	testFile := filepath.Join(e.config.MountPoint, ".nova_test")

	// Try to create test file
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("write access test failed: %w", err)
	}
	file.Close()

	// Clean up test file
	os.Remove(testFile)

	return nil
}

// StoreChunk stores a chunk in network storage
func (e *NetworkStorageEngine) StoreChunk(ctx context.Context, hash string, data []byte) (string, error) {
	if !e.mounted {
		if err := e.Mount(ctx); err != nil {
			return "", fmt.Errorf("failed to mount for chunk storage: %w", err)
		}
	}

	// Create chunks directory if not exists
	chunksDir := filepath.Join(e.config.MountPoint, "chunks")
	if err := os.MkdirAll(chunksDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create chunks directory: %w", err)
	}

	// Write chunk file
	chunkPath := filepath.Join(chunksDir, hash)
	file, err := os.Create(chunkPath)
	if err != nil {
		return "", fmt.Errorf("failed to create chunk file: %w", err)
	}
	defer file.Close()

	// Write data with buffer
	if _, err := io.CopyBuffer(file,
		io.LimitReader(bytes.NewReader(data), int64(len(data))),
		make([]byte, e.config.BufferSize)); err != nil {
		return "", fmt.Errorf("failed to write chunk data: %w", err)
	}

	// Sync to ensure data is written
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("failed to sync chunk file: %w", err)
	}

	return chunkPath, nil
}

// GetChunk retrieves a chunk from network storage
func (e *NetworkStorageEngine) GetChunk(ctx context.Context, hash string) ([]byte, error) {
	if !e.mounted {
		if err := e.Mount(ctx); err != nil {
			return nil, fmt.Errorf("failed to mount for chunk retrieval: %w", err)
		}
	}

	chunkPath := filepath.Join(e.config.MountPoint, "chunks", hash)

	file, err := os.Open(chunkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open chunk file: %w", err)
	}
	defer file.Close()

	// Read file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk data: %w", err)
	}

	return data, nil
}

// ChunkExists checks if a chunk exists in network storage
func (e *NetworkStorageEngine) ChunkExists(ctx context.Context, hash string) (bool, error) {
	if !e.mounted {
		if err := e.Mount(ctx); err != nil {
			return false, fmt.Errorf("failed to mount for chunk check: %w", err)
		}
	}

	chunkPath := filepath.Join(e.config.MountPoint, "chunks", hash)

	_, err := os.Stat(chunkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check chunk existence: %w", err)
	}

	return true, nil
}

// DeleteChunk removes a chunk from network storage
func (e *NetworkStorageEngine) DeleteChunk(ctx context.Context, hash string) error {
	if !e.mounted {
		if err := e.Mount(ctx); err != nil {
			return fmt.Errorf("failed to mount for chunk deletion: %w", err)
		}
	}

	chunkPath := filepath.Join(e.config.MountPoint, "chunks", hash)

	if err := os.Remove(chunkPath); err != nil {
		return fmt.Errorf("failed to delete chunk file: %w", err)
	}

	return nil
}

// ListChunks lists all chunks in network storage
func (e *NetworkStorageEngine) ListChunks(ctx context.Context, prefix string) ([]interface{}, error) {
	if !e.mounted {
		if err := e.Mount(ctx); err != nil {
			return nil, fmt.Errorf("failed to mount for chunk listing: %w", err)
		}
	}

	var objects []interface{}
	chunksDir := filepath.Join(e.config.MountPoint, "chunks")

	err := filepath.Walk(chunksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check prefix filter
		if prefix != "" && !strings.HasPrefix(filepath.Base(path), prefix) {
			return nil
		}

		// Create object info
		relPath, _ := filepath.Rel(chunksDir, path)
		objects = append(objects, NetworkObjectInfo{
			Key:          relPath,
			Size:         info.Size(),
			LastModified: info.ModTime(),
			Mode:         info.Mode(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list chunks: %w", err)
	}

	return objects, nil
}

// Close cleans up network storage resources
func (e *NetworkStorageEngine) Close() error {
	if e.config.AutoUnmount && e.mounted {
		ctx, cancel := context.WithTimeout(context.Background(), e.config.Timeout)
		defer cancel()
		return e.Unmount(ctx)
	}
	return nil
}

// GetStorageInfo returns network storage information
func (e *NetworkStorageEngine) GetStorageInfo(ctx context.Context) (*models.StorageInfo, error) {
	if !e.mounted {
		if err := e.Mount(ctx); err != nil {
			return nil, fmt.Errorf("failed to mount for storage info: %w", err)
		}
	}

	var objectCount int64
	var usedSize int64

	chunksDir := filepath.Join(e.config.MountPoint, "chunks")

	// Walk through chunks directory
	err := filepath.Walk(chunksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			usedSize += info.Size()
			objectCount++
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to calculate storage info: %w", err)
	}

	// Get filesystem stats - platform specific implementation
	var estimatedTotalSize, freeSize int64

	// For Windows, use GetDiskFreeSpaceEx
	// For Unix systems, use syscall.Statfs
	filepath.Walk(e.config.MountPoint, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			estimatedTotalSize += info.Size()
		}
		return nil
	})

	// Simple estimation for total size
	estimatedTotalSize = usedSize * 2
	freeSize = estimatedTotalSize - usedSize

	return &models.StorageInfo{
		Type:        string(e.engineType),
		TotalSize:   estimatedTotalSize,
		UsedSize:    usedSize,
		FreeSize:    freeSize,
		ObjectCount: objectCount,
		Endpoint:    fmt.Sprintf("%s:%s%s", e.config.Host, e.config.Share, e.config.Path),
		Bucket:      e.config.MountPoint,
		Region:      "local",
		LastUpdated: time.Now(),
	}, nil
}

// Cleanup removes old chunks based on retention policy
func (e *NetworkStorageEngine) Cleanup(ctx context.Context, olderThan time.Time) error {
	if !e.mounted {
		if err := e.Mount(ctx); err != nil {
			return fmt.Errorf("failed to mount for cleanup: %w", err)
		}
	}

	chunksDir := filepath.Join(e.config.MountPoint, "chunks")

	err := filepath.Walk(chunksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is older than retention period
		if info.ModTime().Before(olderThan) {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove old chunk %s: %w", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to cleanup old chunks: %w", err)
	}

	return nil
}

// NetworkObjectInfo contains network storage object metadata
type NetworkObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	Mode         os.FileMode
}
