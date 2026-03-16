// Simple Backup - minimal working version without deduplication
package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// SimpleBackup creates a simple ZIP backup without deduplication
func (e *BackupEngine) SimpleBackup(job *BackupJob, session *BackupSession) error {
	e.log(session, "📦 Starting simple ZIP backup...")

	// Create backup directory
	backupDir := filepath.Join(job.Destination, job.Name, time.Now().Format("2006-01-02_150405"))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}
	session.BackupPath = backupDir

	// Create ZIP file
	archivePath := filepath.Join(backupDir, "backup.zip")
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(archiveFile)

	var totalBytes int64
	filesCount := 0

	// Add each source file/folder
	for _, source := range job.Sources {
		count, bytes, err := e.addSourceToZip(zipWriter, source)
		if err != nil {
			zipWriter.Close()
			archiveFile.Close()
			return err
		}
		filesCount += count
		totalBytes += bytes
	}

	// Close ZIP writer to flush data
	if err := zipWriter.Close(); err != nil {
		archiveFile.Close()
		return err
	}
	archiveFile.Close()

	session.FilesProcessed = filesCount
	session.FilesTotal = filesCount
	session.BytesRead = totalBytes
	session.BytesWritten = totalBytes // Simplified - actual compressed size

	// Create metadata
	metadata := map[string]interface{}{
		"job_id":      job.ID,
		"job_name":    job.Name,
		"backup_time": session.StartTime,
		"files_count": filesCount,
		"total_size":  totalBytes,
	}
	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	os.WriteFile(filepath.Join(backupDir, "metadata.json"), metadataJSON, 0644)

	e.log(session, fmt.Sprintf("✅ Backup complete: %d files, %s", filesCount, e.formatBytes(totalBytes)))
	return nil
}

func (e *BackupEngine) addSourceToZip(zw *zip.Writer, source string) (int, int64, error) {
	filesCount := 0
	var totalBytes int64

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Create header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Use relative path
		header.Name, _ = filepath.Rel(source, path)
		header.Method = zip.Deflate // Use compression

		// Create file in ZIP
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		// Open and copy file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		filesCount++
		totalBytes += info.Size()

		return nil
	})

	return filesCount, totalBytes, err
}
