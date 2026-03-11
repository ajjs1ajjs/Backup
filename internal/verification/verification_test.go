package verification

import (
	"context"
	"fmt"
	"testing"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/surebackup"
	"novabackup/internal/storage"
)

// MockTenantManager for AutoVerification testing
type MockTenantManager struct{}

func (m *MockTenantManager) GetTenantFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

func (m *MockTenantManager) WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

func (m *MockTenantManager) GetTenantQuota(tenantID string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"storage_quota": int64(1000000000), // 1GB
		"backup_quota":  int(10),
	}, nil
}

func (m *MockTenantManager) CreateTenant(ctx context.Context, tenant *multitenancy.Tenant) error {
	return nil
}

func (m *MockTenantManager) GetTenant(ctx context.Context, tenantID string) (*multitenancy.Tenant, error) {
	return &multitenancy.Tenant{
		ID:   tenantID,
		Name: "Test Tenant",
	}, nil
}

func (m *MockTenantManager) UpdateTenant(ctx context.Context, tenant *multitenancy.Tenant) error {
	return nil
}

func (m *MockTenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	return nil
}

func (m *MockTenantManager) ListTenants(ctx context.Context) ([]multitenancy.Tenant, error) {
	return []multitenancy.Tenant{}, nil
}

func (m *MockTenantManager) CheckQuota(ctx context.Context, tenantID string, quotaType multitenancy.QuotaType, amount int64) (bool, error) {
	return true, nil
}

func (m *MockTenantManager) GetQuotaUsage(ctx context.Context, tenantID string) (*multitenancy.QuotaUsage, error) {
	return &multitenancy.QuotaUsage{
		BackupsCount:  int64(5),
		StorageUsedGB: 0.5, // 0.5GB
	}, nil
}

func (m *MockTenantManager) UpdateQuota(ctx context.Context, tenantID string, quotas multitenancy.TenantQuotas) error {
	return nil
}

func (m *MockTenantManager) GetTenantResources(ctx context.Context, tenantID string) (*multitenancy.TenantResources, error) {
	return &multitenancy.TenantResources{
		Backups: []multitenancy.TenantResource{},
		VMs: []multitenancy.TenantResource{},
		Storage: []multitenancy.TenantResource{},
		Jobs: []multitenancy.TenantResource{},
		Users: []multitenancy.TenantResource{},
		CustomResources: make(map[string][]multitenancy.TenantResource),
	}, nil
}

func (m *MockTenantManager) AssignResource(ctx context.Context, tenantID string, resourceType string, resourceID string) error {
	return nil
}

func (m *MockTenantManager) UnassignResource(ctx context.Context, tenantID string, resourceType string, resourceID string) error {
	return nil
}

func (m *MockTenantManager) ValidateTenantAccess(ctx context.Context, tenantID string) error {
	return nil
}

// MockSureBackupManager for AutoVerification testing
type MockSureBackupManager struct{}

