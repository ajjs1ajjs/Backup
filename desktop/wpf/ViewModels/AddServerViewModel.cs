using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;
using System;
using System.Threading.Tasks;
using System.Windows;

namespace NovaBackup.GUI.ViewModels
{
    public partial class AddServerViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private string _serverName = string.Empty;

        [ObservableProperty]
        private string _serverAddress = string.Empty;

        [ObservableProperty]
        private string _serverType = "Hyper-V"; // Default

        [ObservableProperty]
        private string _username = string.Empty;

        [ObservableProperty]
        private string _password = string.Empty;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        [ObservableProperty]
        private bool _isBusy;

        public AddServerViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task AddServer(Window window)
        {
            if (string.IsNullOrWhiteSpace(ServerName) || string.IsNullOrWhiteSpace(ServerAddress))
            {
                StatusMessage = "Name and Address are required.";
                return;
            }

            IsBusy = true;
            StatusMessage = "Adding server...";

            try
            {
                bool success = await _apiClient.AddServerAsync(ServerName, ServerAddress, ServerType, Username, Password);
                if (success)
                {
                    StatusMessage = "Server added successfully!";
                    await Task.Delay(1000);
                    window?.Close();
                }
                else
                {
                    StatusMessage = "Failed to add server. Check logs.";
                }
            }
            catch (Exception ex)
            {
                StatusMessage = $"Error: {ex.Message}";
            }
            finally
            {
                IsBusy = false;
            }
        }

        [RelayCommand]
        private void Cancel(Window window)
        {
            window?.Close();
        }
    }
}
