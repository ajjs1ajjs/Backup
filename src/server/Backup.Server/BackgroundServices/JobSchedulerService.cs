using Backup.Server.Database;
using Backup.Server.Database.Entities;
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

                var pendingJobs = await db.Jobs
                    .Where(j => j.Enabled && j.NextRun <= DateTime.UtcNow)
                    .Take(10)
                    .ToListAsync(stoppingToken);

                foreach (var job in pendingJobs)
                {
                    _logger.LogInformation("Triggering job {JobId}: {Name}", job.JobId, job.Name);
                    
                    var runHistory = new JobRunHistory
                    {
                        RunId = Guid.NewGuid().ToString(),
                        JobId = job.JobId,
                        StartTime = DateTime.UtcNow,
                        Status = "running"
                    };
                    db.JobRunHistory.Add(runHistory);
                    
                    job.LastRun = DateTime.UtcNow;
                    if (job.NextRun.HasValue)
                    {
                        job.NextRun = CalculateNextRun(job.NextRun.Value, job.Schedule);
                    }

                    await db.SaveChangesAsync(stoppingToken);
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in job scheduler");
            }

            await Task.Delay(_interval, stoppingToken);
        }
    }

    private DateTime? CalculateNextRun(DateTime lastRun, string? scheduleJson)
    {
        return lastRun.AddHours(24);
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
                var retentionDays = settings.FirstOrDefault(s => s.Key == "backup.retention_days")?.Value ?? "30";
                
                var cutoffDate = DateTime.UtcNow.AddDays(-int.Parse(retentionDays));
                var expiredBackups = await db.BackupPoints
                    .Where(b => b.CreatedAt < cutoffDate && b.Status != "expired")
                    .ToListAsync(stoppingToken);

                foreach (var backup in expiredBackups)
                {
                    _logger.LogInformation("Marking backup {BackupId} as expired", backup.BackupId);
                    backup.Status = "expired";
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