func (m *MockSureBackupManager) CreateSandbox(ctx context.Context, request *surebackup.SandboxRequest) (*surebackup.Sandbox, error) {
	return &surebackup.Sandbox{
		ID:       "test-sandbox-1",
		Name:     request.Name,
		TenantID: request.TenantID,
		Status:   surebackup.SandboxStatusRunning,
		Type:     request.Type,
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockSureBackupManager) GetSandbox(ctx context.Context, sandboxID string) (*surebackup.Sandbox, error) {
	return &surebackup.Sandbox{
		ID:       sandboxID,
		Name:     "test-sandbox",
		TenantID: "test-tenant",
		Status:   surebackup.SandboxStatusRunning,
		Type:     surebackup.SandboxTypeVM,
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockSureBackupManager) ListSandboxes(ctx context.Context, filter *surebackup.SandboxFilter) ([]*surebackup.Sandbox, error) {
	return []*surebackup.Sandbox{
		{
			ID:       "test-sandbox-1",
			Name:     "test-sandbox",
			TenantID: "test-tenant",
			Status:   surebackup.SandboxStatusRunning,
			Type:     surebackup.SandboxTypeVM,
			CreatedAt: time.Now(),
		},
	}, nil
}

func (m *MockSureBackupManager) DeleteSandbox(ctx context.Context, sandboxID string) error {
	return nil
}

func (m *MockSureBackupManager) StartSandbox(ctx context.Context, sandboxID string) error {
	return nil
}

func (m *MockSureBackupManager) StopSandbox(ctx context.Context, sandboxID string) error {
	return nil
}

func (m *MockSureBackupManager) StartVerification(ctx context.Context, request *surebackup.VerificationRequest) (*surebackup.Verification, error) {
	return &surebackup.Verification{
		ID:        "test-verification-1",
		SandboxID: request.SandboxID,
		BackupID:  request.BackupID,
		TenantID:  request.TenantID,
		Status:    surebackup.VerificationStatusPending,
		Type:      request.Type,
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockSureBackupManager) GetVerification(ctx context.Context, verificationID string) (*surebackup.Verification, error) {
	return &surebackup.Verification{
		ID:        verificationID,
		SandboxID: "test-sandbox-1",
		BackupID:  "test-backup-1",
		TenantID:  "test-tenant",
		Status:    surebackup.VerificationStatusPassed,
		Type:      surebackup.VerificationTypeIntegrity,
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockSureBackupManager) ListVerifications(ctx context.Context, filter *surebackup.VerificationFilter) ([]*surebackup.Verification, error) {
	return []*surebackup.Verification{
		{
			ID:        "test-verification-1",
			SandboxID: "test-sandbox-1",
			BackupID:  "test-backup-1",
			TenantID:  "test-tenant",
			Status:    surebackup.VerificationStatusPassed,
			Type:      surebackup.VerificationTypeIntegrity,
			CreatedAt: time.Now(),
		},
	}, nil
}

func (m *MockSureBackupManager) StopVerification(ctx context.Context, verificationID string) error {
	return nil
}

func (m *MockSureBackupManager) GetSureBackupStats(ctx context.Context, tenantID string) (*surebackup.SureBackupStats, error) {
	return &surebackup.SureBackupStats{
		TenantID:             tenantID,
		TotalSandboxes:       2,
		RunningSandboxes:     1,
		TotalVerifications:   5,
		RunningVerifications: 1,
		PassedVerifications:  4,
		FailedVerifications:  1,
		LastUpdated:          time.Now(),
	}, nil
}

func (m *MockSureBackupManager) GetGlobalStats(ctx context.Context) (*surebackup.GlobalSureBackupStats, error) {
	return &surebackup.GlobalSureBackupStats{
		TotalTenants:          3,
		TotalSandboxes:        6,
		RunningSandboxes:      3,
		TotalVerifications:    15,
		RunningVerifications:  3,
		PassedVerifications:   12,
		FailedVerifications:   3,
		LastUpdated:           time.Now(),
	}, nil
}

// TestAutoVerificationManager tests auto verification management functionality
func TestAutoVerificationManager(t *testing.T) {
	mockTenantMgr := &MockTenantManager{}
	mockStorageMgr := *storage.NewEngine()
	mockSureBackupMgr := &MockSureBackupManager{}
	manager := NewInMemoryAutoVerificationManager(mockTenantMgr, mockStorageMgr, mockSureBackupMgr)

	t.Run("CreateSchedule", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		request := &ScheduleRequest{
			Name:        "daily-verification",
			TenantID:    "test-tenant",
			Description: "Daily backup verification schedule",
			Type:        ScheduleTypeRecurring,
			Trigger: ScheduleTrigger{
				Type: TriggerTypeCron,
				Config: TriggerConfig{
					CronExpression: "0 2 * * *", // Daily at 2 AM
				},
				Timezone: "UTC",
			},
			BackupFilter: BackupFilter{
				RepositoryIDs: []string{"repo-1", "repo-2"},
				BackupTypes:   []string{"full", "incremental"},
				MinAge:        1 * time.Hour,
				MaxAge:        24 * time.Hour,
			},
			Verification: VerificationConfig{
				Timeout: 30 * time.Minute,
				Tests: []surebackup.VerificationTest{
					{
						Name:    "integrity_check",
						Type:    "integrity",
						Command: "checksum --verify",
						Expected: "OK",
						Timeout: 30 * time.Second,
						RetryCount: 3,
					},
				},
				Scripts: []surebackup.VerificationScript{
					{
						Name: "mount_check",
						Path: "/scripts/mount_check.sh",
						Args: []string{"--verbose"},
						Environment: map[string]string{"PATH": "/usr/bin:/bin"},
						RunAs: "root",
						Timeout: 60 * time.Second,
					},
				},
				Notifications: NotificationConfig{
					OnSuccess: []string{"admin@example.com"},
					OnFailure: []string{"admin@example.com", "ops@example.com"},
				},
			},
			Retention: RetentionPolicy{
				ResultsDays:   30,
				ReportsDays:   90,
				LogsDays:      7,
				ArtifactsDays: 14,
				MaxResults:    1000,
				MaxReports:    100,
				MaxLogs:       10000,
				MaxArtifacts:  500,
			},
			Concurrency: ConcurrencySettings{
				MaxConcurrentJobs:      5,
				MaxConcurrentPerBackup: 2,
				QueueSize:              100,
				Priority:               JobPriorityNormal,
				ResourceLimits: ResourceLimits{
					MaxCPU:     4,
					MaxMemory:  8192,
					MaxStorage: 100,
					MaxNetwork: 1000,
				},
			},
			Enabled: true,
			Metadata: map[string]string{
				"environment": "production",
				"team":        "backup-ops",
			},
		}

		schedule, err := manager.CreateSchedule(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create schedule: %v", err)
		}

		if schedule.ID == "" {
			t.Error("Schedule ID should not be empty")
		}

		if schedule.Name != request.Name {
			t.Errorf("Expected name %s, got %s", request.Name, schedule.Name)
		}

		if schedule.TenantID != request.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", request.TenantID, schedule.TenantID)
		}

		if schedule.Type != request.Type {
			t.Errorf("Expected type %s, got %s", request.Type, schedule.Type)
		}

		if !schedule.Enabled {
			t.Error("Schedule should be enabled")
		}

		if schedule.NextRunAt == nil {
			t.Error("NextRunAt should not be nil for enabled schedule")
		}
	})

	t.Run("GetSchedule", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// First create a schedule
		request := &ScheduleRequest{
			Name:     "get-test-schedule",
			TenantID: "test-tenant",
			Type:     ScheduleTypeOnDemand,
			Enabled:  false,
		}

		created, err := manager.CreateSchedule(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create schedule: %v", err)
		}

		// Retrieve the schedule
		retrieved, err := manager.GetSchedule(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get schedule: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}

		if retrieved.Name != created.Name {
			t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
		}
	})

	t.Run("ListSchedules", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create multiple schedules
		for i := 0; i < 3; i++ {
			request := &ScheduleRequest{
				Name:     fmt.Sprintf("list-test-schedule-%d", i),
				TenantID: "test-tenant",
				Type:     ScheduleTypeRecurring,
				Enabled:  i%2 == 0, // Enable every other schedule
			}

			_, err := manager.CreateSchedule(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create schedule %d: %v", i, err)
			}
		}

		// List all schedules for tenant
		schedules, err := manager.ListSchedules(ctx, &ScheduleFilter{
			TenantID: "test-tenant",
		})
		if err != nil {
			t.Fatalf("Failed to list schedules: %v", err)
		}

		if len(schedules) < 3 {
			t.Errorf("Expected at least 3 schedules, got %d", len(schedules))
		}

		// Filter by enabled status
		enabledSchedules, err := manager.ListSchedules(ctx, &ScheduleFilter{
			TenantID: "test-tenant",
			Enabled:  &[]bool{true}[0],
		})
		if err != nil {
			t.Fatalf("Failed to list enabled schedules: %v", err)
		}

		// Should have some enabled schedules
		if len(enabledSchedules) < 1 {
			t.Errorf("Expected at least 1 enabled schedule, got %d", len(enabledSchedules))
		}
	})

	t.Run("UpdateSchedule", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a schedule
		request := &ScheduleRequest{
			Name:     "update-test-schedule",
			TenantID: "test-tenant",
			Type:     ScheduleTypeRecurring,
			Enabled:  false,
		}

		schedule, err := manager.CreateSchedule(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create schedule: %v", err)
		}

		// Update the schedule
		newName := "updated-schedule-name"
		newDescription := "Updated description"
		enabled := true

		updateRequest := &UpdateScheduleRequest{
			Name:        &newName,
			Description: &newDescription,
			Enabled:     &enabled,
		}

		updated, err := manager.UpdateSchedule(ctx, schedule.ID, updateRequest)
		if err != nil {
			t.Fatalf("Failed to update schedule: %v", err)
		}

		if updated.Name != newName {
			t.Errorf("Expected name %s, got %s", newName, updated.Name)
		}

		if updated.Description != newDescription {
			t.Errorf("Expected description %s, got %s", newDescription, updated.Description)
		}

		if !updated.Enabled {
			t.Error("Schedule should be enabled after update")
		}
	})

	t.Run("EnableSchedule", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a disabled schedule
		request := &ScheduleRequest{
			Name:     "enable-test-schedule",
			TenantID: "test-tenant",
			Type:     ScheduleTypeRecurring,
			Enabled:  false,
		}

		schedule, err := manager.CreateSchedule(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create schedule: %v", err)
		}

		// Enable the schedule
		err = manager.EnableSchedule(ctx, schedule.ID)
		if err != nil {
			t.Fatalf("Failed to enable schedule: %v", err)
		}

		// Verify it's enabled
		enabled, err := manager.GetSchedule(ctx, schedule.ID)
		if err != nil {
			t.Fatalf("Failed to get schedule: %v", err)
		}

		if !enabled.Enabled {
			t.Error("Schedule should be enabled")
		}
	})

	t.Run("DisableSchedule", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create an enabled schedule
		request := &ScheduleRequest{
			Name:     "disable-test-schedule",
			TenantID: "test-tenant",
			Type:     ScheduleTypeRecurring,
			Enabled:  true,
		}

		schedule, err := manager.CreateSchedule(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create schedule: %v", err)
		}

		// Disable the schedule
		err = manager.DisableSchedule(ctx, schedule.ID)
		if err != nil {
			t.Fatalf("Failed to disable schedule: %v", err)
		}

		// Verify it's disabled
		disabled, err := manager.GetSchedule(ctx, schedule.ID)
		if err != nil {
			t.Fatalf("Failed to get schedule: %v", err)
		}

		if disabled.Enabled {
			t.Error("Schedule should be disabled")
		}

		if disabled.NextRunAt != nil {
			t.Error("NextRunAt should be nil for disabled schedule")
		}
	})

	t.Run("DeleteSchedule", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a schedule
		request := &ScheduleRequest{
			Name:     "delete-test-schedule",
			TenantID: "test-tenant",
			Type:     ScheduleTypeOnDemand,
		}

		schedule, err := manager.CreateSchedule(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create schedule: %v", err)
		}

		// Delete the schedule
		err = manager.DeleteSchedule(ctx, schedule.ID)
		if err != nil {
			t.Fatalf("Failed to delete schedule: %v", err)
		}

		// Verify it's gone
		_, err = manager.GetSchedule(ctx, schedule.ID)
		if err == nil {
			t.Error("Expected error when getting deleted schedule")
		}
	})

	t.Run("GetVerificationStats", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create some schedules and jobs
		for i := 0; i < 2; i++ {
			scheduleRequest := &ScheduleRequest{
				Name:     fmt.Sprintf("stats-schedule-%d", i),
				TenantID: "test-tenant",
				Type:     ScheduleTypeRecurring,
				Enabled:  true,
			}

			schedule, err := manager.CreateSchedule(ctx, scheduleRequest)
			if err != nil {
				t.Fatalf("Failed to create schedule %d: %v", i, err)
			}

			// Create a verification job
			job := &VerificationJob{
				ID:         fmt.Sprintf("stats-job-%d", i),
				ScheduleID: schedule.ID,
				TenantID:   "test-tenant",
				BackupID:   fmt.Sprintf("stats-backup-%d", i),
				Status:     JobStatusCompleted,
				Priority:   JobPriorityNormal,
				Config: VerificationConfig{
					Timeout: 30 * time.Minute,
				},
				Progress: JobProgress{
					Percentage:     100,
					CurrentStep:    "completed",
					TotalSteps:     1,
					CompletedSteps: 1,
				},
				Timing: JobTiming{
					ExecutionDuration: 2 * time.Minute,
					TotalDuration:     2 * time.Minute + 30*time.Second,
				},
				CreatedAt:   time.Now().Add(-1 * time.Hour),
				StartedAt:    &[]time.Time{time.Now().Add(-58 * time.Minute)}[0],
				CompletedAt:  &[]time.Time{time.Now().Add(-56 * time.Minute)}[0],
			}

			// Add job to manager (simulate job creation)
			manager.jobs[job.ID] = job
		}

		// Get tenant stats
		timeRange := TimeRange{
			From: time.Now().Add(-2 * time.Hour),
			To:   time.Now(),
		}

		stats, err := manager.GetVerificationStats(ctx, "test-tenant", timeRange)
		if err != nil {
			t.Fatalf("Failed to get verification stats: %v", err)
		}

		if stats.TenantID != "test-tenant" {
			t.Errorf("Expected tenant ID %s, got %s", "test-tenant", stats.TenantID)
		}

		if stats.TotalJobs < 2 {
			t.Errorf("Expected at least 2 total jobs, got %d", stats.TotalJobs)
		}

		if stats.CompletedJobs < 2 {
			t.Errorf("Expected at least 2 completed jobs, got %d", stats.CompletedJobs)
		}

		if stats.SuccessRate <= 0 {
			t.Errorf("Expected positive success rate, got %f", stats.SuccessRate)
		}
	})

	t.Run("GetSystemHealth", func(t *testing.T) {
		ctx := context.Background()

		// Get system health
		health, err := manager.GetSystemHealth(ctx)
		if err != nil {
			t.Fatalf("Failed to get system health: %v", err)
		}

		if health.Status != HealthStatusHealthy {
			t.Errorf("Expected status %s, got %s", HealthStatusHealthy, health.Status)
		}

		if len(health.WorkerNodes) == 0 {
			t.Error("Expected at least one worker node")
		}

		if health.ErrorRate < 0 || health.ErrorRate > 1 {
			t.Errorf("Expected error rate between 0 and 1, got %f", health.ErrorRate)
		}

		if health.ResponseTime <= 0 {
			t.Errorf("Expected positive response time, got %v", health.ResponseTime)
		}
	})

	t.Run("GetPendingJobs", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create some pending jobs
		for i := 0; i < 2; i++ {
			job := &VerificationJob{
				ID:         fmt.Sprintf("pending-job-%d", i),
				ScheduleID: "test-schedule",
				TenantID:   "test-tenant",
				BackupID:   fmt.Sprintf("pending-backup-%d", i),
				Status:     JobStatusPending,
				Priority:   JobPriorityNormal,
				Config: VerificationConfig{
					Timeout: 30 * time.Minute,
				},
				CreatedAt: time.Now(),
			}

			manager.jobs[job.ID] = job
		}

		// Get pending jobs
		pendingJobs, err := manager.GetPendingJobs(ctx)
		if err != nil {
			t.Fatalf("Failed to get pending jobs: %v", err)
		}

		if len(pendingJobs) < 2 {
			t.Errorf("Expected at least 2 pending jobs, got %d", len(pendingJobs))
		}

		for _, job := range pendingJobs {
			if job.Status != JobStatusPending && job.Status != JobStatusQueued {
				t.Errorf("Expected pending or queued status, got %s", job.Status)
			}
		}
	})
}
