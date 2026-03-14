using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class SyntheticBackupViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<SyntheticBackupModel> _backups = new();

        [ObservableProperty]
        private SyntheticBackupModel _selectedBackup = new();

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        [ObservableProperty]
        private string _filterText = string.Empty;

        public SyntheticBackupViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading synthetic backups...";
            try
            {
                var backups = await _apiClient.GetSyntheticBackupsAsync();
                Backups.Clear();
                foreach (var b in backups)
                {
                    Backups.Add(b);
                }
                StatusMessage = $"Loaded {Backups.Count} synthetic backups";
            }
            catch (System.Exception ex)
            {
                StatusMessage = $"Error loading data: {ex.Message}";
            }
            finally
            {
                IsLoading = false;
            }
        }

        [RelayCommand]
        private async Task CreateSyntheticBackupAsync()
        {
            var request = new SyntheticBackupRequest
            {
                SourceRepo = "Default",
                TargetRepo = "Default",
                BackupType = "Full",
                Compression = true,
                RetentionDays = 30
            };

            try
            {
                var success = await _apiClient.CreateSyntheticBackupAsync(request);
                StatusMessage = success ? "Synthetic backup created successfully" : "Failed to create synthetic backup";
                if (success) await LoadDataAsync();
            }
            catch (System.Exception ex)
            {
                StatusMessage = $"Error: {ex.Message}";
            }
        }

        [RelayCommand]
        private async Task DeleteSyntheticBackupAsync()
        {
            if (SelectedBackup == null || string.IsNullOrEmpty(SelectedBackup.Id))
            {
                StatusMessage = "No backup selected";
                return;
            }

            var result = System.Windows.MessageBox.Show(
                $"Are you sure you want to delete synthetic backup '{SelectedBackup.Name}'?",
                "Confirm Delete",
                System.Windows.MessageBoxButton.YesNo,
                System.Windows.MessageBoxImage.Warning);

            if (result == System.Windows.MessageBoxResult.Yes)
            {
                try
                {
                    var success = await _apiClient.DeleteSyntheticBackupAsync(SelectedBackup.Id);
                    StatusMessage = success ? "Backup deleted successfully" : "Failed to delete backup";
                    if (success) await LoadDataAsync();
                }
                catch (System.Exception ex)
                {
                    StatusMessage = $"Error: {ex.Message}";
                }
            }
        }

        [RelayCommand]
        private async Task MergeIncrementalsAsync()
        {
            if (SelectedBackup == null || string.IsNullOrEmpty(SelectedBackup.Id))
            {
                StatusMessage = "No backup selected";
                return;
            }

            try
            {
                var request = new MergeIncrementalsRequest
                {
                    BackupId = SelectedBackup.Id,
                    Since = System.DateTime.Now.AddDays(-7)
                };

                var success = await _apiClient.MergeIncrementalsAsync(request);
                StatusMessage = success ? "Incremental merge completed" : "Failed to merge incrementals";
                if (success) await LoadDataAsync();
            }
            catch (System.Exception ex)
            {
                StatusMessage = $"Error: {ex.Message}";
            }
        }

        [RelayCommand]
        private async Task RefreshAsync()
        {
            await LoadDataAsync();
        }
    }

    public class SyntheticBackupModel
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string SourceRepo { get; set; } = string.Empty;
        public string TargetRepo { get; set; } = string.Empty;
        public string BackupType { get; set; } = string.Empty;
        public string Status { get; set; } = string.Empty;
        public long Size { get; set; }
        public double CompressionRatio { get; set; }
        public System.DateTime CreatedAt { get; set; }
        public System.DateTime? CompletedAt { get; set; }
    }

    public class SyntheticBackupRequest
    {
        public string SourceRepo { get; set; } = string.Empty;
        public string TargetRepo { get; set; } = string.Empty;
        public string BackupType { get; set; } = string.Empty;
        public bool Compression { get; set; }
        public int RetentionDays { get; set; }
        public System.DateTime? IncrementalSince { get; set; }
    }

    public class MergeIncrementalsRequest
    {
        public string BackupId { get; set; } = string.Empty;
        public System.DateTime Since { get; set; }
    }
}
