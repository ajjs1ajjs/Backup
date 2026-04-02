using Backup.Contracts;
using Google.Protobuf.WellKnownTypes;

namespace Backup.Server.Services;

public class DashboardServiceImpl : DashboardService.DashboardServiceBase
{
    private readonly ILogger<DashboardServiceImpl> _logger;

    public DashboardServiceImpl(ILogger<DashboardServiceImpl> logger)
    {
        _logger = logger;
    }

    public override Task<DashboardStats> GetStats(Empty request, ServerCallContext context)
    {
        return Task.FromResult(new DashboardStats
        {
            TotalJobs = 0,
            RunningJobs = 0,
            FailedJobs = 0,
            TotalBackups = 0,
            TotalRepositories = 0,
            TotalAgents = 0,
            OnlineAgents = 0,
            TotalSizeBytes = 0,
            ProtectedVms = 0,
            SuccessRate = 100.0,
            LastBackupTimestamp = DateTimeOffset.UtcNow.ToUnixTimeSeconds()
        });
    }

    public override Task<ActivityResponse> GetActivity(ActivityRequest request, ServerCallContext context)
    {
        return Task.FromResult(new ActivityResponse { Activities = { } });
    }
}
