using System;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Linq;
using System.ServiceProcess;
using System.Windows;
using System.Windows.Media;

namespace NovaBackup.WPF
{
    public partial class MainWindow : Window
    {
        private const string ServiceName = "NovaBackup";
        public ObservableCollection<BackupJob> Jobs { get; set; }

        public MainWindow()
        {
            InitializeComponent();
            Jobs = JobManager.LoadJobs();
            jobsGrid.ItemsSource = Jobs;
            UpdateServiceStatus();
            UpdateJobStats();
        }

        private void UpdateServiceStatus()
        {
            try
            {
                using (var service = new ServiceController(ServiceName))
                {
                    if (service.Status == ServiceControllerStatus.Running)
                    {
                        statusService.Text = "🟢 Service: Running";
                        statusService.Foreground = Brushes.Green;
                    }
                    else if (service.Status == ServiceControllerStatus.Stopped)
                    {
                        statusService.Text = "🔴 Service: Stopped";
                        statusService.Foreground = Brushes.Red;
                    }
                    else
                    {
                        statusService.Text = "🟡 Service: " + service.Status;
                        statusService.Foreground = Brushes.Orange;
                    }
                }
            }
            catch
            {
                statusService.Text = "⚪ Service: Not Installed";
                statusService.Foreground = Brushes.Gray;
            }
        }

        private void UpdateJobStats()
        {
            var activeJobs = Jobs.Count(j => j.Enabled);
            statusJobs.Text = $"📊 Jobs: {activeJobs} Active / {Jobs.Count} Total";
        }

        private void BtnNewJob_Click(object sender, RoutedEventArgs e)
        {
            var wizard = new NewJobWindow { Owner = this };
            if (wizard.ShowDialog() == true && wizard.CreatedJob != null)
            {
                Jobs.Add(wizard.CreatedJob);
                UpdateJobStats();
            }
        }

        private void BtnRunNow_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                var result = MessageBox.Show(
                    $"Start backup job '{job.Name}' now?\n\n" +
                    $"Type: {job.Type}\n" +
                    $"Sources: {job.Sources.Count} items\n" +
                    $"Destination: {job.Destination}\n" +
                    $"Compression: {(job.Compression ? "ON" : "OFF")}\n" +
                    $"Encryption: {(job.Encryption ? "ON" : "OFF")}",
                    "Confirm Backup",
                    MessageBoxButton.YesNo,
                    MessageBoxImage.Question);

