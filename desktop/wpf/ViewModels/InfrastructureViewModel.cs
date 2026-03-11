using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;
using NovaBackup.GUI.Services;
using NovaBackup.GUI.Views;
using NovaBackup.GUI.Models;
using Microsoft.Extensions.DependencyInjection;

namespace NovaBackup.GUI.ViewModels
{
    public partial class InfrastructureViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;
        private readonly IServiceProvider _serviceProvider;

        [ObservableProperty]
        private string _pageTitle = "Backup Infrastructure";

        [ObservableProperty]
        private ObservableCollection<InfrastructureNode> _infrastructureTree = new();

        [ObservableProperty]
        private InfrastructureNode? _selectedNode;

        public InfrastructureViewModel(IApiClient apiClient, IServiceProvider serviceProvider)
        {
            _apiClient = apiClient;
            _serviceProvider = serviceProvider;
            _ = LoadDataAsync();
        }

        [RelayCommand]
        private void SelectNode(InfrastructureNode? node)
        {
            SelectedNode = node;
        }

        [RelayCommand]
        private void AddServer()
        {
            var addServerWindow = _serviceProvider.GetRequiredService<AddServerWindow>();
            addServerWindow.Owner = App.Current.MainWindow;
            addServerWindow.ShowDialog();
            _ = LoadDataAsync(); // Refresh tree
        }

        [RelayCommand]
        private async Task RescanNode(InfrastructureNode? node)
        {
            if (node == null || string.IsNullOrEmpty(node.Id)) return;

            try
            {
                var success = await _apiClient.DiscoverNodeAsync(node.Id);
                if (success)
                {
                    await LoadDataAsync();
                }
            }
            catch (Exception) { /* Handle error */ }
        }

        private async Task LoadDataAsync()
        {
            try
            {
                var nodes = await _apiClient.GetInfrastructureTreeAsync();
                InfrastructureTree.Clear();
                foreach (var node in nodes)
                {
                    // For each node, fetch children objects (VMs)
                    if (!string.IsNullOrEmpty(node.Id))
                    {
                        var objects = await _apiClient.GetDiscoveredObjectsAsync(node.Id);
                        foreach (var obj in objects)
                        {
                            node.Children.Add(new InfrastructureNode {
                                Name = obj.Name,
                                IconKind = obj.IconKind,
                                NodeType = "VM"
                            });
                        }
                    }
                    InfrastructureTree.Add(node);
                }
            }
            catch (Exception) { }
        }
    }
}
