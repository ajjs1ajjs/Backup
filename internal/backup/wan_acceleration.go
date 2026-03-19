// WAN Acceleration - Optimized data transfer for remote backups
package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// WANAccelerator optimizes data transfer over WAN
type WANAccelerator struct {
	BlockCache map[string]bool // Hash of blocks already at target
	CacheSize  int
	Hits       int
	Misses     int
	BytesSaved int64
	BytesTotal int64
	mu         sync.RWMutex
}

// NewWANAccelerator creates a new WAN accelerator
func NewWANAccelerator(cacheSize int) *WANAccelerator {
	return &WANAccelerator{
		BlockCache: make(map[string]bool),
		CacheSize:  cacheSize,
	}
}

// ShouldTransfer checks if block needs to be transferred
func (w *WANAccelerator) ShouldTransfer(data []byte) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	hash := w.calculateHash(data)
	exists := w.BlockCache[hash]

	if exists {
		w.Hits++
		w.BytesSaved += int64(len(data))
	} else {
		w.Misses++
		w.BytesTotal += int64(len(data))
	}

	return !exists
}

// AddBlock marks block as existing at target
func (w *WANAccelerator) AddBlock(data []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()

	hash := w.calculateHash(data)
	w.BlockCache[hash] = true

	// Limit cache size
	if len(w.BlockCache) > w.CacheSize {
		// Remove oldest entries (simplified - use LRU in production)
		for key := range w.BlockCache {
			delete(w.BlockCache, key)
			if len(w.BlockCache) <= w.CacheSize/2 {
				break
			}
		}
	}
}

// calculateHash calculates SHA256 hash of data
func (w *WANAccelerator) calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// GetStatistics returns accelerator statistics
func (w *WANAccelerator) GetStatistics() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	hitRatio := 0.0
	if w.Hits+w.Misses > 0 {
		hitRatio = float64(w.Hits) / float64(w.Hits+w.Misses) * 100
	}

	bandwidthSavings := 0.0
	if w.BytesTotal+w.BytesSaved > 0 {
		bandwidthSavings = float64(w.BytesSaved) / float64(w.BytesTotal+w.BytesSaved) * 100
	}

	return map[string]interface{}{
		"cache_size":        len(w.BlockCache),
		"max_cache_size":    w.CacheSize,
		"hits":              w.Hits,
		"misses":            w.Misses,
		"hit_ratio":         hitRatio,
		"bytes_transferred": w.BytesTotal,
		"bytes_saved":       w.BytesSaved,
		"bandwidth_savings": bandwidthSavings,
	}
}

// Reset resets statistics
func (w *WANAccelerator) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.Hits = 0
	w.Misses = 0
	w.BytesSaved = 0
	w.BytesTotal = 0
}

// SyncCache synchronizes cache with remote target
func (w *WANAccelerator) SyncCache(remoteHashes []string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Build remote cache
	remoteCache := make(map[string]bool)
	for _, hash := range remoteHashes {
		remoteCache[hash] = true
	}

	// Merge with local cache
	for hash := range w.BlockCache {
		if !remoteCache[hash] {
			delete(w.BlockCache, hash)
		}
	}
}

// OptimalBlocksize returns optimal block size for WAN acceleration
func OptimalBlocksize(averageFileSize int64) int {
	// Larger blocks = fewer hashes, but more data transferred per miss
	// Smaller blocks = more hashes, but less data per miss

	if averageFileSize < 1024*1024 { // < 1MB
		return 64 * 1024 // 64KB
	}
	if averageFileSize < 10*1024*1024 { // < 10MB
		return 256 * 1024 // 256KB
	}
	return 1024 * 1024 // 1MB
}

// EstimateTransferTime estimates time to transfer data over WAN
func EstimateTransferTime(dataSize int64, bandwidthMbps float64, latencyMs int) string {
	// Account for TCP window size and latency
	// Simplified calculation

	if bandwidthMbps <= 0 {
		return "Unknown"
	}

	// Convert to bits
	dataBits := float64(dataSize) * 8

	// Bandwidth in bits per second
	bandwidthBps := bandwidthMbps * 1_000_000

	// Time in seconds
	timeSeconds := dataBits / bandwidthBps

	// Add latency overhead (simplified)
	latencyOverhead := float64(latencyMs) / 1000.0 * 2 // Round trip

	totalTime := timeSeconds + latencyOverhead

	if totalTime < 60 {
		return fmt.Sprintf("%.1fs", totalTime)
	}
	if totalTime < 3600 {
		return fmt.Sprintf("%.1fm", totalTime/60)
	}
	return fmt.Sprintf("%.1fh", totalTime/3600)
}

// CompressForWAN compresses data for WAN transfer
func CompressForWAN(data []byte) ([]byte, error) {
	// Use fast compression algorithm (lz4, snappy)
	// Optimized for speed over compression ratio

	// Placeholder - use actual compression library
	return data, nil
}

// DecompressFromWAN decompresses received data
func DecompressFromWAN(data []byte) ([]byte, error) {
	// Decompress lz4/snappy data
	return data, nil
}
