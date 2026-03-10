package providers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"novabackup/pkg/models"
)

// FileBackupProvider handles file-based backup operations
type FileBackupProvider struct {
	chunkSize int64
}

// NewFileBackupProvider creates a new file backup provider
func NewFileBackupProvider(chunkSize int64) *FileBackupProvider {
	if chunkSize <= 0 {
		chunkSize = 4 * 1024 * 1024 // Default 4MB
	}
	return &FileBackupProvider{chunkSize: chunkSize}
}

// Backup performs backup of a directory or file
func (f *FileBackupProvider) Backup(ctx context.Context, source string, dest string) ([]models.FileInfo, error) {
	var files []models.FileInfo

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, only process files
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Process file
		fileInfo, err := f.processFile(ctx, path, filepath.Join(dest, relPath))
		if err != nil {
			return fmt.Errorf("failed to process file %s: %w", path, err)
		}

		files = append(files, *fileInfo)
		return nil
	})

	return files, err
}

// processFile reads a file, splits it into chunks, and computes hash
func (f *FileBackupProvider) processFile(ctx context.Context, sourcePath, destPath string) (*models.FileInfo, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Create destination directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, err
	}

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, err
	}
	defer destFile.Close()

	// Calculate hash of entire file
	hash := sha256.New()
	teeReader := io.TeeReader(file, hash)

	// Copy data with chunking logic (for future dedupe)
	buf := make([]byte, f.chunkSize)
	bytesWritten := int64(0)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, err := teeReader.Read(buf)
		if n > 0 {
			if _, writeErr := destFile.Write(buf[:n]); writeErr != nil {
				return nil, writeErr
			}
			bytesWritten += int64(n)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	fileChecksum := fmt.Sprintf("%x", hash.Sum(nil))

	return &models.FileInfo{
		Path:     sourcePath,
		Name:     filepath.Base(sourcePath),
		Size:     info.Size(),
		ModTime:  info.ModTime(),
		IsDir:    false,
		Checksum: fileChecksum,
	}, nil
}

// ChunkFile splits a file into chunks and returns their hashes
func (f *FileBackupProvider) ChunkFile(ctx context.Context, sourcePath string) ([]string, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var chunkHashes []string
	buf := make([]byte, f.chunkSize)
	chunkNum := 0

	for {
		n, err := file.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			hash := sha256.Sum256(chunk)
			chunkHashes = append(chunkHashes, fmt.Sprintf("%x", hash))
			chunkNum++
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return chunkHashes, nil
}

// ListFiles returns all files in a directory
func (f *FileBackupProvider) ListFiles(ctx context.Context, source string) ([]models.FileInfo, error) {
	var files []models.FileInfo

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		files = append(files, models.FileInfo{
			Path:    relPath,
			Name:    info.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   false,
		})

		return nil
	})

	return files, err
}
