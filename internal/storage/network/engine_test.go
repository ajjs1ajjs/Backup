package network

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// testConfig returns a test configuration for network storage
func testConfig(storageType NetworkStorageType) *NetworkStorageConfig {
	return &NetworkStorageConfig{
		Type:        storageType,
		Host:        getEnvOrDefault("NETWORK_HOST", "localhost"),
		Share:       getEnvOrDefault("NETWORK_SHARE", "test"),
		Path:        getEnvOrDefault("NETWORK_PATH", "/"),
		Username:    getEnvOrDefault("NETWORK_USER", "testuser"),
		Password:    getEnvOrDefault("NETWORK_PASS", "testpass"),
		Domain:      getEnvOrDefault("NETWORK_DOMAIN", ""),
		MountPoint:  getEnvOrDefault("NETWORK_MOUNT", "/tmp/nova-test-mount"),
		AutoMount:   true,
		AutoUnmount: true,
		Timeout:     10 * time.Second,
		RetryCount:  2,
		BufferSize:  32 * 1024,
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestNetworkStorageEngine tests network storage functionality
func TestNetworkStorageEngine(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network storage tests in short mode")
	}

	// Test NFS
	t.Run("NFS", func(t *testing.T) {
		cfg := testConfig(StorageNFS)
		if cfg.Host == "localhost" {
			t.Skip("NFS test host not configured")
		}

		testNetworkStorage(t, cfg)
	})

	// Test SMB
	t.Run("SMB", func(t *testing.T) {
		cfg := testConfig(StorageSMB)
		if cfg.Host == "localhost" {
			t.Skip("SMB test host not configured")
		}

		testNetworkStorage(t, cfg)
	})
}

// testNetworkStorage performs common network storage tests
func testNetworkStorage(t *testing.T, cfg *NetworkStorageConfig) {
	ctx := context.Background()

	// Create engine
	engine, err := NewNetworkStorageEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create network storage engine: %v", err)
	}
	defer engine.Close()

	// Test connection
	err = engine.TestConnection(ctx)
	if err != nil {
		t.Errorf("Connection test failed: %v", err)
	}

	// Test store and retrieve
	testHash := "test-chunk-network-12345"
	testData := []byte("This is test data for network storage engine testing")

	// Store chunk
	location, err := engine.StoreChunk(ctx, testHash, testData)
	if err != nil {
		t.Errorf("Failed to store chunk: %v", err)
	}
	if location == "" {
		t.Error("StoreChunk returned empty location")
	}

	// Check if chunk exists
	exists, err := engine.ChunkExists(ctx, testHash)
	if err != nil {
		t.Errorf("Failed to check chunk existence: %v", err)
	}
	if !exists {
		t.Error("Chunk should exist after storage")
	}

	// Retrieve chunk
	retrievedData, err := engine.GetChunk(ctx, testHash)
	if err != nil {
		t.Errorf("Failed to retrieve chunk: %v", err)
	}
	if string(retrievedData) != string(testData) {
		t.Errorf("Retrieved data mismatch. Expected: %s, Got: %s", string(testData), string(retrievedData))
	}

	// List chunks
	chunks, err := engine.ListChunks(ctx, "")
	if err != nil {
		t.Errorf("Failed to list chunks: %v", err)
	}
	if len(chunks) == 0 {
		t.Error("Expected at least one chunk in list")
	}

	// Get storage info
	info, err := engine.GetStorageInfo(ctx)
	if err != nil {
		t.Errorf("Failed to get storage info: %v", err)
	}
	if info.Type != string(cfg.Type) {
		t.Errorf("Expected storage type %s, got: %s", string(cfg.Type), info.Type)
	}
	if info.Endpoint == "" {
		t.Error("Expected non-empty endpoint")
	}

	// Delete chunk
	err = engine.DeleteChunk(ctx, testHash)
	if err != nil {
		t.Errorf("Failed to delete chunk: %v", err)
	}

	// Verify chunk is deleted
	exists, err = engine.ChunkExists(ctx, testHash)
	if err != nil {
		t.Errorf("Failed to check chunk existence after delete: %v", err)
	}
	if exists {
		t.Error("Chunk should not exist after deletion")
	}
}

