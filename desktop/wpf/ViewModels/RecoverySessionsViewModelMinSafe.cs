using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    // Minimal ViewModel for Recovery Sessions without MVVM Toolkit
    public class RecoverySessionsViewModelMinSafe
    {
        private readonly IApiClient _apiClient;

        public ObservableCollection<RecoverySessionModel> Sessions { get; } = new ObservableCollection<RecoverySessionModel>();
        public string StatusMessage { get; private set; } = string.Empty;

        public RecoverySessionsViewModelMinSafe(IApiClient apiClient)
        {
            _apiClient = apiClient;
            _ = LoadAsync();
        }

        private async Task LoadAsync()
        {
            try
            {
                var list = await _apiClient.GetInstantRecoverySessionsAsync();
                Sessions.Clear();
                foreach (var s in list) Sessions.Add(s);
            }
            catch (Exception ex)
            {
                StatusMessage = $"Error: {ex.Message}";
            }
        }

        public async Task RefreshAsync()
        {
            await LoadAsync();
        }

        public async Task StopAsync(string sessionId)
        {
            await _apiClient.StopInstantRecoveryAsync(sessionId);
            await LoadAsync();
        }
    }
}
