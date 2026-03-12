using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;
using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;
using System.Windows;

namespace NovaBackup.GUI.ViewModels
{
    public partial class JobWizardViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private int _currentStep = 1;

        [ObservableProperty]
        private string _wizardTitle = "Create New Backup Job";

        [ObservableProperty]
        private string _jobName = string.Empty;

        [ObservableProperty]
        private string _jobDescription = string.Empty;

        [ObservableProperty]
        private ObservableCollection<InfrastructureNode> _availableSources = new();

        [ObservableProperty]
        private InfrastructureNode? _selectedSource;

        [ObservableProperty]
        private ObservableCollection<RepositoryModel> _availableRepositories = new();

        [ObservableProperty]
        private RepositoryModel? _selectedRepository;

        [ObservableProperty]
        private int _retentionDays = 30;

        [ObservableProperty]
        private bool _enableGuestProcessing;

        [ObservableProperty]
        private string _guestCredentialsId = string.Empty;

        [ObservableProperty]
        private bool _isBusy;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        public JobWizardViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
            _ = LoadInitialDataAsync();
        }

        private async Task LoadInitialDataAsync()
        {
            try
            {
                var sources = await _apiClient.GetInfrastructureTreeAsync();
                AvailableSources.Clear();
                foreach (var src in sources) AvailableSources.Add(src);

                var repos = await _apiClient.GetRepositoriesAsync();
                AvailableRepositories.Clear();
                foreach (var repo in repos) AvailableRepositories.Add(repo);
            }
            catch { /* Log error */ }
        }

        [RelayCommand]
        private void NextStep()
        {
            if (CurrentStep < 6)
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
            if (string.IsNullOrWhiteSpace(JobName))
            {
                StatusMessage = "Job name is required.";
                return;
            }

            if (_selectedSource == null)
            {
                StatusMessage = "Please select a backup source.";
                return;
            }

            if (_selectedRepository == null)
            {
                StatusMessage = "Please select a backup destination.";
                return;
            }

            IsBusy = true;
            StatusMessage = "Creating job...";

            try
            {
                var newJob = new JobModel
                {
                    Name = JobName,
                    Platform = SelectedSource?.Name ?? "General",
                    RetentionDays = RetentionDays,
                    EnableGuestProcessing = EnableGuestProcessing,
                    GuestCredentialsId = GuestCredentialsId
                };

                bool success = await _apiClient.CreateJobAsync(newJob);
                if (success)
                {
                    StatusMessage = "Job created successfully!";
                    await Task.Delay(1000);
                    window?.Close();
                }
                else
                {
                    StatusMessage = "Failed to create job.";
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
