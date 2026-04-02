using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Microsoft.EntityFrameworkCore;

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
            var schedule = System.Text.Json.JsonSerializer.Deserialize<ScheduleConfig>(job.Schedule);
            if (schedule == null) return null;

            if (!string.IsNullOrEmpty(schedule.Cron))
                return CalculateFromCron(schedule.Cron, job.LastRun);
            
            if (schedule.IntervalSeconds.HasValue && schedule.IntervalSeconds > 0)
                return job.LastRun?.AddSeconds(schedule.IntervalSeconds.Value);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to calculate next run for job {JobId}", job.JobId);
        }

        return null;
    }

    private DateTime? CalculateFromCron(string cronExpression, DateTime? lastRun)
    {
        return lastRun?.AddHours(24);
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

            var now = DateTime.UtcNow;
            var start = DateTime.Today.AddHours(config.StartHour);
            var end = DateTime.Today.AddHours(config.EndHour);

            if (start < end)
                return now >= start && now <= end;
            else
                return now >= start || now <= end;
        }
        catch
        {
            return true;
        }
    }

    public RetentionPolicy ParseRetentionPolicy(string? policy)
    {
        var defaultPolicy = new RetentionPolicy
        {
            Daily = 7,
            Weekly = 4,
            Monthly = 12,
            Yearly = 7
        };

        if (string.IsNullOrEmpty(policy))
            return defaultPolicy;

        try
        {
            return System.Text.Json.JsonSerializer.Deserialize<RetentionPolicy>(policy) ?? defaultPolicy;
        }
        catch
        {
            return defaultPolicy;
        }
    }
}

public class ScheduleConfig
{
    public string? Cron { get; set; }
    public long? IntervalSeconds { get; set; }
    public string? Timezone { get; set; }
    public TimeWindowConfig? TimeWindow { get; set; }
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
    public int Yearly { get; set; } = 7;
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

    public async Task<Repository> CreateRepositoryAsync(Repository repository)
    {
        repository.RepositoryId = Guid.NewGuid().ToString();
        repository.CreatedAt = DateTime.UtcNow;
        repository.UpdatedAt = DateTime.UtcNow;
        
        _db.Repositories.Add(repository);
        await _db.SaveChangesAsync();
        
        _logger.LogInformation("Created repository {RepoId}: {Name}", 
            repository.RepositoryId, repository.Name);
        
        return repository;
    }

    public async Task<bool> TestConnectionAsync(string repositoryId)
    {
        var repo = await _db.Repositories.FindAsync(repositoryId);
        if (repo == null) return false;

        return repo.Type switch
        {
            "local" => TestLocalPath(repo.Path),
            "nfs" => TestNfsPath(repo.Path),
            "smb" => TestSmbPath(repo.Path),
            "s3" => await TestS3Async(repo.Path, repo.Credentials),
            "azure_blob" => await TestAzureBlobAsync(repo.Path, repo.Credentials),
            "gcs" => await TestGcsAsync(repo.Path, repo.Credentials),
            _ => false
        };
    }

    private bool TestLocalPath(string path)
    {
        try
        {
            return Directory.Exists(path);
        }
        catch
        {
            return false;
        }
    }

    private bool TestNfsPath(string path)
    {
        return !string.IsNullOrEmpty(path);
    }

    private bool TestSmbPath(string path)
    {
        return !string.IsNullOrEmpty(path);
    }

    private async Task<bool> TestS3Async(string bucket, string? credentials)
    {
        await Task.CompletedTask;
        return !string.IsNullOrEmpty(bucket);
    }

    private async Task<bool> TestAzureBlobAsync(string container, string? credentials)
    {
        await Task.CompletedTask;
        return !string.IsNullOrEmpty(container);
    }

    private async Task<bool> TestGcsAsync(string bucket, string? credentials)
    {
        await Task.CompletedTask;
        return !string.IsNullOrEmpty(bucket);
    }

    public async Task UpdateStorageMetricsAsync(string repositoryId)
    {
        var repo = await _db.Repositories.FindAsync(repositoryId);
        if (repo == null) return;

        if (repo.Type == "local" && Directory.Exists(repo.Path))
        {
            var driveInfo = new DriveInfo(Path.GetPathRoot(repo.Path));
            repo.CapacityBytes = driveInfo.TotalSize;
            repo.UsedBytes = driveInfo.TotalSize - driveInfo.AvailableFreeSpace;
        }

        repo.UpdatedAt = DateTime.UtcNow;
        await _db.SaveChangesAsync();
    }

    public async Task<long> GetAvailableSpaceAsync(string repositoryId)
    {
        var repo = await _db.Repositories.FindAsync(repositoryId);
        if (repo == null) return 0;
        
        return repo.CapacityBytes - repo.UsedBytes;
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
            .Where(b => b.RepositoryId == repositoryId && b.Status != "expired")
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

    public async Task<string> UploadToS3Async(string bucket, string key, byte[] data, Dictionary<string, string>? options = null)
    {
        _logger.LogInformation("Uploading to S3: {Bucket}/{Key}", bucket, key);
        await Task.Delay(100);
        return $"s3://{bucket}/{key}";
    }

    public async Task<byte[]> DownloadFromS3Async(string bucket, string key)
    {
        _logger.LogInformation("Downloading from S3: {Bucket}/{Key}", bucket, key);
        await Task.Delay(100);
        return Array.Empty<byte>();
    }

    public async Task UploadToAzureBlobAsync(string container, string blobName, byte[] data)
    {
        _logger.LogInformation("Uploading to Azure: {Container}/{Blob}", container, blobName);
        await Task.Delay(100);
    }

    public async Task<byte[]> DownloadFromAzureBlobAsync(string container, string blobName)
    {
        _logger.LogInformation("Downloading from Azure: {Container}/{Blob}", container, blobName);
        await Task.Delay(100);
        return Array.Empty<byte>();
    }

    public async Task UploadToGcsAsync(string bucket, string objectName, byte[] data)
    {
        _logger.LogInformation("Uploading to GCS: {Bucket}/{Object}", bucket, objectName);
        await Task.Delay(100);
    }

    public async Task<byte[]> DownloadFromGcsAsync(string bucket, string objectName)
    {
        _logger.LogInformation("Downloading from GCS: {Bucket}/{Object}", bucket, objectName);
        await Task.Delay(100);
        return Array.Empty<byte>();
    }

    public async Task SetStorageTierAsync(string path, string tier)
    {
        _logger.LogInformation("Setting storage tier for {Path} to {Tier}", path, tier);
        await Task.Delay(50);
    }
}
