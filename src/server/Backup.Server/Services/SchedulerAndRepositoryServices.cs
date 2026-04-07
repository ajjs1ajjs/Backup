using System.Text.Json;
using Amazon.S3;
using Amazon.S3.Model;
using Azure.Storage.Blobs;
using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Cronos;
using Google.Cloud.Storage.V1;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.Services;

public class SchedulerService
{
    private readonly ILogger<SchedulerService> _logger;
    private static readonly JsonSerializerOptions ScheduleJsonOptions = new()
    {
        PropertyNameCaseInsensitive = true
    };

    public SchedulerService(ILogger<SchedulerService> logger)
    {
        _logger = logger;
    }

    public DateTime? CalculateNextRun(Job job, DateTime? referenceTime = null)
    {
        if (string.IsNullOrWhiteSpace(job.Schedule))
        {
            return null;
        }

        var now = referenceTime ?? DateTime.UtcNow;

        try
        {
            var cron = CronExpression.Parse(job.Schedule);
            return cron.GetNextOccurrence(now, TimeZoneInfo.Utc);
        }
        catch
        {
            try
            {
                var config = JsonSerializer.Deserialize<ScheduleConfig>(job.Schedule, ScheduleJsonOptions);
                if (!string.IsNullOrWhiteSpace(config?.Cron))
                {
                    var cron = CronExpression.Parse(config.Cron);
                    return cron.GetNextOccurrence(now, TimeZoneInfo.Utc);
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
        if (string.IsNullOrWhiteSpace(timeWindow))
        {
            return true;
        }

        try
        {
            var config = JsonSerializer.Deserialize<TimeWindowConfig>(timeWindow, ScheduleJsonOptions);
            if (config == null || !config.Enabled)
            {
                return true;
            }

            var currentHour = DateTime.Now.Hour;
            return config.StartHour < config.EndHour
                ? currentHour >= config.StartHour && currentHour < config.EndHour
                : currentHour >= config.StartHour || currentHour < config.EndHour;
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
        if (repo == null)
        {
            return false;
        }

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

    public async Task<long> GetAvailableSpaceAsync(string repositoryId)
    {
        var repo = await _db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == repositoryId);
        if (repo == null)
        {
            return 0;
        }

        return Math.Max(0, (repo.CapacityBytes ?? 0) - repo.UsedBytes);
    }

    public async Task<List<string>> GetExpiredBackupsAsync(string repositoryId, RetentionPolicy policy)
    {
        var cutoffDates = new Dictionary<string, DateTime>
        {
            ["daily"] = DateTime.UtcNow.AddDays(-policy.Daily),
            ["weekly"] = DateTime.UtcNow.AddDays(-policy.Weekly * 7),
            ["monthly"] = DateTime.UtcNow.AddDays(-policy.Monthly * 30),
            ["yearly"] = DateTime.UtcNow.AddDays(-policy.Yearly * 365)
        };

        var backups = await _db.BackupPoints
            .Where(b => b.RepositoryId == repositoryId && b.Status != BackupStatus.Expired)
            .ToListAsync();

        return backups
            .Where(backup =>
            {
                var metadata = JsonSerializer.Deserialize<Dictionary<string, string>>(backup.Metadata ?? "{}");
                var retentionType = metadata?.GetValueOrDefault("retention", "daily") ?? "daily";
                return backup.CreatedAt < cutoffDates.GetValueOrDefault(retentionType, cutoffDates["daily"]);
            })
            .Select(b => b.BackupId)
            .ToList();
    }

    public async Task UpdateStorageMetricsAsync(string repositoryId)
    {
        var repo = await _db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == repositoryId);
        if (repo == null)
        {
            return;
        }

        if (repo.Type == RepositoryType.Local && Directory.Exists(repo.Path))
        {
            var root = Path.GetPathRoot(repo.Path);
            if (!string.IsNullOrWhiteSpace(root))
            {
                var driveInfo = new DriveInfo(root);
                repo.CapacityBytes = driveInfo.TotalSize;
                repo.UsedBytes = driveInfo.TotalSize - driveInfo.AvailableFreeSpace;
            }
        }

        repo.UpdatedAt = DateTime.UtcNow;
        await _db.SaveChangesAsync();
    }

    private static bool TestNetworkPath(string path)
    {
        try
        {
            return new DirectoryInfo(path).Exists;
        }
        catch
        {
            return false;
        }
    }

    private static async Task<bool> TestS3Async(string bucket, string? credentials)
    {
        var creds = JsonSerializer.Deserialize<S3Credentials>(credentials ?? "{}") ?? new S3Credentials();
        var config = new AmazonS3Config
        {
            RegionEndpoint = Amazon.RegionEndpoint.GetBySystemName(creds.Region ?? "us-east-1")
        };

        using var client = new AmazonS3Client(creds.AccessKey, creds.SecretKey, config);
        await client.ListObjectsV2Async(new ListObjectsV2Request { BucketName = bucket, MaxKeys = 1 });
        return true;
    }

    private static async Task<bool> TestAzureBlobAsync(string container, string? connectionString)
    {
        if (string.IsNullOrWhiteSpace(connectionString))
        {
            return false;
        }

        var client = new BlobContainerClient(connectionString, container);
        return await client.ExistsAsync();
    }

    private static async Task<bool> TestGcsAsync(string bucket, string? credentialsJson)
    {
        StorageClient client;
        if (string.IsNullOrWhiteSpace(credentialsJson))
        {
            client = await StorageClient.CreateAsync();
        }
        else
        {
            using var stream = new MemoryStream(System.Text.Encoding.UTF8.GetBytes(credentialsJson));
            client = await StorageClient.CreateAsync(Google.Apis.Auth.OAuth2.GoogleCredential.FromStream(stream));
        }

        await client.GetBucketAsync(bucket);
        return true;
    }
}

public class CloudStorageService
{
    public async Task<string> UploadToS3Async(string bucket, string key, byte[] data, string? credentials)
    {
        var creds = JsonSerializer.Deserialize<S3Credentials>(credentials ?? "{}") ?? new S3Credentials();
        var config = new AmazonS3Config
        {
            RegionEndpoint = Amazon.RegionEndpoint.GetBySystemName(creds.Region ?? "us-east-1")
        };

        using var client = new AmazonS3Client(creds.AccessKey, creds.SecretKey, config);
        using var stream = new MemoryStream(data);
        await client.PutObjectAsync(new PutObjectRequest
        {
            BucketName = bucket,
            Key = key,
            InputStream = stream
        });

        return $"s3://{bucket}/{key}";
    }

    public async Task<byte[]> DownloadFromS3Async(string bucket, string key, string? credentials)
    {
        var creds = JsonSerializer.Deserialize<S3Credentials>(credentials ?? "{}") ?? new S3Credentials();
        var config = new AmazonS3Config
        {
            RegionEndpoint = Amazon.RegionEndpoint.GetBySystemName(creds.Region ?? "us-east-1")
        };

        using var client = new AmazonS3Client(creds.AccessKey, creds.SecretKey, config);
        using var response = await client.GetObjectAsync(bucket, key);
        using var ms = new MemoryStream();
        await response.ResponseStream.CopyToAsync(ms);
        return ms.ToArray();
    }

    public async Task UploadToAzureBlobAsync(string container, string blobName, byte[] data, string? connectionString)
    {
        if (string.IsNullOrWhiteSpace(connectionString))
        {
            throw new InvalidOperationException("Azure Blob connection string is required");
        }

        var client = new BlobContainerClient(connectionString, container);
        using var stream = new MemoryStream(data);
        await client.UploadBlobAsync(blobName, stream);
    }

    public async Task<byte[]> DownloadFromAzureBlobAsync(string container, string blobName, string? connectionString)
    {
        if (string.IsNullOrWhiteSpace(connectionString))
        {
            throw new InvalidOperationException("Azure Blob connection string is required");
        }

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

public class S3Credentials
{
    public string? AccessKey { get; set; }
    public string? SecretKey { get; set; }
    public string? Region { get; set; }
}

public class ScheduleConfig
{
    public string? Cron { get; set; }
}

public class TimeWindowConfig
{
    public bool Enabled { get; set; }
    public int StartHour { get; set; }
    public int EndHour { get; set; }
}

public class RetentionPolicy
{
    public int Daily { get; set; } = 7;
    public int Weekly { get; set; } = 4;
    public int Monthly { get; set; } = 12;
    public int Yearly { get; set; } = 3;
}
