// Package cloud provides cloud storage integrations
package cloud

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"go.uber.org/zap"
)

// AzureBlobProvider provides Azure Blob Storage integration
type AzureBlobProvider struct {
	logger     *zap.Logger
	accountName string
	containerName string
	client     *azblob.Client
}

// AzureBlobConfig contains Azure Blob configuration
type AzureBlobConfig struct {
	AccountName   string
	ContainerName string
	TenantID      string
	ClientID      string
	ClientSecret  string
	UseManagedIdentity bool
}

// NewAzureBlobProvider creates a new Azure Blob provider
func NewAzureBlobProvider(logger *zap.Logger, config *AzureBlobConfig) (*AzureBlobProvider, error) {
	logger.Info("Creating Azure Blob provider",
		zap.String("account", config.AccountName),
		zap.String("container", config.ContainerName))

	var client *azblob.Client
	var err error

	accountURL := fmt.Sprintf("https://%s.blob.core.windows.net", config.AccountName)

	if config.UseManagedIdentity {
		// Use managed identity (for Azure VMs, AKS, etc.)
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure credential: %w", err)
		}
		client, err = azblob.NewClient(accountURL, cred, nil)
	} else {
		// Use service principal
		cred, err := azidentity.NewClientSecretCredential(
			config.TenantID,
			config.ClientID,
			config.ClientSecret,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure credential: %w", err)
		}
		client, err = azblob.NewClient(accountURL, cred, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}

	return &AzureBlobProvider{
		logger:        logger.With(zap.String("provider", "azure-blob")),
		accountName:   config.AccountName,
		containerName: config.ContainerName,
		client:        client,
	}, nil
}

// TestConnection tests Azure Blob connection
func (a *AzureBlobProvider) TestConnection(ctx context.Context) error {
	a.logger.Info("Testing Azure Blob connection")

	// Try to list containers to verify connection
	pager := a.client.NewListContainersPager(nil)
	_, err := pager.NextPage(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Azure Blob: %w", err)
	}

	a.logger.Info("Azure Blob connection successful")
	return nil
}

// Upload uploads data to Azure Blob
func (a *AzureBlobProvider) Upload(ctx context.Context, key string, data io.Reader, size int64) error {
	a.logger.Info("Uploading to Azure Blob",
		zap.String("container", a.containerName),
		zap.String("blob", key),
		zap.Int64("size", size))

	blockBlobClient := a.client.ServiceClient().
		NewContainerClient(a.containerName).
		NewBlockBlobClient(key)

	_, err := blockBlobClient.UploadStream(ctx, data, nil)
	if err != nil {
		return fmt.Errorf("failed to upload to Azure Blob: %w", err)
	}

	a.logger.Info("Upload successful", zap.String("blob", key))
	return nil
}

// Download downloads data from Azure Blob
func (a *AzureBlobProvider) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	a.logger.Info("Downloading from Azure Blob",
		zap.String("container", a.containerName),
		zap.String("blob", key))

	blockBlobClient := a.client.ServiceClient().
		NewContainerClient(a.containerName).
		NewBlockBlobClient(key)

	resp, err := blockBlobClient.DownloadStream(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download from Azure Blob: %w", err)
	}

	return resp.Body, nil
}

// Delete deletes a blob from Azure Blob
func (a *AzureBlobProvider) Delete(ctx context.Context, key string) error {
	a.logger.Info("Deleting from Azure Blob",
		zap.String("container", a.containerName),
		zap.String("blob", key))

	blockBlobClient := a.client.ServiceClient().
		NewContainerClient(a.containerName).
		NewBlockBlobClient(key)

	_, err := blockBlobClient.Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete from Azure Blob: %w", err)
	}

	return nil
}

