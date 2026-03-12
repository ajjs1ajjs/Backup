using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class TapeView : Window
    {
        public TapeView()
        {
            InitializeComponent();
            DataContext = new TapeViewModel(new MockApiClient());
        }

        private void CloseButton_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }
}
