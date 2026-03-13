using System.Collections.ObjectModel;
using System.Linq;
using System.Threading.Tasks;
using System.Windows;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class ProxiesViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<ProxyModel> _proxies = new();

        [ObservableProperty]
        private ProxyModel _selectedProxy = new();

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        [ObservableProperty]
        private string _searchText = string.Empty;

        [ObservableProperty]
        private string _selectedTypeFilter = "All";

        [ObservableProperty]
        private string _selectedStatusFilter = "All";

        public string[] TypeFilters => new[] { "All", "vmware", "hyperv" };
        public string[] StatusFilters => new[] { "All", "Online", "Offline", "Warning" };

        public ProxiesViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading backup proxies...";
            try
            {
                var proxies = await _apiClient.GetProxiesAsync();
                Proxies.Clear();
                foreach (var p in proxies)
                {
                    Proxies.Add(p);
                }
                StatusMessage = $"Loaded {Proxies.Count} backup proxies";
            }
            catch (System.Exception ex)
            {
                StatusMessage = $"Error loading proxies: {ex.Message}";
            }
            finally
            {
                IsLoading = false;
            }
        }

        [RelayCommand]
        private async Task AddProxyAsync()
        {
            var newProxy = new ProxyModel
            {
                Id = System.Guid.NewGuid().ToString(),
                Name = "New Proxy",
                Host = "",
                Port = 9090,
                Type = "vmware",
                MaxTasks = 4,
                Enabled = true,
                Status = "Offline"
            };

            if (await _apiClient.CreateProxyAsync(newProxy))
            {
                await LoadDataAsync();
                StatusMessage = "Proxy created successfully";
            }
            else
            {
                StatusMessage = "Failed to create proxy";
            }
        }

        [RelayCommand]
        private async Task SaveProxyAsync()
        {
            if (SelectedProxy == null || string.IsNullOrEmpty(SelectedProxy.Id))
            {
                StatusMessage = "No proxy selected";
                return;
            }

            if (await _apiClient.UpdateProxyAsync(SelectedProxy))
            {
                await LoadDataAsync();
                StatusMessage = "Proxy updated successfully";
            }
            else
            {
                StatusMessage = "Failed to update proxy";
            }
        }

        [RelayCommand]
        private async Task DeleteProxyAsync()
        {
            if (SelectedProxy == null || string.IsNullOrEmpty(SelectedProxy.Id))
            {
                StatusMessage = "No proxy selected";
                return;
            }

            var result = MessageBox.Show(
                $"Are you sure you want to delete proxy '{SelectedProxy.Name}'?",
                "Confirm Delete",
                MessageBoxButton.YesNo,
                MessageBoxImage.Warning);

            if (result == MessageBoxResult.Yes)
            {
                if (await _apiClient.DeleteProxyAsync(SelectedProxy.Id))
                {
                    await LoadDataAsync();
                    StatusMessage = "Proxy deleted successfully";
                }
                else
                {
                    StatusMessage = "Failed to delete proxy";
                }
            }
        }

        [RelayCommand]
        private async Task RefreshAsync()
        {
            await LoadDataAsync();
        }

        partial void OnSearchTextChanged(string value)
        {
            ApplyFilters();
        }

        partial void OnSelectedTypeFilterChanged(string value)
        {
            ApplyFilters();
        }

        partial void OnSelectedStatusFilterChanged(string value)
        {
            ApplyFilters();
        }

        private void ApplyFilters()
        {
            var filtered = _apiClient.GetProxiesAsync().Result
                .Where(p =>
                {
                    var matchesSearch = string.IsNullOrWhiteSpace(SearchText) ||
                        p.Name.Contains(SearchText, System.StringComparison.OrdinalIgnoreCase) ||
                        p.Host.Contains(SearchText, System.StringComparison.OrdinalIgnoreCase);

                    var matchesType = SelectedTypeFilter == "All" || p.Type == SelectedTypeFilter;
                    var matchesStatus = SelectedStatusFilter == "All" || p.Status == SelectedStatusFilter;

                    return matchesSearch && matchesType && matchesStatus;
                })
                .ToList();

            Application.Current.Dispatcher.Invoke(() =>
            {
                Proxies.Clear();
                foreach (var p in filtered)
                {
                    Proxies.Add(p);
                }
            });
        }
    }
}