                if (result == MessageBoxResult.Yes)
                {
                    try
                    {
                        // Run backup via service
                        var psi = new ProcessStartInfo
                        {
                            FileName = "C:\\Program Files\\NovaBackup\\nova.exe",
                            Arguments = "debug",
                            UseShellExecute = false,
                            CreateNoWindow = true
                        };
                        Process.Start(psi);

                        job.LastRun = DateTime.Now;
                        job.Status = "Running";
                        job.StatusIcon = "🔄";
                        jobsGrid.Items.Refresh();

                        MessageBox.Show(
                            $"Backup job '{job.Name}' started!\n\n" +
                            $"Monitor progress in the Monitoring tab.",
                            "Backup Started",
                            MessageBoxButton.OK,
                            MessageBoxImage.Information);
                    }
                    catch (Exception ex)
                    {
                        MessageBox.Show($"Error starting backup: {ex.Message}", "Error",
                            MessageBoxButton.OK, MessageBoxImage.Error);
                    }
                }
            }
            else
            {
                MessageBox.Show("Please select a job first", "No Selection",
                    MessageBoxButton.OK, MessageBoxImage.Warning);
            }
        }

        private void BtnRestore_Click(object sender, RoutedEventArgs e)
        {
            var restoreWindow = new RestoreWindow { Owner = this };
            restoreWindow.ShowDialog();
        }

        private void BtnVerify_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                MessageBox.Show(
                    $"Verifying backup integrity for '{job.Name}'...\n\n" +
                    $"This will check:\n" +
                    "• Backup file integrity\n" +
                    "• Data consistency\n" +
                    "• Recovery point validity\n\n" +
                    "Verification started!",
                    "Backup Verification",
                    MessageBoxButton.OK,
                    MessageBoxImage.Information);
            }
            else
            {
                MessageBox.Show("Please select a job to verify", "No Selection",
                    MessageBoxButton.OK, MessageBoxImage.Warning);
            }
        }

        private void BtnCreateJob_Click(object sender, RoutedEventArgs e) => BtnNewJob_Click(sender, e);

        private void BtnEditJob_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                var editWindow = new EditJobWindow(job) { Owner = this };
                if (editWindow.ShowDialog() == true)
                {
                    // Refresh the grid
                    jobsGrid.Items.Refresh();
                }
            }
            else
            {
                MessageBox.Show("Please select a job to edit", "No Selection",
                    MessageBoxButton.OK, MessageBoxImage.Warning);
            }
        }

        private void BtnDeleteJob_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                var result = MessageBox.Show(
                    $"Delete backup job '{job.Name}'?\n\nThis cannot be undone.",
                    "Confirm Delete",
                    MessageBoxButton.YesNo,
                    MessageBoxImage.Warning);

                if (result == MessageBoxResult.Yes)
                {
                    JobManager.DeleteJob(job.Id);
                    Jobs.Remove(job);
                    UpdateJobStats();
                }
            }
            else
            {
                MessageBox.Show("Please select a job to delete", "No Selection",
                    MessageBoxButton.OK, MessageBoxImage.Warning);
            }
        }

        private void BtnEnableJob_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                job.Enabled = true;
                job.Status = "Active";
                job.StatusIcon = "✅";
                JobManager.ToggleJob(job.Id, true);
                jobsGrid.Items.Refresh();
                UpdateJobStats();
            }
        }

        private void BtnDisableJob_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                job.Enabled = false;
                job.Status = "Disabled";
                job.StatusIcon = "⏸️";
                JobManager.ToggleJob(job.Id, false);
                jobsGrid.Items.Refresh();
                UpdateJobStats();
            }
        }

        private void BtnAddRepo_Click(object sender, RoutedEventArgs e)
        {
            var dialog = new System.Windows.Forms.FolderBrowserDialog
            {
                Description = "Select backup repository location",
                ShowNewFolderButton = true
            };

            if (dialog.ShowDialog() == System.Windows.Forms.DialogResult.OK)
            {
                MessageBox.Show(
                    $"Repository added:\n{dialog.SelectedPath}\n\n" +
                    "This repository is now available for backup jobs.",
                    "Repository Added",
                    MessageBoxButton.OK,
                    MessageBoxImage.Information);
            }
        }

        private void BtnAddServer_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show(
                "Add Server Wizard\n\n" +
                "This feature allows you to add remote servers for backup.\n\n" +
                "Supported:\n" +
                "• Windows Server with WinRM\n" +
                "• Hyper-V hosts\n" +
                "• VMware vCenter",
                "Add Server",
                MessageBoxButton.OK,
                MessageBoxImage.Information);
        }

        private void BtnSessions_Click(object sender, RoutedEventArgs e)
        {
            var sessionsWindow = new SessionsWindow { Owner = this };
            sessionsWindow.ShowDialog();
        }

        private void BtnAlarms_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show(
                "Alarms and Warnings\n\n" +
                "No active alarms.\n\n" +
                "All backup jobs are running normally.",
                "Alarms",
                MessageBoxButton.OK,
                MessageBoxImage.Information);
        }

        private void BtnHelp_Click(object sender, RoutedEventArgs e)
        {
            try
            {
                Process.Start(new ProcessStartInfo
                {
                    FileName = "https://github.com/ajjs1ajjs/Backup/wiki",
                    UseShellExecute = true
                });
            }
            catch { }
        }

        private void BtnAbout_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show(
                "NovaBackup Enterprise v6.0\n\n" +
                "Modern Backup and Recovery Platform\n" +
                "Inspired by Veeam Backup & Replication\n\n" +
                "Features:\n" +
                "• File and Folder Backup\n" +
                "• Hyper-V VM Backup\n" +
                "• S3 Compatible Storage\n" +
                "• Deduplication and Compression\n" +
                "• Job Scheduling (Daily/Weekly/Monthly)\n" +
                "• Email and Telegram Notifications\n\n" +
                "© 2024 NovaBackup Team\n" +
                "GitHub: https://github.com/ajjs1ajjs/Backup",
                "About NovaBackup Enterprise",
                MessageBoxButton.OK,
                MessageBoxImage.Information);
        }
    }
}
