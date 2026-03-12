using System.Collections.Generic;
using System.Windows;
using NovaBackup.GUI.Models;

namespace NovaBackup.GUI.Views
{
    public partial class RecoverySessionsWindowMinimal : Window
    {
        public RecoverySessionsWindowMinimal()
        {
            InitializeComponent();
            // Demo data
            var items = new List<RecoverySessionModel>
            {
                new RecoverySessionModel { SessionID = "s1", VMName = "DemoVM-01", Status = "Running", Progress = 0.25 },
                new RecoverySessionModel { SessionID = "s2", VMName = "DemoVM-02", Status = "Completed", Progress = 1.0 }
            };
            SessionsList.ItemsSource = items;
        }

        private void Refresh_Click(object sender, RoutedEventArgs e)
        {
            var items = new List<RecoverySessionModel>
            {
                new RecoverySessionModel { SessionID = "s1", VMName = "DemoVM-01", Status = "Running", Progress = 0.40 },
                new RecoverySessionModel { SessionID = "s2", VMName = "DemoVM-02", Status = "Completed", Progress = 1.0 }
            };
            SessionsList.ItemsSource = items;
        }

        private void Stop_Click(object sender, RoutedEventArgs e)
        {
            // Minimal placeholder
        }

        private void Close_Click(object sender, RoutedEventArgs e)
        {
            this.Close();
        }
    }
}
