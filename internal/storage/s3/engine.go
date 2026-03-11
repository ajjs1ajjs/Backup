package s3

import (
	"context"
	"fmt"
	"strings"
	"time"

	"novabackup/pkg/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Engine handles S3-compatible storage operations
type S3Engine struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	config     *S3Config
	bucket     string
}

// S3Config contains S3 storage configuration
type S3Config struct {
	Endpoint             string `json:"endpoint"`
	Region               string `json:"region"`
	AccessKeyID          string `json:"access_key_id"`
	SecretAccessKey      string `json:"secret_access_key"`
	Bucket               string `json:"bucket"`
	UseSSL               bool   `json:"use_ssl"`
	DisableSSL           bool   `json:"disable_ssl"`
	ForcePathStyle       bool   `json:"force_path_style"`
	PartSize             int64  `json:"part_size"`              // Multipart upload part size in bytes
	Concurrency          int    `json:"concurrency"`            // Upload concurrency
	ServerSideEncryption string `json:"server_side_encryption"` // AES256, aws:kms
	ObjectLock           bool   `json:"object_lock"`            // Enable WORM
	RetentionDays        int    `json:"retention_days"`         // Object lock retention
}

// S3ObjectInfo contains S3 object metadata
type S3ObjectInfo struct {
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
	StorageClass string
}

// NewS3Engine creates a new S3 storage engine
func NewS3Engine(cfg *S3Config) (*S3Engine, error) {
	if cfg == nil {
		return nil, fmt.Errorf("S3 config cannot be nil")
	}

	// Set defaults
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.PartSize == 0 {
		cfg.PartSize = 5 * 1024 * 1024 // 5MB default part size
	}
	if cfg.Concurrency == 0 {
		cfg.Concurrency = 5
	}
	if cfg.ServerSideEncryption == "" {
		cfg.ServerSideEncryption = "AES256"
	}

	// Create AWS config
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.AccessKeyID,
				SecretAccessKey: cfg.SecretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create S3 client with custom options
	clientOptions := []func(*s3.Options){
		func(o *s3.Options) {
			if cfg.Endpoint != "" && !strings.Contains(cfg.Endpoint, "amazonaws.com") {
				o.BaseEndpoint = aws.String(cfg.Endpoint)
			}
			if cfg.ForcePathStyle {
				o.UsePathStyle = true
			}
		},
	}

	// Configure SSL
	if cfg.DisableSSL {
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.RetryMaxAttempts = 3
			o.RetryMode = aws.RetryModeAdaptive
		})
	}

	client := s3.NewFromConfig(awsCfg, clientOptions...)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if bucket exists
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to access bucket %s: %w", cfg.Bucket, err)
	}

	// Create uploader and downloader
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = cfg.PartSize
		u.Concurrency = cfg.Concurrency
	})

	downloader := manager.NewDownloader(client, func(d *manager.Downloader) {
		d.PartSize = cfg.PartSize
		d.Concurrency = cfg.Concurrency
	})

	return &S3Engine{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		config:     cfg,
		bucket:     cfg.Bucket,
	}, nil
}

// StoreChunk stores a chunk in S3 using multipart upload
func (e *S3Engine) StoreChunk(ctx context.Context, hash string, data []byte) (string, error) {
	key := fmt.Sprintf("chunks/%s", hash)

	// Prepare upload input
	input := &s3.PutObjectInput{
		Bucket: aws.String(e.bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(string(data)),
	}

	// Set server-side encryption
	if e.config.ServerSideEncryption != "" {
		input.ServerSideEncryption = types.ServerSideEncryption(e.config.ServerSideEncryption)
	}

	// Set object lock if enabled
	if e.config.ObjectLock {
		input.ObjectLockMode = types.ObjectLockModeCompliance
		if e.config.RetentionDays > 0 {
			retentionTime := time.Now().AddDate(0, 0, e.config.RetentionDays)
			input.ObjectLockRetainUntilDate = &retentionTime
		}
	}

	// Set metadata
	input.Metadata = map[string]string{
		"nova-chunk-hash": hash,
		"nova-version":    "6.0",
	}

	// Upload using multipart uploader
	result, err := e.uploader.Upload(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload chunk %s: %w", hash, err)
	}

	return result.Location, nil
}

// GetChunk retrieves a chunk from S3
func (e *S3Engine) GetChunk(ctx context.Context, hash string) ([]byte, error) {
	key := fmt.Sprintf("chunks/%s", hash)

	input := &s3.GetObjectInput{
		Bucket: aws.String(e.bucket),
		Key:    aws.String(key),
	}

	// Download using range requests for large files
	buf := manager.NewWriteAtBuffer([]byte{})
	_, err := e.downloader.Download(ctx, buf, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download chunk %s: %w", hash, err)
	}

	return buf.Bytes(), nil
}

// ChunkExists checks if a chunk exists in S3
func (e *S3Engine) ChunkExists(ctx context.Context, hash string) (bool, error) {
	key := fmt.Sprintf("chunks/%s", hash)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(e.bucket),
		Key:    aws.String(key),
	}

	_, err := e.client.HeadObject(ctx, input)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") ||
			strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check chunk existence: %w", err)
	}

	return true, nil
}

