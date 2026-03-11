using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.Views;
using System.Collections.ObjectModel;
using System;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;

namespace NovaBackup.GUI.ViewModels
{
    public partial class JobsViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;
        private readonly IServiceProvider _serviceProvider;

        [ObservableProperty]
        private string _pageTitle = "Backup Jobs";

        [ObservableProperty]
        private ObservableCollection<JobModel> _jobs = new();

        [ObservableProperty]
        private JobModel? _selectedJob;

        [ObservableProperty]
        private ObservableCollection<RestorePointModel> _restorePoints = new();

        [ObservableProperty]
        private RestorePointModel? _selectedRestorePoint;

        public JobsViewModel(IApiClient apiClient, IServiceProvider serviceProvider)
        {
            _apiClient = apiClient;
            _serviceProvider = serviceProvider;
            _ = LoadJobsAsync();
        }

        async partial void OnSelectedJobChanged(JobModel? value)
        {
            if (value != null)
            {
                await LoadRestorePointsAsync(value.Id);
            }
            else
            {
                RestorePoints.Clear();
            }
        }

        private async Task LoadRestorePointsAsync(string jobId)
        {
            try
            {
                var points = await _apiClient.GetRestorePointsAsync(jobId);
                RestorePoints.Clear();
                foreach (var point in points)
                {
                    RestorePoints.Add(point);
                }
            }
            catch (Exception) { }
        }

        private async Task LoadJobsAsync()
        {
            try
            {
                var jobs = await _apiClient.GetJobsAsync();
                Jobs.Clear();
                foreach (var job in jobs)
                {
                    Jobs.Add(job);
                }
            }
            catch (Exception) { }
        }

        [RelayCommand]
        private void CreateBackupJob()
        {
            var wizard = _serviceProvider.GetRequiredService<JobWizardWindow>();
            wizard.Owner = System.Windows.Application.Current.MainWindow;
            wizard.ShowDialog();
            _ = LoadJobsAsync();
        }

        [RelayCommand]
        private async Task StartJob(JobModel? job)
        {
            if (job == null || string.IsNullOrEmpty(job.Id)) return;
            
            try
            {
                var success = await _apiClient.RunJobAsync(job.Id);
                if (success)
                {
                    job.Status = "Running";
                    System.Windows.MessageBox.Show($"Job {job.Name} started successfully.", "NovaBackup");
                }
                else
                {
                    System.Windows.MessageBox.Show($"Failed to start job {job.Name}.", "Error", System.Windows.MessageBoxButton.OK, System.Windows.MessageBoxImage.Error);
                }
            }
            catch (Exception ex)
            {
                System.Windows.MessageBox.Show($"Error: {ex.Message}", "Error");
            }
        }

        [RelayCommand]
        private void EditJob(JobModel? job)
        {
            if (job == null) return;
            System.Windows.MessageBox.Show($"Editing job: {job.Name}", "Information");
        }

        [RelayCommand]
        private void DeleteJob(JobModel? job)
        {
            if (job == null) return;
            var result = System.Windows.MessageBox.Show($"Are you sure you want to delete {job.Name}?", "Confirm Delete", System.Windows.MessageBoxButton.YesNo, System.Windows.MessageBoxImage.Warning);
            if (result == System.Windows.MessageBoxResult.Yes)
            {
                Jobs.Remove(job);
            }
        }
        [RelayCommand]
        private async Task InstantRecover(RestorePointModel? rp)
        {
            if (rp == null || SelectedJob == null) return;

            var vmName = SelectedJob.Name + "_Recovered_" + DateTime.Now.ToString("HHmm");
            var result = System.Windows.MessageBox.Show($"Do you want to start Instant Recovery for {SelectedJob.Name} from {rp.PointTime}?\n\nA new VM '{vmName}' will be created.", 
                "Instant Recovery", System.Windows.MessageBoxButton.YesNo, System.Windows.MessageBoxImage.Question);
            
            if (result == System.Windows.MessageBoxResult.Yes)
            {
                try
                {
                    var success = await _apiClient.StartInstantRecoveryAsync(rp.Id, vmName);
                    if (success)
                    {
                        System.Windows.MessageBox.Show($"Instant Recovery session started. VHDX is mounted via NFS.\n\nPoint: {rp.PointTime}", "Success");
                    }
                    else
                    {
                        System.Windows.MessageBox.Show("Failed to start Instant Recovery session.", "Error", System.Windows.MessageBoxButton.OK, System.Windows.MessageBoxImage.Error);
                    }
                }
                catch (Exception ex)
                {
                    System.Windows.MessageBox.Show($"Error: {ex.Message}", "Error");
                }
            }
        }
    }
}
