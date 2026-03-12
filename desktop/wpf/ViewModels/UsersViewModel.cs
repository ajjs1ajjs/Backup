using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class UsersViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<UserModel> _users = new();

        [ObservableProperty]
        private UserModel? _selectedUser;

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        [ObservableProperty]
        private string _newUsername = string.Empty;

        [ObservableProperty]
        private string _newEmail = string.Empty;

        [ObservableProperty]
        private string _newPassword = string.Empty;

        public UsersViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading users...";
            try
            {
                var users = await _apiClient.GetUsersAsync();
                Users.Clear();
                foreach (var u in users) Users.Add(u);
                StatusMessage = $"Loaded {Users.Count} users";
            }
            catch { StatusMessage = "Error loading data"; }
            finally { IsLoading = false; }
        }

        [RelayCommand]
        private async Task CreateUserAsync()
        {
            if (string.IsNullOrWhiteSpace(NewUsername) || string.IsNullOrWhiteSpace(NewEmail))
            {
                StatusMessage = "Username and email are required";
                return;
            }

            var user = new UserModel
            {
                Username = NewUsername,
                Email = NewEmail,
                Active = true
            };

            var success = await _apiClient.CreateUserAsync(user);
            if (success)
            {
                StatusMessage = "User created successfully";
                NewUsername = string.Empty;
                NewEmail = string.Empty;
                NewPassword = string.Empty;
                await LoadDataAsync();
            }
            else
            {
                StatusMessage = "Error creating user";
            }
        }

        [RelayCommand]
        private async Task DeleteUserAsync()
        {
            if (SelectedUser == null) return;

            var success = await _apiClient.DeleteUserAsync(SelectedUser.Id);
            if (success)
            {
                StatusMessage = "User deleted successfully";
                await LoadDataAsync();
            }
            else
            {
                StatusMessage = "Error deleting user";
            }
        }
    }
}
