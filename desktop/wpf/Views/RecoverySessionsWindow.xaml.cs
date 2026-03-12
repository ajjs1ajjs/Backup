using System;
using System.Windows;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class RecoverySessionsWindow : Window
    {
        public RecoverySessionsWindow(IApiClient apiClient)
        {
            InitializeComponent();
            DataContext = new RecoverySessionsViewModel(apiClient);
        }
    }
}
