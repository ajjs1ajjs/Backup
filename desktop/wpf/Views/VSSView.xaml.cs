using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class VSSView : Window
    {
        public VSSView()
        {
            InitializeComponent();
            DataContext = new VSSViewModel(new MockApiClient());
        }

        private void CloseButton_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }
}
