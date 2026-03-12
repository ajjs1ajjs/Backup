using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class RecoverySessionsWindowMVVM : Window
    {
        private readonly RecoverySessionsViewModelMVVM _vm;
        public RecoverySessionsWindowMVVM(NovaBackup.GUI.Services.IApiClient apiClient)
        {
            InitializeComponent();
            _vm = new RecoverySessionsViewModelMVVM(apiClient);
            DataContext = _vm;
        }
    }
}
