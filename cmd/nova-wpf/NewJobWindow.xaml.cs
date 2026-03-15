using System;
using System.Collections.ObjectModel;
using System.Windows;
using System.Windows.Controls;
using Microsoft.Win32;
using System.IO;

namespace NovaBackup.WPF
{
    public partial class NewJobWindow : Window
    {
        public BackupJob CreatedJob { get; private set; }
        public ObservableCollection<string> SourceItems { get; set; }

        public NewJobWindow()
        {
            InitializeComponent();
            SourceItems = new ObservableCollection<string>();
            lstSource.ItemsSource = SourceItems;
        }

        private void BtnAddFolder_Click(object sender, RoutedEventArgs e)
        {
            var dialog = new System.Windows.Forms.FolderBrowserDialog
            {
                Description = "Select folder to backup",
                ShowNewFolderButton = false
            };

            if (dialog.ShowDialog() == System.Windows.Forms.DialogResult.OK)
            {
                SourceItems.Add($"📁 {dialog.SelectedPath}");
            }
        }

        private void BtnAddFile_Click(object sender, RoutedEventArgs e)
        {
            var dialog = new OpenFileDialog
            {
                Title = "Select file to backup",
                Multiselect = true
            };

            if (dialog.ShowDialog() == true)
            {
                foreach (var file in dialog.FileNames)
                {
                    SourceItems.Add($"📄 {file}");
                }
            }
        }

        private void BtnCancel_Click(object sender, RoutedEventArgs e)
        {
            DialogResult = false;
            Close();
        }

        private void BtnFinish_Click(object sender, RoutedEventArgs e)
        {
            if (string.IsNullOrWhiteSpace(txtJobName.Text))
            {
                MessageBox.Show("Please enter a job name", "Validation Error",
                    MessageBoxButton.OK, MessageBoxImage.Warning);
                return;
            }

            if (SourceItems.Count == 0)
            {
                MessageBox.Show("Please add at least one source file or folder", "Validation Error",
                    MessageBoxButton.OK, MessageBoxImage.Warning);
                return;
            }

            // Extract actual paths from source items
            var sources = new System.Collections.Generic.List<string>();
            foreach (var item in SourceItems)
            {
                var path = item.StartsWith("📁 ") || item.StartsWith("📄 ")
                    ? item.Substring(3)
                    : item;
                sources.Add(path);
            }

            // Create job object
            CreatedJob = new BackupJob
            {
                Name = txtJobName.Text,
                Description = txtDescription.Text,
                Type = ((ComboBoxItem)cmbBackupType.SelectedItem)?.Content?.ToString() ?? "File Backup",
                Sources = sources,
                Destination = ((ComboBoxItem)cmbRepository.SelectedItem)?.Content?.ToString() ?? "C:\\ProgramData\\NovaBackup\\Backups",
                Compression = chkCompression.IsChecked == true,
                Encryption = chkEncryption.IsChecked == true,
                ScheduleType = rbDaily.IsChecked == true ? "Daily" : "Weekly",
                ScheduleTime = rbDaily.IsChecked == true ? txtDailyTime.Text : txtWeeklyTime.Text,
                ScheduleDays = new System.Collections.Generic.List<string>()
            };

            if (rbWeekly.IsChecked == true)
            {
                if (chkMonday.IsChecked == true) CreatedJob.ScheduleDays.Add("Monday");
                if (chkWednesday.IsChecked == true) CreatedJob.ScheduleDays.Add("Wednesday");
                if (chkFriday.IsChecked == true) CreatedJob.ScheduleDays.Add("Friday");
            }

            // Calculate next run
            if (CreatedJob.ScheduleType == "Daily")
            {
                var time = TimeSpan.Parse(CreatedJob.ScheduleTime);
                var next = DateTime.Now.Date.AddDays(1).Add(time);
                CreatedJob.NextRun = next;
            }
            else
            {
                CreatedJob.NextRun = DateTime.Now.AddDays(7); // Simplified
            }

            // Save to file
            JobManager.AddJob(CreatedJob);

            MessageBox.Show($"Job '{CreatedJob.Name}' created successfully!\n\n" +
                          $"Type: {CreatedJob.Type}\n" +
                          $"Sources: {CreatedJob.Sources.Count} items\n" +
                          $"Destination: {CreatedJob.Destination}\n" +
                          $"Compression: {(CreatedJob.Compression ? "Enabled" : "Disabled")}\n" +
                          $"Encryption: {(CreatedJob.Encryption ? "Enabled" : "Disabled")}\n" +
                          $"Schedule: {CreatedJob.ScheduleType} at {CreatedJob.ScheduleTime}",
                "Job Created", MessageBoxButton.OK, MessageBoxImage.Information);

            DialogResult = true;
            Close();
        }
    }
}
