using System.Collections.Generic;
using System.Threading.Tasks;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.Services
{
    public class ApiFacadePatch18
    {
        private readonly IApiClient _client;
        public ApiFacadePatch18(IApiClient client) { _client = client; }
        public Task<DashboardStats> GetDashboardStatsAsync() => _client.GetDashboardStatsAsync();
        public Task<List<JobModel>> GetJobsAsync() => _client.GetJobsAsync();
        public Task<List<InfrastructureNode>> GetInfrastructureTreeAsync() => _client.GetInfrastructureTreeAsync();
        public Task<List<RepositoryModel>> GetRepositoriesAsync() => _client.GetRepositoriesAsync();
        public Task<bool> CreateJobAsync(JobModel job) => _client.CreateJobAsync(job);
        public Task<List<RecoverySessionModel>> GetInstantRecoverySessionsAsync() => _client.GetInstantRecoverySessionsAsync();
        public Task<bool> StopInstantRecoveryAsync(string sessionId) => _client.StopInstantRecoveryAsync(sessionId);
    }
}
