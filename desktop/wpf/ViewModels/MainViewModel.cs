using System;
using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Runtime.CompilerServices;
using System.Threading.Tasks;
using System.Windows.Input;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;

namespace NovaBackup.GUI.ViewModels
{
    /// <summary>
    /// Main ViewModel for the application
    /// </summary>
    public partial class MainViewModel : ObservableObject
    {
        private readonly Services.INovaBackupApiService _apiService;

        [ObservableProperty]
        private string _title = "NovaBackup v6.0 Enterprise";

        [ObservableProperty]
        private string _currentView = "Dashboard";

        [ObservableProperty]
        private bool _isConnected;

        [ObservableProperty]
        private string _connectionStatus = "Disconnected";

        [ObservableProperty]
        private ObservableCollection<JobViewModel> _backupJobs = new();

        [ObservableProperty]
        private ObservableCollection<VMViewModel> _virtualMachines = new();

        [ObservableProperty]
        private DashboardStatsViewModel _dashboardStats = new();

        public MainViewModel(Services.INovaBackupApiService apiService)
        {
            _apiService = apiService;
            InitializeCommands();
            LoadDataAsync();
        }

        private void InitializeCommands()
        {
            ConnectCommand = new AsyncRelayCommand(ConnectAsync);
            RefreshCommand = new AsyncRelayCommand(LoadDataAsync);
            CreateJobCommand = new AsyncRelayCommand(CreateJobAsync);
            StartBackupCommand = new AsyncRelayCommand<JobViewModel>(StartBackupAsync);
        }

        public ICommand ConnectCommand { get; private set; }
        public ICommand RefreshCommand { get; private set; }
        public ICommand CreateJobCommand { get; private set; }
        public ICommand StartBackupCommand { get; private set; }

        private async Task ConnectAsync()
        {
            try
            {
                ConnectionStatus = "Connecting...";
                var result = await _apiService.ConnectAsync();
                IsConnected = result;
                ConnectionStatus = result ? "Connected" : "Failed";
                
                if (result)
                {
                    await LoadDataAsync();
                }
            }
            catch (Exception ex)
            {
                ConnectionStatus = $"Error: {ex.Message}";
            }
        }

        private async Task LoadDataAsync()
        {
            try
            {
                // Load dashboard stats
                DashboardStats = await _apiService.GetDashboardStatsAsync();

                // Load jobs
                var jobs = await _apiService.GetBackupJobsAsync();
                BackupJobs.Clear();
                foreach (var job in jobs)
                {
                    BackupJobs.Add(new JobViewModel(job));
                }

                // Load VMs
                var vms = await _apiService.GetVirtualMachinesAsync();
                VirtualMachines.Clear();
                foreach (var vm in vms)
                {
                    VirtualMachines.Add(new VMViewModel(vm));
                }
            }
            catch (Exception ex)
            {
                // Log error
                System.Diagnostics.Debug.WriteLine($"Error loading data: {ex.Message}");
            }
        }

        private async Task CreateJobAsync()
        {
            // Open job creation wizard
            // TODO: Implement job creation dialog
            await Task.CompletedTask;
        }

        private async Task StartBackupAsync(JobViewModel job)
        {
            if (job == null) return;

            try
            {
                job.Status = "Running";
                job.Progress = 0;

                await _apiService.StartBackupAsync(job.Id, progress =>
                {
                    job.Progress = progress;
                });

                job.Status = "Completed";
                job.LastRun = DateTime.Now;
            }
            catch (Exception ex)
            {
                job.Status = $"Failed: {ex.Message}";
            }
        }
    }

    /// <summary>
    /// Dashboard statistics ViewModel
    /// </summary>
    public partial class DashboardStatsViewModel : ObservableObject
    {
        [ObservableProperty]
        private int _activeJobsCount;

        [ObservableProperty]
        private int _totalVMs;

        [ObservableProperty]
        private double _storageUsedGB;

        [ObservableProperty]
        private double _storageTotalGB = 1024; // 1 TB default

        [ObservableProperty]
        private double _storageUsagePercent;

        [ObservableProperty]
        private double _compressionRatio = 3.2;

        [ObservableProperty]
        private double _deduplicationRatio = 5.1;

        partial void OnStorageUsedGBChanged(double value)
        {
            StorageUsagePercent = (value / StorageTotalGB) * 100;
        }
    }

