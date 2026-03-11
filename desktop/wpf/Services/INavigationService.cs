using CommunityToolkit.Mvvm.ComponentModel;
using System;

namespace NovaBackup.GUI.Services
{
    public interface INavigationService
    {
        ObservableObject? CurrentView { get; }
        void NavigateTo<TViewModel>() where TViewModel : ObservableObject;
        event Action? CurrentViewChanged;
    }

    public class NavigationService : INavigationService
    {
        private readonly Func<Type, ObservableObject> _viewModelFactory;
        private ObservableObject? _currentView;

        public NavigationService(Func<Type, ObservableObject> viewModelFactory)
        {
            _viewModelFactory = viewModelFactory;
        }

        public ObservableObject? CurrentView
        {
            get => _currentView;
            private set
            {
                _currentView = value;
                CurrentViewChanged?.Invoke();
            }
        }

        public event Action? CurrentViewChanged;

        public void NavigateTo<TViewModel>() where TViewModel : ObservableObject
        {
            ObservableObject viewModel = _viewModelFactory.Invoke(typeof(TViewModel));
            CurrentView = viewModel;
        }
    }
}
