package wan

import (
	"context"
	"fmt"
	"testing"
	"time"

	"novabackup/internal/deduplication"
	"novabackup/internal/multitenancy"
)

// MockTenantManager for WAN testing
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

// MockDeduplicationManager for WAN testing
type MockDeduplicationManager struct {
	chunks map[string]*deduplication.Chunk
}

func (m *MockDeduplicationManager) StoreChunk(ctx context.Context, chunk *deduplication.Chunk) error {
	if m.chunks == nil {
		m.chunks = make(map[string]*deduplication.Chunk)
	}
	m.chunks[chunk.Hash] = chunk
	return nil
}

func (m *MockDeduplicationManager) GetChunk(ctx context.Context, hash string) (*deduplication.Chunk, error) {
	if m.chunks == nil {
		return nil, nil
	}
	if chunk, exists := m.chunks[hash]; exists {
		return chunk, nil
	}
	return nil, nil
}

func (m *MockDeduplicationManager) DeleteChunk(ctx context.Context, hash string) error {
	if m.chunks != nil {
		delete(m.chunks, hash)
	}
	return nil
}

func (m *MockDeduplicationManager) ChunkExists(ctx context.Context, hash string) (bool, error) {
	if m.chunks == nil {
		return false, nil
	}
	_, exists := m.chunks[hash]
	return exists, nil
}

