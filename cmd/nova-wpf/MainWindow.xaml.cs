using System;
using System.Diagnostics;
using System.ServiceProcess;
using System.Windows;
using System.Windows.Media;

namespace NovaBackup.WPF
{
    public partial class MainWindow : Window
    {
        private const string ServiceName = "NovaBackup";

        public MainWindow()
        {
            InitializeComponent();
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
                        btnRunNow.IsEnabled = true;
                        btnNewJob.IsEnabled = true;
                    }
                    else if (service.Status == ServiceControllerStatus.Stopped)
                    {
                        statusService.Text = "🔴 Service: Stopped";
                        statusService.Foreground = Brushes.Red;
                        btnRunNow.IsEnabled = false;
                        btnNewJob.IsEnabled = false;
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
                btnRunNow.IsEnabled = false;
                btnNewJob.IsEnabled = false;
            }
        }

        private void BtnNewJob_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show("Job creation dialog would open here.", "New Job",
                MessageBoxButton.OK, MessageBoxImage.Information);
        }

        private void BtnRunNow_Click(object sender, RoutedEventArgs e)
        {
            try
            {
                var process = new Process
                {
                    StartInfo = new ProcessStartInfo
                    {
                        FileName = "C:\\Program Files\\NovaBackup\\nova-cli.exe",
                        Arguments = "job run --all",
                        UseShellExecute = false,
                        CreateNoWindow = true
                    }
                };
                process.Start();

                MessageBox.Show("Backup job started!", "Success",
                    MessageBoxButton.OK, MessageBoxImage.Information);
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Error: {ex.Message}", "Error",
                    MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }

        private void BtnRefresh_Click(object sender, RoutedEventArgs e)
        {
            UpdateServiceStatus();
        }

        private void BtnAbout_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show(
                "NovaBackup Enterprise v6.0\n\n" +
                "Modern Backup and Recovery Platform\n\n" +
                "Features:\n" +
                "• File and Folder Backup\n" +
                "• Hyper-V VM Backup\n" +
                "• S3 Compatible Storage\n" +
                "• Deduplication and Compression\n" +
                "• Job Scheduling\n\n" +
                "© 2024 NovaBackup Team",
                "About NovaBackup",
                MessageBoxButton.OK,
                MessageBoxImage.Information);
        }
    }
}
