using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class RecoverySessionsViewModelMVVM : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<RecoverySessionModel> _sessions = new();

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        [ObservableProperty]
        private bool _isBusy;

        public RecoverySessionsViewModelMVVM(IApiClient apiClient)
        {
            _apiClient = apiClient;
            _ = LoadAsync();
        }

        public async Task LoadAsync()
        {
            IsBusy = true;
            StatusMessage = "Loading sessions...";
            try
            {
                var list = await _apiClient.GetInstantRecoverySessionsAsync();
                Sessions.Clear();
                foreach (var s in list) Sessions.Add(s);
                StatusMessage = string.Empty;
            }
            catch (Exception ex)
            {
                StatusMessage = $"Error: {ex.Message}";
            }
            finally
            {
                IsBusy = false;
            }
        }

        [RelayCommand]
        private async Task RefreshAsync()
        {
            await LoadAsync();
        }

        [RelayCommand]
        private async Task StopSessionAsync(string sessionId)
        {
            await _apiClient.StopInstantRecoveryAsync(sessionId);
            await LoadAsync();
        }
    }
}
