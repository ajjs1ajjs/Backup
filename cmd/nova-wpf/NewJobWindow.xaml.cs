using System;
using System.Collections.ObjectModel;
using System.Windows;
using System.Windows.Controls;
using Microsoft.Win32;

namespace NovaBackup.WPF
{
    public partial class NewJobWindow : Window
    {
        public BackupJob? CreatedJob { get; private set; }
        public ObservableCollection<string> SourceItems { get; set; }

        public NewJobWindow()
        {
            InitializeComponent();
            SourceItems = new ObservableCollection<string>();
            lstSource.ItemsSource = SourceItems;
        }

        private void BtnAddFolder_Click(object sender, RoutedEventArgs e)
        {
            var dialog = new System.Windows.Forms.FolderBrowserDialog { Description = "Оберіть папку", ShowNewFolderButton = false };
            if (dialog.ShowDialog() == System.Windows.Forms.DialogResult.OK)
            {
                SourceItems.Add($"📁 {dialog.SelectedPath}");
            }
        }

        private void BtnAddFile_Click(object sender, RoutedEventArgs e)
        {
            var dialog = new OpenFileDialog { Title = "Оберіть файли", Multiselect = true };
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
                MessageBox.Show("Введіть назву завдання", "Помилка", MessageBoxButton.OK, MessageBoxImage.Warning);
                return;
            }

            if (SourceItems.Count == 0)
            {
                MessageBox.Show("Додайте хоча б один файл або папку", "Помилка", MessageBoxButton.OK, MessageBoxImage.Warning);
                return;
            }

            var sources = new System.Collections.Generic.List<string>();
            foreach (var item in SourceItems)
            {
                var path = item.StartsWith("📁 ") || item.StartsWith("📄 ") ? item.Substring(3) : item;
                sources.Add(path);
            }

            CreatedJob = new BackupJob
            {
                Name = txtJobName.Text,
                Type = ((ComboBoxItem)cmbBackupType.SelectedItem)?.Content?.ToString() ?? "Файли",
                Sources = sources,
                Destination = ((ComboBoxItem)cmbRepository.SelectedItem)?.Content?.ToString() ?? "D:\\Backups",
                Compression = chkCompression.IsChecked == true,
                Encryption = chkEncryption.IsChecked == true,
                ScheduleType = rbDaily.IsChecked == true ? "Щодня" : "Щотижня",
                Schedule = rbDaily.IsChecked == true ? $"Щодня {txtDailyTime.Text}" : $"Щотижня {txtWeeklyTime.Text}",
                StatusIcon = "✅",
                Status = "Активно",
                LastRun = "—",
                NextRun = rbDaily.IsChecked == true ? $"Завтра {txtDailyTime.Text}" : "Наступного тижня"
            };

            JobManager.AddJob(CreatedJob);
            MessageBox.Show($"Завдання '{CreatedJob.Name}' створено!", "Успіх", MessageBoxButton.OK, MessageBoxImage.Information);
            DialogResult = true;
            Close();
        }
    }
}