// BenchmarkNetworkStore benchmarks network chunk storage
func BenchmarkNetworkStore(b *testing.B) {
	cfg := testConfig(StorageNFS)
	if cfg.Host == "localhost" {
		b.Skip("Network test host not configured")
	}

	ctx := context.Background()
	engine, err := NewNetworkStorageEngine(cfg)
	if err != nil {
		b.Fatalf("Failed to create network storage engine: %v", err)
	}
	defer engine.Close()

	testData := make([]byte, 1024*1024) // 1MB test data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := fmt.Sprintf("bench-chunk-%d", i)
		_, err := engine.StoreChunk(ctx, hash, testData)
		if err != nil {
			b.Errorf("Failed to store chunk %d: %v", i, err)
		}
	}
}

// BenchmarkNetworkRetrieve benchmarks network chunk retrieval
func BenchmarkNetworkRetrieve(b *testing.B) {
	cfg := testConfig(StorageNFS)
	if cfg.Host == "localhost" {
		b.Skip("Network test host not configured")
	}

	ctx := context.Background()
	engine, err := NewNetworkStorageEngine(cfg)
	if err != nil {
		b.Fatalf("Failed to create network storage engine: %v", err)
	}
	defer engine.Close()

	// Pre-store test data
	testData := make([]byte, 1024*1024) // 1MB test data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	for i := 0; i < 10; i++ {
		hash := fmt.Sprintf("bench-retrieve-%d", i)
		_, err := engine.StoreChunk(ctx, hash, testData)
		if err != nil {
			b.Fatalf("Failed to pre-store chunk %d: %v", i, err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := fmt.Sprintf("bench-retrieve-%d", i%10)
		_, err := engine.GetChunk(ctx, hash)
		if err != nil {
			b.Errorf("Failed to retrieve chunk %d: %v", i, err)
		}
	}
}

// TestMountUnmount tests mount/unmount functionality
func TestMountUnmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mount/unmount tests in short mode")
	}

	cfg := testConfig(StorageNFS)
	if cfg.Host == "localhost" {
		t.Skip("Network test host not configured")
	}

	// Disable auto-mount for this test
	cfg.AutoMount = false

	ctx := context.Background()
	engine, err := NewNetworkStorageEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create network storage engine: %v", err)
	}
	defer engine.Close()

	// Should not be mounted initially
	if engine.IsMounted() {
		t.Error("Engine should not be mounted initially")
	}

	// Test mount
	err = engine.Mount(ctx)
	if err != nil {
		t.Errorf("Failed to mount: %v", err)
	}

	// Should be mounted now
	if !engine.IsMounted() {
		t.Error("Engine should be mounted after mount call")
	}

	// Test unmount
	err = engine.Unmount(ctx)
	if err != nil {
		t.Errorf("Failed to unmount: %v", err)
	}

	// Should not be mounted now
	if engine.IsMounted() {
		t.Error("Engine should not be mounted after unmount call")
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *NetworkStorageConfig
		wantErr bool
	}{
		{
			name:    "Valid NFS config",
			config:  testConfig(StorageNFS),
			wantErr: false,
		},
		{
			name:    "Valid SMB config",
			config:  testConfig(StorageSMB),
			wantErr: false,
		},
		{
			name:    "Nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "Empty host",
			config: &NetworkStorageConfig{
				Type:       StorageNFS,
				Host:       "",
				Share:      "test",
				MountPoint: "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "Empty share",
			config: &NetworkStorageConfig{
				Type:       StorageNFS,
				Host:       "localhost",
				Share:      "",
				MountPoint: "/tmp/test",
			},
			wantErr: true,
		},
		{
			name: "Empty mount point",
			config: &NetworkStorageConfig{
				Type:       StorageNFS,
				Host:       "localhost",
				Share:      "test",
				MountPoint: "",
			},
			wantErr: true,
		},
		{
			name: "SMB without username",
			config: &NetworkStorageConfig{
				Type:       StorageSMB,
				Host:       "localhost",
				Share:      "test",
				MountPoint: "/tmp/test",
				Username:   "",
				Password:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &NetworkStorageEngine{config: tt.config}
			err := engine.ValidateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