// List lists blobs in container with prefix
func (a *AzureBlobProvider) List(ctx context.Context, prefix string) ([]BlobInfo, error) {
	a.logger.Info("Listing Azure Blobs", zap.String("prefix", prefix))

	containerClient := a.client.ServiceClient().NewContainerClient(a.containerName)

	var blobs []BlobInfo
	pager := containerClient.NewListBlobsFlatPager(&azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %w", err)
		}

		for _, blob := range resp.Segment.BlobItems {
			blobs = append(blobs, BlobInfo{
				Key:          *blob.Name,
				Size:         *blob.Properties.ContentLength,
				LastModified: *blob.Properties.LastModified,
				ETag:         string(*blob.Properties.Etag),
			})
		}
	}

	return blobs, nil
}

// GetObjectInfo gets blob information
func (a *AzureBlobProvider) GetObjectInfo(ctx context.Context, key string) (*BlobInfo, error) {
	blockBlobClient := a.client.ServiceClient().
		NewContainerClient(a.containerName).
		NewBlockBlobClient(key)

	props, err := blockBlobClient.GetProperties(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob properties: %w", err)
	}

	return &BlobInfo{
		Key:          key,
		Size:         *props.ContentLength,
		LastModified: *props.LastModified,
		ETag:         string(*props.ETag),
	}, nil
}

// ArchiveToArchiveTier archives blob to Archive tier
func (a *AzureBlobProvider) ArchiveToArchiveTier(ctx context.Context, key string) error {
	a.logger.Info("Archiving blob to Archive tier", zap.String("blob", key))

	blockBlobClient := a.client.ServiceClient().
		NewContainerClient(a.containerName).
		NewBlockBlobClient(key)

	accessTier := azblob.AccessTierArchive
	_, err := blockBlobClient.SetAccessTier(ctx, accessTier, nil)
	if err != nil {
		return fmt.Errorf("failed to archive blob: %w", err)
	}

	return nil
}

// RestoreFromArchiveTier rehydrates blob from Archive tier
func (a *AzureBlobProvider) RestoreFromArchiveTier(ctx context.Context, key string, tier azblob.AccessTier, rehydratePriority azblob.RehydratePriority) error {
	a.logger.Info("Restoring blob from Archive",
		zap.String("blob", key),
		zap.String("tier", string(tier)))

	blockBlobClient := a.client.ServiceClient().
		NewContainerClient(a.containerName).
		NewBlockBlobClient(key)

	_, err := blockBlobClient.SetAccessTier(ctx, tier, &azblob.SetAccessTierOptions{
		RehydratePriority: &rehydratePriority,
	})
	if err != nil {
		return fmt.Errorf("failed to restore blob: %w", err)
	}

	return nil
}

// CreateContainer creates a container if it doesn't exist
func (a *AzureBlobProvider) CreateContainer(ctx context.Context) error {
	a.logger.Info("Creating container", zap.String("container", a.containerName))

	containerClient := a.client.ServiceClient().NewContainerClient(a.containerName)
	_, err := containerClient.Create(ctx, nil)
	if err != nil {
		// Container might already exist
		a.logger.Debug("Container may already exist", zap.Error(err))
	}

	return nil
}

// CopyBlob copies a blob within Azure Blob Storage
func (a *AzureBlobProvider) CopyBlob(ctx context.Context, sourceKey, destKey string) error {
	a.logger.Info("Copying blob",
		zap.String("source", sourceKey),
		zap.String("dest", destKey))

	sourceBlobClient := a.client.ServiceClient().
		NewContainerClient(a.containerName).
		NewBlobClient(sourceKey)

	destBlobClient := a.client.ServiceClient().
		NewContainerClient(a.containerName).
		NewBlockBlobClient(destKey)

	sourceURL := sourceBlobClient.URL()
	copyOptions := &azblob.StartCopyFromURLOptions{}
	
	_, err := destBlobClient.StartCopyFromURL(ctx, sourceURL, copyOptions)
	if err != nil {
		return fmt.Errorf("failed to copy blob: %w", err)
	}

	return nil
}

// GenerateSASURL generates a SAS URL for temporary access
func (a *AzureBlobProvider) GenerateSASURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// In production, implement SAS token generation
	// This is a placeholder implementation
	a.logger.Info("Generating SAS URL", zap.String("blob", key), zap.Duration("expiry", expiry))
	return "", fmt.Errorf("SAS URL generation not implemented")
}
