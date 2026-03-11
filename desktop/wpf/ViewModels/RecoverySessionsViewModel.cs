using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;
using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;

namespace NovaBackup.GUI.ViewModels
{
    public partial class RecoverySessionsViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<RecoverySessionModel> _sessions = new();

        [ObservableProperty]
        private bool _isBusy;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        public RecoverySessionsViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
            _ = RefreshSessionsAsync();
        }

        [RelayCommand]
        public async Task RefreshSessionsAsync()
        {
            IsBusy = true;
            try
            {
                var sessions = await _apiClient.GetInstantRecoverySessionsAsync();
                Sessions.Clear();
                foreach (var session in sessions)
                {
                    Sessions.Add(session);
                }
                StatusMessage = $"Updated at {DateTime.Now:HH:mm:ss}";
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
        private async Task StopSession(RecoverySessionModel? session)
        {
            if (session == null) return;

            IsBusy = true;
            StatusMessage = $"Stopping session {session.SessionID}...";
            try
            {
                bool success = await _apiClient.StopInstantRecoveryAsync(session.SessionID);
                if (success)
                {
                    Sessions.Remove(session);
                    StatusMessage = "Session stopped successfully.";
                }
                else
                {
                    StatusMessage = "Failed to stop session.";
                }
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
    }
}
