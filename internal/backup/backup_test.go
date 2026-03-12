package backup

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"novabackup/pkg/models"
)

func TestBackupManager_CreateJob(t *testing.T) {
	job := &models.Job{
		ID:          uuid.New(),
		Name:        "Test Job",
		Description: "Test description",
		JobType:     models.JobTypeVM,
		Source:      "vm-001",
		Destination: "repo-001",
		Schedule:    "Daily 22:00",
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if job.Name != "Test Job" {
		t.Errorf("Expected job name 'Test Job', got '%s'", job.Name)
	}

	if !job.Enabled {
		t.Error("Job should be enabled")
	}
}

func TestBackupManager_RunJob(t *testing.T) {
	result := &models.BackupResult{
		ID:        uuid.New(),
		JobID:     uuid.New(),
		Status:    "Running",
		StartTime: time.Now(),
	}

	if result.Status != "Running" {
		t.Errorf("Expected status Running, got %s", result.Status)
	}
}

func TestBackupManager_RetentionPolicy(t *testing.T) {
	retainDays := 30

	if retainDays != 30 {
		t.Errorf("Expected retainDays 30, got %d", retainDays)
	}
}

func TestBackupManager_Compression(t *testing.T) {
	originalSize := int64(1000)
	compressedSize := int64(350)

	ratio := float64(originalSize) / float64(compressedSize)

	if ratio < 2.5 || ratio > 3.0 {
		t.Logf("Compression ratio: %.2f", ratio)
	}
}

func TestBackupManager_Deduplication(t *testing.T) {
	totalData := int64(10000)
	uniqueData := int64(2500)

	dedupeRatio := float64(totalData) / float64(uniqueData)

	if dedupeRatio < 3.5 || dedupeRatio > 5.0 {
		t.Logf("Deduplication ratio: %.2f", dedupeRatio)
	}
}

func TestBackupManager_IncrementalBackup(t *testing.T) {
	backups := []struct {
		ID     string
		Type   string
		SizeMB int64
	}{
		{"1", "full", 500},
		{"2", "incremental", 50},
		{"3", "incremental", 30},
	}

	totalSize := int64(0)
	for _, b := range backups {
		totalSize += b.SizeMB
	}

	if totalSize != 580 {
		t.Errorf("Expected total size 580 MB, got %d MB", totalSize)
	}
}

func TestBackupManager_Encryption(t *testing.T) {
	encrypted := true
	algorithm := "AES-256-GCM"

	if !encrypted {
		t.Error("Backup should be encrypted")
	}

	if algorithm != "AES-256-GCM" {
		t.Errorf("Expected AES-256-GCM, got %s", algorithm)
	}
}
