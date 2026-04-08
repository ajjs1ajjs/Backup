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
    private readonly ICloudStorageService _cloudStorage;
    private readonly IEncryptionService _encryption;
    private readonly ILogger<RepositoryService> _logger;

    public RepositoryService(
        BackupDbContext db, 
        ICloudStorageService cloudStorage,
        IEncryptionService encryption,
        ILogger<RepositoryService> logger)
    {
        _db = db;
        _cloudStorage = cloudStorage;
        _encryption = encryption;
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
            var decryptedCreds = string.IsNullOrEmpty(repo.Credentials) 
                ? null 
                : _encryption.Decrypt(repo.Credentials);

            var result = repo.Type switch
            {
                RepositoryType.Local => Directory.Exists(repo.Path),
                RepositoryType.NFS or RepositoryType.SMB => TestNetworkPath(repo.Path),
                RepositoryType.S3 => await _cloudStorage.TestS3ConnectionAsync(repo.Path, decryptedCreds),
                RepositoryType.AzureBlob => await _cloudStorage.TestAzureConnectionAsync(repo.Path, decryptedCreds),
                _ => false
            };

            repo.Status = result ? "online" : "error";
            repo.UpdatedAt = DateTime.UtcNow;
            await _db.SaveChangesAsync();
            return result;
        }
        catch (Exception ex)
        {
            _logger.LogWarning(ex, "Connection test failed for {RepoId}", repositoryId);
            repo.Status = "error";
            await _db.SaveChangesAsync();
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
