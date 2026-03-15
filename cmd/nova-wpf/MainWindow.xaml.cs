using System;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.ServiceProcess;
using System.Windows;
using System.Windows.Media;

namespace NovaBackup.WPF
{
    public partial class MainWindow : Window
    {
        private const string ServiceName = "NovaBackup";
        public ObservableCollection<JobItem> Jobs { get; set; }

        public MainWindow()
        {
            InitializeComponent();
            Jobs = new ObservableCollection<JobItem>
            {
                new JobItem { Name = "Daily Documents Backup", Type = "File Backup", StatusIcon = "✅", Status = "Active", LastRun = "2026-03-14 02:00", NextRun = "2026-03-15 02:00", Schedule = "Daily 2AM" },
                new JobItem { Name = "Weekly System Backup", Type = "System", StatusIcon = "✅", Status = "Active", LastRun = "2026-03-09 22:00", NextRun = "2026-03-16 22:00", Schedule = "Weekly Sun" },
                new JobItem { Name = "Cloud Sync Job", Type = "Cloud", StatusIcon = "⏸️", Status = "Disabled", LastRun = "2026-03-10 06:00", NextRun = "-", Schedule = "Manual" }
            };
            jobsGrid.ItemsSource = Jobs;
            UpdateServiceStatus();
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

        private void BtnNewJob_Click(object sender, RoutedEventArgs e) => ShowMessage("New Backup Job Wizard");
        private void BtnRunNow_Click(object sender, RoutedEventArgs e) => RunSelectedJob();
        private void BtnRestore_Click(object sender, RoutedEventArgs e) => ShowMessage("Restore Wizard");
        private void BtnVerify_Click(object sender, RoutedEventArgs e) => ShowMessage("Backup Verification");
        private void BtnCreateJob_Click(object sender, RoutedEventArgs e) => ShowMessage("Create Job Wizard");
        private void BtnEditJob_Click(object sender, RoutedEventArgs e) => ShowMessage("Edit Job");
        private void BtnDeleteJob_Click(object sender, RoutedEventArgs e) => ShowMessage("Delete Job Confirmation");
        private void BtnEnableJob_Click(object sender, RoutedEventArgs e) => ShowMessage("Job Enabled");
        private void BtnDisableJob_Click(object sender, RoutedEventArgs e) => ShowMessage("Job Disabled");
        private void BtnAddRepo_Click(object sender, RoutedEventArgs e) => ShowMessage("Add Repository Wizard");
        private void BtnAddServer_Click(object sender, RoutedEventArgs e) => ShowMessage("Add Server Wizard");
        private void BtnSessions_Click(object sender, RoutedEventArgs e) => ShowMessage("Sessions History");
        private void BtnAlarms_Click(object sender, RoutedEventArgs e) => ShowMessage("Alarms and Warnings");
        private void BtnHelp_Click(object sender, RoutedEventArgs e) => Process.Start(new ProcessStartInfo { FileName = "https://github.com/ajjs1ajjs/Backup/wiki", UseShellExecute = true });

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

        private void RunSelectedJob()
        {
            if (jobsGrid.SelectedItem is JobItem job)
            {
                try
                {
                    var process = new Process
                    {
                        StartInfo = new ProcessStartInfo
                        {
                            FileName = "C:\\Program Files\\NovaBackup\\nova.exe",
                            Arguments = "debug",
                            UseShellExecute = false,
                            CreateNoWindow = true
                        }
                    };
                    process.Start();
                    MessageBox.Show($"Started backup job: {job.Name}", "Job Started",
                        MessageBoxButton.OK, MessageBoxImage.Information);
                }
                catch (Exception ex)
                {
                    MessageBox.Show($"Error starting job: {ex.Message}", "Error",
                        MessageBoxButton.OK, MessageBoxImage.Error);
                }
            }
            else
            {
                MessageBox.Show("Please select a job first", "No Selection",
                    MessageBoxButton.OK, MessageBoxImage.Warning);
            }
        }

        private void ShowMessage(string message)
        {
            MessageBox.Show(message + "\n\n(This is a demo - full implementation coming soon)",
                "NovaBackup Enterprise", MessageBoxButton.OK, MessageBoxImage.Information);
        }
    }

    public class JobItem
    {
        public string Name { get; set; }
        public string Type { get; set; }
        public string StatusIcon { get; set; }
        public string Status { get; set; }
        public string LastRun { get; set; }
        public string NextRun { get; set; }
        public string Schedule { get; set; }
    }
}
