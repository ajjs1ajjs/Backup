using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Json;
using System.Threading.Tasks;
using Newtonsoft.Json;

namespace NovaBackup.GUI.Services
{
    /// <summary>
    /// Interface for NovaBackup API service
    /// </summary>
    public interface INovaBackupApiService
    {
        Task<bool> ConnectAsync();
        Task<DashboardStats> GetDashboardStatsAsync();
        Task<List<BackupJob>> GetBackupJobsAsync();
        Task<List<VirtualMachine>> GetVirtualMachinesAsync();
        Task StartBackupAsync(Guid jobId, Action<int> progressCallback);
        Task<bool> CreateBackupJobAsync(CreateJobRequest request);
    }

    /// <summary>
    /// Implementation of NovaBackup API service
    /// </summary>
    public class NovaBackupApiService : INovaBackupApiService
    {
        private readonly HttpClient _httpClient;
        private const string BaseUrl = "http://localhost:8080/api/v1";

        public NovaBackupApiService()
        {
            _httpClient = new HttpClient
            {
                BaseAddress = new Uri(BaseUrl),
                Timeout = TimeSpan.FromSeconds(30)
            };
        }

        public async Task<bool> ConnectAsync()
        {
            try
            {
                var response = await _httpClient.GetAsync("/health");
                return response.IsSuccessStatusCode;
            }
            catch
            {
                return false;
            }
        }

        public async Task<DashboardStats> GetDashboardStatsAsync()
        {
            try
            {
                var response = await _httpClient.GetAsync("/dashboard/stats");
                response.EnsureSuccessStatusCode();
                return await response.Content.ReadFromJsonAsync<DashboardStats>() ?? new DashboardStats();
            }
            catch (Exception ex)
            {
                System.Diagnostics.Debug.WriteLine($"Error getting dashboard stats: {ex.Message}");
                // Return default stats for demo
                return new DashboardStats
                {
                    ActiveJobsCount = 3,
                    TotalVMs = 12,
                    StorageUsedGB = 697,
                    StorageTotalGB = 1000,
                    CompressionRatio = 3.2,
                    DeduplicationRatio = 5.1
                };
            }
        }

        public async Task<List<BackupJob>> GetBackupJobsAsync()
        {
            try
            {
                var response = await _httpClient.GetAsync("/jobs");
                response.EnsureSuccessStatusCode();
                return await response.Content.ReadFromJsonAsync<List<BackupJob>>() ?? new List<BackupJob>();
            }
            catch (Exception ex)
            {
                System.Diagnostics.Debug.WriteLine($"Error getting backup jobs: {ex.Message}");
                // Return demo data
                return new List<BackupJob>
                {
                    new BackupJob
                    {
                        Id = Guid.NewGuid(),
                        Name = "Daily Documents Backup",
                        Type = "Files",
                        Status = "Active",
                        LastRun = DateTime.Now.AddHours(-2),
                        NextRun = DateTime.Now.AddHours(22),
                        Source = "C:\\Users\\Documents",
                        Destination = "\\\\backup\\documents",
                        IsIncremental = true,
                        IsEnabled = true
                    },
                    new BackupJob
                    {
                        Id = Guid.NewGuid(),
                        Name = "Weekly System Backup",
                        Type = "System",
                        Status = "Active",
                        LastRun = DateTime.Now.AddDays(-3),
                        NextRun = DateTime.Now.AddDays(4),
                        Source = "C:\\",
                        Destination = "\\\\backup\\system",
                        IsIncremental = false,
                        IsEnabled = true
                    },
                    new BackupJob
                    {
                        Id = Guid.NewGuid(),
                        Name = "Database Backup",
                        Type = "SQL Server",
                        Status = "Active",
                        LastRun = DateTime.Now.AddHours(-1),
                        NextRun = DateTime.Now.AddHours(23),
                        Source = "SQL Server",
                        Destination = "\\\\backup\\sql",
                        IsIncremental = true,
                        IsEnabled = true
                    }
                };
            }
        }

        public async Task<List<VirtualMachine>> GetVirtualMachinesAsync()
        {
            try
            {
                var response = await _httpClient.GetAsync("/vmware/vms");
                response.EnsureSuccessStatusCode();
                return await response.Content.ReadFromJsonAsync<List<VirtualMachine>>() ?? new List<VirtualMachine>();
            }
            catch (Exception ex)
            {
                System.Diagnostics.Debug.WriteLine($"Error getting VMs: {ex.Message}");
                // Return demo data
                return new List<VirtualMachine>
                {
                    new VirtualMachine
                    {
                        Id = Guid.NewGuid(),
                        Name = "VM-DC01",
                        PowerState = "Powered On",
                        GuestOS = "Windows Server 2022",
                        IpAddress = "192.168.1.10",
                        CpuCount = 4,
                        MemoryMB = 8192,
                        DiskCount = 2
                    },
                    new VirtualMachine
                    {
                        Id = Guid.NewGuid(),
                        Name = "VM-SQL01",
                        PowerState = "Powered On",
                        GuestOS = "Windows Server 2022",
                        IpAddress = "192.168.1.20",
                        CpuCount = 8,
                        MemoryMB = 16384,
                        DiskCount = 4
                    },
                    new VirtualMachine
                    {
                        Id = Guid.NewGuid(),
                        Name = "VM-WEB01",
                        PowerState = "Powered On",
                        GuestOS = "Ubuntu 22.04",
                        IpAddress = "192.168.1.30",
                        CpuCount = 2,
                        MemoryMB = 4096,
                        DiskCount = 1
                    }
                };
            }
        }

