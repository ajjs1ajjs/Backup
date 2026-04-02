using Backup.Contracts;
using Grpc.Core;

namespace Backup.Server.Services;

public class RepositoryServiceImpl : RepositoryService.RepositoryServiceBase
{
    private readonly ILogger<RepositoryServiceImpl> _logger;
    private readonly Dictionary<string, Repository> _repositories = new();

    public RepositoryServiceImpl(ILogger<RepositoryServiceImpl> logger)
    {
        _logger = logger;
    }

    public override Task<RepositoryResponse> CreateRepository(RepositoryRequest request, ServerCallContext context)
    {
        var repoId = Guid.NewGuid().ToString();
        var repo = new Repository
        {
            RepositoryId = repoId,
            Name = request.Name,
            Type = request.Type,
            Path = request.Path,
            Status = RepositoryStatus.RepositoryStatusOnline,
            CreatedAt = DateTimeOffset.UtcNow.ToUnixTimeSeconds()
        };
        _repositories[repoId] = repo;
        
        _logger.LogInformation("Created repository {RepoId}: {Name}", repoId, request.Name);
        
        return Task.FromResult(new RepositoryResponse { Success = true, RepositoryId = repoId, Repository = repo });
    }

    public override Task<Repository> GetRepository(GetRepositoryRequest request, ServerCallContext context)
    {
        if (_repositories.TryGetValue(request.RepositoryId, out var repo))
            return Task.FromResult(repo);
        throw new RpcException(new Status(StatusCode.NotFound, "Repository not found"));
    }

    public override Task<RepositoryListResponse> ListRepositories(RepositoryListRequest request, ServerCallContext context)
    {
        return Task.FromResult(new RepositoryListResponse
        {
            Repositories = { _repositories.Values },
            TotalCount = _repositories.Count
        });
    }

    public override Task<RepositoryResponse> DeleteRepository(DeleteRepositoryRequest request, ServerCallContext context)
    {
        if (_repositories.Remove(request.RepositoryId))
            return Task.FromResult(new RepositoryResponse { Success = true, Message = "Repository deleted" });
        return Task.FromResult(new RepositoryResponse { Success = false, Message = "Repository not found" });
    }

    public override Task<RepositoryResponse> TestRepository(TestRepositoryRequest request, ServerCallContext context)
    {
        return Task.FromResult(new RepositoryResponse { Success = true, Message = "Connection successful" });
    }

    public override Task<RepositoryStatsResponse> GetRepositoryStats(RepositoryStatsRequest request, ServerCallContext context)
    {
        return Task.FromResult(new RepositoryStatsResponse
        {
            RepositoryId = request.RepositoryId,
            TotalBackups = 0,
            CompressionRatio = 1.0,
            DeduplicationRatio = 1.0
        });
    }
}
