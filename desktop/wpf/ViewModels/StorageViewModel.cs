using CommunityToolkit.Mvvm.ComponentModel;
using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.Models;

namespace NovaBackup.GUI.ViewModels
{
    public partial class StorageViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private string _pageTitle = "Backup Repositories";

        [ObservableProperty]
        private ObservableCollection<RepositoryModel> _repositories = new();

        public StorageViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
            _ = LoadDataAsync();
        }

        private async Task LoadDataAsync()
        {
            try
            {
                var repos = await _apiClient.GetRepositoriesAsync();
                Repositories.Clear();
                foreach(var repo in repos)
                {
                    Repositories.Add(repo);
                }
            }
            catch (Exception ex)
            {
                // Handle error
            }
        }
    }
}
