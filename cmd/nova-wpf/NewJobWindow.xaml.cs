using System.Collections.ObjectModel;
using System.Windows;
using System.Windows.Controls;
using Microsoft.Win32;

namespace NovaBackup.WPF
{
    public partial class NewJobWindow : Window
    {
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

            // Save job configuration
            var jobConfig = new
            {
                Name = txtJobName.Text,
                Description = txtDescription.Text,
                Type = ((ComboBoxItem)cmbBackupType.SelectedItem)?.Content?.ToString(),
                Sources = SourceItems,
                Repository = ((ComboBoxItem)cmbRepository.SelectedItem)?.Content?.ToString(),
                Compression = chkCompression.IsChecked == true,
                Encryption = chkEncryption.IsChecked == true,
                Schedule = rbDaily.IsChecked == true ? "Daily" : "Weekly"
            };

            // In real implementation, save to database
            MessageBox.Show($"Job '{jobConfig.Name}' created successfully!\n\n" +
                          $"Sources: {jobConfig.Sources.Count} items\n" +
                          $"Compression: {(jobConfig.Compression ? "Enabled" : "Disabled")}\n" +
                          $"Encryption: {(jobConfig.Encryption ? "Enabled" : "Disabled")}",
                "Job Created", MessageBoxButton.OK, MessageBoxImage.Information);

            DialogResult = true;
            Close();
        }
    }
}
