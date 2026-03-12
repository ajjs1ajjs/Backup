using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class RolesView : Window
    {
        public RolesView()
        {
            InitializeComponent();
            DataContext = new RolesViewModel(new MockApiClient());
        }

        private void CloseButton_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }
}
