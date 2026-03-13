using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class VSSView : Window
    {
        public VSSView(IApiClient apiClient)
        {
            InitializeComponent();
            DataContext = new VSSViewModel(apiClient);
        }

        private void CloseButton_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }

    // Alias for backward compatibility
    public class VSSWindow : VSSView
    {
        public VSSWindow(IApiClient apiClient) : base(apiClient)
        {
        }
    }
}
