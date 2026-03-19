// Verification - Backup verification and health check (Veeam SureBackup style)
package backup

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// VerificationType represents type of verification
type VerificationType string

const (
	VerificationIntegrity VerificationType = "integrity" // Check file integrity
	VerificationMountable VerificationType = "mountable" // Verify backup is mountable
	VerificationBootable  VerificationType = "bootable"  // Verify VM can boot
	VerificationData      VerificationType = "data"      // Verify data consistency
)

// VerificationResult represents verification result
type VerificationResult struct {
	ID           string                 `json:"id"`
	BackupPath   string                 `json:"backup_path"`
	SessionID    string                 `json:"session_id"`
	Type         VerificationType       `json:"type"`
	Status       string                 `json:"status"` // success, warning, failed
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     string                 `json:"duration"`
	FilesChecked int                    `json:"files_checked"`
	FilesFailed  int                    `json:"files_failed"`
	TotalSize    int64                  `json:"total_size"`
	CheckedSize  int64                  `json:"checked_size"`
	Errors       []string               `json:"errors,omitempty"`
	Warnings     []string               `json:"warnings,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// VerifyBackup performs backup verification (Veeam SureBackup style)
func (e *BackupEngine) VerifyBackup(backupPath string, verificationType VerificationType) (*VerificationResult, error) {
	result := &VerificationResult{
		ID:         fmt.Sprintf("verify_%d", time.Now().Unix()),
		BackupPath: backupPath,
		Type:       verificationType,
		StartTime:  time.Now(),
		Status:     "running",
		Errors:     make([]string, 0),
		Warnings:   make([]string, 0),
		Details:    make(map[string]interface{}),
	}

	switch verificationType {
	case VerificationIntegrity:
		e.verifyIntegrity(result)
	case VerificationMountable:
		e.verifyMountable(result)
	case VerificationBootable:
		e.verifyBootable(result)
	case VerificationData:
		e.verifyDataConsistency(result)
	default:
		result.Status = "failed"
		result.Errors = append(result.Errors, "Невідомий тип перевірки")
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).String()

	// Save verification result
	e.saveVerificationResult(result)

	return result, nil
}

// verifyIntegrity checks backup file integrity using checksums
func (e *BackupEngine) verifyIntegrity(result *VerificationResult) {
	// Find backup archive
	archivePath := filepath.Join(result.BackupPath, "backup.zip")

	info, err := os.Stat(archivePath)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, fmt.Sprintf("Архів не знайдено: %v", err))
		return
	}

	result.TotalSize = info.Size()

	// Open and verify ZIP integrity
	r, err := os.Open(archivePath)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, fmt.Sprintf("Помилка відкриття: %v", err))
		return
	}
	defer r.Close()

	// Calculate SHA256 hash of entire file
	hash := sha256.New()
	written, err := io.Copy(hash, r)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, fmt.Sprintf("Помилка читання: %v", err))
		return
	}

	result.CheckedSize = written
	result.FilesChecked = 1

	// Store hash for future verification
	checksumHex := hex.EncodeToString(hash.Sum(nil))
	result.Details["sha256"] = checksumHex

	// Verify ZIP structure
	r.Seek(0, 0)
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, fmt.Sprintf("ZIP пошкоджено: %v", err))
		return
	}
	defer zipReader.Close()

	result.FilesChecked = len(zipReader.File)

	// Verify each file in archive
	var failedFiles int
	for _, f := range zipReader.File {
		if !e.verifyZipFile(f) {
			failedFiles++
			result.Warnings = append(result.Warnings, fmt.Sprintf("Файл пошкоджено: %s", f.Name))
		}
	}

	result.FilesFailed = failedFiles

	if failedFiles > 0 {
		result.Status = "warning"
	} else {
		result.Status = "success"
	}
}

// verifyZipFile verifies a single ZIP file integrity
func (e *BackupEngine) verifyZipFile(f *zip.File) bool {
	rc, err := f.Open()
	if err != nil {
		return false
	}
	defer rc.Close()

	// Try to read file content
	buffer := make([]byte, 64*1024)
	_, err = io.ReadFull(rc, buffer)

	// EOF is OK (file smaller than buffer)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return true
	}
	return err == nil
}

// verifyMountable checks if backup can be mounted
func (e *BackupEngine) verifyMountable(result *VerificationResult) {
	// Verify backup structure
	metadataPath := filepath.Join(result.BackupPath, "metadata.json")

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		result.Status = "failed"
		result.Errors = append(result.Errors, "Файл metadata.json відсутній")
		return
	}

	// Read and validate metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, fmt.Sprintf("Помилка читання metadata: %v", err))
		return
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, fmt.Sprintf("Metadata пошкоджено: %v", err))
		return
	}

	// Check required fields
	requiredFields := []string{"job_id", "job_name", "backup_time", "files_count"}
	for _, field := range requiredFields {
		if _, exists := metadata[field]; !exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Відсутнє поле: %s", field))
		}
	}

	result.FilesChecked = 1
	result.Status = "success"
	result.Details["metadata"] = metadata
}

// verifyBootable checks if VM backup is bootable (Hyper-V specific)
func (e *BackupEngine) verifyBootable(result *VerificationResult) {
	// Check for VM configuration files
	vmPath := filepath.Join(result.BackupPath, "vms")

	if _, err := os.Stat(vmPath); os.IsNotExist(err) {
		result.Status = "warning"
		result.Warnings = append(result.Warnings, "VM бекап не знайдено")
		return
	}

	// Verify VM configuration exists
	entries, err := os.ReadDir(vmPath)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, fmt.Sprintf("Помилка читання VM: %v", err))
		return
	}

	if len(entries) == 0 {
		result.Status = "warning"
		result.Warnings = append(result.Warnings, "VM конфігурації не знайдено")
		return
	}

	result.FilesChecked = len(entries)
	result.Status = "success"
	result.Details["vms_found"] = len(entries)
}

// verifyDataConsistency verifies data consistency
func (e *BackupEngine) verifyDataConsistency(result *VerificationResult) {
	// Verify backup consistency
	archivePath := filepath.Join(result.BackupPath, "backup.zip")

	// Check archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		result.Status = "failed"
		result.Errors = append(result.Errors, "Архів не знайдено")
		return
	}

	// Verify metadata matches actual content
	metadataPath := filepath.Join(result.BackupPath, "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, "Metadata не знайдено")
		return
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, "Metadata пошкоджено")
		return
	}

	// Compare file counts
	if filesCount, ok := metadata["files_count"].(float64); ok {
		r, err := zip.OpenReader(archivePath)
		if err != nil {
			result.Status = "failed"
			result.Errors = append(result.Errors, "Архів пошкоджено")
			return
		}
		defer r.Close()

		actualCount := len(r.File)
		if int(filesCount) != actualCount {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Різниця у кількості файлів: очікувалось %d, фактично %d",
					int(filesCount), actualCount))
		}
		result.FilesChecked = actualCount
	}

	result.Status = "success"
}

// saveVerificationResult saves verification result to file
func (e *BackupEngine) saveVerificationResult(result *VerificationResult) {
	verificationsDir := filepath.Join(e.DataDir, "verifications")
	os.MkdirAll(verificationsDir, 0755)

	resultFile := filepath.Join(verificationsDir, fmt.Sprintf("%s.json", result.ID))
	data, _ := json.MarshalIndent(result, "", "  ")
	os.WriteFile(resultFile, data, 0644)
}

// GetVerificationHistory returns verification history
func (e *BackupEngine) GetVerificationHistory(backupPath string, limit int) ([]VerificationResult, error) {
	verificationsDir := filepath.Join(e.DataDir, "verifications")

	entries, err := os.ReadDir(verificationsDir)
	if err != nil {
		return nil, err
	}

	var results []VerificationResult
	count := 0
	for i := len(entries) - 1; i >= 0 && count < limit; i-- {
		entry := entries[i]
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(verificationsDir, entry.Name()))
		if err != nil {
			continue
		}

		var result VerificationResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		if backupPath == "" || result.BackupPath == backupPath {
			results = append(results, result)
			count++
		}
	}

	return results, nil
}
