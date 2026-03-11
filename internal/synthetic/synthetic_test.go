package synthetic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"novabackup/internal/deduplication"
	"novabackup/internal/multitenancy"
	"testing"
	"time"
)

// MockTenantManager for testing
type MockTenantManager struct{}

func (m *MockTenantManager) GetTenantFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return "default-tenant"
}

func (m *MockTenantManager) WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, "tenant_id", tenantID)
}

func (m *MockTenantManager) ValidateTenant(tenantID string) error {
	return nil
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

// MockDeduplicationManager for testing
type MockDeduplicationManager struct{}

func (m *MockDeduplicationManager) IsDuplicate(ctx context.Context, data []byte) (bool, string, error) {
	return false, "", nil
}

func (m *MockDeduplicationManager) StoreHash(ctx context.Context, hash string, dataID string) error {
	return nil
}

func (m *MockDeduplicationManager) GetHash(ctx context.Context, hash string) (string, error) {
	return "", nil
}

func (m *MockDeduplicationManager) CalculateDeduplicationRatio(originalSize, compressedSize int64) float64 {
	if originalSize == 0 {
		return 0
	}
	return float64(compressedSize) / float64(originalSize)
}

func (m *MockDeduplicationManager) CalculateHash(ctx context.Context, data []byte) (string, error) {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (m *MockDeduplicationManager) ChunkExists(ctx context.Context, hash string) (bool, error) {
	return false, nil
}

func (m *MockDeduplicationManager) StoreChunk(ctx context.Context, chunk *deduplication.Chunk) error {
	return nil
}

func (m *MockDeduplicationManager) GetChunk(ctx context.Context, hash string) (*deduplication.Chunk, error) {
	return nil, nil
}

func (m *MockDeduplicationManager) DeleteChunk(ctx context.Context, hash string) error {
	return nil
}

func (m *MockDeduplicationManager) FindDuplicateChunks(ctx context.Context, chunkHashes []string) (map[string]bool, error) {
	return make(map[string]bool), nil
}

func (m *MockDeduplicationManager) GetDeduplicationStats(ctx context.Context, tenantID string) (*deduplication.DeduplicationStats, error) {
	return &deduplication.DeduplicationStats{
		TotalChunks:        100,
		TotalSize:          1000000,
		DeduplicatedSize:   500000,
		DeduplicationRatio: 0.5,
		UniqueChunks:       80,
		SharedChunks:       20,
	}, nil
}

func (m *MockDeduplicationManager) GetGlobalStats(ctx context.Context) (*deduplication.GlobalDeduplicationStats, error) {
	return &deduplication.GlobalDeduplicationStats{
		TotalTenants:       100,
		TotalChunks:        1000,
		TotalSize:          10000000,
		DeduplicatedSize:   5000000,
		DeduplicationRatio: 0.5,
		UniqueChunks:       800,
		SharedChunks:       200,
	}, nil
}

func (m *MockDeduplicationManager) CleanupUnusedChunks(ctx context.Context, olderThan time.Duration) error {
	return nil
}

func (m *MockDeduplicationManager) OptimizeStorage(ctx context.Context) error {
	return nil
}

// MockStorageManager for synthetic backup testing
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
		"total_space": int64(2000000000), // 2GB
		"used_space":  int64(1000000000), // 1GB
	}, nil
}

