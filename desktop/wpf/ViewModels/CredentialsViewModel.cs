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
    public partial class CredentialsViewModel : ObservableObject
    {
        private readonly CredentialService _credentialService;

        [ObservableProperty]
        private ObservableCollection<CredentialModel> _credentials = new();

        [ObservableProperty]
        private CredentialModel _selectedCredential = new();

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        [ObservableProperty]
        private string _searchText = string.Empty;

        [ObservableProperty]
        private string _selectedTypeFilter = "All";

        public string[] TypeFilters => new[] { "All", "windows", "linux", "sql", "exchange", "vcenter" };

        public CredentialsViewModel(CredentialService credentialService)
        {
            _credentialService = credentialService;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading credentials...";
            try
            {
                var credentials = await _credentialService.GetCredentialsAsync();
                Credentials.Clear();
                foreach (var c in credentials)
                {
                    Credentials.Add(c);
                }
                StatusMessage = $"Loaded {Credentials.Count} credentials";
            }
            catch (System.Exception ex)
            {
                StatusMessage = $"Error loading credentials: {ex.Message}";
            }
            finally
            {
                IsLoading = false;
            }
        }

        [RelayCommand]
        private async Task AddCredentialAsync()
        {
            var newCredential = new CredentialModel
            {
                Id = System.Guid.NewGuid().ToString(),
                Name = "New Credential",
                Username = "",
                Domain = "",
                Type = "windows",
                Description = ""
            };

            if (await _credentialService.CreateCredentialAsync(newCredential))
            {
                await LoadDataAsync();
                StatusMessage = "Credential created successfully";
            }
            else
            {
                StatusMessage = "Failed to create credential";
            }
        }

        [RelayCommand]
        private async Task DeleteCredentialAsync()
        {
            if (SelectedCredential == null || string.IsNullOrEmpty(SelectedCredential.Id))
            {
                StatusMessage = "No credential selected";
                return;
            }

            var result = MessageBox.Show(
                $"Are you sure you want to delete credential '{SelectedCredential.Name}'?",
                "Confirm Delete",
                MessageBoxButton.YesNo,
                MessageBoxImage.Warning);

            if (result == MessageBoxResult.Yes)
            {
                if (await _credentialService.DeleteCredentialAsync(SelectedCredential.Id))
                {
                    await LoadDataAsync();
                    StatusMessage = "Credential deleted successfully";
                }
                else
                {
                    StatusMessage = "Failed to delete credential";
                }
            }
        }

        [RelayCommand]
        private async Task RefreshAsync()
        {
            _credentialService.ClearCache();
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

        private async Task ApplyFiltersAsync()
        {
            await Task.Run(() => ApplyFilters());
        }

        private void ApplyFilters()
        {
            var filtered = _credentialService.GetCredentialsAsync().Result
                .Where(c =>
                {
                    var matchesSearch = string.IsNullOrWhiteSpace(SearchText) ||
                        c.Name.Contains(SearchText, System.StringComparison.OrdinalIgnoreCase) ||
                        c.Username.Contains(SearchText, System.StringComparison.OrdinalIgnoreCase) ||
                        c.Description.Contains(SearchText, System.StringComparison.OrdinalIgnoreCase);

                    var matchesType = SelectedTypeFilter == "All" || c.Type == SelectedTypeFilter;

                    return matchesSearch && matchesType;
                })
                .ToList();

            Application.Current.Dispatcher.Invoke(() =>
            {
                Credentials.Clear();
                foreach (var c in filtered)
                {
                    Credentials.Add(c);
                }
            });
        }
    }
}
