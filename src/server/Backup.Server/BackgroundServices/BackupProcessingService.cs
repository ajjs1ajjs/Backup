using System.Threading.Channels;
using Backup.Server.Database;
using Backup.Server.Services;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.BackgroundServices;

public interface IBackupQueue
{
    ValueTask QueueAsync(string runId, CancellationToken cancellationToken = default);
    ValueTask<string> DequeueAsync(CancellationToken cancellationToken);
}

public class BackupQueue : IBackupQueue
{
    private readonly Channel<string> _channel = Channel.CreateUnbounded<string>();

    public ValueTask QueueAsync(string runId, CancellationToken cancellationToken = default)
        => _channel.Writer.WriteAsync(runId, cancellationToken);

    public ValueTask<string> DequeueAsync(CancellationToken cancellationToken)
        => _channel.Reader.ReadAsync(cancellationToken);
}

public class BackupProcessingService : BackgroundService
{
    private readonly IServiceProvider _services;
    private readonly IBackupQueue _backupQueue;
    private readonly ILogger<BackupProcessingService> _logger;

    public BackupProcessingService(
        IServiceProvider services,
        IBackupQueue backupQueue,
        ILogger<BackupProcessingService> logger)
    {
        _services = services;
        _backupQueue = backupQueue;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Backup processing service started");

        await RequeuePendingRunsAsync(stoppingToken);

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                var runId = await _backupQueue.DequeueAsync(stoppingToken);

                using var scope = _services.CreateScope();
                var backupService = scope.ServiceProvider.GetRequiredService<BackupExecutionService>();
                var result = await backupService.ExecuteRunAsync(runId, stoppingToken);

                if (!result.Success)
                {
                    _logger.LogWarning("Backup run {RunId} finished unsuccessfully: {Message}", runId, result.Message);
                }
            }
            catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
            {
                break;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error while processing backup queue");
            }
        }
    }

    private async Task RequeuePendingRunsAsync(CancellationToken cancellationToken)
    {
        using var scope = _services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();

        var pendingRunIds = await db.JobRunHistory
            .Where(r => r.Status == "queued" || r.Status == "running")
            .OrderBy(r => r.StartTime)
            .Select(r => r.RunId)
            .ToListAsync(cancellationToken);

        foreach (var runId in pendingRunIds)
        {
            await _backupQueue.QueueAsync(runId, cancellationToken);
        }

        if (pendingRunIds.Count > 0)
        {
            _logger.LogInformation("Re-queued {Count} backup run(s) on startup", pendingRunIds.Count);
        }
    }
}
