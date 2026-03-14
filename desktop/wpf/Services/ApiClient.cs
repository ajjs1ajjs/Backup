using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Json;
using System.Threading.Tasks;
using NovaBackup.GUI.ViewModels;
using NovaBackup.GUI.Models;

namespace NovaBackup.GUI.Services
{
    public class ApiClient : IApiClient
    {
        private readonly HttpClient _httpClient;

        public ApiClient(HttpClient httpClient)
        {
            _httpClient = httpClient;
            _httpClient.BaseAddress = new Uri("http://localhost:8080/api/v1/");
        }

        public async Task<DashboardStats> GetDashboardStatsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<DashboardStats>("dashboard/stats");
            return response ?? new DashboardStats();
        }

        public async Task<List<JobModel>> GetJobsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<JobModel>>("jobs");
            return response ?? new List<JobModel>();
        }

        public async Task<List<InfrastructureNode>> GetInfrastructureTreeAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<InfrastructureNode>>("infrastructure/tree");
            return response ?? new List<InfrastructureNode>();
        }

        public async Task<List<RepositoryModel>> GetRepositoriesAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<RepositoryModel>>("storage/repositories");
            return response ?? new List<RepositoryModel>();
        }

        public async Task<bool> CreateRepositoryAsync(RepositoryModel repo)
        {
            var response = await _httpClient.PostAsJsonAsync("storage/repositories", repo);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> AddServerAsync(string name, string address, string type, string username, string password)
        {
            var response = await _httpClient.PostAsJsonAsync("infrastructure/add", new {
                name, address, type, username, password
            });
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> CreateJobAsync(JobModel job, string schedule)
        {
            var response = await _httpClient.PostAsJsonAsync("jobs", new {
                name = job.Name,
                description = "Created via WPF Wizard",
                job_type = job.Platform.ToLower() == "hyper-v" ? "vm" : "file",
                schedule = schedule,
                enabled = true,
                retention_days = job.RetentionDays,
                guest_processing = job.EnableGuestProcessing,
                guest_credentials_id = job.GuestCredentialsId
            });
            return response.IsSuccessStatusCode;
        }

        // Overload for simpler WPF call without explicit schedule parameter
        public async Task<bool> CreateJobAsync(JobModel job)
        {
            var response = await _httpClient.PostAsJsonAsync("jobs", new {
                name = job.Name,
                description = "Created via WPF Wizard",
                job_type = job.Platform.ToLower() == "hyper-v" ? "vm" : "file",
                schedule = "Daily 22:00",
                enabled = true,
                retention_days = job.RetentionDays,
                guest_processing = job.EnableGuestProcessing,
                guest_credentials_id = job.GuestCredentialsId
            });
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> DiscoverNodeAsync(string nodeId)
        {
            var response = await _httpClient.GetAsync($"infrastructure/nodes/{nodeId}/discover");
            return response.IsSuccessStatusCode;
        }

        public async Task<List<InfrastructureObject>> GetDiscoveredObjectsAsync(string nodeId)
        {
            var response = await _httpClient.GetFromJsonAsync<List<InfrastructureObject>>($"infrastructure/nodes/{nodeId}/objects");
            return response ?? new List<InfrastructureObject>();
        }

        public async Task<bool> RunJobAsync(string jobId)
        {
            var response = await _httpClient.PostAsync($"jobs/{jobId}/run", null);
            return response.IsSuccessStatusCode;
        }

        public async Task<List<RestorePointModel>> GetRestorePointsAsync(string jobId)
        {
            var response = await _httpClient.GetFromJsonAsync<List<RestorePointModel>>("restore/points");
            return response ?? new List<RestorePointModel>();
        }

        public async Task<bool> StartInstantRecoveryAsync(string rpId, string vmName)
        {
            var response = await _httpClient.PostAsJsonAsync("recovery/instant", new {
                restore_point_id = rpId,
                vm_name = vmName
            });
            return response.IsSuccessStatusCode;
        }

        public async Task<List<RecoverySessionModel>> GetInstantRecoverySessionsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<RecoverySessionModel>>("recovery/sessions");
            return response ?? new List<RecoverySessionModel>();
        }

        public async Task<bool> StopInstantRecoveryAsync(string sessionId)
        {
            var response = await _httpClient.PostAsync($"recovery/sessions/{sessionId}/stop", null);
            return response.IsSuccessStatusCode;
        }

        // Job History
        public async Task<List<JobHistoryItem>> GetJobHistoryAsync(string jobId)
        {
            var response = await _httpClient.GetFromJsonAsync<List<JobHistoryItem>>($"jobs/{jobId}/history");
            return response ?? new List<JobHistoryItem>();
        }

        // Backup Sessions
        public async Task<List<BackupSession>> GetBackupSessionsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<BackupSession>>("backup/sessions");
            return response ?? new List<BackupSession>();
        }

        public async Task<bool> StopBackupSessionAsync(string sessionId)
        {
            var response = await _httpClient.DeleteAsync($"backup/sessions/{sessionId}");
            return response.IsSuccessStatusCode;
        }

        // Credentials
        public async Task<List<CredentialModel>> GetCredentialsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<CredentialModel>>("credentials");
            return response ?? new List<CredentialModel>();
        }

        public async Task<bool> CreateCredentialAsync(CredentialModel credential)
        {
            var response = await _httpClient.PostAsJsonAsync("credentials", credential);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> DeleteCredentialAsync(string id)
        {
            var response = await _httpClient.DeleteAsync($"credentials/{id}");
            return response.IsSuccessStatusCode;
        }

        // Proxies
        public async Task<List<ProxyModel>> GetProxiesAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<ProxyModel>>("proxies");
            return response ?? new List<ProxyModel>();
        }

        public async Task<bool> CreateProxyAsync(ProxyModel proxy)
        {
            var response = await _httpClient.PostAsJsonAsync("proxies", proxy);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> UpdateProxyAsync(ProxyModel proxy)
        {
            var response = await _httpClient.PutAsJsonAsync($"proxies/{proxy.Id}", proxy);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> DeleteProxyAsync(string id)
        {
            var response = await _httpClient.DeleteAsync($"proxies/{id}");
            return response.IsSuccessStatusCode;
        }

        // Reports
        public async Task<List<ReportModel>> GetReportsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<ReportModel>>("reports");
            return response ?? new List<ReportModel>();
        }

        public async Task<ReportModel> GetReportAsync(string id)
        {
            var response = await _httpClient.GetFromJsonAsync<ReportModel>($"reports/{id}");
            return response ?? new ReportModel();
        }

        public async Task<bool> GenerateReportAsync(ReportRequest request)
        {
            var response = await _httpClient.PostAsJsonAsync("reports/generate", request);
            return response.IsSuccessStatusCode;
        }

        // Notifications
        public async Task<List<NotificationModel>> GetNotificationsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<NotificationModel>>("notifications");
            return response ?? new List<NotificationModel>();
        }

        public async Task<bool> MarkNotificationReadAsync(string id)
        {
            var response = await _httpClient.PutAsync($"notifications/{id}/read", null);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> DeleteNotificationAsync(string id)
        {
            var response = await _httpClient.DeleteAsync($"notifications/{id}");
            return response.IsSuccessStatusCode;
        }

        // Settings
        public async Task<AppSettings> GetSettingsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<AppSettings>("settings");
            return response ?? new AppSettings();
        }

        public async Task<bool> UpdateSettingsAsync(AppSettings settings)
        {
            var response = await _httpClient.PutAsJsonAsync("settings", settings);
            return response.IsSuccessStatusCode;
        }

        // Tape
        public async Task<List<TapeLibraryModel>> GetTapeLibrariesAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<TapeLibraryModel>>("tape/libraries");
            return response ?? new List<TapeLibraryModel>();
        }

        public async Task<List<TapeCartridgeModel>> GetTapeCartridgesAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<TapeCartridgeModel>>("tape/cartridges");
            return response ?? new List<TapeCartridgeModel>();
        }

        public async Task<List<TapeVaultModel>> GetTapeVaultsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<TapeVaultModel>>("tape/vaults");
            return response ?? new List<TapeVaultModel>();
        }

        public async Task<bool> CreateTapeVaultAsync(TapeVaultModel vault)
        {
            var response = await _httpClient.PostAsJsonAsync("tape/vaults", vault);
            return response.IsSuccessStatusCode;
        }

        public async Task<List<TapeJobModel>> GetTapeJobsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<TapeJobModel>>("tape/jobs");
            return response ?? new List<TapeJobModel>();
        }

        public async Task<bool> CreateTapeJobAsync(TapeJobModel job)
        {
            var response = await _httpClient.PostAsJsonAsync("tape/jobs", job);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> RunTapeJobAsync(string jobId)
        {
            var response = await _httpClient.PostAsync($"tape/jobs/{jobId}/run", null);
            return response.IsSuccessStatusCode;
        }

        // VSS
        public async Task<List<VSSWriterModel>> GetVSSWritersAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<VSSWriterModel>>("vss/writers");
            return response ?? new List<VSSWriterModel>();
        }

        // Replication
        public async Task<List<ReplicationJobModel>> GetReplicationJobsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<ReplicationJobModel>>("replication/jobs");
            return response ?? new List<ReplicationJobModel>();
        }

        // RBAC - Users
        public async Task<List<UserModel>> GetUsersAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<UserModel>>("rbac/users");
            return response ?? new List<UserModel>();
        }

        public async Task<bool> CreateUserAsync(UserModel user)
        {
            var response = await _httpClient.PostAsJsonAsync("rbac/users", user);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> UpdateUserAsync(UserModel user)
        {
            var response = await _httpClient.PutAsJsonAsync($"rbac/users/{user.Id}", user);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> DeleteUserAsync(string id)
        {
            var response = await _httpClient.DeleteAsync($"rbac/users/{id}");
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> AssignRoleToUserAsync(string userId, string roleId)
        {
            var response = await _httpClient.PostAsJsonAsync($"rbac/users/{userId}/roles", new { role_id = roleId });
            return response.IsSuccessStatusCode;
        }

        // RBAC - Roles
        public async Task<List<RoleModel>> GetRolesAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<RoleModel>>("rbac/roles");
            return response ?? new List<RoleModel>();
        }

        public async Task<bool> CreateRoleAsync(RoleModel role)
        {
            var response = await _httpClient.PostAsJsonAsync("rbac/roles", role);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> UpdateRoleAsync(RoleModel role)
        {
            var response = await _httpClient.PutAsJsonAsync($"rbac/roles/{role.Id}", role);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> DeleteRoleAsync(string id)
        {
            var response = await _httpClient.DeleteAsync($"rbac/roles/{id}");
            return response.IsSuccessStatusCode;
        }

        // RBAC - Permissions
        public async Task<List<string>> GetPermissionsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<List<string>>("rbac/permissions");
            return response ?? new List<string>();
        }

        // Synthetic Backup
        public async Task<List<SyntheticBackupModel>> GetSyntheticBackupsAsync()
        {
            var response = await _httpClient.GetFromJsonAsync<Dictionary<string, object>>("synthetic");
            if (response != null && response.ContainsKey("backups"))
            {
                var backups = new List<SyntheticBackupModel>();
                var backupArray = response["backups"] as System.Text.Json.JsonElement?;
                if (backupArray.HasValue)
                {
                    foreach (var item in backupArray.Value.EnumerateArray())
                    {
                        var backup = new SyntheticBackupModel
                        {
                            Id = item.GetProperty("id").GetString() ?? "",
                            Name = item.GetProperty("backup_type").GetString() ?? "Synthetic Backup",
                            SourceRepo = item.GetProperty("source_repo").GetString() ?? "",
                            TargetRepo = item.GetProperty("target_repo").GetString() ?? "",
                            BackupType = item.GetProperty("backup_type").GetString() ?? "",
                            Status = item.GetProperty("status").GetString() ?? "",
                            Size = item.GetProperty("size").GetInt64(),
                            CompressionRatio = item.GetProperty("compression_ratio").GetDouble(),
                            CreatedAt = item.GetProperty("created_at").GetDateTime(),
                        };
                        backups.Add(backup);
                    }
                }
                return backups;
            }
            return new List<SyntheticBackupModel>();
        }

        public async Task<bool> CreateSyntheticBackupAsync(SyntheticBackupRequest request)
        {
            var response = await _httpClient.PostAsJsonAsync("synthetic", request);
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> DeleteSyntheticBackupAsync(string id)
        {
            var response = await _httpClient.DeleteAsync($"synthetic/{id}");
            return response.IsSuccessStatusCode;
        }

        public async Task<bool> MergeIncrementalsAsync(MergeIncrementalsRequest request)
        {
            var response = await _httpClient.PostAsJsonAsync("synthetic/merge", request);
            return response.IsSuccessStatusCode;
        }
    }
}
