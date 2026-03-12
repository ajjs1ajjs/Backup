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
    }
}
