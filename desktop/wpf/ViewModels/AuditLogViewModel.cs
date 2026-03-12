using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class AuditLogViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<AuditLogEntry> _entries = new();

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        [ObservableProperty]
        private string _filterText = string.Empty;

        public AuditLogViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading audit log...";
            try
            {
                Entries.Clear();
                Entries.Add(new AuditLogEntry { Timestamp = System.DateTime.Now, User = "admin", Action = "Login", Details = "User logged in successfully" });
                Entries.Add(new AuditLogEntry { Timestamp = System.DateTime.Now.AddMinutes(-5), User = "admin", Action = "CreateJob", Details = "Created backup job 'Daily Backup'" });
                Entries.Add(new AuditLogEntry { Timestamp = System.DateTime.Now.AddMinutes(-15), User = "operator", Action = "RunJob", Details = "Started backup job 'Weekly Full'" });
                Entries.Add(new AuditLogEntry { Timestamp = System.DateTime.Now.AddHours(-1), User = "admin", Action = "UpdateSettings", Details = "Changed backup retention to 30 days" });
                Entries.Add(new AuditLogEntry { Timestamp = System.DateTime.Now.AddHours(-2), User = "viewer", Action = "ViewReport", Details = "Viewed backup summary report" });
                StatusMessage = $"Loaded {Entries.Count} audit entries";
            }
            catch { StatusMessage = "Error loading data"; }
            finally { IsLoading = false; }
        }
    }

    public class AuditLogEntry
    {
        public System.DateTime Timestamp { get; set; }
        public string User { get; set; } = string.Empty;
        public string Action { get; set; } = string.Empty;
        public string Details { get; set; } = string.Empty;
    }
}
