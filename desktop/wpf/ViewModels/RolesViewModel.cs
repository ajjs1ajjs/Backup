using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class RolesViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<RoleModel> _roles = new();

        [ObservableProperty]
        private RoleModel? _selectedRole;

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        public RolesViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading roles...";
            try
            {
                var roles = await _apiClient.GetRolesAsync();
                Roles.Clear();
                foreach (var r in roles) Roles.Add(r);
                StatusMessage = $"Loaded {Roles.Count} roles";
            }
            catch { StatusMessage = "Error loading data"; }
            finally { IsLoading = false; }
        }

        [RelayCommand]
        private async Task DeleteRoleAsync()
        {
            if (SelectedRole == null) return;

            var success = await _apiClient.DeleteRoleAsync(SelectedRole.Id);
            if (success)
            {
                StatusMessage = "Role deleted successfully";
                await LoadDataAsync();
            }
            else
            {
                StatusMessage = "Error deleting role";
            }
        }
    }
}
