using Backup.Contracts;
using Grpc.Core;

namespace Backup.Server.Services;

public class RestoreServiceImpl : RestoreService.RestoreServiceBase
{
    private readonly ILogger<RestoreServiceImpl> _logger;
    private readonly Dictionary<string, RestoreProgress> _restores = new();

    public RestoreServiceImpl(ILogger<RestoreServiceImpl> logger)
    {
        _logger = logger;
    }

    public override Task<RestoreResponse> StartRestore(RestoreRequest request, ServerCallContext context)
    {
        var restoreId = Guid.NewGuid().ToString();
        _restores[restoreId] = new RestoreProgress
        {
            RestoreId = restoreId,
            Status = RestoreStatus.RestoreStatusPending,
            TotalBytes = 0
        };
        
        _logger.LogInformation("Started restore {RestoreId}", restoreId);
        
        return Task.FromResult(new RestoreResponse
        {
            RestoreId = restoreId,
            Success = true,
            Message = "Restore started",
            Status = RestoreStatus.RestoreStatusPending
        });
    }

    public override Task<RestoreProgress> GetRestoreProgress(RestoreProgressRequest request, ServerCallContext context)
    {
        if (_restores.TryGetValue(request.RestoreId, out var progress))
            return Task.FromResult(progress);
        throw new RpcException(new Status(StatusCode.NotFound, "Restore not found"));
    }

    public override Task<RestoreResponse> CancelRestore(CancelRestoreRequest request, ServerCallContext context)
    {
        if (_restores.TryGetValue(request.RestoreId, out var progress))
        {
            progress.Status = RestoreStatus.RestoreStatusCancelled;
            return Task.FromResult(new RestoreResponse { Success = true, Message = "Restore cancelled" });
        }
        return Task.FromResult(new RestoreResponse { Success = false, Message = "Restore not found" });
    }

    public override Task<RestoreListResponse> ListRestores(RestoreListRequest request, ServerCallContext context)
    {
        return Task.FromResult(new RestoreListResponse
        {
            Restores = { _restores.Values.Select(r => new RestoreInfo { RestoreId = r.RestoreId, Status = r.Status }) },
            TotalCount = _restores.Count
        });
    }

    public override Task<BrowseBackupFilesResponse> BrowseBackupFiles(BrowseBackupFilesRequest request, ServerCallContext context)
    {
        return Task.FromResult(new BrowseBackupFilesResponse
        {
            CurrentPath = request.Path,
            HasMore = false
        });
    }
}