        public async Task StartBackupAsync(Guid jobId, Action<int> progressCallback)
        {
            try
            {
                var response = await _httpClient.PostAsync($"/jobs/{jobId}/start", null);
                response.EnsureSuccessStatusCode();

                // Simulate progress updates
                for (int i = 0; i <= 100; i += 10)
                {
                    await Task.Delay(500);
                    progressCallback?.Invoke(i);
                }
            }
            catch (Exception ex)
            {
                System.Diagnostics.Debug.WriteLine($"Error starting backup: {ex.Message}");
                throw;
            }
        }

        public async Task<bool> CreateBackupJobAsync(CreateJobRequest request)
        {
            try
            {
                var response = await _httpClient.PostAsJsonAsync("/jobs", request);
                return response.IsSuccessStatusCode;
            }
            catch (Exception ex)
            {
                System.Diagnostics.Debug.WriteLine($"Error creating job: {ex.Message}");
                return false;
            }
        }
    }

    /// <summary>
    /// Interface for VMware service
    /// </summary>
    public interface IVMwareService
    {
        Task<List<Host>> GetHostsAsync();
        Task<List<Datastore>> GetDatastoresAsync();
        Task<List<VirtualMachine>> GetVMsAsync();
    }

    /// <summary>
    /// Implementation of VMware service
    /// </summary>
    public class VMwareService : IVMwareService
    {
        private readonly HttpClient _httpClient;

        public VMwareService()
        {
            _httpClient = new HttpClient
            {
                BaseAddress = new Uri("http://localhost:8080/api/v1"),
                Timeout = TimeSpan.FromSeconds(30)
            };
        }

        public async Task<List<Host>> GetHostsAsync()
        {
            // Return demo data
            return new List<Host>
            {
                new Host { Name = "esxi01.local", Status = "Connected", CpuUsagePercent = 45, MemoryUsagePercent = 60 },
                new Host { Name = "esxi02.local", Status = "Connected", CpuUsagePercent = 30, MemoryUsagePercent = 45 }
            };
        }

        public async Task<List<Datastore>> GetDatastoresAsync()
        {
            // Return demo data
            return new List<Datastore>
            {
                new Datastore { Name = "Datastore1", TotalCapacityGB = 2000, FreeSpaceGB = 850 },
                new Datastore { Name = "Datastore2", TotalCapacityGB = 3000, FreeSpaceGB = 1200 }
            };
        }

        public async Task<List<VirtualMachine>> GetVMsAsync()
        {
            // Return demo data
            return new List<VirtualMachine>
            {
                new VirtualMachine
                {
                    Id = Guid.NewGuid(),
                    Name = "VM-DC01",
                    PowerState = "Powered On",
                    GuestOS = "Windows Server 2022",
                    IpAddress = "192.168.1.10",
                    CpuCount = 4,
                    MemoryMB = 8192,
                    DiskCount = 2
                }
            };
        }
    }

    // Model classes
    public class DashboardStats
    {
        public int ActiveJobsCount { get; set; }
        public int TotalVMs { get; set; }
        public double StorageUsedGB { get; set; }
        public double StorageTotalGB { get; set; }
        public double StorageUsagePercent => StorageTotalGB > 0 ? (StorageUsedGB / StorageTotalGB) * 100 : 0;
        public double CompressionRatio { get; set; }
        public double DeduplicationRatio { get; set; }
    }

    public class BackupJob
    {
        public Guid Id { get; set; }
        public string Name { get; set; } = "";
        public string Type { get; set; } = "";
        public string Status { get; set; } = "";
        public DateTime? LastRun { get; set; }
        public DateTime? NextRun { get; set; }
        public string Source { get; set; } = "";
        public string Destination { get; set; } = "";
        public bool IsIncremental { get; set; }
        public bool IsEnabled { get; set; }
    }

    public class VirtualMachine
    {
        public Guid Id { get; set; }
        public string Name { get; set; } = "";
        public string PowerState { get; set; } = "";
        public string GuestOS { get; set; } = "";
        public string IpAddress { get; set; } = "";
        public int CpuCount { get; set; }
        public long MemoryMB { get; set; }
        public int DiskCount { get; set; }
    }

    public class Host
    {
        public string Name { get; set; } = "";
        public string Status { get; set; } = "";
        public int CpuUsagePercent { get; set; }
        public int MemoryUsagePercent { get; set; }
    }

    public class Datastore
    {
        public string Name { get; set; } = "";
        public double TotalCapacityGB { get; set; }
        public double FreeSpaceGB { get; set; }
    }

    public class CreateJobRequest
    {
        public string Name { get; set; } = "";
        public string Type { get; set; } = "";
        public string Source { get; set; } = "";
        public string Destination { get; set; } = "";
        public bool IsIncremental { get; set; }
        public string Schedule { get; set; } = "";
    }
}
