package deduplication

import (
	"context"
	"testing"

	"novabackup/internal/multitenancy"
)

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

// TestDeduplicationManager tests the deduplication manager
func TestDeduplicationManager(t *testing.T) {
	config := NewDeduplicationConfig()
	// Adjust min chunk size for testing
	config.MinChunkSize = 1
	config.MaxChunkSize = 1024 * 1024 // 1MB

	mockTenantMgr := &MockTenantManager{}
	dm := NewInMemoryDeduplicationManager(config, mockTenantMgr)

	t.Run("CalculateHash", func(t *testing.T) {
		data := []byte("test data for hashing")
		ctx := context.Background()

		hash, err := dm.CalculateHash(ctx, data)
		if err != nil {
			t.Fatalf("Failed to calculate hash: %v", err)
		}

		if hash == "" {
			t.Error("Hash should not be empty")
		}

		// Verify hash is consistent
		hash2, err := dm.CalculateHash(ctx, data)
		if err != nil {
			t.Fatalf("Failed to calculate hash: %v", err)
		}

		if hash != hash2 {
			t.Error("Hash should be consistent for same data")
		}
	})

	t.Run("StoreAndGetChunk", func(t *testing.T) {
		data := []byte("test chunk data")
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Calculate hash
		hash, err := dm.CalculateHash(ctx, data)
		if err != nil {
			t.Fatalf("Failed to calculate hash: %v", err)
		}

		// Store chunk
		chunk := &Chunk{
			Hash:     hash,
			Size:     int64(len(data)),
			Data:     data,
			Metadata: map[string]string{"type": "test"},
		}

		err = dm.StoreChunk(ctx, chunk)
		if err != nil {
			t.Fatalf("Failed to store chunk: %v", err)
		}

		// Get chunk
		retrieved, err := dm.GetChunk(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to get chunk: %v", err)
		}

		if retrieved.Hash != hash {
			t.Errorf("Expected hash %s, got %s", hash, retrieved.Hash)
		}

		if retrieved.Size != int64(len(data)) {
			t.Errorf("Expected size %d, got %d", len(data), retrieved.Size)
		}

		if retrieved.RefCount != 1 {
			t.Errorf("Expected ref count 1, got %d", retrieved.RefCount)
		}

		if retrieved.TenantID != "test-tenant" {
			t.Errorf("Expected tenant ID 'test-tenant', got '%s'", retrieved.TenantID)
		}
	})

	t.Run("ChunkExists", func(t *testing.T) {
		data := []byte("test chunk existence")
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Calculate hash
		hash, err := dm.CalculateHash(ctx, data)
		if err != nil {
			t.Fatalf("Failed to calculate hash: %v", err)
		}

		// Check non-existent chunk
		exists, err := dm.ChunkExists(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to check chunk existence: %v", err)
		}

		if exists {
			t.Error("Chunk should not exist before storage")
		}

		// Store chunk
		chunk := &Chunk{
			Hash: hash,
			Size: int64(len(data)),
			Data: data,
		}

		err = dm.StoreChunk(ctx, chunk)
		if err != nil {
			t.Fatalf("Failed to store chunk: %v", err)
		}

		// Check existing chunk
		exists, err = dm.ChunkExists(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to check chunk existence: %v", err)
		}

		if !exists {
			t.Error("Chunk should exist after storage")
		}
	})

	t.Run("FindDuplicateChunks", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Store some chunks
		chunks := []struct {
			data []byte
			hash string
		}{
			{data: []byte("chunk1"), hash: ""},
			{data: []byte("chunk2"), hash: ""},
			{data: []byte("chunk3"), hash: ""},
		}

		for i := range chunks {
			hash, err := dm.CalculateHash(ctx, chunks[i].data)
			if err != nil {
				t.Fatalf("Failed to calculate hash: %v", err)
			}
			chunks[i].hash = hash

			chunk := &Chunk{
				Hash: hash,
				Size: int64(len(chunks[i].data)),
				Data: chunks[i].data,
			}

			err = dm.StoreChunk(ctx, chunk)
			if err != nil {
				t.Fatalf("Failed to store chunk: %v", err)
			}
		}

		// Test finding duplicates
		hashes := []string{chunks[0].hash, chunks[1].hash, "nonexistent"}
		duplicates, err := dm.FindDuplicateChunks(ctx, hashes)
		if err != nil {
			t.Fatalf("Failed to find duplicates: %v", err)
		}

		if len(duplicates) != 3 {
			t.Errorf("Expected 3 results, got %d", len(duplicates))
		}

		if !duplicates[chunks[0].hash] {
			t.Error("First chunk should be found")
		}

		if !duplicates[chunks[1].hash] {
			t.Error("Second chunk should be found")
		}

		if duplicates["nonexistent"] {
			t.Error("Non-existent chunk should not be found")
		}
	})

	t.Run("DeleteChunk", func(t *testing.T) {
		data := []byte("test chunk deletion")
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Calculate hash
		hash, err := dm.CalculateHash(ctx, data)
		if err != nil {
			t.Fatalf("Failed to calculate hash: %v", err)
		}

		// Store chunk
		chunk := &Chunk{
			Hash: hash,
			Size: int64(len(data)),
			Data: data,
		}

		err = dm.StoreChunk(ctx, chunk)
		if err != nil {
			t.Fatalf("Failed to store chunk: %v", err)
		}

		// Verify chunk exists
		exists, err := dm.ChunkExists(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to check chunk existence: %v", err)
		}

		if !exists {
			t.Error("Chunk should exist before deletion")
		}

		// Delete chunk
		err = dm.DeleteChunk(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to delete chunk: %v", err)
		}

		// Verify chunk is deleted
		exists, err = dm.ChunkExists(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to check chunk existence: %v", err)
		}

		if exists {
			t.Error("Chunk should not exist after deletion")
		}
	})

	t.Run("ReferenceCounting", func(t *testing.T) {
		data := []byte("test reference counting")
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Calculate hash
		hash, err := dm.CalculateHash(ctx, data)
		if err != nil {
			t.Fatalf("Failed to calculate hash: %v", err)
		}

		// Store chunk twice
		chunk := &Chunk{
			Hash: hash,
			Size: int64(len(data)),
			Data: data,
		}

		err = dm.StoreChunk(ctx, chunk)
		if err != nil {
			t.Fatalf("Failed to store chunk: %v", err)
		}

		err = dm.StoreChunk(ctx, chunk)
		if err != nil {
			t.Fatalf("Failed to store chunk: %v", err)
		}

		// Check reference count
		retrieved, err := dm.GetChunk(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to get chunk: %v", err)
		}

		if retrieved.RefCount != 2 {
			t.Errorf("Expected ref count 2, got %d", retrieved.RefCount)
		}

		// Delete once
		err = dm.DeleteChunk(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to delete chunk: %v", err)
		}

		// Should still exist
		exists, err := dm.ChunkExists(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to check chunk existence: %v", err)
		}

		if !exists {
			t.Error("Chunk should still exist after one deletion")
		}

		// Delete second time
		err = dm.DeleteChunk(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to delete chunk: %v", err)
		}

		// Should be deleted now
		exists, err = dm.ChunkExists(ctx, hash)
		if err != nil {
			t.Fatalf("Failed to check chunk existence: %v", err)
		}

		if exists {
			t.Error("Chunk should be deleted after second deletion")
		}
	})
}

