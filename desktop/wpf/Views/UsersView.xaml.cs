using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class UsersView : Window
    {
        public UsersView()
        {
            InitializeComponent();
            DataContext = new UsersViewModel(new MockApiClient());
        }

        private void CloseButton_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }
}
