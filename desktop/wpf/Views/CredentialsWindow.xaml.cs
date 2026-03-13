using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class CredentialsWindow : Window
    {
        private readonly CredentialsViewModel _viewModel;

        public CredentialsWindow(CredentialService credentialService)
        {
            InitializeComponent();
            _viewModel = new CredentialsViewModel(credentialService);
            DataContext = _viewModel;
            Loaded += async (s, e) => await _viewModel.LoadDataCommand.ExecuteAsync(null);
        }
    }
}
