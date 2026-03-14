using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using NovaBackup.GUI.ViewModels;
using NovaBackup.GUI.Models;

namespace NovaBackup.GUI.Services
{
    public interface IApiClient
    {
        // Dashboard
        Task<DashboardStats> GetDashboardStatsAsync();

        // Jobs
        Task<List<JobModel>> GetJobsAsync();
        Task<bool> CreateJobAsync(JobModel job);
        Task<List<JobHistoryItem>> GetJobHistoryAsync(string jobId);

        // Infrastructure
        Task<List<InfrastructureNode>> GetInfrastructureTreeAsync();
        Task<bool> AddServerAsync(string name, string address, string type, string username, string password);
        Task<bool> DiscoverNodeAsync(string nodeId);
        Task<List<InfrastructureObject>> GetDiscoveredObjectsAsync(string nodeId);

        // Repositories
        Task<List<RepositoryModel>> GetRepositoriesAsync();
        Task<bool> CreateRepositoryAsync(RepositoryModel repo);

        // Job Execution
        Task<bool> RunJobAsync(string jobId);
        Task<List<BackupSession>> GetBackupSessionsAsync();
        Task<bool> StopBackupSessionAsync(string sessionId);

        // Restore Points
        Task<List<RestorePointModel>> GetRestorePointsAsync(string jobId);

        // Instant Recovery
        Task<bool> StartInstantRecoveryAsync(string rpId, string vmName);
        Task<List<RecoverySessionModel>> GetInstantRecoverySessionsAsync();
        Task<bool> StopInstantRecoveryAsync(string sessionId);

        // Credentials
        Task<List<CredentialModel>> GetCredentialsAsync();
        Task<bool> CreateCredentialAsync(CredentialModel credential);
        Task<bool> DeleteCredentialAsync(string id);

        // Proxies
        Task<List<ProxyModel>> GetProxiesAsync();
        Task<bool> CreateProxyAsync(ProxyModel proxy);
        Task<bool> UpdateProxyAsync(ProxyModel proxy);
        Task<bool> DeleteProxyAsync(string id);

        // Reports
        Task<List<ReportModel>> GetReportsAsync();
        Task<ReportModel> GetReportAsync(string id);
        Task<bool> GenerateReportAsync(ReportRequest request);

        // Notifications
        Task<List<NotificationModel>> GetNotificationsAsync();
        Task<bool> MarkNotificationReadAsync(string id);
        Task<bool> DeleteNotificationAsync(string id);

        // Settings
        Task<AppSettings> GetSettingsAsync();
        Task<bool> UpdateSettingsAsync(AppSettings settings);

        // Tape
        Task<List<TapeLibraryModel>> GetTapeLibrariesAsync();
        Task<List<TapeCartridgeModel>> GetTapeCartridgesAsync();
        Task<List<TapeVaultModel>> GetTapeVaultsAsync();
        Task<bool> CreateTapeVaultAsync(TapeVaultModel vault);
        Task<List<TapeJobModel>> GetTapeJobsAsync();
        Task<bool> CreateTapeJobAsync(TapeJobModel job);
        Task<bool> RunTapeJobAsync(string jobId);

        // VSS
        Task<List<VSSWriterModel>> GetVSSWritersAsync();

        // Replication
        Task<List<ReplicationJobModel>> GetReplicationJobsAsync();

        // RBAC - Users
        Task<List<UserModel>> GetUsersAsync();
        Task<bool> CreateUserAsync(UserModel user);
        Task<bool> UpdateUserAsync(UserModel user);
        Task<bool> DeleteUserAsync(string id);
        Task<bool> AssignRoleToUserAsync(string userId, string roleId);

        // RBAC - Roles
        Task<List<RoleModel>> GetRolesAsync();
        Task<bool> CreateRoleAsync(RoleModel role);
        Task<bool> UpdateRoleAsync(RoleModel role);
        Task<bool> DeleteRoleAsync(string id);

        // RBAC - Permissions
        Task<List<string>> GetPermissionsAsync();

        // Synthetic Backup
        Task<List<SyntheticBackupModel>> GetSyntheticBackupsAsync();
        Task<bool> CreateSyntheticBackupAsync(SyntheticBackupRequest request);
        Task<bool> DeleteSyntheticBackupAsync(string id);
        Task<bool> MergeIncrementalsAsync(MergeIncrementalsRequest request);
    }

