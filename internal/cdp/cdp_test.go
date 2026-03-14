package cdp

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"novabackup/internal/deduplication"
	"novabackup/internal/multitenancy"
)

// MockDeduplicationManager for testing
type MockDeduplicationManager struct {
	chunks map[string]*deduplication.Chunk
}

func (m *MockDeduplicationManager) StoreChunk(ctx context.Context, chunk *deduplication.Chunk) error {
	m.chunks[chunk.Hash] = chunk
	return nil
}

func (m *MockDeduplicationManager) GetChunk(ctx context.Context, hash string) (*deduplication.Chunk, error) {
	if chunk, exists := m.chunks[hash]; exists {
		return chunk, nil
	}
	return nil, nil
}

func (m *MockDeduplicationManager) DeleteChunk(ctx context.Context, hash string) error {
	delete(m.chunks, hash)
	return nil
}

func (m *MockDeduplicationManager) ChunkExists(ctx context.Context, hash string) (bool, error) {
	_, exists := m.chunks[hash]
	return exists, nil
}

func (m *MockDeduplicationManager) FindDuplicateChunks(ctx context.Context, chunkHashes []string) (map[string]bool, error) {
	result := make(map[string]bool)
	for _, hash := range chunkHashes {
		result[hash] = m.chunks[hash] != nil
	}
	return result, nil
}

func (m *MockDeduplicationManager) CalculateHash(ctx context.Context, data []byte) (string, error) {
	return "mock_hash", nil
}

func (m *MockDeduplicationManager) GetDeduplicationStats(ctx context.Context, tenantID string) (*deduplication.DeduplicationStats, error) {
	return &deduplication.DeduplicationStats{TenantID: tenantID}, nil
}

func (m *MockDeduplicationManager) GetGlobalStats(ctx context.Context) (*deduplication.GlobalDeduplicationStats, error) {
	return &deduplication.GlobalDeduplicationStats{}, nil
}

func (m *MockDeduplicationManager) CleanupUnusedChunks(ctx context.Context, olderThan time.Duration) error {
	return nil
}

func (m *MockDeduplicationManager) OptimizeStorage(ctx context.Context) error {
	return nil
}

func (m *MockDeduplicationManager) UpdateQuota(ctx context.Context, tenantID string, quotas multitenancy.TenantQuotas) error {
	return nil
}

// MockTenantManager for testing
type MockTenantManager struct{}

func (m *MockTenantManager) CreateTenant(ctx context.Context, tenant *multitenancy.Tenant) error {
	return nil
}

func (m *MockTenantManager) GetTenant(ctx context.Context, tenantID string) (*multitenancy.Tenant, error) {
	return &multitenancy.Tenant{ID: tenantID}, nil
}

func (m *MockTenantManager) UpdateTenant(ctx context.Context, tenant *multitenancy.Tenant) error {
	return nil
}

func (m *MockTenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	return nil
}

func (m *MockTenantManager) ListTenants(ctx context.Context) ([]multitenancy.Tenant, error) {
	return nil, nil
}

func (m *MockTenantManager) CheckQuota(ctx context.Context, tenantID string, quotaType multitenancy.QuotaType, amount int64) (bool, error) {
	return true, nil
}

func (m *MockTenantManager) GetQuotaUsage(ctx context.Context, tenantID string) (*multitenancy.QuotaUsage, error) {
	return &multitenancy.QuotaUsage{}, nil
}

func (m *MockTenantManager) UpdateQuota(ctx context.Context, tenantID string, quotas multitenancy.TenantQuotas) error {
	return nil
}

func (m *MockTenantManager) GetTenantResources(ctx context.Context, tenantID string) (*multitenancy.TenantResources, error) {
	return &multitenancy.TenantResources{}, nil
}

func (m *MockTenantManager) AssignResource(ctx context.Context, tenantID string, resourceType string, resourceID string) error {
	return nil
}

func (m *MockTenantManager) UnassignResource(ctx context.Context, tenantID string, resourceType string, resourceID string) error {
	return nil
}

func (m *MockTenantManager) WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, multitenancy.TenantContextKey("tenant_id"), tenantID)
}

func (m *MockTenantManager) GetTenantFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value(multitenancy.TenantContextKey("tenant_id")).(string); ok {
		return tenantID
	}
	return ""
}

func (m *MockTenantManager) ValidateTenantAccess(ctx context.Context, tenantID string) error {
	return nil
}

