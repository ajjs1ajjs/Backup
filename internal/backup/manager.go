package backup

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"novabackup/internal/compression"
	"novabackup/internal/database"
	"novabackup/internal/deduplication"
	"novabackup/internal/providers"
	"novabackup/pkg/models"

	"github.com/google/uuid"
)

// BackupManager orchestrates backup jobs
type BackupManager struct {
	db     *database.Connection
	dedupe *deduplication.PersistentDeduplicationManager
	comp   compression.Compressor
	ret    *RetentionService
	gc     *deduplication.GarbageCollector
	lastJobID uuid.UUID
}

// NewBackupManager creates a new backup manager
func NewBackupManager(db *database.Connection) *BackupManager {
	storagePath := "C:\\Backup\\Chunks" // Default path for now
	return &BackupManager{
		db:     db,
		dedupe: deduplication.NewPersistentDeduplicationManager(db, storagePath),
		comp:   compression.NewGzipCompressor(6),
		ret:    NewRetentionService(db),
		gc:     deduplication.NewGarbageCollector(db),
	}
}

func (bm *BackupManager) GetCompressor() compression.Compressor {
	return bm.comp
}

// RunJob executes a backup job manually or by schedule
func (bm *BackupManager) RunJob(jobID uuid.UUID) error {
	bm.lastJobID = jobID
	// 1. Get job from DB
	job, err := bm.db.GetJobByID(jobID)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	// 2. Create backup result record (Starting)
	result := &models.BackupResult{
		ID:        uuid.New(),
		JobID:     jobID,
		Status:    models.JobStatusRunning,
		StartTime: time.Now(),
	}

	if err := bm.db.CreateBackupResult(result); err != nil {
		return fmt.Errorf("failed to create backup result: %w", err)
	}

	go bm.executeBackup(job, result)

	return nil
}

func (bm *BackupManager) executeBackup(job *models.Job, result *models.BackupResult) {
	log.Printf("[BackupManager] Starting job: %s (%s)", job.Name, job.ID)

	ctx := context.Background()
	var err error

	// 3. Selection of provider (Only Hyper-V for now)
	if job.JobType == models.JobTypeVM {
		hp := providers.NewHyperVBackupProvider(providers.HyperVConfig{
			Host: "localhost",
		})
		
		// 4. Guest Processing (Application-Aware)
		var vss *VSSRequestor
		if job.Enabled { // Assuming 'Enabled' is the flag for now, or use BackupConfig logic
			log.Printf("[BackupManager] Enabling Guest Processing for VM: %s", job.Source)
			vss = NewVSSRequestor("localhost", job.Source, "", "") // TODO: Load credentials
			if err := vss.Freeze(ctx); err != nil {
				log.Printf("[BackupManager] Guest Freeze failed (non-critical): %v", err)
			}
		}

		var finalResult *models.BackupResult
		finalResult, err = hp.Backup(ctx, job.Source, job.Destination)
		
		if err == nil {
			result.Status = models.JobStatusCompleted
			result.BytesWritten = finalResult.BytesWritten
			result.FilesTotal = finalResult.FilesTotal
			result.FilesSuccess = finalResult.FilesSuccess

			// 5. Post-Backup Guest Processing (Log Truncation)
			if vss != nil {
				log.Printf("[BackupManager] Performing log truncation for VM: %s", job.Source)
				if err := vss.TruncateLogs(ctx); err != nil {
					log.Printf("[BackupManager] Log truncation failed (non-critical): %v", err)
				}
			}

			// Process exported files for deduplication
			bm.processExportedFiles(ctx, job.Source, job.Destination)
		}
	} else {
		err = fmt.Errorf("unsupported job type: %s", job.JobType)
	}

	result.EndTime = time.Now()
	if err != nil {
		result.Status = models.JobStatusFailed
		result.ErrorMessage = err.Error()
		log.Printf("[BackupManager] Job failed: %v", err)
	} else {
		log.Printf("[BackupManager] Job completed: %s", job.Name)
	}

	// 4. Update result in DB
	if err := bm.db.UpdateBackupResult(result); err != nil {
		log.Printf("[BackupManager] Failed to update final status: %v", err)
	}

	// 5. Apply Retention Policy
	if result.Status == models.JobStatusCompleted {
		if err := bm.ret.ApplyPolicy(ctx, job.ID, job.RetentionDays); err != nil {
			log.Printf("[BackupManager] Retention failed: %v", err)
		}

		// 6. Run Garbage Collection
		if err := bm.gc.Run(ctx); err != nil {
			log.Printf("[BackupManager] GC failed: %v", err)
		}
	}
}

func (bm *BackupManager) processExportedFiles(ctx context.Context, vmName string, dest string) {
	exportPath := filepath.Join(dest, vmName)
	log.Printf("[BackupManager] Processing exported data for deduplication: %s", exportPath)

	// Create a new Restore Point
	rp := &models.RestorePoint{
		ID:        uuid.New(),
		JobID:     bm.lastJobID, // We need to store this in bm during RunJob
		PointTime: time.Now(),
		Status:    "completed",
	}
	
	if err := bm.db.CreateRestorePoint(rp); err != nil {
		log.Printf("[BackupManager] Failed to create Restore Point: %v", err)
		return
	}

	err := filepath.Walk(exportPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		relPath, _ := filepath.Rel(exportPath, path)
		chunkSize := 1024 * 1024 // 1MB
		buffer := make([]byte, chunkSize)
		sequence := 0

		for {
			n, err := file.Read(buffer)
			if n > 0 {
				// 1. Compress (For now we compress each chunk independently for random access)
				compressed, _ := bm.comp.Compress(buffer[:n])
				
				// 2. Deduplicate/Store
				hash, err := bm.dedupe.StoreChunk(ctx, compressed)
				if err != nil {
					return err
				}

				// 3. Save mapping
				if err := bm.db.SaveRestorePointChunk(rp.ID, hash, sequence, relPath); err != nil {
					return err
				}
				sequence++
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("[BackupManager] Deduplication failed: %v", err)
	}
}
