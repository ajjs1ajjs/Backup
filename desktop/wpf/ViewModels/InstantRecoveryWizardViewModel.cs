using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;
using System;
using System.Collections.ObjectModel;
using System.Linq;
using System.Threading.Tasks;
using System.Windows;

namespace NovaBackup.GUI.ViewModels
{
    public partial class InstantRecoveryWizardViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private int _currentStep = 1;

        [ObservableProperty]
        private string _wizardTitle = "Instant VM Recovery Wizard";

        // Step 1: Select VM and Restore Point
        [ObservableProperty]
        private ObservableCollection<JobModel> _availableJobs = new();

        [ObservableProperty]
        private JobModel? _selectedJob;

        [ObservableProperty]
        private ObservableCollection<RestorePointModel> _restorePoints = new();

        [ObservableProperty]
        private RestorePointModel? _selectedRestorePoint;

        // Step 2: Target Settings
        [ObservableProperty]
        private string _vmName = string.Empty;

        [ObservableProperty]
        private string _recoveryType = "Hyper-V"; // Default to Hyper-V

        // Step 3: Status
        [ObservableProperty]
        private bool _isBusy;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        public InstantRecoveryWizardViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
            _ = LoadJobsAsync();
        }

        private async Task LoadJobsAsync()
        {
            try
            {
                var jobs = await _apiClient.GetJobsAsync();
                AvailableJobs.Clear();
                foreach (var job in jobs) AvailableJobs.Add(job);
            }
            catch { StatusMessage = "Failed to load jobs."; }
        }

        partial void OnSelectedJobChanged(JobModel? value)
        {
            if (value != null)
            {
                _ = LoadRestorePointsAsync(value.Id);
                VmName = value.Name + "_Recovered";
            }
        }

        private async Task LoadRestorePointsAsync(string jobId)
        {
            try
            {
                var rps = await _apiClient.GetRestorePointsAsync(jobId);
                RestorePoints.Clear();
                foreach (var rp in rps) RestorePoints.Add(rp);
                SelectedRestorePoint = RestorePoints.FirstOrDefault();
            }
            catch { StatusMessage = "Failed to load restore points."; }
        }

        [RelayCommand]
        private void NextStep()
        {
            if (CurrentStep < 3)
                CurrentStep++;
        }

        [RelayCommand]
        private void PreviousStep()
        {
            if (CurrentStep > 1)
                CurrentStep--;
        }

        [RelayCommand]
        private async Task Finish(Window window)
        {
            if (SelectedRestorePoint == null)
            {
                StatusMessage = "Please select a restore point.";
                return;
            }

            if (string.IsNullOrWhiteSpace(VmName))
            {
                StatusMessage = "VM name is required.";
                return;
            }

            IsBusy = true;
            StatusMessage = "Starting instant recovery session...";

            try
            {
                bool success = await _apiClient.StartInstantRecoveryAsync(SelectedRestorePoint.Id, VmName);
                if (success)
                {
                    StatusMessage = "Instant recovery started! Mounting to " + RecoveryType;
                    await Task.Delay(2000);
                    window?.Close();
                }
                else
                {
                    StatusMessage = "Failed to start instant recovery.";
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

        [RelayCommand]
        private void Cancel(Window window)
        {
            window?.Close();
        }
    }
}
