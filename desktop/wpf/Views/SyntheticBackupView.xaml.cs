using System.Windows;

namespace NovaBackup.GUI.Views
{
    public partial class SyntheticBackupView : Window
    {
        public SyntheticBackupView()
        {
            InitializeComponent();
        }

        private void CloseButton_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }
}
