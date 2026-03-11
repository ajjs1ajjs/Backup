// Package cloud provides cloud storage integrations
package cloud

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"go.uber.org/zap"
)

// GCSProvider provides Google Cloud Storage integration
type GCSProvider struct {
	logger     *zap.Logger
	projectID  string
	bucketName string
	client     *storage.Client
}

// GCSConfig contains GCS configuration
type GCSConfig struct {
	ProjectID      string
	BucketName     string
	CredentialsJSON string // Service account key JSON
	UseADC         bool   // Use Application Default Credentials
}

// NewGCSProvider creates a new GCS provider
func NewGCSProvider(logger *zap.Logger, config *GCSConfig) (*GCSProvider, error) {
	logger.Info("Creating GCS provider",
		zap.String("project", config.ProjectID),
		zap.String("bucket", config.BucketName))

	ctx := context.Background()
	var client *storage.Client
	var err error

	if config.UseADC {
		// Use Application Default Credentials
		client, err = storage.NewClient(ctx)
	} else if config.CredentialsJSON != "" {
		// Use service account credentials
		client, err = storage.NewClient(ctx, option.WithCredentialsJSON([]byte(config.CredentialsJSON)))
	} else {
		return nil, fmt.Errorf("either UseADC or CredentialsJSON must be provided")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSProvider{
		logger:     logger.With(zap.String("provider", "gcs")),
		projectID:  config.ProjectID,
		bucketName: config.BucketName,
		client:     client,
	}, nil
}

// TestConnection tests GCS connection
func (g *GCSProvider) TestConnection(ctx context.Context) error {
	g.logger.Info("Testing GCS connection")

	// Try to get bucket attributes to verify connection
	bucket := g.client.Bucket(g.bucketName)
	_, err := bucket.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to GCS: %w", err)
	}

	g.logger.Info("GCS connection successful")
	return nil
}

// Upload uploads data to GCS
func (g *GCSProvider) Upload(ctx context.Context, key string, data io.Reader, size int64) error {
	g.logger.Info("Uploading to GCS",
		zap.String("bucket", g.bucketName),
		zap.String("object", key),
		zap.Int64("size", size))

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(key)

	writer := obj.NewWriter(ctx)
	writer.ContentType = "application/octet-stream"

	if _, err := io.Copy(writer, data); err != nil {
		writer.Close()
		return fmt.Errorf("failed to upload to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}

	g.logger.Info("Upload successful", zap.String("object", key))
	return nil
}

// Download downloads data from GCS
func (g *GCSProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	g.logger.Info("Downloading from GCS",
		zap.String("bucket", g.bucketName),
		zap.String("object", key))

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(key)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to download from GCS: %w", err)
	}

	return reader, nil
}

// Delete deletes an object from GCS
func (g *GCSProvider) Delete(ctx context.Context, key string) error {
	g.logger.Info("Deleting from GCS",
		zap.String("bucket", g.bucketName),
		zap.String("object", key))

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(key)

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete from GCS: %w", err)
	}

	return nil
}

// List lists objects in bucket with prefix
func (g *GCSProvider) List(ctx context.Context, prefix string) ([]BlobInfo, error) {
	g.logger.Info("Listing GCS objects", zap.String("prefix", prefix))

	bucket := g.client.Bucket(g.bucketName)
	query := &storage.Query{Prefix: prefix}

	var blobs []BlobInfo
	it := bucket.Objects(ctx, query)

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list GCS objects: %w", err)
		}

		blobs = append(blobs, BlobInfo{
			Key:          attrs.Name,
			Size:         attrs.Size,
			LastModified: attrs.Updated,
			ETag:         attrs.Etag,
			StorageClass: attrs.StorageClass,
		})
	}

	return blobs, nil
}