// TestChunker tests the chunker functionality
func TestChunker(t *testing.T) {
	config := NewDeduplicationConfig()
	chunker := NewChunker(config)

	t.Run("ChunkData", func(t *testing.T) {
		data := []byte("this is a test string for chunking")
		chunks := chunker.ChunkData(data)

		if len(chunks) == 0 {
			t.Error("Should produce at least one chunk")
		}

		// Verify total size
		totalSize := 0
		for _, chunk := range chunks {
			totalSize += len(chunk)
		}

		if totalSize != len(data) {
			t.Errorf("Expected total size %d, got %d", len(data), totalSize)
		}
	})

	t.Run("ChunkDataEmpty", func(t *testing.T) {
		chunks := chunker.ChunkData([]byte{})

		if len(chunks) != 0 {
			t.Error("Empty data should produce no chunks")
		}
	})

	t.Run("ReassembleData", func(t *testing.T) {
		original := []byte("this is test data for reassembly")
		chunks := chunker.ChunkData(original)

		reassembled := chunker.ReassembleData(chunks)

		if string(reassembled) != string(original) {
			t.Errorf("Expected '%s', got '%s'", string(original), string(reassembled))
		}
	})
}

// TestDeduplicationConfig tests deduplication configuration
func TestDeduplicationConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := NewDeduplicationConfig()

		if !config.Enabled {
			t.Error("Deduplication should be enabled by default")
		}

		if config.ChunkSize != 64*1024 {
			t.Errorf("Expected chunk size 64KB, got %d", config.ChunkSize)
		}

		if config.Algorithm != string(HashAlgorithmSHA256) {
			t.Errorf("Expected algorithm SHA256, got %s", config.Algorithm)
		}

		if config.MaxChunkSize != 1024*1024 {
			t.Errorf("Expected max chunk size 1MB, got %d", config.MaxChunkSize)
		}

		if config.MinChunkSize != 1024 {
			t.Errorf("Expected min chunk size 1KB, got %d", config.MinChunkSize)
		}
	})
}

