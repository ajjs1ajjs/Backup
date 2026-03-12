using System.Collections.Generic;
using System.Threading.Tasks;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;
using Xunit;

namespace NovaBackup.UI.Tests.Integration
{
    public class RecoverySessionsIntegrationTests
    {
        private class MockApiClientForIntegration : IApiClient
        {
            public Task<DashboardStats> GetDashboardStatsAsync() => Task.FromResult(new DashboardStats());
            public Task<List<JobModel>> GetJobsAsync() => Task.FromResult(new List<JobModel>());
            public Task<List<InfrastructureNode>> GetInfrastructureTreeAsync() => Task.FromResult(new List<InfrastructureNode>());
            public Task<List<RepositoryModel>> GetRepositoriesAsync() => Task.FromResult(new List<RepositoryModel>());
            public Task<bool> CreateRepositoryAsync(RepositoryModel repo) => Task.FromResult(true);
            public Task<bool> AddServerAsync(string name, string address, string type, string username, string password) => Task.FromResult(true);
            public Task<bool> DiscoverNodeAsync(string nodeId) => Task.FromResult(true);
            public Task<List<InfrastructureObject>> GetDiscoveredObjectsAsync(string nodeId) => Task.FromResult(new List<InfrastructureObject>());
            public Task<bool> RunJobAsync(string jobId) => Task.FromResult(true);
            public Task<bool> CreateJobAsync(JobModel job) => Task.FromResult(true);
            public Task<List<RestorePointModel>> GetRestorePointsAsync(string jobId) => Task.FromResult(new List<RestorePointModel>());
            public Task<bool> StartInstantRecoveryAsync(string rpId, string vmName) => Task.FromResult(true);
            public Task<List<RecoverySessionModel>> GetInstantRecoverySessionsAsync() => Task.FromResult(new List<RecoverySessionModel> { new RecoverySessionModel { SessionID = "mock1", VMName = "MockVM", Status = "Running", Progress = 0.2 } });
            public Task<bool> StopInstantRecoveryAsync(string sessionId) => Task.FromResult(true);

            public Task<List<JobHistoryItem>> GetJobHistoryAsync(string jobId) => Task.FromResult(new List<JobHistoryItem>());
            public Task<List<BackupSession>> GetBackupSessionsAsync() => Task.FromResult(new List<BackupSession>());
            public Task<bool> StopBackupSessionAsync(string sessionId) => Task.FromResult(true);
            public Task<List<CredentialModel>> GetCredentialsAsync() => Task.FromResult(new List<CredentialModel>());
            public Task<bool> CreateCredentialAsync(CredentialModel credential) => Task.FromResult(true);
            public Task<bool> DeleteCredentialAsync(string id) => Task.FromResult(true);
            public Task<List<ProxyModel>> GetProxiesAsync() => Task.FromResult(new List<ProxyModel>());
            public Task<bool> CreateProxyAsync(ProxyModel proxy) => Task.FromResult(true);
            public Task<bool> UpdateProxyAsync(ProxyModel proxy) => Task.FromResult(true);
            public Task<bool> DeleteProxyAsync(string id) => Task.FromResult(true);
            public Task<List<ReportModel>> GetReportsAsync() => Task.FromResult(new List<ReportModel>());
            public Task<ReportModel> GetReportAsync(string id) => Task.FromResult(new ReportModel());
            public Task<bool> GenerateReportAsync(ReportRequest request) => Task.FromResult(true);
            public Task<List<NotificationModel>> GetNotificationsAsync() => Task.FromResult(new List<NotificationModel>());
            public Task<bool> MarkNotificationReadAsync(string id) => Task.FromResult(true);
            public Task<bool> DeleteNotificationAsync(string id) => Task.FromResult(true);
            public Task<AppSettings> GetSettingsAsync() => Task.FromResult(new AppSettings());
            public Task<bool> UpdateSettingsAsync(AppSettings settings) => Task.FromResult(true);

            public Task<List<TapeLibraryModel>> GetTapeLibrariesAsync() => Task.FromResult(new List<TapeLibraryModel>());
            public Task<List<TapeCartridgeModel>> GetTapeCartridgesAsync() => Task.FromResult(new List<TapeCartridgeModel>());
            public Task<List<TapeVaultModel>> GetTapeVaultsAsync() => Task.FromResult(new List<TapeVaultModel>());
            public Task<bool> CreateTapeVaultAsync(TapeVaultModel vault) => Task.FromResult(true);
            public Task<List<TapeJobModel>> GetTapeJobsAsync() => Task.FromResult(new List<TapeJobModel>());
            public Task<bool> CreateTapeJobAsync(TapeJobModel job) => Task.FromResult(true);
            public Task<bool> RunTapeJobAsync(string jobId) => Task.FromResult(true);

            public Task<List<VSSWriterModel>> GetVSSWritersAsync() => Task.FromResult(new List<VSSWriterModel>());

            public Task<List<ReplicationJobModel>> GetReplicationJobsAsync() => Task.FromResult(new List<ReplicationJobModel>());

            public Task<List<UserModel>> GetUsersAsync() => Task.FromResult(new List<UserModel>());
            public Task<bool> CreateUserAsync(UserModel user) => Task.FromResult(true);
            public Task<bool> UpdateUserAsync(UserModel user) => Task.FromResult(true);
            public Task<bool> DeleteUserAsync(string id) => Task.FromResult(true);
            public Task<bool> AssignRoleToUserAsync(string userId, string roleId) => Task.FromResult(true);

            public Task<List<RoleModel>> GetRolesAsync() => Task.FromResult(new List<RoleModel>());
            public Task<bool> CreateRoleAsync(RoleModel role) => Task.FromResult(true);
            public Task<bool> UpdateRoleAsync(RoleModel role) => Task.FromResult(true);
            public Task<bool> DeleteRoleAsync(string id) => Task.FromResult(true);

            public Task<List<string>> GetPermissionsAsync() => Task.FromResult(new List<string>());
        }

        [Fact]
        public async Task LoadSessions_WithMockApi_ReturnsNonEmptyList()
        {
            var api = new MockApiClientForIntegration();
            var vm = new NovaBackup.GUI.ViewModels.RecoverySessionsViewModelMVVM(api);
            await vm.LoadAsync();
            Assert.NotNull(vm.Sessions);
        }
    }
}