func (m *MockDeduplicationManager) FindDuplicateChunks(ctx context.Context, chunkHashes []string) (map[string]bool, error) {
	result := make(map[string]bool)
	if m.chunks == nil {
		return result, nil
	}
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

// TestWANAccelerator tests the WAN accelerator
func TestWANAccelerator(t *testing.T) {
	config := NewWANConfig()
	mockTenantMgr := &MockTenantManager{}
	mockDedupeMgr := &MockDeduplicationManager{}

	accelerator := NewInMemoryWANAccelerator(config, mockTenantMgr, mockDedupeMgr)

	t.Run("OptimizeTransfer", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(mockTenantMgr.WithTenant(context.Background(), "test-tenant"), 10*time.Second)
		defer cancel()

		// Create a transfer request
		transfer := &TransferRequest{
			ID:          "test-transfer-1",
			Source:      "source-server",
			Destination: "dest-server",
			Data:        []byte("This is test data for WAN optimization"),
			Size:        int64(len("This is test data for WAN optimization")),
			Priority:    PriorityNormal,
			Metadata: map[string]string{
				"test": "true",
			},
			CreatedAt: time.Now(),
		}

		// Optimize the transfer
		result, err := accelerator.OptimizeTransfer(ctx, transfer)
		if err != nil {
			t.Fatalf("Failed to optimize transfer: %v", err)
		}

		// Verify results
		if result.RequestID != transfer.ID {
			t.Errorf("Expected request ID %s, got %s", transfer.ID, result.RequestID)
		}

		if result.OriginalSize != transfer.Size {
			t.Errorf("Expected original size %d, got %d", transfer.Size, result.OriginalSize)
		}

		if result.CompressedSize >= result.OriginalSize {
			t.Error("Compressed size should be smaller than original size")
		}

		if result.CompressionRatio <= 1.0 {
			t.Error("Compression ratio should be greater than 1.0")
		}

		if len(result.Optimizations) == 0 {
			t.Error("Should have at least one optimization applied")
		}
	})

	t.Run("Caching", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(mockTenantMgr.WithTenant(context.Background(), "test-tenant"), 5*time.Second)
		defer cancel()

		// Test cache miss
		_, err := accelerator.GetCachedData(ctx, "non-existent-key")
		if err == nil {
			t.Error("Expected error for cache miss")
		}

		// Test cache store
		testData := []byte("cached test data")
		err = accelerator.CacheData(ctx, "test-key", testData, 1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to cache data: %v", err)
		}

		// Test cache hit
		cachedData, err := accelerator.GetCachedData(ctx, "test-key")
		if err != nil {
			t.Fatalf("Failed to get cached data: %v", err)
		}

		if string(cachedData) != string(testData) {
			t.Error("Cached data should match original data")
		}

		// Test cache invalidation
		err = accelerator.InvalidateCache(ctx, "test-key")
		if err != nil {
			t.Fatalf("Failed to invalidate cache: %v", err)
		}

		// Verify cache is invalidated
		_, err = accelerator.GetCachedData(ctx, "test-key")
		if err == nil {
			t.Error("Expected error after cache invalidation")
		}
	})

	t.Run("Compression", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Test compression
		testData := []byte("This is test data for compression testing. It should compress reasonably well.")
		compressed, err := accelerator.CompressData(ctx, testData, CompressionGzip)
		if err != nil {
			t.Fatalf("Failed to compress data: %v", err)
		}

		if compressed.Algorithm != CompressionGzip {
			t.Errorf("Expected algorithm %s, got %s", CompressionGzip, compressed.Algorithm)
		}

		if compressed.OriginalSize != int64(len(testData)) {
			t.Errorf("Expected original size %d, got %d", len(testData), compressed.OriginalSize)
		}

		if compressed.CompressedSize >= compressed.OriginalSize {
			t.Logf("Compressed size %d is not smaller than original size %d (this can happen with small data)", compressed.CompressedSize, compressed.OriginalSize)
		}

		// Test decompression
		decompressed, err := accelerator.DecompressData(ctx, compressed)
		if err != nil {
			t.Fatalf("Failed to decompress data: %v", err)
		}

		if string(decompressed) != string(testData) {
			t.Error("Decompressed data should match original data")
		}
	})

	t.Run("TrafficShaping", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Get current bandwidth
		bandwidth, err := accelerator.GetCurrentBandwidth(ctx)
		if err != nil {
			t.Fatalf("Failed to get current bandwidth: %v", err)
		}

		if bandwidth.AvailableBandwidth <= 0 {
			t.Error("Available bandwidth should be positive")
		}

		// Apply new traffic shaping config
		newConfig := &TrafficShapingConfig{
			MaxBandwidth:      5000000, // 5MB/s
			BurstSize:         500000,  // 500KB
			PriorityLevels:    []Priority{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow},
			ThrottlingEnabled: true,
			QualityOfService: QoSProfile{
				Name:         "test-profile",
				MinBandwidth: 1000000, // 1MB/s
				MaxBandwidth: 5000000, // 5MB/s
				MaxLatency:   100 * time.Millisecond,
				PacketLoss:   0.01,
				Jitter:       10 * time.Millisecond,
			},
		}

		err = accelerator.ApplyTrafficShaping(ctx, newConfig)
		if err != nil {
			t.Fatalf("Failed to apply traffic shaping: %v", err)
		}

		// Verify config was applied
		config, err := accelerator.GetConfiguration(ctx)
		if err != nil {
			t.Fatalf("Failed to get configuration: %v", err)
		}

		if config.TrafficShaping.MaxBandwidth != 5000000 {
			t.Errorf("Expected max bandwidth 5000000, got %d", config.TrafficShaping.MaxBandwidth)
		}
	})

	t.Run("Statistics", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(mockTenantMgr.WithTenant(context.Background(), "test-tenant"), 5*time.Second)
		defer cancel()

		// Perform some transfers to generate statistics
		for i := 0; i < 5; i++ {
			transfer := &TransferRequest{
				ID:          fmt.Sprintf("stats-transfer-%d", i),
				Source:      "source-server",
				Destination: "dest-server",
				Data:        []byte(fmt.Sprintf("Test data %d for statistics", i)),
				Size:        int64(len(fmt.Sprintf("Test data %d for statistics", i))),
				Priority:    PriorityNormal,
				CreatedAt:   time.Now(),
			}

			_, err := accelerator.OptimizeTransfer(ctx, transfer)
			if err != nil {
				t.Fatalf("Failed to optimize transfer %d: %v", i, err)
			}
		}

		// Get optimization statistics
		stats, err := accelerator.GetOptimizationStats(ctx)
		if err != nil {
			t.Fatalf("Failed to get optimization stats: %v", err)
		}

		if stats.TotalTransfers != 5 {
			t.Errorf("Expected 5 total transfers, got %d", stats.TotalTransfers)
		}

		if stats.TotalDataTransferred <= 0 {
			t.Error("Total data transferred should be positive")
		}

		if stats.AverageCompression <= 1.0 {
			t.Error("Average compression should be greater than 1.0")
		}

		if stats.LastUpdated.IsZero() {
			t.Error("Last updated time should not be zero")
		}
	})
}

