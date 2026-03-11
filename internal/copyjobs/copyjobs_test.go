package copyjobs

import (
	"context"
	"testing"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/deduplication"
)

// MockStorageManager for copy jobs testing
type MockStorageManager struct{}

func (m *MockStorageManager) StoreData(ctx context.Context, key string, data []byte) error {
	return nil
}

func (m *MockStorageManager) RetrieveData(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (m *MockStorageManager) DeleteData(ctx context.Context, key string) error {
	return nil
}

func (m *MockStorageManager) ListData(ctx context.Context, prefix string) (map[string][]byte, error) {
	return make(map[string][]byte), nil
}

func (m *MockStorageManager) GetStorageInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"total_space": int64(1000000000), // 1GB
		"used_space":  int64(100000000),  // 1GB
	}, nil
}

// TestBackupCopyManager tests backup copy job management
func TestBackupCopyManager(t *testing.T) {
	mockTenantMgr := &MockTenantManager{}
	mockDedupeMgr := &MockDeduplicationManager{}
	mockStorageMgr := &MockStorageManager{}
	
	t.Run("CreateJob", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		
		job := &BackupCopyJob{
			Name:       "Test Backup Copy",
			SourceRepo: "primary-repo",
			TargetRepo: "secondary-repo",
			BackupType: "full",
			Schedule:   "0 2 * * *", // Daily at 2 AM
			Enabled:    true,
			Priority:   PriorityNormal,
			Settings: map[string]interface{}{
				"compression": true,
				"encryption": false,
			},
			Metadata: map[string]string{
				"description": "Test backup copy job",
				"environment": "test",
			},
		}
		
		manager := NewInMemoryBackupCopyManager(mockStorageMgr, mockTenantMgr, mockDedupeMgr)
		
		err := manager.CreateJob(ctx, job)
		if err != nil {
			t.Fatalf("Failed to create backup copy job: %v", err)
		}
		
		if job.ID == "" {
			t.Error("Job ID should not be empty")
		}
		
		if job.TenantID != "test-tenant" {
			t.Errorf("Expected tenant ID %s, got %s", "test-tenant", job.TenantID)
		}
		
		if job.Status != JobStatusPending {
			t.Errorf("Expected status %s, got %s", JobStatusPending, job.Status)
		}
	})

	t.Run("GetJob", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemoryBackupCopyManager(mockStorageMgr, mockTenantMgr, mockDedupeMgr)
		
		// First create a job
		job := &BackupCopyJob{
			Name:       "Test Job",
			SourceRepo: "primary-repo",
			TargetRepo: "secondary-repo",
			BackupType: "incremental",
			Enabled:    true,
			Priority:   PriorityHigh,
			TenantID:   "test-tenant",
		}
		
		err := manager.CreateJob(ctx, job)
		if err != nil {
			t.Fatalf("Failed to create backup copy job: %v", err)
		}
		
		// Retrieve the job
		retrievedJob, err := manager.GetJob(ctx, job.ID)
		if err != nil {
			t.Fatalf("Failed to get backup copy job: %v", err)
		}
		
		if retrievedJob.ID != job.ID {
			t.Errorf("Expected job ID %s, got %s", job.ID, retrievedJob.ID)
		}
		
		if retrievedJob.Name != job.Name {
			t.Errorf("Expected job name %s, got %s", job.Name, retrievedJob.Name)
		}
	})

	t.Run("UpdateJob", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemoryBackupCopyManager(mockStorageMgr, mockTenantMgr, mockDedupeMgr)
		
		// Create a job first
		job := &BackupCopyJob{
			Name:       "Original Job",
			SourceRepo: "primary-repo",
			TargetRepo: "secondary-repo",
			BackupType: "full",
			Enabled:    true,
			Priority:   PriorityNormal,
			TenantID:   "test-tenant",
		}
		
		err := manager.CreateJob(ctx, job)
		if err != nil {
			t.Fatalf("Failed to create backup copy job: %v", err)
		}
		
		// Update the job
		updatedJob := &BackupCopyJob{
			ID:          job.ID,
			Name:        "Updated Job",
			SourceRepo:  "updated-primary-repo",
			TargetRepo:  "updated-secondary-repo",
			BackupType:  "incremental",
			Enabled:     false,
			Priority:    PriorityHigh,
			Settings: map[string]interface{}{
				"retention_days": 30,
			},
		}
		
		err = manager.UpdateJob(ctx, updatedJob)
		if err != nil {
			t.Fatalf("Failed to update backup copy job: %v", err)
		}
		
		// Verify the update
		retrievedJob, err := manager.GetJob(ctx, job.ID)
		if err != nil {
			t.Fatalf("Failed to get updated backup copy job: %v", err)
		}
		
		if retrievedJob.Name != "Updated Job" {
			t.Errorf("Expected updated job name %s, got %s", "Updated Job", retrievedJob.Name)
		}
		
		if retrievedJob.Enabled != false {
			t.Errorf("Expected job to be disabled, got %t", retrievedJob.Enabled)
		}
	})

	t.Run("ListJobs", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemoryBackupCopyManager(mockStorageMgr, mockTenantMgr, mockDedupeMgr)
		
		// Create multiple jobs
		for i := 0; i < 5; i++ {
			job := &BackupCopyJob{
				ID:          fmt.Sprintf("list-job-%d", i),
				Name:        fmt.Sprintf("List Test Job %d", i),
				SourceRepo: "primary-repo",
				TargetRepo: "secondary-repo",
				BackupType: "full",
				Enabled:    i < 3, // First 3 enabled
				Priority:   Priority(i % 2 == 0 ? PriorityHigh : PriorityNormal),
				TenantID:   "test-tenant",
				Status:     JobStatus(i % 2 == 1 ? JobStatusCompleted : JobStatusPending),
			}
			
			err := manager.CreateJob(ctx, job)
			if err != nil {
				t.Fatalf("Failed to create list job %d: %v", i, err)
			}
		}
		
		// Test listing without filter
		jobs, err := manager.ListJobs(ctx, &JobFilter{})
		if err != nil {
			t.Fatalf("Failed to list jobs: %v", err)
		}
		
		if len(jobs) != 5 {
			t.Errorf("Expected 5 jobs, got %d", len(jobs))
		}
		
		// Test listing with filter
		filter := &JobFilter{
			Status:   JobStatusCompleted,
			Priority: PriorityHigh,
			Limit:    3,
		}
		
		filteredJobs, err := manager.ListJobs(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list filtered jobs: %v", err)
		}
		
		if len(filteredJobs) != 2 {
			t.Errorf("Expected 2 filtered jobs, got %d", len(filteredJobs))
		}
		
		// Verify tenant isolation
		otherTenantCtx := mockTenantMgr.WithTenant(context.Background(), "other-tenant")
		otherJobs, err := manager.ListJobs(otherTenantCtx, &JobFilter{})
		if err != nil {
			t.Fatalf("Failed to list jobs for other tenant: %v", err)
		}
		
		if len(otherJobs) != 0 {
			t.Errorf("Expected no jobs for other tenant, got %d", len(otherJobs))
		}
	})

	t.Run("JobExecution", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemoryBackupCopyManager(mockStorageMgr, mockTenantMgr, mockDedupeMgr)
		
		// Create a job
		job := &BackupCopyJob{
			Name:       "Execution Test",
			SourceRepo: "primary-repo",
			TargetRepo: "secondary-repo",
			BackupType: "full",
			Enabled:    true,
			Priority:   PriorityNormal,
			TenantID:   "test-tenant",
		}
		
		err := manager.CreateJob(ctx, job)
		if err != nil {
			t.Fatalf("Failed to create execution test job: %v", err)
		}
		
		// Start the job
		err = manager.StartJob(ctx, job.ID)
		if err != nil {
			t.Fatalf("Failed to start job: %v", err)
		}
		
		// Wait a bit for progress
		time.Sleep(100 * time.Millisecond)
		
		// Check progress
		retrievedJob, err := manager.GetJob(ctx, job.ID)
		if err != nil {
			t.Fatalf("Failed to get running job: %v", err)
		}
		
		if retrievedJob.Status != JobStatusRunning {
			t.Errorf("Expected job status %s, got %s", JobStatusRunning, retrievedJob.Status)
		}
		
		if retrievedJob.Progress.Percentage <= 0 {
			t.Error("Job progress should be greater than 0")
		}
		
		// Stop the job
		err = manager.StopJob(ctx, job.ID)
		if err != nil {
			t.Fatalf("Failed to stop job: %v", err)
		}
		
		// Check final status
		retrievedJob, err = manager.GetJob(ctx, job.ID)
		if err != nil {
			t.Fatalf("Failed to get stopped job: %v", err)
		}
		
		if retrievedJob.Status != JobStatusPaused {
			t.Errorf("Expected job status %s, got %s", JobStatusPaused, retrievedJob.Status)
		}
	})

	t.Run("JobStatistics", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemoryBackupCopyManager(mockStorageMgr, mockTenantMgr, mockDedupeMgr)
		
		// Get manager statistics
		stats, err := manager.GetManagerStats(ctx)
		if err != nil {
			t.Fatalf("Failed to get manager stats: %v", err)
		}
		
		if stats.TotalJobs != 5 {
			t.Errorf("Expected 5 total jobs, got %d", stats.TotalJobs)
		}
		
		if stats.ActiveJobs != 0 {
			t.Errorf("Expected 0 active jobs, got %d", stats.ActiveJobs)
		}
		
		if stats.PendingJobs != 3 {
			t.Errorf("Expected 3 pending jobs, got %d", stats.PendingJobs)
		}
		
		if stats.CompletedJobs != 2 {
			t.Errorf("Expected 2 completed jobs, got %d", stats.CompletedJobs)
		}
		
		if stats.FailedJobs != 0 {
			t.Errorf("Expected 0 failed jobs, got %d", stats.FailedJobs)
		}
	})

	t.Run("JobScheduling", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemoryBackupCopyManager(mockStorageMgr, mockTenantMgr, mockDedupeMgr)
		
		// Create a job with schedule
		job := &BackupCopyJob{
			Name:       "Scheduled Job",
			SourceRepo: "primary-repo",
			TargetRepo: "secondary-repo",
			BackupType: "incremental",
			Schedule:   "0 2 * * *", // Daily at 2 AM
			Enabled:    true,
			Priority:   PriorityNormal,
			TenantID:   "test-tenant",
		}
		
		err := manager.CreateJob(ctx, job)
		if err != nil {
			t.Fatalf("Failed to create scheduled job: %v", err)
		}
		
		// Test scheduling
		err = manager.ScheduleJob(ctx, job.ID, "0 4 * * *") // Change to 4 AM
		if err != nil {
			t.Fatalf("Failed to schedule job: %v", err)
		}
		
		// Test unscheduling
		err = manager.UnscheduleJob(ctx, job.ID)
		if err != nil {
			t.Fatalf("Failed to unschedule job: %v", err)
		}
		
		// Verify schedule was removed
		retrievedJob, err := manager.GetJob(ctx, job.ID)
		if err != nil {
			t.Fatalf("Failed to get unscheduled job: %v", err)
		}
		
		if retrievedJob.Schedule != "" {
			t.Errorf("Expected empty schedule, got %s", retrievedJob.Schedule)
		}
	})

	t.Run("DeleteJob", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemoryBackupCopyManager(mockStorageMgr, mockTenantMgr, mockDedupeMgr)
		
		// Create a job
		job := &BackupCopyJob{
			Name:       "Delete Test",
			SourceRepo: "primary-repo",
			TargetRepo: "secondary-repo",
			BackupType: "full",
			Enabled:    true,
			Priority:   PriorityNormal,
			TenantID:   "test-tenant",
		}
		
		err := manager.CreateJob(ctx, job)
		if err != nil {
			t.Fatalf("Failed to create delete test job: %v", err)
		}
		
		// Delete the job
		err = manager.DeleteJob(ctx, job.ID)
		if err != nil {
			t.Fatalf("Failed to delete job: %v", err)
		}
		
		// Verify job is deleted
		_, err = manager.GetJob(ctx, job.ID)
		if err == nil {
			t.Error("Expected error when getting deleted job")
		}
	})
}
