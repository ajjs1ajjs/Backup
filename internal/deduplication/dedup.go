package deduplication

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"novabackup/internal/multitenancy"
)

// DeduplicationManager manages global deduplication operations
type DeduplicationManager interface {
	// Chunk operations
	StoreChunk(ctx context.Context, chunk *Chunk) error
	GetChunk(ctx context.Context, hash string) (*Chunk, error)
	DeleteChunk(ctx context.Context, hash string) error
	ChunkExists(ctx context.Context, hash string) (bool, error)

	// Hash operations
	CalculateHash(ctx context.Context, data []byte) (string, error)
	FindDuplicateChunks(ctx context.Context, chunkHashes []string) (map[string]bool, error)

	// Statistics
	GetDeduplicationStats(ctx context.Context, tenantID string) (*DeduplicationStats, error)
	GetGlobalStats(ctx context.Context) (*GlobalDeduplicationStats, error)

	// Cleanup
	CleanupUnusedChunks(ctx context.Context, olderThan time.Duration) error
	OptimizeStorage(ctx context.Context) error
}

// Chunk represents a deduplicated data chunk
type Chunk struct {
	Hash       string            `json:"hash"`
	Size       int64             `json:"size"`
	Data       []byte            `json:"data,omitempty"`
	RefCount   int               `json:"ref_count"`
	TenantID   string            `json:"tenant_id"`
	CreatedAt  time.Time         `json:"created_at"`
	LastAccess time.Time         `json:"last_access"`
	Metadata   map[string]string `json:"metadata"`
}

// DeduplicationConfig contains deduplication settings
type DeduplicationConfig struct {
	Enabled          bool          `json:"enabled"`
	ChunkSize        int           `json:"chunk_size"`
	Algorithm        string        `json:"algorithm"`
	MaxChunkSize     int64         `json:"max_chunk_size"`
	MinChunkSize     int64         `json:"min_chunk_size"`
	CleanupInterval  time.Duration `json:"cleanup_interval"`
	CompressionLevel int           `json:"compression_level"`
	CacheSize        int           `json:"cache_size"`
}

// DeduplicationStats contains tenant-specific deduplication statistics
type DeduplicationStats struct {
	TenantID           string    `json:"tenant_id"`
	TotalChunks        int64     `json:"total_chunks"`
	TotalSize          int64     `json:"total_size"`
	DeduplicatedSize   int64     `json:"deduplicated_size"`
	DeduplicationRatio float64   `json:"deduplication_ratio"`
	UniqueChunks       int64     `json:"unique_chunks"`
	SharedChunks       int64     `json:"shared_chunks"`
	LastUpdated        time.Time `json:"last_updated"`
}

// GlobalDeduplicationStats contains global deduplication statistics
type GlobalDeduplicationStats struct {
	TotalTenants       int64     `json:"total_tenants"`
	TotalChunks        int64     `json:"total_chunks"`
	TotalSize          int64     `json:"total_size"`
	DeduplicatedSize   int64     `json:"deduplicated_size"`
	DeduplicationRatio float64   `json:"deduplication_ratio"`
	UniqueChunks       int64     `json:"unique_chunks"`
	SharedChunks       int64     `json:"shared_chunks"`
	StorageSaved       int64     `json:"storage_saved"`
	LastUpdated        time.Time `json:"last_updated"`
}

// HashAlgorithm represents different hashing algorithms
type HashAlgorithm string

const (
	HashAlgorithmSHA256 HashAlgorithm = "sha256"
	HashAlgorithmMD5    HashAlgorithm = "md5"
	HashAlgorithmSHA1   HashAlgorithm = "sha1"
)

// InMemoryDeduplicationManager implements DeduplicationManager in memory
type InMemoryDeduplicationManager struct {
	chunks      map[string]*Chunk
	mutex       sync.RWMutex
	config      DeduplicationConfig
	stats       map[string]*DeduplicationStats
	globalStats *GlobalDeduplicationStats
	tenantMgr   multitenancy.TenantManager
}

// NewInMemoryDeduplicationManager creates a new in-memory deduplication manager
func NewInMemoryDeduplicationManager(config DeduplicationConfig, tenantMgr multitenancy.TenantManager) *InMemoryDeduplicationManager {
	return &InMemoryDeduplicationManager{
		chunks:    make(map[string]*Chunk),
		config:    config,
		stats:     make(map[string]*DeduplicationStats),
		tenantMgr: tenantMgr,
		globalStats: &GlobalDeduplicationStats{
			LastUpdated: time.Now(),
		},
	}
}

// CalculateHash calculates hash for given data
func (dm *InMemoryDeduplicationManager) CalculateHash(ctx context.Context, data []byte) (string, error) {
	switch HashAlgorithm(dm.config.Algorithm) {
	case HashAlgorithmSHA256:
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:]), nil
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", dm.config.Algorithm)
	}
}

