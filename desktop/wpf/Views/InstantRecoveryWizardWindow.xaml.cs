using System.Windows;
using NovaBackup.GUI.ViewModels;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.Views
{
    public partial class InstantRecoveryWizardWindow : Window
    {
        public InstantRecoveryWizardWindow(IApiClient apiClient)
        {
            InitializeComponent();
            DataContext = new InstantRecoveryWizardViewModel(apiClient);
        }
    }
}