    public class ReplicationJobModel
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string SourceVM { get; set; } = string.Empty;
        public string TargetHost { get; set; } = string.Empty;
        public string Status { get; set; } = string.Empty;
        public int Progress { get; set; }
    }

    public class UserModel
    {
        public string Id { get; set; } = string.Empty;
        public string Username { get; set; } = string.Empty;
        public string Email { get; set; } = string.Empty;
        public List<string> RoleIds { get; set; } = new();
        public bool Active { get; set; }
        public DateTime LastLogin { get; set; }
    }

    public class RoleModel
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public List<string> Permissions { get; set; } = new();
    }

    public class VSSWriterModel
    {
        public string WriterType { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string State { get; set; } = string.Empty;
    }

    public class TapeLibraryModel
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Model { get; set; } = string.Empty;
        public int Slots { get; set; }
        public int DriveCount { get; set; }
        public string Status { get; set; } = string.Empty;
    }

    public class TapeCartridgeModel
    {
        public string Id { get; set; } = string.Empty;
        public string Barcode { get; set; } = string.Empty;
        public int Slot { get; set; }
        public string MediaType { get; set; } = string.Empty;
        public long CapacityGb { get; set; }
        public long UsedGb { get; set; }
        public string Status { get; set; } = string.Empty;
        public string Label { get; set; } = string.Empty;
    }

    public class TapeVaultModel
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string Location { get; set; } = string.Empty;
        public string Contact { get; set; } = string.Empty;
    }

    public class TapeJobModel
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Source { get; set; } = string.Empty;
        public string TargetVault { get; set; } = string.Empty;
        public string Schedule { get; set; } = string.Empty;
        public int RetentionDays { get; set; }
        public bool Enabled { get; set; }
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

    public class JobHistoryItem
    {
        public string Id { get; set; } = string.Empty;
        public string JobId { get; set; } = string.Empty;
        public string Status { get; set; } = string.Empty;
        public DateTime StartTime { get; set; }
        public DateTime EndTime { get; set; }
        public double SizeGB { get; set; }
        public int DurationSec { get; set; }
        public string Message { get; set; } = string.Empty;
    }

    public class CredentialModel
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Username { get; set; } = string.Empty;
        public string Domain { get; set; } = string.Empty;
        public string Type { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
    }

    public class ProxyModel
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Host { get; set; } = string.Empty;
        public int Port { get; set; }
        public string Type { get; set; } = string.Empty;
        public int MaxTasks { get; set; }
        public bool Enabled { get; set; }
        public string Status { get; set; } = string.Empty;
    }

    public class BackupSession
    {
        public string Id { get; set; } = string.Empty;
        public string JobId { get; set; } = string.Empty;
        public string JobName { get; set; } = string.Empty;
        public string Status { get; set; } = string.Empty;
        public int Progress { get; set; }
        public DateTime StartTime { get; set; }
        public DateTime? EndTime { get; set; }
        public double ProcessedGB { get; set; }
    }

    public class ReportModel
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Type { get; set; } = string.Empty;
        public DateTime Generated { get; set; }
        public string Format { get; set; } = string.Empty;
        public string Url { get; set; } = string.Empty;
    }

    public class ReportRequest
    {
        public string Name { get; set; } = string.Empty;
        public string Type { get; set; } = string.Empty;
        public string From { get; set; } = string.Empty;
        public string To { get; set; } = string.Empty;
        public string Format { get; set; } = "pdf";
    }

    public class NotificationModel
    {
        public string Id { get; set; } = string.Empty;
        public string Type { get; set; } = string.Empty;
        public string Title { get; set; } = string.Empty;
        public string Message { get; set; } = string.Empty;
        public DateTime Time { get; set; }
        public bool Read { get; set; }
    }

    public class AppSettings
    {
        public GeneralSettings General { get; set; } = new();
        public BackupSettings Backup { get; set; } = new();
        public NetworkSettings Network { get; set; } = new();
        public SecuritySettings Security { get; set; } = new();
    }

    public class GeneralSettings
    {
        public string Language { get; set; } = "en";
        public string Theme { get; set; } = "dark";
        public string DateFormat { get; set; } = "yyyy-MM-dd";
        public string TimeZone { get; set; } = "UTC";
    }

    public class BackupSettings
    {
        public string DefaultRepo { get; set; } = string.Empty;
        public int MaxParallel { get; set; } = 4;
        public string TempPath { get; set; } = string.Empty;
        public bool EnableNotifications { get; set; } = true;
    }

    public class NetworkSettings
    {
        public int Timeout { get; set; } = 300;
        public int RetryCount { get; set; } = 3;
        public int MaxBandwidth { get; set; } = 0;
    }

    public class SecuritySettings
    {
        public bool EnableRBAC { get; set; } = true;
        public int SessionTimeout { get; set; } = 60;
        public bool Require2FA { get; set; } = false;
    }
}
