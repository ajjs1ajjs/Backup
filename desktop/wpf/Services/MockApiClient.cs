using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.Services
{
    // Minimal mock API client for CI & UI validation
    public class MockApiClient : IApiClient
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

        // Job History
        public Task<List<JobHistoryItem>> GetJobHistoryAsync(string jobId) => Task.FromResult(new List<JobHistoryItem>());

        // Backup Sessions
        public Task<List<BackupSession>> GetBackupSessionsAsync() => Task.FromResult(new List<BackupSession>());
        public Task<bool> StopBackupSessionAsync(string sessionId) => Task.FromResult(true);

        // Credentials
        public Task<List<CredentialModel>> GetCredentialsAsync() => Task.FromResult(new List<CredentialModel>());
        public Task<bool> CreateCredentialAsync(CredentialModel credential) => Task.FromResult(true);
        public Task<bool> DeleteCredentialAsync(string id) => Task.FromResult(true);

        // Proxies
        public Task<List<ProxyModel>> GetProxiesAsync() => Task.FromResult(new List<ProxyModel>());
        public Task<bool> CreateProxyAsync(ProxyModel proxy) => Task.FromResult(true);
        public Task<bool> UpdateProxyAsync(ProxyModel proxy) => Task.FromResult(true);
        public Task<bool> DeleteProxyAsync(string id) => Task.FromResult(true);

        // Reports
        public Task<List<ReportModel>> GetReportsAsync() => Task.FromResult(new List<ReportModel>());
        public Task<ReportModel> GetReportAsync(string id) => Task.FromResult(new ReportModel());
        public Task<bool> GenerateReportAsync(ReportRequest request) => Task.FromResult(true);

        // Notifications
        public Task<List<NotificationModel>> GetNotificationsAsync() => Task.FromResult(new List<NotificationModel>());
        public Task<bool> MarkNotificationReadAsync(string id) => Task.FromResult(true);
        public Task<bool> DeleteNotificationAsync(string id) => Task.FromResult(true);

        // Settings
        public Task<AppSettings> GetSettingsAsync() => Task.FromResult(new AppSettings());
        public Task<bool> UpdateSettingsAsync(AppSettings settings) => Task.FromResult(true);

        // Tape
        public Task<List<TapeLibraryModel>> GetTapeLibrariesAsync() => Task.FromResult(new List<TapeLibraryModel>());
        public Task<List<TapeCartridgeModel>> GetTapeCartridgesAsync() => Task.FromResult(new List<TapeCartridgeModel>());
        public Task<List<TapeVaultModel>> GetTapeVaultsAsync() => Task.FromResult(new List<TapeVaultModel>());
        public Task<bool> CreateTapeVaultAsync(TapeVaultModel vault) => Task.FromResult(true);
        public Task<List<TapeJobModel>> GetTapeJobsAsync() => Task.FromResult(new List<TapeJobModel>());
        public Task<bool> CreateTapeJobAsync(TapeJobModel job) => Task.FromResult(true);
        public Task<bool> RunTapeJobAsync(string jobId) => Task.FromResult(true);

        // VSS
        public Task<List<VSSWriterModel>> GetVSSWritersAsync() => Task.FromResult(new List<VSSWriterModel>());

        // Replication
        public Task<List<ReplicationJobModel>> GetReplicationJobsAsync() => Task.FromResult(new List<ReplicationJobModel>
        {
            new ReplicationJobModel { Id = "rep1", Name = "WebServer-Replica", SourceVM = "WebServer-Prod", TargetHost = "192.168.1.20", Status = "Running", Progress = 45 },
            new ReplicationJobModel { Id = "rep2", Name = "DBServer-Replica", SourceVM = "DBServer-Prod", TargetHost = "192.168.1.21", Status = "Idle", Progress = 0 },
            new ReplicationJobModel { Id = "rep3", Name = "AppServer-Replica", SourceVM = "AppServer-Prod", TargetHost = "192.168.1.22", Status = "Success", Progress = 100 }
        });

        // RBAC - Users
        public Task<List<UserModel>> GetUsersAsync() => Task.FromResult(new List<UserModel>
        {
            new UserModel { Id = "user1", Username = "admin", Email = "admin@novabackup.local", RoleIds = new List<string> { "role_admin" }, Active = true },
            new UserModel { Id = "user2", Username = "operator", Email = "operator@novabackup.local", RoleIds = new List<string> { "role_operator" }, Active = true },
            new UserModel { Id = "user3", Username = "viewer", Email = "viewer@novabackup.local", RoleIds = new List<string> { "role_viewer" }, Active = false }
        });

        public Task<bool> CreateUserAsync(UserModel user) => Task.FromResult(true);
        public Task<bool> UpdateUserAsync(UserModel user) => Task.FromResult(true);
        public Task<bool> DeleteUserAsync(string id) => Task.FromResult(true);
        public Task<bool> AssignRoleToUserAsync(string userId, string roleId) => Task.FromResult(true);

        // RBAC - Roles
        public Task<List<RoleModel>> GetRolesAsync() => Task.FromResult(new List<RoleModel>
        {
            new RoleModel { Id = "role_admin", Name = "Administrator", Description = "Full system access", Permissions = new List<string> { "job:create", "job:read", "job:update", "job:delete", "backup:create", "backup:read", "system:admin" } },
            new RoleModel { Id = "role_operator", Name = "Backup Operator", Description = "Can manage backups", Permissions = new List<string> { "job:create", "job:read", "backup:create", "backup:read" } },
            new RoleModel { Id = "role_viewer", Name = "Viewer", Description = "Read-only access", Permissions = new List<string> { "job:read", "backup:read" } }
        });

        public Task<bool> CreateRoleAsync(RoleModel role) => Task.FromResult(true);
        public Task<bool> UpdateRoleAsync(RoleModel role) => Task.FromResult(true);
        public Task<bool> DeleteRoleAsync(string id) => Task.FromResult(true);

        // RBAC - Permissions
        public Task<List<string>> GetPermissionsAsync() => Task.FromResult(new List<string>
        {
            "job:create", "job:read", "job:update", "job:delete", "job:execute",
            "backup:create", "backup:read", "backup:delete", "backup:restore",
            "vm:read", "vm:snapshot", "vm:restore",
            "storage:read", "storage:write", "storage:delete",
            "replication:read", "replication:write", "replication:delete",
            "monitoring:read", "monitoring:admin",
            "user:create", "user:read", "user:update", "user:delete",
            "system:config", "system:logs", "system:admin"
        });
    }
}
