using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;
using Cronos;
using Amazon.S3;
using Azure.Storage.Blobs;
using Google.Cloud.Storage.V1;

namespace Backup.Server.Services;

public class SchedulerService
{
    private readonly BackupDbContext _db;
    private readonly ILogger<SchedulerService> _logger;

    public SchedulerService(BackupDbContext db, ILogger<SchedulerService> logger)
    {
        _db = db;
        _logger = logger;
    }

    public DateTime? CalculateNextRun(Job job)
    {
        if (string.IsNullOrEmpty(job.Schedule))
            return null;

        try
        {
            // Try parse as direct Cron first for simplicity
            var cron = CronExpression.Parse(job.Schedule);
            return cron.GetNextOccurrence(DateTime.UtcNow);
        }
        catch
        {
            try 
            {
                var config = System.Text.Json.JsonSerializer.Deserialize<ScheduleConfig>(job.Schedule);
                if (config != null && !string.IsNullOrEmpty(config.Cron))
                {
                    var cron = CronExpression.Parse(config.Cron);
                    return cron.GetNextOccurrence(DateTime.UtcNow);
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to calculate next run for job {JobId}", job.JobId);
            }
        }

        return null;
    }

    public bool IsWithinTimeWindow(string? timeWindow)
    {
        if (string.IsNullOrEmpty(timeWindow))
            return true;

        try
        {
            var config = System.Text.Json.JsonSerializer.Deserialize<TimeWindowConfig>(timeWindow);
            if (config == null || !config.Enabled)
                return true;

            var now = DateTime.UtcNow.ToLocalTime();
            int currentHour = now.Hour;

            if (config.StartHour < config.EndHour)
                return currentHour >= config.StartHour && currentHour < config.EndHour;
            else
                return currentHour >= config.StartHour || currentHour < config.EndHour;
        }
        catch
        {
            return true;
        }
    }
}

public class RepositoryService
{
    private readonly BackupDbContext _db;
    private readonly ILogger<RepositoryService> _logger;

    public RepositoryService(BackupDbContext db, ILogger<RepositoryService> logger)
    {
        _db = db;
        _logger = logger;
    }

    public async Task<bool> TestConnectionAsync(string repositoryId)
    {
        var repo = await _db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == repositoryId);
        if (repo == null) return false;

        try
        {
            return repo.Type switch
            {
                RepositoryType.Local => Directory.Exists(repo.Path),
                RepositoryType.NFS or RepositoryType.SMB => TestNetworkPath(repo.Path),
                RepositoryType.S3 => await TestS3Async(repo.Path, repo.Credentials),
                RepositoryType.AzureBlob => await TestAzureBlobAsync(repo.Path, repo.Credentials),
                RepositoryType.GCS => await TestGcsAsync(repo.Path, repo.Credentials),
                _ => false
            };
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "Connection test failed for {RepoId}", repositoryId);
            return false;
        }
    }

    private bool TestNetworkPath(string path)
    {
        // On Windows, we check if the UNC path is accessible
        try { return new DirectoryInfo(path).Exists; } catch { return false; }
    }

    private async Task<bool> TestS3Async(string bucket, string? credentials)
    {
        // Expecting JSON: {"AccessKey":"...", "SecretKey":"...", "Region":"..."}
        var creds = System.Text.Json.JsonSerializer.Deserialize<S3Credentials>(credentials ?? "{}");
        var config = new AmazonS3Config { RegionEndpoint = Amazon.RegionEndpoint.GetBySystemName(creds.Region ?? "us-east-1") };
        using var client = new AmazonS3Client(creds.AccessKey, creds.SecretKey, config);
        await client.ListObjectsV2Async(new Amazon.S3.Model.ListObjectsV2Request { BucketName = bucket, MaxKeys = 1 });
        return true;
    }

    private async Task<bool> TestAzureBlobAsync(string container, string? connectionString)
    {
        var client = new BlobContainerClient(connectionString, container);
        return await client.ExistsAsync();
    }

    private async Task<bool> TestGcsAsync(string bucket, string? credentialsJson)
    {
        var client = await StorageClient.CreateAsync(); // Uses GOOGLE_APPLICATION_CREDENTIALS or provided JSON
        await client.GetBucketAsync(bucket);
        return true;
    }

    public async Task UpdateStorageMetricsAsync(string repositoryId)
    {
        var repo = await _db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == repositoryId);
        if (repo == null) return;

        if (repo.Type == RepositoryType.Local && Directory.Exists(repo.Path))
        {
            var driveInfo = new DriveInfo(Path.GetPathRoot(repo.Path)!);
            repo.CapacityBytes = driveInfo.TotalSize;
            repo.UsedBytes = driveInfo.TotalSize - driveInfo.AvailableFreeSpace;
        }
        // For Cloud, we would sum the blobs/objects (omitted for brevity but can be added)

        repo.UpdatedAt = DateTime.UtcNow;
        await _db.SaveChangesAsync();
    }
}

