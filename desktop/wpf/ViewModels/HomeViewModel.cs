using CommunityToolkit.Mvvm.ComponentModel;
using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;
using NovaBackup.GUI.Services;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Views;

namespace NovaBackup.GUI.ViewModels
{
    public partial class HomeViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;
        private readonly INavigationService _navigationService;

        [ObservableProperty]
        private string _systemStatus = "Loading...";
        
        [ObservableProperty]
        private int _activeJobs = 0;

        [ObservableProperty]
        private int _successJobs = 0;

        [ObservableProperty]
        private int _warningJobs = 0;

        [ObservableProperty]
        private int _failedJobs = 0;

        [ObservableProperty]
        private double _storageUsed = 0;

        [ObservableProperty]
        private double _totalProcessedGB = 0;

        [ObservableProperty]
        private string _backupBottleneck = "Waiting for data...";

        [ObservableProperty]
        private ObservableCollection<string> _recentActivity = new();

        [ObservableProperty]
        private ObservableCollection<Models.RecoverySessionModel> _activeSessions = new();

        [ObservableProperty]
        private bool _hasActiveSessions = false;

        public HomeViewModel(IApiClient apiClient, INavigationService navigationService)
        {
            _apiClient = apiClient;
            _navigationService = navigationService;
            _ = LoadDataAsync();
        }

        private async Task LoadDataAsync()
        {
            try
            {
                var stats = await _apiClient.GetDashboardStatsAsync();
                
                SystemStatus = stats.SystemStatus;
                ActiveJobs = stats.ActiveJobs;
                SuccessJobs = stats.SuccessJobs;
                WarningJobs = stats.WarningJobs;
                FailedJobs = stats.FailedJobs;
                StorageUsed = stats.StorageUsed;
                TotalProcessedGB = stats.TotalProcessedGB;
                BackupBottleneck = stats.BackupBottleneck;
                
                RecentActivity.Clear();
                foreach(var activity in stats.RecentActivity)
                {
                    RecentActivity.Add(activity);
                }

                // Load Active Recovery Sessions
                var sessions = await _apiClient.GetInstantRecoverySessionsAsync();
                ActiveSessions.Clear();
                foreach (var session in sessions)
                {
                    ActiveSessions.Add(session);
                }
                HasActiveSessions = ActiveSessions.Count > 0;
            }
            catch (Exception ex)
            {
                SystemStatus = $"Offline - Cannot connect to backend service";
                RecentActivity.Add($"Connection Error: {ex.Message}");
            }
        }

        [RelayCommand]
        private void OpenInstantRecoveryWizard()
        {
            var wizard = new InstantRecoveryWizardWindow(_apiClient);
            wizard.Owner = System.Windows.Application.Current.MainWindow;
            wizard.ShowDialog();
        }
        [RelayCommand]
        private void NavigateToSessions()
        {
            _navigationService.NavigateTo<RecoverySessionsViewModel>();
        }

        [RelayCommand]
        private async Task StopSessionAsync(string sessionId)
        {
            try
            {
                await _apiClient.StopInstantRecoveryAsync(sessionId);
                await LoadDataAsync(); // Refresh dashboard
            }
            catch (Exception ex)
            {
                RecentActivity.Add($"Failed to stop session: {ex.Message}");
            }
        }
        [RelayCommand]
        private void OpenLogs(string sessionId)
        {
            RecentActivity.Add($"Opening logs for session: {sessionId}");
            // In a real implementation:
            // var logViewer = new LogViewerWindow(sessionId);
            // logViewer.Show();
        }

        [RelayCommand]
        private void OpenRecoverySessionsWindowMinimal()
        {
            var window = new NovaBackup.GUI.Views.RecoverySessionsWindowMinimal();
            window.Owner = System.Windows.Application.Current.MainWindow;
            window.ShowDialog();
        }

        [RelayCommand]
        private void OpenRecoverySessionsWindowMVVM()
        {
            var window = new NovaBackup.GUI.Views.RecoverySessionsWindowMVVM(_apiClient);
            window.Owner = System.Windows.Application.Current.MainWindow;
            window.ShowDialog();
        }

        [RelayCommand]
        private void OpenRecoverySessionsWindow()
        {
            var window = new RecoverySessionsWindow(_apiClient);
            window.Owner = System.Windows.Application.Current.MainWindow;
            window.ShowDialog();
        }
    }
}
