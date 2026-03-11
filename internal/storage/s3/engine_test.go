package s3

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// TestConfig returns a test configuration for S3
func testConfig() *S3Config {
	return &S3Config{
		Endpoint:             os.Getenv("S3_ENDPOINT"),
		Region:               getEnvOrDefault("S3_REGION", "us-east-1"),
		AccessKeyID:          os.Getenv("S3_ACCESS_KEY"),
		SecretAccessKey:      os.Getenv("S3_SECRET_KEY"),
		Bucket:               getEnvOrDefault("S3_BUCKET", "novabackup-test"),
		UseSSL:               getEnvOrDefault("S3_USE_SSL", "true") == "true",
		DisableSSL:           false,
		ForcePathStyle:       getEnvOrDefault("S3_FORCE_PATH_STYLE", "true") == "true",
		PartSize:             5 * 1024 * 1024, // 5MB
		Concurrency:          3,
		ServerSideEncryption: "AES256",
		ObjectLock:           false,
		RetentionDays:        30,
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateTestBucket creates a test bucket for testing
func CreateTestBucket(ctx context.Context, cfg *S3Config) error {
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return fmt.Errorf("S3 test credentials not configured")
	}

	engine, err := NewS3Engine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create S3 engine: %w", err)
	}
	defer engine.Close()

	// Try to create bucket
	_, err = engine.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		// Bucket might already exist
		if !strings.Contains(err.Error(), "BucketAlreadyExists") &&
			!strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
			return fmt.Errorf("failed to create test bucket: %w", err)
		}
	}

	return nil
}

// CleanupTestBucket cleans up test bucket
func CleanupTestBucket(ctx context.Context, cfg *S3Config) error {
	engine, err := NewS3Engine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create S3 engine: %w", err)
	}
	defer engine.Close()

	// Delete all objects in bucket
	paginator := s3.NewListObjectsV2Paginator(engine.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(cfg.Bucket),
	})

	var objectsToDelete []types.ObjectIdentifier

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects for cleanup: %w", err)
		}

		for _, obj := range page.Contents {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}
	}

	// Delete objects in batches
	for i := 0; i < len(objectsToDelete); i += 1000 {
		end := i + 1000
		if end > len(objectsToDelete) {
			end = len(objectsToDelete)
		}

		batch := objectsToDelete[i:end]
		_, err := engine.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(cfg.Bucket),
			Delete: &types.Delete{
				Objects: batch,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete batch of objects: %w", err)
		}
	}

	// Delete bucket
	_, err = engine.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to delete test bucket: %w", err)
	}

	return nil
}

// TestS3Engine runs comprehensive S3 engine tests
func TestS3Engine(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping S3 integration tests in short mode")
	}

	cfg := testConfig()
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		t.Skip("S3 test credentials not configured")
	}

	// Setup test bucket
	ctx := context.Background()
	testBucket := cfg.Bucket + "-" + fmt.Sprintf("%d", time.Now().Unix())
	cfg.Bucket = testBucket

	err := CreateTestBucket(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create test bucket: %v", err)
	}
	defer CleanupTestBucket(ctx, cfg)

	// Create engine
	engine, err := NewS3Engine(cfg)
	if err != nil {
		t.Fatalf("Failed to create S3 engine: %v", err)
	}
	defer engine.Close()

	// Test connection
	err = engine.TestConnection(ctx)
	if err != nil {
		t.Errorf("Connection test failed: %v", err)
	}

	// Test store and retrieve
	testHash := "test-chunk-12345"
	testData := []byte("This is test data for S3 storage engine testing")

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
	if info.Type != "S3" {
		t.Errorf("Expected storage type S3, got: %s", info.Type)
	}
	if info.Bucket != testBucket {
		t.Errorf("Expected bucket %s, got: %s", testBucket, info.Bucket)
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

// BenchmarkS3Store benchmarks S3 chunk storage
func BenchmarkS3Store(b *testing.B) {
	cfg := testConfig()
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		b.Skip("S3 test credentials not configured")
	}

	ctx := context.Background()
	engine, err := NewS3Engine(cfg)
	if err != nil {
		b.Fatalf("Failed to create S3 engine: %v", err)
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

// BenchmarkS3Retrieve benchmarks S3 chunk retrieval
func BenchmarkS3Retrieve(b *testing.B) {
	cfg := testConfig()
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		b.Skip("S3 test credentials not configured")
	}

	ctx := context.Background()
	engine, err := NewS3Engine(cfg)
	if err != nil {
		b.Fatalf("Failed to create S3 engine: %v", err)
	}
	defer engine.Close()

	// Pre-store test data
	testData := make([]byte, 1024*1024) // 1MB test data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	for i := 0; i < 100; i++ {
		hash := fmt.Sprintf("bench-retrieve-%d", i)
		_, err := engine.StoreChunk(ctx, hash, testData)
		if err != nil {
			b.Fatalf("Failed to pre-store chunk %d: %v", i, err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := fmt.Sprintf("bench-retrieve-%d", i%100)
		_, err := engine.GetChunk(ctx, hash)
		if err != nil {
			b.Errorf("Failed to retrieve chunk %d: %v", i, err)
		}
	}
}
