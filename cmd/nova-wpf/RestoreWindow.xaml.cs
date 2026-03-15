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
                MessageBox.Show("Оберіть резервну копію", "Не обрано", MessageBoxButton.OK, MessageBoxImage.Warning);
                return;
            }

            MessageBox.Show($"Відновлення запущено!\n\nЗ: {selectedBackup.Content}\nВ: {txtPath.Text}", "Відновлення", MessageBoxButton.OK, MessageBoxImage.Information);
            DialogResult = true;
            Close();
        }
    }
}
