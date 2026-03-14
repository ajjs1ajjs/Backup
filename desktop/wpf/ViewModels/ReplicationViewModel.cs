using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class ReplicationViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<Services.ReplicationJobModel> _jobs = new();

        [ObservableProperty]
        private Services.ReplicationJobModel _selectedJob = new();

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        public ReplicationViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading replication jobs...";
            try
            {
                var jobs = await _apiClient.GetReplicationJobsAsync();
                Jobs.Clear();
                foreach (var j in jobs) Jobs.Add(j);
                StatusMessage = $"Loaded {Jobs.Count} replication jobs";
            }
            catch { StatusMessage = "Error loading data"; }
            finally { IsLoading = false; }
        }
    }
}