// TestCDPEngine tests the CDP engine
func TestCDPEngine(t *testing.T) {
	config := NewCDPConfig()
	mockTenantMgr := &MockTenantManager{}
	mockDedupeMgr := &MockDeduplicationManager{chunks: make(map[string]*deduplication.Chunk)}

	cdpEngine := NewInMemoryCDPEngine(config, mockTenantMgr, mockDedupeMgr)

	t.Run("StartStopWatching", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Test starting
		paths := []string{"/test/path", "/another/path"}
		err := cdpEngine.StartWatching(ctx, paths)
		if err != nil {
			t.Fatalf("Failed to start watching: %v", err)
		}

		if !cdpEngine.IsWatching() {
			t.Error("Engine should be watching")
		}

		// Test protected paths
		protectedPaths, err := cdpEngine.GetProtectedPaths(ctx)
		if err != nil {
			t.Fatalf("Failed to get protected paths: %v", err)
		}

		if len(protectedPaths) != 2 {
			t.Errorf("Expected 2 protected paths, got %d", len(protectedPaths))
		}

		// Test stopping
		err = cdpEngine.StopWatching(ctx)
		if err != nil {
			t.Fatalf("Failed to stop watching: %v", err)
		}

		if cdpEngine.IsWatching() {
			t.Error("Engine should not be watching")
		}
	})

	t.Run("ProcessEvent", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Enable protection for a path first
		err := cdpEngine.EnableProtection(ctx, "/test/path")
		if err != nil {
			t.Fatalf("Failed to enable protection: %v", err)
		}

		// Process a create event - use same path as protected
		event := &FileEvent{
			ID:       "test-event-1",
			Type:     EventCreate,
			Path:     "/test/path",
			Size:     1024,
			ModTime:  time.Now(),
			Checksum: "test-checksum",
			Metadata: map[string]string{
				"test": "true",
			},
		}

		err = cdpEngine.ProcessEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to process event: %v", err)
		}

		// Check if event was processed
		if event.ProcessedAt == nil {
			t.Error("Event should have been processed")
		}

		// Get events
		events, err := cdpEngine.GetEvents(ctx, 10)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}

		if len(events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(events))
		}

		if events[0].ID != "test-event-1" {
			t.Errorf("Expected event ID 'test-event-1', got '%s'", events[0].ID)
		}
	})

	t.Run("RecoveryPoints", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Enable protection
		err := cdpEngine.EnableProtection(ctx, "/test/recovery")
		if err != nil {
			t.Fatalf("Failed to enable protection: %v", err)
		}

		// Process multiple events to create recovery points - use same path as protected
		events := []*FileEvent{
			{
				ID:       "recovery-event-1",
				Type:     EventCreate,
				Path:     "/test/recovery",
				Size:     512,
				ModTime:  time.Now().Add(-2 * time.Hour),
				Checksum: "checksum-1",
			},
			{
				ID:       "recovery-event-2",
				Type:     EventModify,
				Path:     "/test/recovery",
				Size:     1024,
				ModTime:  time.Now().Add(-1 * time.Hour),
				Checksum: "checksum-2",
			},
			{
				ID:       "recovery-event-3",
				Type:     EventModify,
				Path:     "/test/recovery",
				Size:     1536,
				ModTime:  time.Now(),
				Checksum: "checksum-3",
			},
		}

		for _, event := range events {
			err = cdpEngine.ProcessEvent(ctx, event)
			if err != nil {
				t.Fatalf("Failed to process event %s: %v", event.ID, err)
			}
		}

		// Get recovery points
		since := time.Now().Add(-3 * time.Hour)
		points, err := cdpEngine.GetRecoveryPoints(ctx, "/test/recovery", since)
		if err != nil {
			t.Fatalf("Failed to get recovery points: %v", err)
		}

		if len(points) != 3 {
			t.Errorf("Expected 3 recovery points, got %d", len(points))
		}

		// Test restore to point
		err = cdpEngine.RestoreToPoint(ctx, "/test/recovery", points[1])
		if err != nil {
			t.Fatalf("Failed to restore to point: %v", err)
		}
	})

	t.Run("ProtectionManagement", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Test enable protection
		err := cdpEngine.EnableProtection(ctx, "/test/protection")
		if err != nil {
			t.Fatalf("Failed to enable protection: %v", err)
		}

		protectedPaths, err := cdpEngine.GetProtectedPaths(ctx)
		if err != nil {
			t.Fatalf("Failed to get protected paths: %v", err)
		}

		// Check for exact match or absolute path
		found := false
		for _, path := range protectedPaths {
			// Convert to absolute path for comparison
			absPath, _ := filepath.Abs("/test/protection")
			if path == "/test/protection" || path == absPath {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Protected path should be in the list. Got paths: %v", protectedPaths)
		}

		// Test disable protection
		err = cdpEngine.DisableProtection(ctx, "/test/protection")
		if err != nil {
			t.Fatalf("Failed to disable protection: %v", err)
		}

		protectedPaths, err = cdpEngine.GetProtectedPaths(ctx)
		if err != nil {
			t.Fatalf("Failed to get protected paths: %v", err)
		}

		found = false
		for _, path := range protectedPaths {
			absPath, _ := filepath.Abs("/test/protection")
			if path == "/test/protection" || path == absPath {
				found = true
				break
			}
		}

		if found {
			t.Error("Protected path should not be in the list after disabling")
		}
	})

	t.Run("Statistics", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Enable protection and process some events
		err := cdpEngine.EnableProtection(ctx, "/test/stats")
		if err != nil {
			t.Fatalf("Failed to enable protection: %v", err)
		}

		// Process events - use same path as protected
		for i := 0; i < 5; i++ {
			event := &FileEvent{
				ID:       fmt.Sprintf("stats-event-%d", i),
				Type:     EventModify,
				Path:     "/test/stats",
				Size:     int64(1024 * (i + 1)),
				ModTime:  time.Now().Add(time.Duration(i) * time.Minute),
				Checksum: fmt.Sprintf("checksum-%d", i),
			}

			err = cdpEngine.ProcessEvent(ctx, event)
			if err != nil {
				t.Fatalf("Failed to process event %d: %v", i, err)
			}
		}

		// Get CDP stats
		stats, err := cdpEngine.GetCDPStats(ctx)
		if err != nil {
			t.Fatalf("Failed to get CDP stats: %v", err)
		}

		if stats.TotalEvents != 5 {
			t.Errorf("Expected 5 total events, got %d", stats.TotalEvents)
		}

		if stats.ProcessedEvents != 5 {
			t.Errorf("Expected 5 processed events, got %d", stats.ProcessedEvents)
		}

		if stats.ProtectedPaths < 1 {
			t.Errorf("Expected at least 1 protected path, got %d", stats.ProtectedPaths)
		}

		// Get RPO stats
		rpoStats, err := cdpEngine.GetRPOStats(ctx, "/test/stats")
		if err != nil {
			t.Fatalf("Failed to get RPO stats: %v", err)
		}

		if rpoStats.Path != "/test/stats" {
			t.Errorf("Expected path '/test/stats', got '%s'", rpoStats.Path)
		}

		if rpoStats.RecoveryPoints != 5 {
			t.Errorf("Expected 5 recovery points, got %d", rpoStats.RecoveryPoints)
		}

		if rpoStats.ProtectedSize != 15360 { // 1024 + 2048 + 3072 + 4096 + 5120
			t.Errorf("Expected protected size 15360, got %d", rpoStats.ProtectedSize)
		}
	})
}

