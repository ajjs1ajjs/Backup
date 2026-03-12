using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class TapeViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<TapeLibraryModel> _libraries = new();

        [ObservableProperty]
        private ObservableCollection<TapeCartridgeModel> _cartridges = new();

        [ObservableProperty]
        private ObservableCollection<TapeVaultModel> _vaults = new();

        [ObservableProperty]
        private ObservableCollection<TapeJobModel> _jobs = new();

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        [ObservableProperty]
        private string _newVaultName = string.Empty;

        public TapeViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading tape data...";

            try
            {
                var libs = await _apiClient.GetTapeLibrariesAsync();
                Libraries.Clear();
                foreach (var lib in libs) Libraries.Add(lib);

                var carts = await _apiClient.GetTapeCartridgesAsync();
                Cartridges.Clear();
                foreach (var cart in carts) Cartridges.Add(cart);

                var vaults = await _apiClient.GetTapeVaultsAsync();
                Vaults.Clear();
                foreach (var vault in vaults) Vaults.Add(vault);

                var jobs = await _apiClient.GetTapeJobsAsync();
                Jobs.Clear();
                foreach (var job in jobs) Jobs.Add(job);

                StatusMessage = $"Loaded {Libraries.Count} libraries, {Cartridges.Count} cartridges";
            }
            catch
            {
                StatusMessage = "Error loading tape data";
            }
            finally
            {
                IsLoading = false;
            }
        }

        [RelayCommand]
        private async Task CreateVaultAsync()
        {
            if (string.IsNullOrEmpty(NewVaultName)) return;

            var vault = new TapeVaultModel { Name = NewVaultName };
            await _apiClient.CreateTapeVaultAsync(vault);
            NewVaultName = string.Empty;
            await LoadDataAsync();
        }

        [RelayCommand]
        private async Task RunJobAsync(string jobId)
        {
            if (string.IsNullOrEmpty(jobId)) return;
            await _apiClient.RunTapeJobAsync(jobId);
            StatusMessage = $"Job {jobId} started";
        }
    }
}
