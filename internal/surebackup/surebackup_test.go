package surebackup

import (
	"context"
	"fmt"
	"testing"
	"time"

	"novabackup/internal/multitenancy"
	"novabackup/internal/storage"
)

// MockTenantManager for SureBackup testing
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
		Backups:         []multitenancy.TenantResource{},
		VMs:             []multitenancy.TenantResource{},
		Storage:         []multitenancy.TenantResource{},
		Jobs:            []multitenancy.TenantResource{},
		Users:           []multitenancy.TenantResource{},
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

// MockStorageEngine for SureBackup testing
type MockStorageEngine struct {
	storage.Engine
}

func NewMockStorageEngine() *MockStorageEngine {
	return &MockStorageEngine{
		Engine: *storage.NewEngine(),
	}
}

// TestSureBackupManager tests SureBackup management functionality
func TestSureBackupManager(t *testing.T) {
	mockTenantMgr := &MockTenantManager{}
	mockStorageMgr := *storage.NewEngine()
	manager := NewInMemorySureBackupManager(mockTenantMgr, mockStorageMgr)

	t.Run("CreateSandbox", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		request := &SandboxRequest{
			Name:     "test-sandbox",
			TenantID: "test-tenant",
			Type:     SandboxTypeVM,
			Environment: SandboxEnvironment{
				OS:           "linux",
				Version:      "ubuntu-20.04",
				Architecture: "x86_64",
				Tools:        []string{"curl", "wget"},
				Config:       map[string]string{"timezone": "UTC"},
			},
			Resources: SandboxResources{
				CPUCount:    2,
				MemoryMB:    4096,
				StorageGB:   20,
				NetworkMbps: 1000,
			},
			Network: SandboxNetwork{
				Isolated:     true,
				Subnet:       "10.0.1.0/24",
				AllowedPorts: []int{22, 80, 443},
				DNS:          []string{"8.8.8.8", "8.8.4.4"},
			},
			Storage: SandboxStorage{
				BackendType: "local",
				Config:      map[string]string{"path": "/tmp/sandbox"},
				MountPoints: []MountPoint{
					{
						Path:     "/data",
						SizeGB:   10,
						ReadOnly: false,
					},
				},
			},
			AutoStart: true,
			ExpiresIn: 24 * time.Hour,
			Metadata: map[string]string{
				"environment": "test",
				"purpose":     "backup-verification",
			},
		}

		sandbox, err := manager.CreateSandbox(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create sandbox: %v", err)
		}

		if sandbox.ID == "" {
			t.Error("Sandbox ID should not be empty")
		}

		if sandbox.Name != request.Name {
			t.Errorf("Expected name %s, got %s", request.Name, sandbox.Name)
		}

		if sandbox.TenantID != request.TenantID {
			t.Errorf("Expected tenant ID %s, got %s", request.TenantID, sandbox.TenantID)
		}

		if sandbox.Status != SandboxStatusRunning {
			t.Errorf("Expected status %s, got %s", SandboxStatusRunning, sandbox.Status)
		}

		if sandbox.Type != request.Type {
			t.Errorf("Expected type %s, got %s", request.Type, sandbox.Type)
		}

		if sandbox.StartedAt == nil {
			t.Error("StartedAt should not be nil when AutoStart is true")
		}

		if sandbox.ExpiresAt == nil {
			t.Error("ExpiresAt should not be nil when ExpiresIn is set")
		}
	})

	t.Run("GetSandbox", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// First create a sandbox
		request := &SandboxRequest{
			Name:      "get-test-sandbox",
			TenantID:  "test-tenant",
			Type:      SandboxTypeContainer,
			AutoStart: false,
		}

		created, err := manager.CreateSandbox(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create sandbox: %v", err)
		}

		// Retrieve the sandbox
		retrieved, err := manager.GetSandbox(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get sandbox: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}

		if retrieved.Name != created.Name {
			t.Errorf("Expected name %s, got %s", created.Name, retrieved.Name)
		}
	})

	t.Run("ListSandboxes", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create multiple sandboxes
		for i := 0; i < 3; i++ {
			request := &SandboxRequest{
				Name:      fmt.Sprintf("list-test-sandbox-%d", i),
				TenantID:  "test-tenant",
				Type:      SandboxTypeVM,
				AutoStart: false,
			}

			_, err := manager.CreateSandbox(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create sandbox %d: %v", i, err)
			}
		}

		// List all sandboxes for tenant
		sandboxes, err := manager.ListSandboxes(ctx, &SandboxFilter{
			TenantID: "test-tenant",
		})
		if err != nil {
			t.Fatalf("Failed to list sandboxes: %v", err)
		}

		if len(sandboxes) < 3 {
			t.Errorf("Expected at least 3 sandboxes, got %d", len(sandboxes))
		}

		// Filter by status
		runningSandboxes, err := manager.ListSandboxes(ctx, &SandboxFilter{
			TenantID: "test-tenant",
			Status:   SandboxStatusRunning,
		})
		if err != nil {
			t.Fatalf("Failed to list running sandboxes: %v", err)
		}

		// Should have some running sandboxes (from previous tests)
		if len(runningSandboxes) < 1 {
			t.Errorf("Expected at least 1 running sandbox, got %d", len(runningSandboxes))
		}
	})

	t.Run("StartSandbox", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a sandbox without auto-start
		request := &SandboxRequest{
			Name:      "start-test-sandbox",
			TenantID:  "test-tenant",
			Type:      SandboxTypeVM,
			AutoStart: false,
		}

		sandbox, err := manager.CreateSandbox(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create sandbox: %v", err)
		}

		if sandbox.Status != SandboxStatusCreating {
			t.Errorf("Expected status %s, got %s", SandboxStatusCreating, sandbox.Status)
		}

		// Start the sandbox
		err = manager.StartSandbox(ctx, sandbox.ID)
		if err != nil {
			t.Fatalf("Failed to start sandbox: %v", err)
		}

		// Verify it's running
		running, err := manager.GetSandbox(ctx, sandbox.ID)
		if err != nil {
			t.Fatalf("Failed to get sandbox: %v", err)
		}

		if running.Status != SandboxStatusRunning {
			t.Errorf("Expected status %s, got %s", SandboxStatusRunning, running.Status)
		}

		if running.StartedAt == nil {
			t.Error("StartedAt should not be nil after starting")
		}
	})

	t.Run("StopSandbox", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create and start a sandbox
		request := &SandboxRequest{
			Name:      "stop-test-sandbox",
			TenantID:  "test-tenant",
			Type:      SandboxTypeContainer,
			AutoStart: true,
		}

		sandbox, err := manager.CreateSandbox(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create sandbox: %v", err)
		}

		// Stop the sandbox
		err = manager.StopSandbox(ctx, sandbox.ID)
		if err != nil {
			t.Fatalf("Failed to stop sandbox: %v", err)
		}

		// Verify it's stopped
		stopped, err := manager.GetSandbox(ctx, sandbox.ID)
		if err != nil {
			t.Fatalf("Failed to get sandbox: %v", err)
		}

		if stopped.Status != SandboxStatusStopped {
			t.Errorf("Expected status %s, got %s", SandboxStatusStopped, stopped.Status)
		}

		if stopped.StoppedAt == nil {
			t.Error("StoppedAt should not be nil after stopping")
		}
	})

	t.Run("DeleteSandbox", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a sandbox
		request := &SandboxRequest{
			Name:      "delete-test-sandbox",
			TenantID:  "test-tenant",
			Type:      SandboxTypeVM,
			AutoStart: false,
		}

		sandbox, err := manager.CreateSandbox(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create sandbox: %v", err)
		}

		// Delete the sandbox
		err = manager.DeleteSandbox(ctx, sandbox.ID)
		if err != nil {
			t.Fatalf("Failed to delete sandbox: %v", err)
		}

		// Verify it's gone
		_, err = manager.GetSandbox(ctx, sandbox.ID)
		if err == nil {
			t.Error("Expected error when getting deleted sandbox")
		}
	})

	t.Run("StartVerification", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a running sandbox
		sandboxRequest := &SandboxRequest{
			Name:      "verification-test-sandbox",
			TenantID:  "test-tenant",
			Type:      SandboxTypeVM,
			AutoStart: true,
		}

		sandbox, err := manager.CreateSandbox(ctx, sandboxRequest)
		if err != nil {
			t.Fatalf("Failed to create sandbox: %v", err)
		}

		// Start verification
		verificationRequest := &VerificationRequest{
			SandboxID: sandbox.ID,
			BackupID:  "test-backup-123",
			TenantID:  "test-tenant",
			Type:      VerificationTypeIntegrity,
			Config: VerificationConfig{
				Timeout: 10 * time.Minute,
				Tests: []VerificationTest{
					{
						Name:       "file_integrity",
						Type:       "integrity",
						Command:    "checksum --verify",
						Expected:   "OK",
						Timeout:    30 * time.Second,
						RetryCount: 3,
					},
				},
				Scripts: []VerificationScript{
					{
						Name:        "mount_check",
						Path:        "/scripts/mount_check.sh",
						Args:        []string{"--verbose"},
						Environment: map[string]string{"PATH": "/usr/bin:/bin"},
						RunAs:       "root",
						Timeout:     60 * time.Second,
					},
				},
				Notifications: NotificationConfig{
					OnSuccess: []string{"admin@example.com"},
					OnFailure: []string{"admin@example.com", "ops@example.com"},
				},
			},
		}

		verification, err := manager.StartVerification(ctx, verificationRequest)
		if err != nil {
			t.Fatalf("Failed to start verification: %v", err)
		}

		if verification.ID == "" {
			t.Error("Verification ID should not be empty")
		}

		if verification.SandboxID != sandbox.ID {
			t.Errorf("Expected sandbox ID %s, got %s", sandbox.ID, verification.SandboxID)
		}

		if verification.BackupID != verificationRequest.BackupID {
			t.Errorf("Expected backup ID %s, got %s", verificationRequest.BackupID, verification.BackupID)
		}

		if verification.Status != VerificationStatusPending {
			t.Errorf("Expected status %s, got %s", VerificationStatusPending, verification.Status)
		}
	})

	t.Run("GetVerification", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a running sandbox
		sandboxRequest := &SandboxRequest{
			Name:      "get-verification-sandbox",
			TenantID:  "test-tenant",
			Type:      SandboxTypeContainer,
			AutoStart: true,
		}

		sandbox, err := manager.CreateSandbox(ctx, sandboxRequest)
		if err != nil {
			t.Fatalf("Failed to create sandbox: %v", err)
		}

		// Start verification
		verificationRequest := &VerificationRequest{
			SandboxID: sandbox.ID,
			BackupID:  "test-backup-456",
			TenantID:  "test-tenant",
			Type:      VerificationTypeMount,
		}

		created, err := manager.StartVerification(ctx, verificationRequest)
		if err != nil {
			t.Fatalf("Failed to start verification: %v", err)
		}

		// Wait a moment for verification to potentially start
		time.Sleep(100 * time.Millisecond)

		// Retrieve the verification
		retrieved, err := manager.GetVerification(ctx, created.ID)
		if err != nil {
			t.Fatalf("Failed to get verification: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
		}

		if retrieved.SandboxID != created.SandboxID {
			t.Errorf("Expected sandbox ID %s, got %s", created.SandboxID, retrieved.SandboxID)
		}
	})

	t.Run("ListVerifications", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create a running sandbox
		sandboxRequest := &SandboxRequest{
			Name:      "list-verification-sandbox",
			TenantID:  "test-tenant",
			Type:      SandboxTypeVM,
			AutoStart: true,
		}

		sandbox, err := manager.CreateSandbox(ctx, sandboxRequest)
		if err != nil {
			t.Fatalf("Failed to create sandbox: %v", err)
		}

		// Create multiple verifications
		for i := 0; i < 3; i++ {
			verificationRequest := &VerificationRequest{
				SandboxID: sandbox.ID,
				BackupID:  fmt.Sprintf("test-backup-%d", i),
				TenantID:  "test-tenant",
				Type:      VerificationTypeIntegrity,
			}

			_, err := manager.StartVerification(ctx, verificationRequest)
			if err != nil {
				t.Fatalf("Failed to start verification %d: %v", i, err)
			}
		}

		// List all verifications for tenant
		verifications, err := manager.ListVerifications(ctx, &VerificationFilter{
			TenantID: "test-tenant",
		})
		if err != nil {
			t.Fatalf("Failed to list verifications: %v", err)
		}

		if len(verifications) < 3 {
			t.Errorf("Expected at least 3 verifications, got %d", len(verifications))
		}

		// Filter by type
		integrityVerifications, err := manager.ListVerifications(ctx, &VerificationFilter{
			TenantID: "test-tenant",
			Type:     VerificationTypeIntegrity,
		})
		if err != nil {
			t.Fatalf("Failed to list integrity verifications: %v", err)
		}

		if len(integrityVerifications) < 3 {
			t.Errorf("Expected at least 3 integrity verifications, got %d", len(integrityVerifications))
		}
	})

	t.Run("GetSureBackupStats", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Create some sandboxes and verifications
		for i := 0; i < 2; i++ {
			sandboxRequest := &SandboxRequest{
				Name:      fmt.Sprintf("stats-sandbox-%d", i),
				TenantID:  "test-tenant",
				Type:      SandboxTypeVM,
				AutoStart: true,
			}

			sandbox, err := manager.CreateSandbox(ctx, sandboxRequest)
			if err != nil {
				t.Fatalf("Failed to create sandbox %d: %v", i, err)
			}

			verificationRequest := &VerificationRequest{
				SandboxID: sandbox.ID,
				BackupID:  fmt.Sprintf("stats-backup-%d", i),
				TenantID:  "test-tenant",
				Type:      VerificationTypeIntegrity,
			}

			_, err = manager.StartVerification(ctx, verificationRequest)
			if err != nil {
				t.Fatalf("Failed to start verification %d: %v", i, err)
			}
		}

		// Get tenant stats
		stats, err := manager.GetSureBackupStats(ctx, "test-tenant")
		if err != nil {
			t.Fatalf("Failed to get SureBackup stats: %v", err)
		}

		if stats.TenantID != "test-tenant" {
			t.Errorf("Expected tenant ID %s, got %s", "test-tenant", stats.TenantID)
		}

		if stats.TotalSandboxes < 2 {
			t.Errorf("Expected at least 2 total sandboxes, got %d", stats.TotalSandboxes)
		}

		if stats.RunningSandboxes < 2 {
			t.Errorf("Expected at least 2 running sandboxes, got %d", stats.RunningSandboxes)
		}

		if stats.TotalVerifications < 2 {
			t.Errorf("Expected at least 2 total verifications, got %d", stats.TotalVerifications)
		}
	})

	t.Run("GetGlobalStats", func(t *testing.T) {
		ctx := context.Background()

		// Get global stats
		stats, err := manager.GetGlobalStats(ctx)
		if err != nil {
			t.Fatalf("Failed to get global stats: %v", err)
		}

		if stats.TotalTenants < 1 {
			t.Errorf("Expected at least 1 tenant, got %d", stats.TotalTenants)
		}

		if stats.TotalSandboxes < 1 {
			t.Errorf("Expected at least 1 total sandbox, got %d", stats.TotalSandboxes)
		}
	})
}
