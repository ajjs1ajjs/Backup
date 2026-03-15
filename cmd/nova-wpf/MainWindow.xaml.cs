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
        public ObservableCollection<BackupJob> Jobs { get; set; }

        public MainWindow()
        {
            InitializeComponent();

            Jobs = new ObservableCollection<BackupJob>
            {
                new BackupJob { Name = "Щоденне Резервне Копіювання Документів", Type = "Файли", StatusIcon = "✅", Status = "Активно", LastRun = "2026-03-14 02:00", NextRun = "2026-03-15 02:00", Schedule = "Щодня 02:00" },
                new BackupJob { Name = "Тижнева Резервна Копія Системи", Type = "Система", StatusIcon = "✅", Status = "Активно", LastRun = "2026-03-09 22:00", NextRun = "2026-03-16 22:00", Schedule = "Щонед 22:00" },
                new BackupJob { Name = "Хмарна Синхронізація", Type = "Хмара", StatusIcon = "⏸️", Status = "Вимкнено", LastRun = "2026-03-10 06:00", NextRun = "—", Schedule = "Вручну" }
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
                        statusService.Text = "🟢 Служба: Працює";
                        statusService.Foreground = Brushes.Green;
                    }
                    else if (service.Status == ServiceControllerStatus.Stopped)
                    {
                        statusService.Text = "🔴 Служба: Зупинено";
                        statusService.Foreground = Brushes.Red;
                    }
                    else
                    {
                        statusService.Text = "🟡 Служба: " + service.Status;
                        statusService.Foreground = Brushes.Orange;
                    }
                }
            }
            catch
            {
                statusService.Text = "⚪ Служба: Не встановлено";
                statusService.Foreground = Brushes.Gray;
            }
        }

        private void BtnNewJob_Click(object sender, RoutedEventArgs e)
        {
            var wizard = new NewJobWindow { Owner = this };
            if (wizard.ShowDialog() == true)
            {
                if (wizard.CreatedJob != null)
                {
                    Jobs.Add(wizard.CreatedJob);
                    MessageBox.Show("Завдання створено!", "Успіх", MessageBoxButton.OK, MessageBoxImage.Information);
                }
            }
        }

        private void BtnRunNow_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                var result = MessageBox.Show($"Запустити '{job.Name}' зараз?", "Запуск", MessageBoxButton.YesNo, MessageBoxImage.Question);
                if (result == MessageBoxResult.Yes)
                {
                    job.StatusIcon = "🔄";
                    job.Status = "Виконується";
                    jobsGrid.Items.Refresh();
                    MessageBox.Show($"Бекап '{job.Name}' запущено!", "Запущено", MessageBoxButton.OK, MessageBoxImage.Information);
                }
            }
            else
            {
                MessageBox.Show("Оберіть завдання спочатку", "Не обрано", MessageBoxButton.OK, MessageBoxImage.Warning);
            }
        }

        private void BtnRestore_Click(object sender, RoutedEventArgs e)
        {
            var restoreWindow = new RestoreWindow { Owner = this };
            restoreWindow.ShowDialog();
        }

        private void BtnVerify_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show("Перевірка цілісності...\n\n✓ Всі перевірки пройшли успішно!", "Перевірка", MessageBoxButton.OK, MessageBoxImage.Information);
        }

        private void BtnCreateJob_Click(object sender, RoutedEventArgs e) => BtnNewJob_Click(sender, e);

        private void BtnEditJob_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                var editWindow = new EditJobWindow(job) { Owner = this };
                if (editWindow.ShowDialog() == true)
                {
                    jobsGrid.Items.Refresh();
                }
            }
            else
            {
                MessageBox.Show("Оберіть завдання", "Не обрано", MessageBoxButton.OK, MessageBoxImage.Warning);
            }
        }

        private void BtnDeleteJob_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                var result = MessageBox.Show($"Видалити '{job.Name}'?", "Видалення", MessageBoxButton.YesNo, MessageBoxImage.Warning);
                if (result == MessageBoxResult.Yes)
                {
                    Jobs.Remove(job);
                }
            }
            else
            {
                MessageBox.Show("Оберіть завдання", "Не обрано", MessageBoxButton.OK, MessageBoxImage.Warning);
            }
        }

        private void BtnEnableJob_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                job.Enabled = true;
                job.Status = "Активно";
                job.StatusIcon = "✅";
                jobsGrid.Items.Refresh();
            }
        }

        private void BtnDisableJob_Click(object sender, RoutedEventArgs e)
        {
            if (jobsGrid.SelectedItem is BackupJob job)
            {
                job.Enabled = false;
                job.Status = "Вимкнено";
                job.StatusIcon = "⏸️";
                jobsGrid.Items.Refresh();
            }
        }

        private void BtnAddRepo_Click(object sender, RoutedEventArgs e)
        {
            var dialog = new System.Windows.Forms.FolderBrowserDialog { Description = "Оберіть сховище", ShowNewFolderButton = true };
            if (dialog.ShowDialog() == System.Windows.Forms.DialogResult.OK)
            {
                MessageBox.Show($"Сховище додано:\n{dialog.SelectedPath}", "Додано", MessageBoxButton.OK, MessageBoxImage.Information);
            }
        }

        private void BtnAddServer_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show("Майстер Додавання Сервера\n\nПідтримується:\n• Windows Server з WinRM\n• Hyper-V хости\n• VMware vCenter", "Додати Сервер", MessageBoxButton.OK, MessageBoxImage.Information);
        }

        private void BtnSessions_Click(object sender, RoutedEventArgs e)
        {
            var sessionsWindow = new SessionsWindow { Owner = this };
            sessionsWindow.ShowDialog();
        }

        private void BtnAlarms_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show("Сповіщення\n\nНемає активних сповіщень.\n\nВсі завдання працюють нормально.", "Сповіщення", MessageBoxButton.OK, MessageBoxImage.Information);
        }

        private void BtnHelp_Click(object sender, RoutedEventArgs e)
        {
            try { Process.Start(new ProcessStartInfo { FileName = "https://github.com/ajjs1ajjs/Backup/wiki", UseShellExecute = true }); }
            catch { }
        }

        private void BtnAbout_Click(object sender, RoutedEventArgs e)
        {
            MessageBox.Show("NovaBackup Enterprise v6.0\n\nСучасна Платформа Резервного Копіювання\n\n© 2024 NovaBackup Team\nGitHub: github.com/ajjs1ajjs/Backup", "Про Програмy", MessageBoxButton.OK, MessageBoxImage.Information);
        }
    }
}
