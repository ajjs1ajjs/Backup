// Package cloud provides cloud storage integration for NovaBackup
package cloud

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
)

// S3Provider provides AWS S3 integration
type S3Provider struct {
	logger *zap.Logger
	client *s3.Client
	bucket string
	prefix string
}

// S3Config holds S3 configuration
type S3Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	Prefix          string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

// BlobInfo contains information about a cloud storage object
type BlobInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	StorageClass string
}

// NewS3Provider creates a new S3 provider
func NewS3Provider(logger *zap.Logger, cfg *S3Config) (*S3Provider, error) {
	ctx := context.Background()

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Provider{
		logger: logger.With(zap.String("provider", "s3")),
		client: client,
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
	}, nil
}

// Upload uploads data to S3
func (s *S3Provider) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	fullKey := s.fullKey(key)
	s.logger.Info("Uploading to S3", zap.String("bucket", s.bucket), zap.String("key", fullKey))

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(fullKey),
		Body:          reader,
		ContentLength: aws.Int64(size),
	})

	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	return nil
}

// Download downloads data from S3
func (s *S3Provider) Download(ctx context.Context, key string, writer io.WriterAt) error {
	fullKey := s.fullKey(key)
	s.logger.Info("Downloading from S3", zap.String("bucket", s.bucket), zap.String("key", fullKey))

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(writer.(io.Writer), resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// Delete deletes an object from S3
func (s *S3Provider) Delete(ctx context.Context, key string) error {
	fullKey := s.fullKey(key)
	s.logger.Info("Deleting from S3", zap.String("bucket", s.bucket), zap.String("key", fullKey))

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
	})

	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}

// List lists objects in S3
func (s *S3Provider) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	fullPrefix := s.fullKey(prefix)
	s.logger.Info("Listing S3 objects", zap.String("bucket", s.bucket), zap.String("prefix", fullPrefix))

	resp, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(fullPrefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list: %w", err)
	}

	objects := make([]ObjectInfo, 0, len(resp.Contents))
	for _, obj := range resp.Contents {
		objects = append(objects, ObjectInfo{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: aws.ToTime(obj.LastModified),
		})
	}

	return objects, nil
}

// ObjectInfo contains information about an S3 object
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
}

// fullKey returns the full key with prefix
func (s *S3Provider) fullKey(key string) string {
	if s.prefix == "" {
		return key
	}
	return s.prefix + "/" + key
}

// GetBucketStats returns bucket statistics
func (s *S3Provider) GetBucketStats(ctx context.Context) (*BucketStats, error) {
	resp, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket stats: %w", err)
	}

	stats := &BucketStats{
		ObjectCount: int64(*resp.KeyCount),
	}

	for _, obj := range resp.Contents {
		stats.TotalSize += aws.ToInt64(obj.Size)
	}

	return stats, nil
}

// BucketStats contains S3 bucket statistics
type BucketStats struct {
	ObjectCount int64
	TotalSize   int64
}

// ArchiveTier archives objects to Glacier
func (s *S3Provider) ArchiveTier(ctx context.Context, key string) error {
	fullKey := s.fullKey(key)
	s.logger.Info("Archiving to Glacier", zap.String("key", fullKey))

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:       aws.String(s.bucket),
		CopySource:   aws.String(fmt.Sprintf("%s/%s", s.bucket, fullKey)),
		Key:          aws.String(fullKey),
		StorageClass: types.StorageClassGlacier,
	})

	if err != nil {
		return fmt.Errorf("failed to archive: %w", err)
	}

	return nil
}

// RestoreFromArchive restores objects from Glacier
func (s *S3Provider) RestoreFromArchive(ctx context.Context, key string, days int32) error {
	fullKey := s.fullKey(key)
	s.logger.Info("Restoring from Glacier", zap.String("key", fullKey), zap.Int32("days", days))

	_, err := s.client.RestoreObject(ctx, &s3.RestoreObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fullKey),
		RestoreRequest: &types.RestoreRequest{
			Days: aws.Int32(days),
			GlacierJobParameters: &types.GlacierJobParameters{
				Tier: types.TierStandard,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}

	return nil
}

// CloudManager manages cloud storage providers
type CloudManager struct {
	logger    *zap.Logger
	providers map[string]CloudProvider
}

// CloudProvider interface for cloud storage
type CloudProvider interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error
	Download(ctx context.Context, key string, writer io.WriterAt) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)
}

// NewCloudManager creates a new cloud manager
func NewCloudManager(logger *zap.Logger) *CloudManager {
	return &CloudManager{
		logger:    logger.With(zap.String("component", "cloud-manager")),
		providers: make(map[string]CloudProvider),
	}
}

// RegisterProvider registers a cloud provider
func (m *CloudManager) RegisterProvider(name string, provider CloudProvider) {
	m.providers[name] = provider
}

// GetProvider returns a cloud provider by name
func (m *CloudManager) GetProvider(name string) (CloudProvider, error) {
	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("cloud provider not found: %s", name)
	}
	return provider, nil
}
