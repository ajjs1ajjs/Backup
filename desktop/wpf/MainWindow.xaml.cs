using System;
using System.Collections.Generic;
using System.Windows;
using System.Windows.Controls;

namespace NovaBackup.GUI
{
    /// <summary>
    /// Interaction logic for MainWindow.xaml
    /// </summary>
    public partial class MainWindow : Window
    {
        public MainWindow()
        {
            InitializeComponent();
            InitializeSampleData();
        }

        private void InitializeSampleData()
        {
            // Sample jobs data
            var jobs = new List<JobInfo>
            {
                new JobInfo { Name = "Daily Documents Backup", Type = "Files", Status = "Active", LastRun = "2026-03-11 02:00", NextRun = "2026-03-12 02:00" },
                new JobInfo { Name = "Weekly System Backup", Type = "System", Status = "Active", LastRun = "2026-03-08 22:00", NextRun = "2026-03-15 22:00" },
                new JobInfo { Name = "Database Backup", Type = "SQL Server", Status = "Active", LastRun = "2026-03-11 01:00", NextRun = "2026-03-12 01:00" }
            };

            dgJobs.ItemsSource = jobs;
        }

        // Navigation button click handler
        private void NavButton_Click(object sender, RoutedEventArgs e)
        {
            // Reset all navigation buttons
            btnDashboard.Tag = null;
            btnJobs.Tag = null;
            btnStorage.Tag = null;
            btnReports.Tag = null;
            btnInfrastructure.Tag = null;
            btnHistory.Tag = null;

            // Set selected button
            if (sender is Button button)
            {
                button.Tag = "Selected";
            }
        }

        // Start backup button click
        private void BtnStartBackup_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show("Starting backup process...", "NovaBackup", MessageBoxButton.OK, MessageBoxImage.Information);
        }

        // Create job button click
        private void BtnCreateJob_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show("Opening job creation wizard...", "NovaBackup", MessageBoxButton.OK, MessageBoxImage.Information);
        }

        // Settings button click
        private void BtnSettings_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show("Opening settings...", "NovaBackup", MessageBoxButton.OK, MessageBoxImage.Information);
        }
    }

    // Sample job info class
    public class JobInfo
    {
        public string Name { get; set; } = "";
        public string Type { get; set; } = "";
        public string Status { get; set; } = "";
        public string LastRun { get; set; } = "";
        public string NextRun { get; set; } = "";
    }
}
