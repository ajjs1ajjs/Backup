using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public class RecoverySessionsViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<RecoverySessionModel> _sessions = new();

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        [ObservableProperty]
        private bool _isBusy;

        public RecoverySessionsViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
            _ = LoadSessionsAsync();
        }

        private async Task LoadSessionsAsync()
        {
            IsBusy = true;
            StatusMessage = "Loading sessions...";
            try
            {
                var sessions = await _apiClient.GetInstantRecoverySessionsAsync();
                Sessions.Clear();
                foreach (var s in sessions) Sessions.Add(s);
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
        private async Task StopSessionAsync(string sessionId)
        {
            try
            {
                await _apiClient.StopInstantRecoveryAsync(sessionId);
                await LoadSessionsAsync();
            }
            catch (Exception ex)
            {
                StatusMessage = ex.Message;
            }
        }
    }
}