// GetObjectInfo gets object information
func (g *GCSProvider) GetObjectInfo(ctx context.Context, key string) (*BlobInfo, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(key)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCS object attributes: %w", err)
	}

	return &BlobInfo{
		Key:          attrs.Name,
		Size:         attrs.Size,
		LastModified: attrs.Updated,
		ETag:         attrs.Etag,
		StorageClass: attrs.StorageClass,
	}, nil
}

// ArchiveToColdline archives object to Coldline storage class
func (g *GCSProvider) ArchiveToColdline(ctx context.Context, key string) error {
	g.logger.Info("Archiving to Coldline", zap.String("object", key))

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(key)

	// Copy to same location with Coldline storage class
	copier := obj.CopierFrom(obj)
	copier.StorageClass = "COLDLINE"
	_, err := copier.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to archive to Coldline: %w", err)
	}

	return nil
}

// ArchiveToArchive archives object to Archive storage class
func (g *GCSProvider) ArchiveToArchive(ctx context.Context, key string) error {
	g.logger.Info("Archiving to Archive", zap.String("object", key))

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(key)

	// Copy to same location with Archive storage class
	copier := obj.CopierFrom(obj)
	copier.StorageClass = "ARCHIVE"
	_, err := copier.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to archive to Archive: %w", err)
	}

	return nil
}

// RestoreFromArchive rehydrates object to standard storage class
func (g *GCSProvider) RestoreFromArchive(ctx context.Context, key string) error {
	g.logger.Info("Restoring from Archive", zap.String("object", key))

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(key)

	// Copy to same location with Standard storage class
	copier := obj.CopierFrom(obj)
	copier.StorageClass = "STANDARD"
	_, err := copier.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to restore from Archive: %w", err)
	}

	return nil
}

// CreateBucket creates a bucket if it doesn't exist
func (g *GCSProvider) CreateBucket(ctx context.Context, location string) error {
	g.logger.Info("Creating bucket",
		zap.String("bucket", g.bucketName),
		zap.String("location", location))

	bucket := g.client.Bucket(g.bucketName)

	if err := bucket.Create(ctx, g.projectID, &storage.BucketAttrs{
		Location: location,
	}); err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

// CopyObject copies an object within GCS
func (g *GCSProvider) CopyObject(ctx context.Context, sourceKey, destKey string) error {
	g.logger.Info("Copying object",
		zap.String("source", sourceKey),
		zap.String("dest", destKey))

	bucket := g.client.Bucket(g.bucketName)
	srcObj := bucket.Object(sourceKey)
	dstObj := bucket.Object(destKey)

	if _, err := dstObj.CopierFrom(srcObj).Run(ctx); err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	return nil
}

// ComposeObjects composes multiple objects into one
func (g *GCSProvider) ComposeObjects(ctx context.Context, sourceKeys []string, destKey string) error {
	g.logger.Info("Composing objects",
		zap.Strings("sources", sourceKeys),
		zap.String("dest", destKey))

	bucket := g.client.Bucket(g.bucketName)

	var sources []*storage.ObjectHandle
	for _, key := range sourceKeys {
		sources = append(sources, bucket.Object(key))
	}

	dstObj := bucket.Object(destKey)
	if _, err := dstObj.ComposerFrom(sources...).Run(ctx); err != nil {
		return fmt.Errorf("failed to compose objects: %w", err)
	}

	return nil
}

// GenerateSignedURL generates a signed URL for temporary access
func (g *GCSProvider) GenerateSignedURL(ctx context.Context, key string, method string, expiry time.Duration) (string, error) {
	g.logger.Info("Generating signed URL",
		zap.String("object", key),
		zap.String("method", method),
		zap.Duration("expiry", expiry))

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  method,
		Expires: time.Now().Add(expiry),
	}

	url, err := storage.SignedURL(g.bucketName, key, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// Close closes the GCS client
func (g *GCSProvider) Close() error {
	if err := g.client.Close(); err != nil {
		return fmt.Errorf("failed to close GCS client: %w", err)
	}
	return nil
}
