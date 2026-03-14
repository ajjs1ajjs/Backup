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

        [RelayCommand]
        private void NavigateVSS() => _navigationService.NavigateTo<VSSViewModel>();

        [RelayCommand]
        private void NavigateReplication() => _navigationService.NavigateTo<ReplicationViewModel>();

        [RelayCommand]
        private void NavigateReports() => _navigationService.NavigateTo<ReportsViewModel>();

        [RelayCommand]
        private void NavigateAuditLog() => _navigationService.NavigateTo<AuditLogViewModel>();

        [RelayCommand]
        private void NavigateUsers() => _navigationService.NavigateTo<UsersViewModel>();

        [RelayCommand]
        private void NavigateRoles() => _navigationService.NavigateTo<RolesViewModel>();

        [RelayCommand]
        private void NavigateTape() => _navigationService.NavigateTo<TapeViewModel>();

        [RelayCommand]
        private void NavigateCredentials() => _navigationService.NavigateTo<CredentialsViewModel>();

        [RelayCommand]
        private void NavigateProxies() => _navigationService.NavigateTo<ProxiesViewModel>();

        [RelayCommand]
        private void NavigateSynthetic() => _navigationService.NavigateTo<SyntheticBackupViewModel>();
    }
}
