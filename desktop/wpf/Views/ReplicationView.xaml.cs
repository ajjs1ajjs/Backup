using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class ReplicationView : Window
    {
        public ReplicationView(IApiClient apiClient)
        {
            InitializeComponent();
            DataContext = new ReplicationViewModel(apiClient);
        }

        private void CloseButton_Click(object sender, RoutedEventArgs e)
        {
            Close();
        }
    }

    // Alias for backward compatibility
    public class ReplicationWindow : ReplicationView
    {
        public ReplicationWindow(IApiClient apiClient) : base(apiClient)
        {
        }
    }
}