// TestCDPConfig tests CDP configuration
func TestCDPConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := NewCDPConfig()

		if !config.Enabled {
			t.Error("CDP should be enabled by default")
		}

		if config.MaxEventsPerSecond != 1000 {
			t.Errorf("Expected max events per second 1000, got %d", config.MaxEventsPerSecond)
		}

		if config.RPOTarget != 1*time.Minute {
			t.Errorf("Expected RPO target 1 minute, got %v", config.RPOTarget)
		}

		if config.MaxRecoveryPoints != 100 {
			t.Errorf("Expected max recovery points 100, got %d", config.MaxRecoveryPoints)
		}

		if len(config.ExcludePatterns) == 0 {
			t.Error("Should have default exclude patterns")
		}

		if len(config.IncludePatterns) == 0 {
			t.Error("Should have default include patterns")
		}
	})
}

// TestEventProcessor tests the event processor
func TestEventProcessor(t *testing.T) {
	config := NewCDPConfig()
	mockDedupeMgr := &MockDeduplicationManager{chunks: make(map[string]*deduplication.Chunk)}

	processor := NewEventProcessor(config, mockDedupeMgr)

	if processor == nil {
		t.Error("Event processor should not be nil")
	}
	_ = config // Use config
}

// TestFileWatcher tests the file watcher
func TestFileWatcher(t *testing.T) {
	// Skip file watcher test - requires real file system
	t.Skip("File watcher test requires real file system")
}
