using System;
using System.Collections.Generic;
using System.Windows;
using System.Windows.Controls;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.Views;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI
{
    /// <summary>
    /// Interaction logic for MainWindow.xaml
    /// </summary>
    public partial class MainWindow : Window
    {
        private readonly IApiClient _apiClient;
        private readonly CredentialService _credentialService;

        public MainWindow(IApiClient apiClient, CredentialService credentialService)
        {
            InitializeComponent();
            _apiClient = apiClient;
            _credentialService = credentialService;
        }

        private void btnCredentials_Click(object sender, RoutedEventArgs e)
        {
            var window = new CredentialsWindow(_credentialService);
            window.Owner = this;
            window.ShowDialog();
        }

        private void btnProxies_Click(object sender, RoutedEventArgs e)
        {
            var window = new ProxiesWindow(_apiClient);
            window.Owner = this;
            window.ShowDialog();
        }

        private void btnVSS_Click(object sender, RoutedEventArgs e)
        {
            var window = new VSSWindow(_apiClient);
            window.Owner = this;
            window.ShowDialog();
        }

        private void btnReplication_Click(object sender, RoutedEventArgs e)
        {
            var window = new ReplicationWindow(_apiClient);
            window.Owner = this;
            window.ShowDialog();
        }
    }
}
