using System.Windows;
using System.Windows.Controls;

namespace NovaBackup.WPF
{
    public partial class RestoreWindow : Window
    {
        public RestoreWindow()
        {
            InitializeComponent();
        }

        private void BtnCancel_Click(object sender, RoutedEventArgs e)
        {
            DialogResult = false;
            Close();
        }

        private void BtnRestore_Click(object sender, RoutedEventArgs e)
        {
            var selectedBackup = lstBackups.SelectedItem as ListBoxItem;
            if (selectedBackup == null)
            {
                MessageBox.Show("Please select a backup to restore from", "No Selection",
                    MessageBoxButton.OK, MessageBoxImage.Warning);
                return;
            }

            var destination = rbOriginal.IsChecked == true
                ? "Original location"
                : txtPath.Text;

            MessageBox.Show(
                $"Restore started!\n\n" +
                $"From: {selectedBackup.Content}\n" +
                $"To: {destination}\n\n" +
                "Monitor progress in the Monitoring tab.",
                "Restore Started",
                MessageBoxButton.OK,
                MessageBoxImage.Information);

            DialogResult = true;
            Close();
        }
    }
}
