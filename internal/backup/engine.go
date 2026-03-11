package backup

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"novabackup/internal/database"
	"novabackup/internal/storage"
	"novabackup/pkg/models"

	"github.com/klauspost/compress/zstd"
)

// BackupEngine handles backup operations
type BackupEngine struct {
	db        *database.Connection
	config    *models.BackupConfig
	storage   *storage.Engine
	dedupeMap map[string]bool
	mu        sync.RWMutex
}

// NewBackupEngine creates a new backup engine
func NewBackupEngine(db *database.Connection, config *models.BackupConfig) *BackupEngine {
	return &BackupEngine{
		db:        db,
		config:    config,
		storage:   storage.NewEngine(),
		dedupeMap: make(map[string]bool),
	}
}

// Chunk represents a data chunk with metadata
type Chunk struct {
	Hash     string
	Data     []byte
	Original []byte
}

// PerformBackup executes a complete backup operation
func (e *BackupEngine) PerformBackup(ctx context.Context, source string) (*models.BackupResult, error) {
	result := &models.BackupResult{
		Status:    models.JobStatusRunning,
		StartTime: models.TimeNow(),
	}

	// Walk through source directory
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Process file
		fileResult, err := e.processFile(ctx, path)
		if err != nil {
			result.FilesFailed++
			return nil // Continue with other files
		}

		result.FilesTotal++
		result.FilesSuccess++
		result.BytesRead += fileResult.BytesRead
		result.BytesWritten += fileResult.BytesWritten

		return nil
	})

	result.EndTime = models.TimeNow()
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
	} else {
		result.Status = models.JobStatusCompleted
	}

	return result, nil
}

// processFile reads a file, chunks it, and stores unique chunks
func (e *BackupEngine) processFile(ctx context.Context, path string) (*FileResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	result := &FileResult{
		Path: path,
		Size: info.Size(),
	}

	// Read file in chunks
	buf := make([]byte, e.config.ChunkSize)
	totalWritten := int64(0)

	for {
		n, err := file.Read(buf)
		if n > 0 {
			chunkData := buf[:n]

			// Calculate hash
			hash := sha256.Sum256(chunkData)
			hashStr := fmt.Sprintf("%x", hash)

			// Check deduplication
			isNew := e.checkDedupe(hashStr)

			if isNew {
				// Compress
				compressed, err := e.compress(chunkData)
				if err != nil {
					return nil, err
				}

				// Encrypt if enabled
				finalData := compressed
				if e.config.Encrypt {
					finalData, err = e.encrypt(compressed)
					if err != nil {
						return nil, err
					}
				}

				// Store chunk
				storagePath, repoID, tier, err := e.storage.StoreChunk(hashStr, finalData)
				if err != nil {
					return nil, err
				}

				// Update database
				e.db.AddChunk(hashStr, int64(len(chunkData)), len(finalData), storagePath, repoID, tier)
				totalWritten += int64(len(finalData))
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	result.BytesRead = info.Size()
	result.BytesWritten = totalWritten

	return result, nil
}

// checkDedupe checks if a chunk hash already exists
func (e *BackupEngine) checkDedupe(hash string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, exists := e.dedupeMap[hash]
	if exists {
		return false
	}

	// Also check database
	path, _ := e.db.GetChunkByHash(hash)
	if path != "" {
		return false
	}

	return true
}

// compress compresses data using zstd
func (e *BackupEngine) compress(data []byte) ([]byte, error) {
	if !e.config.Compress {
		return data, nil
	}

	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	return encoder.EncodeAll(data, nil), nil
}

// encrypt encrypts data using AES-256-GCM
func (e *BackupEngine) encrypt(data []byte) ([]byte, error) {
	// Generate a random nonce
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Create cipher
	block, err := aes.NewCipher(e.config.EncryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Prepend nonce to ciphertext
	return append(nonce, ciphertext...), nil
}

// FileResult contains file processing results
type FileResult struct {
	Path         string
	Size         int64
	BytesRead    int64
	BytesWritten int64
}
