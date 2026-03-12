using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class VSSViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<VSSWriterModel> _writers = new();

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        public VSSViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading VSS writers...";
            try
            {
                var writers = await _apiClient.GetVSSWritersAsync();
                Writers.Clear();
                foreach (var w in writers) Writers.Add(w);
                StatusMessage = $"Loaded {Writers.Count} VSS writers";
            }
            catch { StatusMessage = "Error loading VSS data"; }
            finally { IsLoading = false; }
        }
    }
}