// MockDeduplicationManager for testing synthetic backup management
func TestSyntheticBackupManager(t *testing.T) {
	mockTenantMgr := &MockTenantManager{}
	mockDedupeMgr := &MockDeduplicationManager{}

	t.Run("CreateSyntheticBackup", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		request := &SyntheticBackupRequest{
			SourceRepo:    "primary-repo",
			TargetRepo:    "secondary-repo",
			BackupType:    "synthetic",
			Compression:   true,
			RetentionDays: 30,
			Settings: map[string]interface{}{
				"compression_level": "high",
			},
			Metadata: map[string]string{
				"description": "Test synthetic backup",
				"environment": "test",
			},
		}

		manager := NewInMemorySyntheticBackupManager(mockTenantMgr, mockDedupeMgr)

		result, err := manager.CreateSyntheticBackup(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create synthetic backup: %v", err)
		}

		if result.BackupID == "" {
			t.Error("Backup ID should not be empty")
		}

		if result.Success != true {
			t.Error("Expected success to be true")
		}
	})

	t.Run("GetSyntheticBackup", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemorySyntheticBackupManager(mockTenantMgr, mockDedupeMgr)

		// First create a backup
		request := &SyntheticBackupRequest{
			SourceRepo:  "primary-repo",
			TargetRepo:  "secondary-repo",
			BackupType:  "synthetic",
			Compression: true,
			TenantID:    "test-tenant",
		}

		result, err := manager.CreateSyntheticBackup(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create synthetic backup: %v", err)
		}

		// Retrieve the backup
		retrievedBackup, err := manager.GetSyntheticBackup(ctx, result.BackupID)
		if err != nil {
			t.Fatalf("Failed to get synthetic backup: %v", err)
		}

		if retrievedBackup.ID != result.BackupID {
			t.Errorf("Expected backup ID %s, got %s", result.BackupID, retrievedBackup.ID)
		}

		if retrievedBackup.SourceRepo != request.SourceRepo {
			t.Errorf("Expected source repo %s, got %s", request.SourceRepo, retrievedBackup.SourceRepo)
		}
	})

	t.Run("ListSyntheticBackups", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemorySyntheticBackupManager(mockTenantMgr, mockDedupeMgr)

		// Create multiple backups
		for i := 0; i < 5; i++ {
			request := &SyntheticBackupRequest{
				SourceRepo: "primary-repo",
				TargetRepo: "secondary-repo",
				BackupType: "incremental",
				TenantID:   "test-tenant",
			}

			_, err := manager.CreateSyntheticBackup(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create synthetic backup %d: %v", i, err)
			}
		}

		// Create a synthetic backup
		syntheticRequest := &SyntheticBackupRequest{
			SourceRepo:  "primary-repo",
			TargetRepo:  "secondary-repo",
			BackupType:  "synthetic",
			Compression: true,
			TenantID:    "test-tenant",
		}

		_, err := manager.CreateSyntheticBackup(ctx, syntheticRequest)
		if err != nil {
			t.Fatalf("Failed to create synthetic backup: %v", err)
		}

		// Test listing without filter
		backups, err := manager.ListSyntheticBackups(ctx, &SyntheticBackupFilter{})
		if err != nil {
			t.Fatalf("Failed to list synthetic backups: %v", err)
		}

		if len(backups) != 6 {
			t.Errorf("Expected 6 backups, got %d", len(backups))
		}

		// Test listing with filter
		filter := &SyntheticBackupFilter{
			BackupType: "synthetic",
			Limit:      3,
		}

		filteredBackups, err := manager.ListSyntheticBackups(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to list filtered synthetic backups: %v", err)
		}

		if len(filteredBackups) != 1 {
			t.Errorf("Expected 1 filtered backup, got %d", len(filteredBackups))
		}

		// Verify tenant isolation
		otherTenantCtx := mockTenantMgr.WithTenant(context.Background(), "other-tenant")
		otherBackups, err := manager.ListSyntheticBackups(otherTenantCtx, &SyntheticBackupFilter{})
		if err != nil {
			t.Fatalf("Failed to list synthetic backups for other tenant: %v", err)
		}

		if len(otherBackups) != 0 {
			t.Errorf("Expected no backups for other tenant, got %d", len(otherBackups))
		}
	})

	t.Run("MergeIncrementals", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemorySyntheticBackupManager(mockTenantMgr, mockDedupeMgr)

		// Create incremental backups to merge
		var incrementalBackupIDs []string
		for i := 0; i < 3; i++ {
			request := &SyntheticBackupRequest{
				SourceRepo: "primary-repo",
				TargetRepo: "secondary-repo",
				BackupType: "incremental",
				TenantID:   "test-tenant",
			}

			result, err := manager.CreateSyntheticBackup(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create incremental backup %d: %v", i, err)
			}

			incrementalBackupIDs = append(incrementalBackupIDs, result.BackupID)
		}

		// Merge incrementals
		mergeRequest := &MergeRequest{
			ChainID:     "test-chain",
			Compression: true,
			TenantID:    "test-tenant",
		}

		result, err := manager.MergeIncrementals(ctx, mergeRequest)
		if err != nil {
			t.Fatalf("Failed to merge incrementals: %v", err)
		}

		if !result.Success {
			t.Errorf("Expected merge to succeed")
		}

		if len(result.MergedBackups) != 3 {
			t.Errorf("Expected 3 merged backups, got %d", len(result.MergedBackups))
		}

		if result.BytesReduced <= 0 {
			t.Errorf("Expected bytes reduced to be positive")
		}
	})

	t.Run("GetBackupChain", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemorySyntheticBackupManager(mockTenantMgr, mockDedupeMgr)

		// First create a chain
		// Create incremental backups to merge
		for i := 0; i < 3; i++ {
			request := &SyntheticBackupRequest{
				SourceRepo: "primary-repo",
				TargetRepo: "secondary-repo",
				BackupType: "incremental",
				TenantID:   "test-tenant",
			}
			_, err := manager.CreateSyntheticBackup(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create incremental backup %d: %v", i, err)
			}
		}

		// Get the chain
		chain, err := manager.GetBackupChain(ctx, &ChainFilter{ChainID: "test-chain"})
		if err != nil {
			t.Fatalf("Failed to get backup chain: %v", err)
		}

		if chain.ID != "test-chain" {
			t.Errorf("Expected chain ID %s, got %s", "test-chain", chain.ID)
		}

		if len(chain.Backups) != 3 {
			t.Errorf("Expected 3 backups in chain, got %d", len(chain.Backups))
		}
	})

	t.Run("GetSyntheticStats", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemorySyntheticBackupManager(mockTenantMgr, mockDedupeMgr)

		// Create some backups to generate statistics
		for i := 0; i < 5; i++ {
			request := &SyntheticBackupRequest{
				SourceRepo:  "primary-repo",
				TargetRepo:  "secondary-repo",
				BackupType:  "synthetic",
				Compression: i%2 == 0, // Compress every other backup
				TenantID:    "test-tenant",
			}

			_, err := manager.CreateSyntheticBackup(ctx, request)
			if err != nil {
				t.Fatalf("Failed to create synthetic backup %d: %v", i, err)
			}
		}

		// Create a synthetic backup
		syntheticRequest := &SyntheticBackupRequest{
			SourceRepo:  "primary-repo",
			TargetRepo:  "secondary-repo",
			BackupType:  "synthetic",
			Compression: true,
			TenantID:    "test-tenant",
		}

		_, err := manager.CreateSyntheticBackup(ctx, syntheticRequest)
		if err != nil {
			t.Fatalf("Failed to create synthetic backup: %v", err)
		}

		// Get statistics
		stats, err := manager.GetSyntheticStats(ctx)
		if err != nil {
			t.Fatalf("Failed to get synthetic stats: %v", err)
		}

		if stats.TotalBackups != 6 {
			t.Errorf("Expected 6 total backups, got %d", stats.TotalBackups)
		}

		if stats.TotalChains != 1 {
			t.Errorf("Expected 1 total chain, got %d", stats.TotalChains)
		}

		if stats.ActiveBackups < 0 {
			t.Errorf("Expected active backups to be positive")
		}

		if stats.TotalSize <= 0 {
			t.Errorf("Expected total size to be positive")
		}

		if stats.AverageCompressionRatio <= 0 {
			t.Errorf("Expected compression ratio to be positive")
		}
	})

	t.Run("DeleteSyntheticBackup", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")
		manager := NewInMemorySyntheticBackupManager(mockTenantMgr, mockDedupeMgr)

		// Create a backup
		request := &SyntheticBackupRequest{
			SourceRepo: "primary-repo",
			TargetRepo: "secondary-repo",
			BackupType: "synthetic",
			TenantID:   "test-tenant",
		}

		result, err := manager.CreateSyntheticBackup(ctx, request)
		if err != nil {
			t.Fatalf("Failed to create synthetic backup: %v", err)
		}

		// Delete the backup
		err = manager.DeleteSyntheticBackup(ctx, result.BackupID)
		if err != nil {
			t.Fatalf("Failed to delete synthetic backup: %v", err)
		}

		// Verify backup is deleted
		_, err = manager.GetSyntheticBackup(ctx, result.BackupID)
		if err == nil {
			t.Error("Expected error when getting deleted synthetic backup")
		}
	})
}
