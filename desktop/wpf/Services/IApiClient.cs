using System.Collections.Generic;
using System.Threading.Tasks;
using NovaBackup.GUI.ViewModels;
using NovaBackup.GUI.Models;

namespace NovaBackup.GUI.Services
{
    public interface IApiClient
    {
        Task<DashboardStats> GetDashboardStatsAsync();
        Task<List<JobModel>> GetJobsAsync();
        Task<List<InfrastructureNode>> GetInfrastructureTreeAsync();
        Task<List<RepositoryModel>> GetRepositoriesAsync();
        Task<bool> CreateRepositoryAsync(RepositoryModel repo);
        Task<bool> AddServerAsync(string name, string address, string type, string username, string password);
        Task<bool> DiscoverNodeAsync(string nodeId);
        Task<List<InfrastructureObject>> GetDiscoveredObjectsAsync(string nodeId);
        Task<bool> RunJobAsync(string jobId);
        Task<bool> CreateJobAsync(JobModel job);
        Task<List<RestorePointModel>> GetRestorePointsAsync(string jobId);
        Task<bool> StartInstantRecoveryAsync(string rpId, string vmName);
        Task<List<RecoverySessionModel>> GetInstantRecoverySessionsAsync();
        Task<bool> StopInstantRecoveryAsync(string sessionId);
    }

    public class DashboardStats
    {
        public string SystemStatus { get; set; } = string.Empty;
        public int ActiveJobs { get; set; }
        public int SuccessJobs { get; set; }
        public int WarningJobs { get; set; }
        public int FailedJobs { get; set; }
        public double StorageUsed { get; set; }
        public double TotalProcessedGB { get; set; }
        public string BackupBottleneck { get; set; } = string.Empty;
        public List<string> RecentActivity { get; set; } = new();
    }
}
