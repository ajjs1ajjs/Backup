// Synthetic Full Backups - Create full backup from incremental chain
package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// SyntheticFullConfig represents synthetic full backup configuration
type SyntheticFullConfig struct {
	BaseBackupPath   string   // Path to last full backup
	IncrementalPaths []string // Paths to incremental backups in order
	OutputPath       string   // Where to create synthetic full
	Compression      bool     // Compress output
	VerifyIntegrity  bool     // Verify after creation
}

// SyntheticFullResult represents the result of synthetic full creation
type SyntheticFullResult struct {
	Success        bool      `json:"success"`
	OutputPath     string    `json:"output_path"`
	Size           int64     `json:"size"`
	Duration       string    `json:"duration"`
	FilesProcessed int       `json:"files_processed"`
	BlocksMerged   int       `json:"blocks_merged"`
	Error          string    `json:"error,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateSyntheticFull creates a synthetic full backup from incremental chain
func (e *BackupEngine) CreateSyntheticFull(config *SyntheticFullConfig) (*SyntheticFullResult, error) {
	startTime := time.Now()

	result := &SyntheticFullResult{
		Success:    false,
		OutputPath: config.OutputPath,
		CreatedAt:  startTime,
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputPath, 0755); err != nil {
		result.Error = fmt.Sprintf("Failed to create output directory: %v", err)
		return result, err
	}

	// Create synthetic full archive
	outputPath := filepath.Join(config.OutputPath, "backup.zip")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create output file: %v", err)
		return result, err
	}
	defer outputFile.Close()

	// Merge logic using manifests
	filesProcessed, blocksMerged, err := e.mergeBackups(outputFile, config)
	if err != nil {
		result.Error = fmt.Sprintf("Merge failed: %v", err)
		return result, err
	}

	// Verify integrity if requested
	if config.VerifyIntegrity {
		if err := e.verifySyntheticFull(outputPath); err != nil {
			result.Error = fmt.Sprintf("Verification failed: %v", err)
			return result, err
		}
	}

	// Get output size
	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get output size: %v", err)
		return result, err
	}

	// Success!
	result.Success = true
	result.Size = outputInfo.Size()
	result.Duration = time.Since(startTime).String()
	result.FilesProcessed = filesProcessed
	result.BlocksMerged = blocksMerged

	return result, nil
}

// mergeBackups merges base and incremental backups into single archive
func (e *BackupEngine) mergeBackups(output *os.File, config *SyntheticFullConfig) (int, int, error) {
	zipWriter := zip.NewWriter(output)
	defer zipWriter.Close()

	// We need to keep track of the latest version of every file
	latestFiles := make(map[string]*zip.File)

	// Chain order: Base (Full) -> Inc1 -> Inc2 -> ...
	// Later files in the chain override earlier ones
	allPaths := append([]string{config.BaseBackupPath}, config.IncrementalPaths...)

	for _, path := range allPaths {
		archivePath := filepath.Join(path, "backup.zip")
		r, err := zip.OpenReader(archivePath)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to open archive %s: %v", archivePath, err)
		}
		// Note: We can't close r yet because we'll need to read file contents later
		// In a production app, we'd close it and re-open or read everything into memory if small
		for _, f := range r.File {
			latestFiles[f.Name] = f
		}
		// For simplicity in this demo, we'll copy immediately to avoid keeping many zip.Readers open
		for _, f := range r.File {
			if latestFiles[f.Name] == f {
				if err := e.copyZipEntry(zipWriter, f); err != nil {
					r.Close()
					return 0, 0, err
				}
			}
		}
		r.Close()
	}

	return len(latestFiles), 0, nil
}

// copyZipEntry copies a zip entry from one archive to another
func (e *BackupEngine) copyZipEntry(zw *zip.Writer, f *zip.File) error {
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := zw.CreateHeader(&f.FileHeader)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	return err
}

// copyFileToArchive copies a file (placeholder for production)
func (e *BackupEngine) copyFileToArchive(zipWriter *zip.Writer, file string) error {
	// In production: copy actual file data
	return nil
}

// verifySyntheticFull verifies the integrity of synthetic full backup
func (e *BackupEngine) verifySyntheticFull(path string) error {
	// Try to open and verify the archive
	_, err := zip.OpenReader(path)
	return err
}

// ShouldCreateSyntheticFull determines if synthetic full should be created
func ShouldCreateSyntheticFull(lastFull time.Time, incrementalCount int, config *BackupJob) bool {
	// Check if it's time for a full backup
	daysSinceFull := int(time.Since(lastFull).Hours() / 24)

	if config.FullBackupEvery > 0 && daysSinceFull >= config.FullBackupEvery {
		return true
	}

	// Or if too many incrementals (chain too long)
	if incrementalCount >= 10 {
		return true
	}

	return false
}

// GetOptimalSyntheticFullSchedule returns optimal schedule for synthetic full
func GetOptimalSyntheticFullSchedule(config *BackupJob) string {
	// Synthetic full should run when:
	// 1. Production load is low (e.g., Sunday 2 AM)
	// 2. After last incremental of the week
	// 3. Enough time before next incremental cycle

	return "0 2 * * 0" // Sunday at 2 AM
}