// TestWANConfig tests WAN configuration
func TestWANConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := NewWANConfig()

		if !config.Enabled {
			t.Error("WAN should be enabled by default")
		}

		if !config.CacheEnabled {
			t.Error("Cache should be enabled by default")
		}

		if !config.CompressionEnabled {
			t.Error("Compression should be enabled by default")
		}

		if !config.TrafficShapingEnabled {
			t.Error("Traffic shaping should be enabled by default")
		}

		if config.DefaultAlgorithm != CompressionLZ4 {
			t.Errorf("Expected default algorithm %s, got %s", CompressionLZ4, config.DefaultAlgorithm)
		}

		if config.CacheSize != 1000 {
			t.Errorf("Expected cache size 1000, got %d", config.CacheSize)
		}

		if config.CacheTTL != 1*time.Hour {
			t.Errorf("Expected cache TTL 1 hour, got %v", config.CacheTTL)
		}

		if config.TrafficShaping == nil {
			t.Error("Traffic shaping config should not be nil")
		}

		if len(config.CompressionProfiles) == 0 {
			t.Error("Compression profiles should not be empty")
		}

		if len(config.QoSProfiles) == 0 {
			t.Error("QoS profiles should not be empty")
		}
	})
}

// TestBandwidthLimiter tests bandwidth limiting
func TestBandwidthLimiter(t *testing.T) {
	t.Run("BasicLimiter", func(t *testing.T) {
		config := &TrafficShapingConfig{
			MaxBandwidth: 1000000, // 1MB/s
		}

		limiter := NewBandwidthLimiter(config)
		if limiter == nil {
			t.Error("Bandwidth limiter should not be nil")
		}

		// Test bandwidth reservation
		err := limiter.ReserveBandwidth(500000, PriorityNormal)
		if err != nil {
			t.Fatalf("Failed to reserve bandwidth: %v", err)
		}

		if limiter.available != 500000 {
			t.Errorf("Expected available bandwidth 500000, got %d", limiter.available)
		}

		if limiter.current != 500000 {
			t.Errorf("Expected current bandwidth 500000, got %d", limiter.current)
		}

		// Test bandwidth release
		limiter.ReleaseBandwidth(500000)

		if limiter.available != 1000000 {
			t.Errorf("Expected available bandwidth 1000000, got %d", limiter.available)
		}

		if limiter.current != 0 {
			t.Errorf("Expected current bandwidth 0, got %d", limiter.current)
		}

		// Test insufficient bandwidth
		err = limiter.ReserveBandwidth(2000000, PriorityNormal)
		if err == nil {
			t.Error("Expected error for insufficient bandwidth")
		}
	})
}

// TestCompressor tests data compression
func TestCompressor(t *testing.T) {
	t.Run("BasicCompression", func(t *testing.T) {
		profiles := map[Priority]CompressionAlgorithm{
			PriorityHigh:   CompressionGzip,
			PriorityNormal: CompressionGzip,
			PriorityLow:    CompressionGzip,
		}

		compressor := NewCompressor(profiles)
		if compressor == nil {
			t.Error("Compressor should not be nil")
		}

		// Test compression
		testData := []byte("This is test data for compression testing. It should compress reasonably well.")
		compressed, err := compressor.Compress(testData, CompressionGzip)
		if err != nil {
			t.Fatalf("Failed to compress data: %v", err)
		}

		if compressed.Algorithm != CompressionGzip {
			t.Errorf("Expected algorithm %s, got %s", CompressionGzip, compressed.Algorithm)
		}

		if compressed.OriginalSize != int64(len(testData)) {
			t.Errorf("Expected original size %d, got %d", len(testData), compressed.OriginalSize)
		}

		// Test decompression
		decompressed, err := compressor.Decompress(compressed)
		if err != nil {
			t.Fatalf("Failed to decompress data: %v", err)
		}

		if string(decompressed) != string(testData) {
			t.Error("Decompressed data should match original data")
		}

		// Test no compression
		noCompress, err := compressor.Compress(testData, CompressionNone)
		if err != nil {
			t.Fatalf("Failed to compress with no compression: %v", err)
		}

		if noCompress.Algorithm != CompressionNone {
			t.Errorf("Expected algorithm %s, got %s", CompressionNone, noCompress.Algorithm)
		}

		if noCompress.CompressedSize != noCompress.OriginalSize {
			t.Error("Compressed size should equal original size for no compression")
		}
	})
}
