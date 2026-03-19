// Changed Block Tracking - Veeam-style CBT for incremental backups
package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// BlockInfo represents a file block
type BlockInfo struct {
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
	Offset    int64  `json:"offset"`
	ModTime   int64  `json:"mod_time"`
	Reference string `json:"reference"` // Reference to block location
}

// FileChangeTracker tracks changes for a single file
type FileChangeTracker struct {
	Path     string                 `json:"path"`
	Size     int64                  `json:"size"`
	ModTime  int64                  `json:"mod_time"`
	Blocks   []BlockInfo            `json:"blocks"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChangeTracker tracks all file changes
type ChangeTracker struct {
	mu           sync.RWMutex
	DataDir      string
	Files        map[string]*FileChangeTracker // Map by file path
	LastScan     time.Time                     `json:"last_scan"`
	BlockSize    int64                         `json:"block_size"` // Default 1MB
	TotalBlocks  int                           `json:"total_blocks"`
	UniqueBlocks int                           `json:"unique_blocks"`
}

// NewChangeTracker creates a new change tracker
func NewChangeTracker(dataDir string) *ChangeTracker {
	tracker := &ChangeTracker{
		DataDir:   dataDir,
		Files:     make(map[string]*FileChangeTracker),
		BlockSize: 1024 * 1024, // 1MB blocks
	}

	// Load existing tracker if exists
	tracker.load()

	return tracker
}

// TrackFile tracks changes for a file
func (ct *ChangeTracker) TrackFile(filePath string) (*FileChangeTracker, error) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	fileTracker, exists := ct.Files[filePath]

	if !exists {
		// New file - track all blocks
		fileTracker = &FileChangeTracker{
			Path:     filePath,
			Size:     info.Size(),
			ModTime:  info.ModTime().Unix(),
			Blocks:   make([]BlockInfo, 0),
			Metadata: make(map[string]interface{}),
		}
		ct.Files[filePath] = fileTracker
	} else {
		// Existing file - check if modified
		if info.ModTime().Unix() == fileTracker.ModTime && info.Size() == fileTracker.Size {
			// File unchanged
			return fileTracker, nil
		}
	}

	// Calculate blocks for file
	blocks, err := ct.calculateBlocks(filePath, info.Size())
	if err != nil {
		return nil, err
	}

	fileTracker.Blocks = blocks
	fileTracker.Size = info.Size()
	fileTracker.ModTime = info.ModTime().Unix()

	ct.TotalBlocks += len(blocks)

	return fileTracker, nil
}

// GetChangedFiles returns files that changed since last scan
func (ct *ChangeTracker) GetChangedFiles(since time.Time) []string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	var changed []string
	for path, tracker := range ct.Files {
		modTime := time.Unix(tracker.ModTime, 0)
		if modTime.After(since) {
			changed = append(changed, path)
		}
	}

	return changed
}

// GetChangedBlocks returns blocks that changed for a file
func (ct *ChangeTracker) GetChangedBlocks(filePath string, oldTracker *FileChangeTracker) []BlockInfo {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	current, exists := ct.Files[filePath]
	if !exists {
		return nil
	}

	// Build hash map for quick lookup
	oldHashes := make(map[string]bool)
	for _, block := range oldTracker.Blocks {
		oldHashes[block.Hash] = true
	}

	// Find new/changed blocks
	var changed []BlockInfo
	for _, block := range current.Blocks {
		if !oldHashes[block.Hash] {
			changed = append(changed, block)
		}
	}

	return changed
}

// GetUnchangedBlocks returns blocks that haven't changed (for deduplication)
func (ct *ChangeTracker) GetUnchangedBlocks(filePath string, oldTracker *FileChangeTracker) []BlockInfo {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	current, exists := ct.Files[filePath]
	if !exists {
		return nil
	}

	// Build hash map for old blocks
	oldHashes := make(map[string]bool)
	for _, block := range oldTracker.Blocks {
		oldHashes[block.Hash] = true
	}

	// Find unchanged blocks
	var unchanged []BlockInfo
	for _, block := range current.Blocks {
		if oldHashes[block.Hash] {
			unchanged = append(unchanged, block)
		}
	}

	return unchanged
}

// calculateBlocks calculates block hashes for a file
func (ct *ChangeTracker) calculateBlocks(filePath string, size int64) ([]BlockInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var blocks []BlockInfo
	offset := int64(0)
	blockNum := 0

	for offset < size {
		// Calculate block size (last block may be smaller)
		remaining := size - offset
		blockSize := ct.BlockSize
		if remaining < blockSize {
			blockSize = remaining
		}

		// Read block
		buffer := make([]byte, blockSize)
		n, err := file.ReadAt(buffer, offset)
		if err != nil && err != io.ErrUnexpectedEOF {
			return nil, err
		}

		// Calculate hash
		hash := sha256.Sum256(buffer[:n])
		hashHex := hex.EncodeToString(hash[:])

		block := BlockInfo{
			Hash:      hashHex,
			Size:      int64(n),
			Offset:    offset,
			ModTime:   time.Now().Unix(),
			Reference: fmt.Sprintf("block_%d_%d", blockNum, offset),
		}

		blocks = append(blocks, block)
		offset += int64(n)
		blockNum++
	}

	return blocks, nil
}

// GetStatistics returns CBT statistics
func (ct *ChangeTracker) GetStatistics() map[string]interface{} {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	totalFiles := len(ct.Files)
	totalSize := int64(0)
	for _, ft := range ct.Files {
		totalSize += ft.Size
	}

	return map[string]interface{}{
		"total_files":   totalFiles,
		"total_blocks":  ct.TotalBlocks,
		"unique_blocks": ct.UniqueBlocks,
		"total_size":    totalSize,
		"block_size":    ct.BlockSize,
		"last_scan":     ct.LastScan,
		"dedup_ratio":   float64(ct.TotalBlocks) / float64(ct.UniqueBlocks),
		"space_saved":   int64(ct.TotalBlocks-ct.UniqueBlocks) * ct.BlockSize,
	}
}

// Save saves tracker state to disk
func (ct *ChangeTracker) save() error {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	trackerDir := filepath.Join(ct.DataDir, "cbt")
	os.MkdirAll(trackerDir, 0755)

	trackerFile := filepath.Join(trackerDir, "tracker.json")
	data, err := json.MarshalIndent(ct, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(trackerFile, data, 0644)
}

// Load loads tracker state from disk
func (ct *ChangeTracker) load() error {
	trackerFile := filepath.Join(ct.DataDir, "cbt", "tracker.json")

	data, err := os.ReadFile(trackerFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing tracker
		}
		return err
	}

	return json.Unmarshal(data, ct)
}

// Reset resets the change tracker
func (ct *ChangeTracker) Reset() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.Files = make(map[string]*FileChangeTracker)
	ct.TotalBlocks = 0
	ct.UniqueBlocks = 0
	ct.LastScan = time.Time{}
}

// RemoveFile removes a file from tracking
func (ct *ChangeTracker) RemoveFile(filePath string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if tracker, exists := ct.Files[filePath]; exists {
		ct.TotalBlocks -= len(tracker.Blocks)
		delete(ct.Files, filePath)
	}
}

// GetBlockReference returns reference to a block by hash
func (ct *ChangeTracker) GetBlockReference(hash string) string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	for _, ft := range ct.Files {
		for _, block := range ft.Blocks {
			if block.Hash == hash {
				return block.Reference
			}
		}
	}

	return ""
}

// CountUniqueBlocks counts unique blocks across all files
func (ct *ChangeTracker) CountUniqueBlocks() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	hashes := make(map[string]bool)
	for _, ft := range ct.Files {
		for _, block := range ft.Blocks {
			hashes[block.Hash] = true
		}
	}

	return len(hashes)
}
