using System.Text.Json;
using Amazon.S3;
using Amazon.S3.Model;
using Amazon.S3.Transfer;
using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Models;
using Google.Cloud.Storage.V1;

namespace Backup.Server.Services;

public interface ICloudStorageService
{
    // S3
    Task<string> UploadToS3Async(string bucket, string key, Stream stream, string? credentials, CancellationToken ct = default);
    Task<Stream> DownloadFromS3Async(string bucket, string key, string? credentials, CancellationToken ct = default);
    Task DeleteFromS3Async(string bucket, string key, string? credentials, CancellationToken ct = default);
    Task<bool> TestS3ConnectionAsync(string bucket, string? credentials, CancellationToken ct = default);

    // Azure Blob
    Task<string> UploadToAzureBlobAsync(string container, string blobName, Stream stream, string? connectionString, CancellationToken ct = default);
    Task<Stream> DownloadFromAzureBlobAsync(string container, string blobName, string? connectionString, CancellationToken ct = default);
    Task DeleteFromAzureBlobAsync(string container, string blobName, string? connectionString, CancellationToken ct = default);
    Task<bool> TestAzureConnectionAsync(string container, string? connectionString, CancellationToken ct = default);
}

public class CloudStorageService : ICloudStorageService
{
    private readonly ILogger<CloudStorageService> _logger;

    public CloudStorageService(ILogger<CloudStorageService> logger)
    {
        _logger = logger;
    }

    #region S3 Implementation
    private AmazonS3Client GetS3Client(string? credentials)
    {
        var creds = string.IsNullOrEmpty(credentials) 
            ? new S3Credentials() 
            : JsonSerializer.Deserialize<S3Credentials>(credentials) ?? new S3Credentials();

        var config = new AmazonS3Config
        {
            RegionEndpoint = Amazon.RegionEndpoint.GetBySystemName(creds.Region ?? "us-east-1")
        };

        if (string.IsNullOrEmpty(creds.AccessKey) || string.IsNullOrEmpty(creds.SecretKey))
        {
            return new AmazonS3Client(config);
        }

        return new AmazonS3Client(creds.AccessKey, creds.SecretKey, config);
    }

    public async Task<string> UploadToS3Async(string bucket, string key, Stream stream, string? credentials, CancellationToken ct = default)
    {
        using var client = GetS3Client(credentials);
        var transferUtility = new TransferUtility(client);
        await transferUtility.UploadAsync(stream, bucket, key, ct);
        return $"s3://{bucket}/{key}";
    }

    public async Task<Stream> DownloadFromS3Async(string bucket, string key, string? credentials, CancellationToken ct = default)
    {
        var client = GetS3Client(credentials);
        var response = await client.GetObjectAsync(bucket, key, ct);
        return response.ResponseStream;
    }

    public async Task DeleteFromS3Async(string bucket, string key, string? credentials, CancellationToken ct = default)
    {
        using var client = GetS3Client(credentials);
        await client.DeleteObjectAsync(bucket, key, ct);
    }

    public async Task<bool> TestS3ConnectionAsync(string bucket, string? credentials, CancellationToken ct = default)
    {
        try
        {
            using var client = GetS3Client(credentials);
            await client.ListObjectsV2Async(new ListObjectsV2Request { BucketName = bucket, MaxKeys = 1 }, ct);
            return true;
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "S3 connection test failed for bucket {Bucket}", bucket);
            return false;
        }
    }
    #endregion

    #region Azure Blob Implementation
    public async Task<string> UploadToAzureBlobAsync(string container, string blobName, Stream stream, string? connectionString, CancellationToken ct = default)
    {
        if (string.IsNullOrEmpty(connectionString)) throw new ArgumentNullException(nameof(connectionString));
        var client = new BlobContainerClient(connectionString, container);
        await client.CreateIfNotExistsAsync(PublicAccessType.None, cancellationToken: ct);
        var blobClient = client.GetBlobClient(blobName);
        await blobClient.UploadAsync(stream, true, ct);
        return $"azure://{container}/{blobName}";
    }

    public async Task<Stream> DownloadFromAzureBlobAsync(string container, string blobName, string? connectionString, CancellationToken ct = default)
    {
        if (string.IsNullOrEmpty(connectionString)) throw new ArgumentNullException(nameof(connectionString));
        var client = new BlobContainerClient(connectionString, container);
        var blobClient = client.GetBlobClient(blobName);
        var response = await blobClient.DownloadStreamingAsync(cancellationToken: ct);
        return response.Value.Content;
    }

    public async Task DeleteFromAzureBlobAsync(string container, string blobName, string? connectionString, CancellationToken ct = default)
    {
        if (string.IsNullOrEmpty(connectionString)) throw new ArgumentNullException(nameof(connectionString));
        var client = new BlobContainerClient(connectionString, container);
        var blobClient = client.GetBlobClient(blobName);
        await blobClient.DeleteIfExistsAsync(DeleteSnapshotsOption.IncludeSnapshots, cancellationToken: ct);
    }

    public async Task<bool> TestAzureConnectionAsync(string container, string? connectionString, CancellationToken ct = default)
    {
        try
        {
            if (string.IsNullOrEmpty(connectionString)) return false;
            var client = new BlobContainerClient(connectionString, container);
            await client.ExistsAsync(ct);
            return true;
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "Azure connection test failed for container {Container}", container);
            return false;
        }
    }
    #endregion
}
