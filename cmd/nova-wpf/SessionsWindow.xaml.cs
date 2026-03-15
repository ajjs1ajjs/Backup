using System;
using System.Collections.ObjectModel;
using System.Windows;

namespace NovaBackup.WPF
{
    public partial class SessionsWindow : Window
    {
        public ObservableCollection<SessionInfo> Sessions { get; set; }

        public SessionsWindow()
        {
            InitializeComponent();

            Sessions = new ObservableCollection<SessionInfo>
            {
                new SessionInfo { JobName = "Daily Documents Backup", StartTime = "2026-03-14 02:00", EndTime = "2026-03-14 02:45", Duration = "45 min", Status = "Success", StatusIcon = "✅", Processed = "1,234 files", Transferred = "1.2 GB" },
                new SessionInfo { JobName = "Weekly System Backup", StartTime = "2026-03-09 22:00", EndTime = "2026-03-10 01:30", Duration = "3h 30min", Status = "Success", StatusIcon = "✅", Processed = "45,678 files", Transferred = "45.6 GB" },
                new SessionInfo { JobName = "Daily Documents Backup", StartTime = "2026-03-13 02:00", EndTime = "2026-03-13 02:42", Duration = "42 min", Status = "Success", StatusIcon = "✅", Processed = "1,230 files", Transferred = "1.1 GB" },
                new SessionInfo { JobName = "Cloud Sync", StartTime = "2026-03-10 06:00", EndTime = "2026-03-10 06:15", Duration = "15 min", Status = "Warning", StatusIcon = "⚠️", Processed = "234 files", Transferred = "2.3 GB" },
                new SessionInfo { JobName = "Daily Documents Backup", StartTime = "2026-03-12 02:00", EndTime = "2026-03-12 02:05", Duration = "5 min", Status = "Failed", StatusIcon = "❌", Processed = "12 files", Transferred = "0 MB" }
            };

            DataContext = this;
        }

        private void BtnRefresh_Click(object sender, RoutedEventArgs e)
        {
            // Reload sessions
            MessageBox.Show("Sessions refreshed!", "Info",
                MessageBoxButton.OK, MessageBoxImage.Information);
        }

        private void BtnClose_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }

    public class SessionInfo
    {
        public string JobName { get; set; }
        public string StartTime { get; set; }
        public string EndTime { get; set; }
        public string Duration { get; set; }
        public string Status { get; set; }
        public string StatusIcon { get; set; }
        public string Processed { get; set; }
        public string Transferred { get; set; }
    }
}