// StoreChunk stores a chunk in the deduplication system
func (dm *InMemoryDeduplicationManager) StoreChunk(ctx context.Context, chunk *Chunk) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Get tenant ID from context
	tenantID := dm.tenantMgr.GetTenantFromContext(ctx)
	if tenantID == "" {
		return fmt.Errorf("tenant ID required for chunk storage")
	}

	// Check if chunk already exists
	if existing, exists := dm.chunks[chunk.Hash]; exists {
		// Increment reference count
		existing.RefCount++
		existing.LastAccess = time.Now()
		return nil
	}

	// Validate chunk size
	if int64(len(chunk.Data)) > dm.config.MaxChunkSize {
		return fmt.Errorf("chunk size %d exceeds maximum %d", len(chunk.Data), dm.config.MaxChunkSize)
	}

	if int64(len(chunk.Data)) < dm.config.MinChunkSize {
		return fmt.Errorf("chunk size %d below minimum %d", len(chunk.Data), dm.config.MinChunkSize)
	}

	// Store new chunk
	chunk.TenantID = tenantID
	chunk.RefCount = 1
	chunk.CreatedAt = time.Now()
	chunk.LastAccess = time.Now()

	dm.chunks[chunk.Hash] = chunk

	// Update statistics
	dm.updateStats(tenantID, chunk.Size, true)

	return nil
}

// GetChunk retrieves a chunk by hash
func (dm *InMemoryDeduplicationManager) GetChunk(ctx context.Context, hash string) (*Chunk, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	chunk, exists := dm.chunks[hash]
	if !exists {
		return nil, fmt.Errorf("chunk not found: %s", hash)
	}

	// Update last access time
	chunk.LastAccess = time.Now()

	return chunk, nil
}

// DeleteChunk removes a chunk from the deduplication system
func (dm *InMemoryDeduplicationManager) DeleteChunk(ctx context.Context, hash string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	chunk, exists := dm.chunks[hash]
	if !exists {
		return fmt.Errorf("chunk not found: %s", hash)
	}

	// Decrement reference count
	chunk.RefCount--
	if chunk.RefCount <= 0 {
		// Remove chunk and update statistics
		delete(dm.chunks, hash)
		dm.updateStats(chunk.TenantID, -chunk.Size, false)
	}

	return nil
}

// ChunkExists checks if a chunk exists
func (dm *InMemoryDeduplicationManager) ChunkExists(ctx context.Context, hash string) (bool, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	_, exists := dm.chunks[hash]
	return exists, nil
}

// FindDuplicateChunks finds which chunks already exist
func (dm *InMemoryDeduplicationManager) FindDuplicateChunks(ctx context.Context, chunkHashes []string) (map[string]bool, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	result := make(map[string]bool)
	for _, hash := range chunkHashes {
		result[hash] = dm.chunks[hash] != nil
	}

	return result, nil
}

// GetDeduplicationStats returns tenant-specific deduplication statistics
func (dm *InMemoryDeduplicationManager) GetDeduplicationStats(ctx context.Context, tenantID string) (*DeduplicationStats, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	stats, exists := dm.stats[tenantID]
	if !exists {
		return &DeduplicationStats{
			TenantID:    tenantID,
			LastUpdated: time.Now(),
		}, nil
	}

	return stats, nil
}

// GetGlobalStats returns global deduplication statistics
func (dm *InMemoryDeduplicationManager) GetGlobalStats(ctx context.Context) (*GlobalDeduplicationStats, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	// Recalculate global stats
	totalChunks := int64(len(dm.chunks))
	totalSize := int64(0)
	uniqueChunks := int64(0)
	sharedChunks := int64(0)

	tenantSet := make(map[string]bool)

	for _, chunk := range dm.chunks {
		totalSize += int64(len(chunk.Data))
		tenantSet[chunk.TenantID] = true

		if chunk.RefCount == 1 {
			uniqueChunks++
		} else {
			sharedChunks++
		}
	}

	deduplicatedSize := totalSize
	for _, chunk := range dm.chunks {
		if chunk.RefCount > 1 {
			deduplicatedSize -= int64(len(chunk.Data)) * int64(chunk.RefCount-1)
		}
	}

	deduplicationRatio := float64(0)
	if totalSize > 0 {
		deduplicationRatio = float64(deduplicatedSize) / float64(totalSize)
	}

	storageSaved := totalSize - deduplicatedSize

	dm.globalStats = &GlobalDeduplicationStats{
		TotalTenants:       int64(len(tenantSet)),
		TotalChunks:        totalChunks,
		TotalSize:          totalSize,
		DeduplicatedSize:   deduplicatedSize,
		DeduplicationRatio: deduplicationRatio,
		UniqueChunks:       uniqueChunks,
		SharedChunks:       sharedChunks,
		StorageSaved:       storageSaved,
		LastUpdated:        time.Now(),
	}

	return dm.globalStats, nil
}

