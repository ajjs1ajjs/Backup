using Backup.Server.Database;
using Backup.Server.Database.Entities;
using Backup.Server.Services;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.BackgroundServices;

public class JobSchedulerService : BackgroundService
{
    private readonly IServiceProvider _services;
    private readonly ILogger<JobSchedulerService> _logger;
    private readonly TimeSpan _interval = TimeSpan.FromSeconds(30);

    public JobSchedulerService(IServiceProvider services, ILogger<JobSchedulerService> logger)
    {
        _services = services;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Job Scheduler started");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                using var scope = _services.CreateScope();
                var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();
                var scheduler = scope.ServiceProvider.GetRequiredService<SchedulerService>();
                var backupService = scope.ServiceProvider.GetRequiredService<BackupExecutionService>();
                var backupQueue = scope.ServiceProvider.GetRequiredService<IBackupQueue>();

                var pendingJobs = await db.Jobs
                    .Where(j => j.Enabled && j.NextRun <= DateTime.UtcNow)
                    .Take(10)
                    .ToListAsync(stoppingToken);

                foreach (var job in pendingJobs)
                {
                    _logger.LogInformation("Scheduling job {JobId}: {Name}", job.JobId, job.Name);
                    job.LastRun = DateTime.UtcNow;
                    job.NextRun = scheduler.CalculateNextRun(job, job.LastRun);
                    await db.SaveChangesAsync(stoppingToken);

                    var queueResult = await backupService.QueueJobAsync(job.JobId, stoppingToken);
                    if (queueResult.Success)
                    {
                        await backupQueue.QueueAsync(queueResult.RunId, stoppingToken);
                    }
                    else
                    {
                        _logger.LogWarning("Failed to queue scheduled job {JobId}: {Message}", job.JobId, queueResult.Message);
                    }
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in job scheduler");
            }

            await Task.Delay(_interval, stoppingToken);
        }
    }

}

public class AgentHealthCheckService : BackgroundService
{
    private readonly IServiceProvider _services;
    private readonly ILogger<AgentHealthCheckService> _logger;
    private readonly TimeSpan _interval = TimeSpan.FromMinutes(1);

    public AgentHealthCheckService(IServiceProvider services, ILogger<AgentHealthCheckService> logger)
    {
        _services = services;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Agent Health Check started");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                using var scope = _services.CreateScope();
                var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();

                var timeout = DateTime.UtcNow.AddMinutes(-5);
                var offlineAgents = await db.Agents
                    .Where(a => a.LastHeartbeat < timeout && a.Status != "offline")
                    .ToListAsync(stoppingToken);

                foreach (var agent in offlineAgents)
                {
                    _logger.LogWarning("Agent {AgentId} ({Hostname}) is offline", agent.AgentId, agent.Hostname);
                    agent.Status = "offline";
                }

                if (offlineAgents.Any())
                {
                    await db.SaveChangesAsync(stoppingToken);
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in agent health check");
            }

            await Task.Delay(_interval, stoppingToken);
        }
    }
}

public class RetentionPolicyService : BackgroundService
{
    private readonly IServiceProvider _services;
    private readonly ILogger<RetentionPolicyService> _logger;
    private readonly TimeSpan _interval = TimeSpan.FromHours(1);

    public RetentionPolicyService(IServiceProvider services, ILogger<RetentionPolicyService> logger)
    {
        _services = services;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Retention Policy Service started");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                using var scope = _services.CreateScope();
                var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();

                var settings = await db.Settings.ToListAsync(stoppingToken);
                var retentionDaysValue = settings.FirstOrDefault(s => s.Key == "backup.retention_days")?.Value ?? "30";
                var retentionDays = int.TryParse(retentionDaysValue, out var parsedRetentionDays) ? parsedRetentionDays : 30;

                var cutoffDate = DateTime.UtcNow.AddDays(-retentionDays);
                var expiredBackups = await db.BackupPoints
                    .Where(b => b.CreatedAt < cutoffDate && b.Status != BackupStatus.Expired)
                    .ToListAsync(stoppingToken);

                foreach (var backup in expiredBackups)
                {
                    try 
                    {
                        _logger.LogInformation("Deleting expired backup {BackupId} from {Path}", backup.BackupId, backup.FilePath);
                        
                        var repository = await db.Repositories.FirstOrDefaultAsync(r => r.RepositoryId == backup.RepositoryId, stoppingToken);
                        if (repository != null && !string.IsNullOrEmpty(backup.FilePath))
                        {
                            if (repository.Type == RepositoryType.Local && File.Exists(backup.FilePath))
                            {
                                File.Delete(backup.FilePath);
                            }
                            else if (repository.Type == RepositoryType.S3)
                            {
                                var cloudStorage = scope.ServiceProvider.GetRequiredService<ICloudStorageService>();
                                var encryption = scope.ServiceProvider.GetRequiredService<IEncryptionService>();
                                var bucket = repository.Path;
                                var key = backup.FilePath.Replace($"s3://{bucket}/", string.Empty);
                                var creds = string.IsNullOrEmpty(repository.Credentials) ? null : encryption.Decrypt(repository.Credentials);
                                await cloudStorage.DeleteFromS3Async(bucket, key, creds, stoppingToken);
                            }
                            else if (repository.Type == RepositoryType.AzureBlob)
                            {
                                var cloudStorage = scope.ServiceProvider.GetRequiredService<ICloudStorageService>();
                                var encryption = scope.ServiceProvider.GetRequiredService<IEncryptionService>();
                                var container = repository.Path;
                                var blobName = backup.FilePath.Replace($"azure://{container}/", string.Empty);
                                var connectionString = string.IsNullOrEmpty(repository.Credentials) ? null : encryption.Decrypt(repository.Credentials);
                                await cloudStorage.DeleteFromAzureBlobAsync(container, blobName, connectionString, stoppingToken);
                            }
                        }

                        backup.Status = BackupStatus.Expired;
                    }
                    catch (Exception ex)
                    {
                        _logger.LogError(ex, "Failed to delete backup file for {BackupId}", backup.BackupId);
                    }
                }

                if (expiredBackups.Any())
                {
                    await db.SaveChangesAsync(stoppingToken);
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in retention policy service");
            }

            await Task.Delay(_interval, stoppingToken);
        }
    }
}
