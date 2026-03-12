using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class ReportsView : Window
    {
        public ReportsView()
        {
            InitializeComponent();
            DataContext = new ReportsViewModel(new MockApiClient());
        }

        private void CloseButton_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }
}
