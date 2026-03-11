using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class MainViewModel : ObservableObject
    {
        private readonly INavigationService _navigationService;

        [ObservableProperty]
        private ObservableObject? _currentViewModel;

        public MainViewModel(INavigationService navigationService)
        {
            _navigationService = navigationService;
            _navigationService.CurrentViewChanged += OnCurrentViewChanged;

            // Start on Dashboard
            _navigationService.NavigateTo<HomeViewModel>();
        }

        private void OnCurrentViewChanged()
        {
            CurrentViewModel = _navigationService.CurrentView;
        }

        [RelayCommand]
        private void NavigateHome() => _navigationService.NavigateTo<HomeViewModel>();

        [RelayCommand]
        private void NavigateJobs() => _navigationService.NavigateTo<JobsViewModel>();

        [RelayCommand]
        private void NavigateInfrastructure() => _navigationService.NavigateTo<InfrastructureViewModel>();

        [RelayCommand]
        private void NavigateStorage() => _navigationService.NavigateTo<StorageViewModel>();

        [RelayCommand]
        private void NavigateRecovery() => _navigationService.NavigateTo<RecoverySessionsViewModel>();
    }
}