// TestDeduplicationStats tests deduplication statistics
func TestDeduplicationStats(t *testing.T) {
	config := NewDeduplicationConfig()
	// Adjust min chunk size for testing
	config.MinChunkSize = 1
	config.MaxChunkSize = 1024 * 1024 // 1MB

	mockTenantMgr := &MockTenantManager{}
	dm := NewInMemoryDeduplicationManager(config, mockTenantMgr)

	t.Run("TenantStats", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Store some chunks
		data1 := []byte("test data 1")
		data2 := []byte("test data 2")

		hash1, _ := dm.CalculateHash(ctx, data1)
		hash2, _ := dm.CalculateHash(ctx, data2)

		chunk1 := &Chunk{Hash: hash1, Size: int64(len(data1)), Data: data1}
		chunk2 := &Chunk{Hash: hash2, Size: int64(len(data2)), Data: data2}

		dm.StoreChunk(ctx, chunk1)
		dm.StoreChunk(ctx, chunk2)

		// Get stats
		stats, err := dm.GetDeduplicationStats(ctx, "test-tenant")
		if err != nil {
			t.Fatalf("Failed to get stats: %v", err)
		}

		if stats.TenantID != "test-tenant" {
			t.Errorf("Expected tenant ID 'test-tenant', got '%s'", stats.TenantID)
		}

		if stats.TotalChunks != 2 {
			t.Errorf("Expected 2 chunks, got %d", stats.TotalChunks)
		}

		expectedSize := int64(len(data1) + len(data2))
		if stats.TotalSize != expectedSize {
			t.Errorf("Expected total size %d, got %d", expectedSize, stats.TotalSize)
		}
	})

	t.Run("GlobalStats", func(t *testing.T) {
		ctx := mockTenantMgr.WithTenant(context.Background(), "test-tenant")

		// Get global stats
		stats, err := dm.GetGlobalStats(ctx)
		if err != nil {
			t.Fatalf("Failed to get global stats: %v", err)
		}

		if stats.TotalTenants < 1 {
			t.Error("Should have at least one tenant")
		}

		if stats.TotalChunks < 2 {
			t.Error("Should have at least 2 chunks")
		}

		if stats.TotalSize < 1 {
			t.Error("Total size should be greater than 0")
		}
	})
}

// TestChunkValidation tests chunk validation
func TestChunkValidation(t *testing.T) {
	config := NewDeduplicationConfig()
	// Adjust min chunk size for testing
	config.MinChunkSize = 1
	config.MaxChunkSize = 1024 * 1024 // 1MB

	t.Run("ValidChunk", func(t *testing.T) {
		data := []byte("valid chunk data")
		hash, _ := CalculateChunkHash(data, HashAlgorithmSHA256)

		chunk := &Chunk{
			Hash: hash,
			Size: int64(len(data)),
			Data: data,
		}

		err := ValidateChunk(chunk, config)
		if err != nil {
			t.Errorf("Valid chunk should pass validation: %v", err)
		}
	})

	t.Run("InvalidHash", func(t *testing.T) {
		data := []byte("valid chunk data")

		chunk := &Chunk{
			Hash: "invalidhash",
			Size: int64(len(data)),
			Data: data,
		}

		err := ValidateChunk(chunk, config)
		if err == nil {
			t.Error("Invalid hash should fail validation")
		}
	})

	t.Run("EmptyData", func(t *testing.T) {
		chunk := &Chunk{
			Hash: "somehash",
			Size: 0,
			Data: []byte{},
		}

		err := ValidateChunk(chunk, config)
		if err == nil {
			t.Error("Empty data should fail validation")
		}
	})

	t.Run("ChunkTooLarge", func(t *testing.T) {
		data := make([]byte, 2*1024*1024) // 2MB
		hash, _ := CalculateChunkHash(data, HashAlgorithmSHA256)

		chunk := &Chunk{
			Hash: hash,
			Size: int64(len(data)),
			Data: data,
		}

		err := ValidateChunk(chunk, config)
		if err == nil {
			t.Error("Oversized chunk should fail validation")
		}
	})

	t.Run("ChunkTooSmall", func(t *testing.T) {
		// Use original config for this test to verify min size validation
		originalConfig := NewDeduplicationConfig()

		data := []byte("small")
		hash, _ := CalculateChunkHash(data, HashAlgorithmSHA256)

		chunk := &Chunk{
			Hash: hash,
			Size: int64(len(data)),
			Data: data,
		}

		err := ValidateChunk(chunk, originalConfig)
		if err == nil {
			t.Error("Undersized chunk should fail validation")
		}
	})
}