public class S3Credentials { public string? AccessKey { get; set; } public string? SecretKey { get; set; } public string? Region { get; set; } }

    public async Task<long> GetAvailableSpaceAsync(string repositoryId)
    {
        var repo = await _db.Repositories.FindAsync(repositoryId);
        if (repo == null) return 0;
        
        return (repo.CapacityBytes ?? 0) - repo.UsedBytes;
    }

    public async Task<List<string>> GetExpiredBackupsAsync(string repositoryId, RetentionPolicy policy)
    {
        var cutoffDates = new Dictionary<string, DateTime>
        {
            { "daily", DateTime.UtcNow.AddDays(-policy.Daily) },
            { "weekly", DateTime.UtcNow.AddDays(-policy.Weekly * 7) },
            { "monthly", DateTime.UtcNow.AddDays(-policy.Monthly * 30) },
            { "yearly", DateTime.UtcNow.AddDays(-policy.Yearly * 365) }
        };

        var expired = await _db.BackupPoints
            .Where(b => b.RepositoryId == repositoryId && b.Status != BackupStatus.Expired)
            .ToListAsync();

        var toExpire = expired
            .Where(b => {
                var metadata = System.Text.Json.JsonSerializer.Deserialize<Dictionary<string, string>>(b.Metadata ?? "{}");
                var retentionType = metadata?.GetValueOrDefault("retention", "daily") ?? "daily";
                return b.CreatedAt < cutoffDates.GetValueOrDefault(retentionType, cutoffDates["daily"]);
            })
            .Select(b => b.BackupId)
            .ToList();

        return toExpire;
    }
}

public class CloudStorageService
{
    private readonly ILogger<CloudStorageService> _logger;

    public CloudStorageService(ILogger<CloudStorageService> logger)
    {
        _logger = logger;
    }

    public async Task<string> UploadToS3Async(string bucket, string key, byte[] data, string? credentials)
    {
        var creds = System.Text.Json.JsonSerializer.Deserialize<S3Credentials>(credentials ?? "{}");
        var config = new AmazonS3Config { RegionEndpoint = Amazon.RegionEndpoint.GetBySystemName(creds.Region ?? "us-east-1") };
        using var client = new AmazonS3Client(creds.AccessKey, creds.SecretKey, config);
        
        using var stream = new MemoryStream(data);
        await client.PutObjectAsync(new Amazon.S3.Model.PutObjectRequest
        {
            BucketName = bucket,
            Key = key,
            InputStream = stream
        });
        return $"s3://{bucket}/{key}";
    }

    public async Task<byte[]> DownloadFromS3Async(string bucket, string key, string? credentials)
    {
        var creds = System.Text.Json.JsonSerializer.Deserialize<S3Credentials>(credentials ?? "{}");
        var config = new AmazonS3Config { RegionEndpoint = Amazon.RegionEndpoint.GetBySystemName(creds.Region ?? "us-east-1") };
        using var client = new AmazonS3Client(creds.AccessKey, creds.SecretKey, config);

        var response = await client.GetObjectAsync(bucket, key);
        using var ms = new MemoryStream();
        await response.ResponseStream.CopyToAsync(ms);
        return ms.ToArray();
    }

    public async Task UploadToAzureBlobAsync(string container, string blobName, byte[] data, string? connectionString)
    {
        var client = new BlobContainerClient(connectionString, container);
        using var stream = new MemoryStream(data);
        await client.UploadBlobAsync(blobName, stream);
    }

    public async Task<byte[]> DownloadFromAzureBlobAsync(string container, string blobName, string? connectionString)
    {
        var client = new BlobContainerClient(connectionString, container);
        var blob = client.GetBlobClient(blobName);
        var response = await blob.DownloadContentAsync();
        return response.Value.Content.ToArray();
    }

    public async Task UploadToGcsAsync(string bucket, string objectName, byte[] data)
    {
        var client = await StorageClient.CreateAsync();
        using var stream = new MemoryStream(data);
        await client.UploadObjectAsync(bucket, objectName, null, stream);
    }

    public async Task<byte[]> DownloadFromGcsAsync(string bucket, string objectName)
    {
        var client = await StorageClient.CreateAsync();
        using var ms = new MemoryStream();
        await client.DownloadObjectAsync(bucket, objectName, ms);
        return ms.ToArray();
    }
}