// DeleteChunk removes a chunk from S3
func (e *S3Engine) DeleteChunk(ctx context.Context, hash string) error {
	key := fmt.Sprintf("chunks/%s", hash)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(e.bucket),
		Key:    aws.String(key),
	}

	_, err := e.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete chunk %s: %w", hash, err)
	}

	return nil
}

// ListChunks lists all chunks in S3
func (e *S3Engine) ListChunks(ctx context.Context, prefix string) ([]interface{}, error) {
	var objects []interface{}

	paginator := s3.NewListObjectsV2Paginator(e.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(e.bucket),
		Prefix: aws.String("chunks/" + prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list chunks: %w", err)
		}

		for _, obj := range page.Contents {
			objects = append(objects, S3ObjectInfo{
				Key:          *obj.Key,
				Size:         *obj.Size,
				ETag:         *obj.ETag,
				LastModified: *obj.LastModified,
				StorageClass: string(obj.StorageClass),
			})
		}
	}

	return objects, nil
}

// GetStorageInfo returns S3 storage information
func (e *S3Engine) GetStorageInfo(ctx context.Context) (*models.StorageInfo, error) {
	// Get bucket size and object count
	var totalSize int64
	var objectCount int64

	paginator := s3.NewListObjectsV2Paginator(e.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(e.bucket),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get storage info: %w", err)
		}

		for _, obj := range page.Contents {
			totalSize += *obj.Size
			objectCount++
		}
	}

	// Get bucket location
	location, err := e.client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(e.bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket location: %w", err)
	}

	region := string(location.LocationConstraint)
	if region == "" {
		region = "us-east-1" // Default region
	}

	return &models.StorageInfo{
		Type:        "S3",
		TotalSize:   totalSize,
		UsedSize:    totalSize,
		FreeSize:    0, // S3 doesn't provide this info
		ObjectCount: objectCount,
		Endpoint:    e.config.Endpoint,
		Bucket:      e.bucket,
		Region:      region,
		LastUpdated: time.Now(),
	}, nil
}

// TestConnection tests S3 connectivity
func (e *S3Engine) TestConnection(ctx context.Context) error {
	// Try to list bucket
	_, err := e.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(e.bucket),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("S3 connection test failed: %w", err)
	}

	return nil
}

// Cleanup removes old chunks based on retention policy
func (e *S3Engine) Cleanup(ctx context.Context, olderThan time.Time) error {
	paginator := s3.NewListObjectsV2Paginator(e.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(e.bucket),
		Prefix: aws.String("chunks/"),
	})

	var objectsToDelete []types.ObjectIdentifier

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list chunks for cleanup: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.LastModified.Before(olderThan) {
				objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
					Key: obj.Key,
				})
			}
		}
	}

	// Delete objects in batches of 1000
	for i := 0; i < len(objectsToDelete); i += 1000 {
		end := i + 1000
		if end > len(objectsToDelete) {
			end = len(objectsToDelete)
		}

		batch := objectsToDelete[i:end]
		_, err := e.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(e.bucket),
			Delete: &types.Delete{
				Objects: batch,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete batch of objects: %w", err)
		}
	}

	return nil
}

// Close cleans up S3 engine resources
func (e *S3Engine) Close() error {
	// S3 client doesn't need explicit cleanup
	return nil
}
