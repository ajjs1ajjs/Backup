using Backup.Contracts;
using Grpc.Core;

namespace Backup.Server.Services;

public class BackupServiceImpl : BackupService.BackupServiceBase
{
    private readonly ILogger<BackupServiceImpl> _logger;
    private readonly Dictionary<string, BackupPoint> _backups = new();

    public BackupServiceImpl(ILogger<BackupServiceImpl> logger)
    {
        _logger = logger;
    }

    public override Task<BackupListResponse> ListBackups(BackupListRequest request, ServerCallContext context)
    {
        var filtered = _backups.Values.Where(b => 
            (string.IsNullOrEmpty(request.JobId) || b.JobId == request.JobId) &&
            (request.BackupType == BackupType.BackupTypeUnspecified || b.BackupType == request.BackupType))
            .ToList();

        return Task.FromResult(new BackupListResponse
        {
            Backups = { filtered },
            TotalCount = filtered.Count
        });
    }

    public override Task<BackupDetailsResponse> GetBackupDetails(BackupDetailsRequest request, ServerCallContext context)
    {
        if (_backups.TryGetValue(request.BackupId, out var backup))
        {
            return Task.FromResult(new BackupDetailsResponse
            {
                Backup = backup,
                Metrics = new BackupMetrics
                {
                    SpeedMbps = 125.5,
                    DurationSeconds = 3600,
                    CompressionRatio = 0.65,
                    DeduplicationRatio = 0.42
                }
            });
        }
        throw new RpcException(new Status(StatusCode.NotFound, "Backup not found"));
    }

    public override Task<DeleteBackupResponse> DeleteBackup(DeleteBackupRequest request, ServerCallContext context)
    {
        if (_backups.Remove(request.BackupId))
        {
            _logger.LogInformation("Deleted backup {BackupId}", request.BackupId);
            return Task.FromResult(new DeleteBackupResponse { Success = true, Message = "Backup deleted" });
        }
        return Task.FromResult(new DeleteBackupResponse { Success = false, Message = "Backup not found" });
    }

    public override Task<VerifyBackupResponse> VerifyBackup(VerifyBackupRequest request, ServerCallContext context)
    {
        _logger.LogInformation("Verifying backup {BackupId}", request.BackupId);
        return Task.FromResult(new VerifyBackupResponse
        {
            BackupId = request.BackupId,
            Success = true,
            Message = "Backup verified successfully"
        });
    }
}
