using System.Threading.Channels;
using Backup.Server.Database;
using Backup.Server.Services;
using Microsoft.EntityFrameworkCore;

namespace Backup.Server.BackgroundServices;

public interface IRestoreQueue
{
    ValueTask QueueAsync(string restoreId, CancellationToken cancellationToken = default);
    ValueTask<string> DequeueAsync(CancellationToken cancellationToken);
}

public class RestoreQueue : IRestoreQueue
{
    private readonly Channel<string> _channel = Channel.CreateUnbounded<string>();

    public ValueTask QueueAsync(string restoreId, CancellationToken cancellationToken = default)
        => _channel.Writer.WriteAsync(restoreId, cancellationToken);

    public ValueTask<string> DequeueAsync(CancellationToken cancellationToken)
        => _channel.Reader.ReadAsync(cancellationToken);
}

public class RestoreProcessingService : BackgroundService
{
    private readonly IServiceProvider _services;
    private readonly IRestoreQueue _restoreQueue;
    private readonly ILogger<RestoreProcessingService> _logger;

    public RestoreProcessingService(
        IServiceProvider services,
        IRestoreQueue restoreQueue,
        ILogger<RestoreProcessingService> logger)
    {
        _services = services;
        _restoreQueue = restoreQueue;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("Restore processing service started");

        await RequeuePendingRestoresAsync(stoppingToken);

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                var restoreId = await _restoreQueue.DequeueAsync(stoppingToken);

                using var scope = _services.CreateScope();
                var restoreService = scope.ServiceProvider.GetRequiredService<RestoreService>();
                var result = await restoreService.ExecuteRestoreAsync(restoreId, stoppingToken);

                if (!result.Success)
                {
                    _logger.LogWarning("Restore {RestoreId} finished unsuccessfully: {Message}", restoreId, result.Message);
                }
            }
            catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
            {
                break;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error while processing restore queue");
            }
        }
    }

    private async Task RequeuePendingRestoresAsync(CancellationToken cancellationToken)
    {
        using var scope = _services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<BackupDbContext>();

        var pendingRestoreIds = await db.Restores
            .Where(r => r.Status == "queued" || r.Status == "running")
            .OrderBy(r => r.CreatedAt)
            .Select(r => r.RestoreId)
            .ToListAsync(cancellationToken);

        foreach (var restoreId in pendingRestoreIds)
        {
            await _restoreQueue.QueueAsync(restoreId, cancellationToken);
        }

        if (pendingRestoreIds.Count > 0)
        {
            _logger.LogInformation("Re-queued {Count} restore task(s) on startup", pendingRestoreIds.Count);
        }
    }
}
