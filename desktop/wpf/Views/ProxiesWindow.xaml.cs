using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class ProxiesWindow : Window
    {
        private readonly ProxiesViewModel _viewModel;

        public ProxiesWindow(IApiClient apiClient)
        {
            InitializeComponent();
            _viewModel = new ProxiesViewModel(apiClient);
            DataContext = _viewModel;
            Loaded += async (s, e) => await _viewModel.LoadDataCommand.ExecuteAsync(null);
        }
    }
}