    /// <summary>
    /// Backup job ViewModel
    /// </summary>
    public partial class JobViewModel : ObservableObject
    {
        public Guid Id { get; set; }

        [ObservableProperty]
        private string _name = "";

        [ObservableProperty]
        private string _type = "";

        [ObservableProperty]
        private string _status = "";

        [ObservableProperty]
        private string _lastRun = "";

        [ObservableProperty]
        private string _nextRun = "";

        [ObservableProperty]
        private int _progress;

        [ObservableProperty]
        private string _source = "";

        [ObservableProperty]
        private string _destination = "";

        [ObservableProperty]
        private bool _isIncremental;

        [ObservableProperty]
        private bool _isEnabled = true;

        public JobViewModel() { }

        public JobViewModel(Models.BackupJob model)
        {
            Id = model.Id;
            Name = model.Name;
            Type = model.Type;
            Status = model.Status;
            LastRun = model.LastRun?.ToString("g") ?? "Never";
            NextRun = model.NextRun?.ToString("g") ?? "Not scheduled";
            Source = model.Source;
            Destination = model.Destination;
            IsIncremental = model.IsIncremental;
            IsEnabled = model.IsEnabled;
        }
    }

    /// <summary>
    /// Virtual Machine ViewModel
    /// </summary>
    public partial class VMViewModel : ObservableObject
    {
        public Guid Id { get; set; }

        [ObservableProperty]
        private string _name = "";

        [ObservableProperty]
        private string _powerState = "";

        [ObservableProperty]
        private string _guestOS = "";

        [ObservableProperty]
        private string _ipAddress = "";

        [ObservableProperty]
        private int _cpuCount;

        [ObservableProperty]
        private long _memoryMB;

        [ObservableProperty]
        private int _diskCount;

        [ObservableProperty]
        private bool _isSelected;

        [ObservableProperty]
        private bool _hasBackup;

        public VMViewModel() { }

        public VMViewModel(Models.VirtualMachine model)
        {
            Id = model.Id;
            Name = model.Name;
            PowerState = model.PowerState;
            GuestOS = model.GuestOS;
            IpAddress = model.IpAddress;
            CpuCount = model.CpuCount;
            MemoryMB = model.MemoryMB;
            DiskCount = model.DiskCount;
        }
    }

    /// <summary>
    /// Infrastructure ViewModel
    /// </summary>
    public partial class InfrastructureViewModel : ObservableObject
    {
        private readonly Services.IVMwareService _vmwareService;

        [ObservableProperty]
        private ObservableCollection<HostViewModel> _hosts = new();

        [ObservableProperty]
        private ObservableCollection<DatastoreViewModel> _datastores = new();

        [ObservableProperty]
        private HostViewModel _selectedHost;

        public InfrastructureViewModel(Services.IVMwareService vmwareService)
        {
            _vmwareService = vmwareService;
            LoadInfrastructureAsync();
        }

        private async Task LoadInfrastructureAsync()
        {
            try
            {
                var hosts = await _vmwareService.GetHostsAsync();
                Hosts.Clear();
                foreach (var host in hosts)
                {
                    Hosts.Add(new HostViewModel(host));
                }

                var datastores = await _vmwareService.GetDatastoresAsync();
                Datastores.Clear();
                foreach (var ds in datastores)
                {
                    Datastores.Add(new DatastoreViewModel(ds));
                }
            }
            catch (Exception ex)
            {
                System.Diagnostics.Debug.WriteLine($"Error loading infrastructure: {ex.Message}");
            }
        }
    }

    public partial class HostViewModel : ObservableObject
    {
        public string Name { get; set; } = "";
        public string Status { get; set; } = "";
        public string CpuUsage { get; set; } = "";
        public string MemoryUsage { get; set; } = "";

        public HostViewModel(Models.Host model)
        {
            Name = model.Name;
            Status = model.Status;
            CpuUsage = $"{model.CpuUsagePercent}%";
            MemoryUsage = $"{model.MemoryUsagePercent}%";
        }
    }

    public partial class DatastoreViewModel : ObservableObject
    {
        public string Name { get; set; } = "";
        public double TotalGB { get; set; }
        public double FreeGB { get; set; }
        public double UsedPercent => ((TotalGB - FreeGB) / TotalGB) * 100;

        public DatastoreViewModel(Models.Datastore model)
        {
            Name = model.Name;
            TotalGB = model.TotalCapacityGB;
            FreeGB = model.FreeSpaceGB;
        }
    }
}