// CleanupUnusedChunks removes chunks that haven't been accessed recently
func (dm *InMemoryDeduplicationManager) CleanupUnusedChunks(ctx context.Context, olderThan time.Duration) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	cutoff := time.Now().Add(-olderThan)
	removedCount := 0

	for hash, chunk := range dm.chunks {
		if chunk.LastAccess.Before(cutoff) && chunk.RefCount == 0 {
			delete(dm.chunks, hash)
			removedCount++
		}
	}

	return nil
}

// OptimizeStorage optimizes the deduplication storage
func (dm *InMemoryDeduplicationManager) OptimizeStorage(ctx context.Context) error {
	// In memory implementation, optimization is minimal
	// Could implement chunk consolidation or other optimizations
	return nil
}

// updateStats updates deduplication statistics
func (dm *InMemoryDeduplicationManager) updateStats(tenantID string, sizeDelta int64, isStore bool) {
	stats, exists := dm.stats[tenantID]
	if !exists {
		stats = &DeduplicationStats{
			TenantID:    tenantID,
			LastUpdated: time.Now(),
		}
		dm.stats[tenantID] = stats
	}

	if isStore {
		stats.TotalChunks++
		stats.TotalSize += sizeDelta
	} else {
		stats.TotalChunks--
		stats.TotalSize += sizeDelta // sizeDelta is negative for deletion
	}

	// Recalculate deduplication ratio
	if stats.TotalSize > 0 {
		// Count unique chunks for this tenant
		uniqueCount := int64(0)
		for _, chunk := range dm.chunks {
			if chunk.TenantID == tenantID && chunk.RefCount == 1 {
				uniqueCount++
			}
		}

		sharedCount := int64(0)
		for _, chunk := range dm.chunks {
			if chunk.TenantID == tenantID && chunk.RefCount > 1 {
				sharedCount++
			}
		}

		stats.UniqueChunks = uniqueCount
		stats.SharedChunks = sharedCount

		// Simple deduplication ratio calculation
		stats.DeduplicationRatio = float64(sharedCount) / float64(stats.TotalChunks)
	}

	stats.LastUpdated = time.Now()
}

// Chunker splits data into chunks for deduplication
type Chunker struct {
	config DeduplicationConfig
}

// NewChunker creates a new chunker
func NewChunker(config DeduplicationConfig) *Chunker {
	return &Chunker{config: config}
}

// ChunkData splits data into chunks
func (c *Chunker) ChunkData(data []byte) [][]byte {
	if len(data) == 0 {
		return nil
	}

	chunkSize := c.config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 64 * 1024 // Default 64KB
	}

	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}

	return chunks
}

// ReassembleData reassembles chunks back into original data
func (c *Chunker) ReassembleData(chunks [][]byte) []byte {
	var data []byte
	for _, chunk := range chunks {
		data = append(data, chunk...)
	}
	return data
}

// Utility functions
func NewDeduplicationConfig() DeduplicationConfig {
	return DeduplicationConfig{
		Enabled:          true,
		ChunkSize:        64 * 1024, // 64KB
		Algorithm:        string(HashAlgorithmSHA256),
		MaxChunkSize:     1024 * 1024, // 1MB
		MinChunkSize:     1024,        // 1KB
		CleanupInterval:  24 * time.Hour,
		CompressionLevel: 6,
		CacheSize:        1000,
	}
}

// CalculateChunkHash calculates hash for a chunk
func CalculateChunkHash(data []byte, algorithm HashAlgorithm) (string, error) {
	switch algorithm {
	case HashAlgorithmSHA256:
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:]), nil
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}
}

// ValidateChunk validates chunk data
func ValidateChunk(chunk *Chunk, config DeduplicationConfig) error {
	if chunk.Hash == "" {
		return fmt.Errorf("chunk hash cannot be empty")
	}

	if len(chunk.Data) == 0 {
		return fmt.Errorf("chunk data cannot be empty")
	}

	chunkSize := int64(len(chunk.Data))
	if chunkSize > config.MaxChunkSize {
		return fmt.Errorf("chunk size %d exceeds maximum %d", chunkSize, config.MaxChunkSize)
	}

	if chunkSize < config.MinChunkSize {
		return fmt.Errorf("chunk size %d below minimum %d", chunkSize, config.MinChunkSize)
	}

	// Verify hash
	calculatedHash, err := CalculateChunkHash(chunk.Data, HashAlgorithm(config.Algorithm))
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	if calculatedHash != chunk.Hash {
		return fmt.Errorf("chunk hash mismatch: expected %s, got %s", chunk.Hash, calculatedHash)
	}

	return nil
}
