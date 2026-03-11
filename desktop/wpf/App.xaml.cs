using System;
using System.Windows;
using System.Windows.Media;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

namespace NovaBackup.GUI
{
    /// <summary>
    /// Interaction logic for App.xaml
    /// </summary>
    public partial class App : Application
    {
        private IHost? _host;

        protected override async void OnStartup(StartupEventArgs e)
        {
            base.OnStartup(e);

            _host = Host.CreateDefaultBuilder()
                .ConfigureServices((context, services) =>
                {
                    // Register services
                    services.AddSingleton<Services.INavigationService, Services.NavigationService>();
                    services.AddHttpClient<Services.IApiClient, Services.ApiClient>();
                    
                    // Register ViewModels
                    services.AddSingleton<ViewModels.MainViewModel>();
                    services.AddSingleton<ViewModels.HomeViewModel>();
                    services.AddSingleton<ViewModels.InfrastructureViewModel>();
                    services.AddSingleton<ViewModels.JobsViewModel>();
                    services.AddSingleton<ViewModels.StorageViewModel>();
                    services.AddSingleton<ViewModels.RecoverySessionsViewModel>();
                    services.AddTransient<ViewModels.AddServerViewModel>();
                    services.AddTransient<Views.AddServerWindow>();
                    services.AddTransient<ViewModels.JobWizardViewModel>();
                    services.AddTransient<Views.JobWizardWindow>();
                    
                    // Provides instances of ViewModels for the NavigationService
                    services.AddSingleton<Func<Type, CommunityToolkit.Mvvm.ComponentModel.ObservableObject>>(serviceProvider => viewModelType => {
                        return (CommunityToolkit.Mvvm.ComponentModel.ObservableObject)serviceProvider.GetRequiredService(viewModelType);
                    });

                    // Register Views
                    services.AddSingleton<MainWindow>(provider => new MainWindow {
                        DataContext = provider.GetRequiredService<ViewModels.MainViewModel>()
                    });
                })
                .Build();

            await _host.StartAsync();

            // Show main window
            var mainWindow = _host.Services.GetRequiredService<MainWindow>();
            mainWindow.Show();
        }

        protected override async void OnExit(ExitEventArgs e)
        {
            if (_host != null)
            {
                await _host.StopAsync();
                _host.Dispose();
            }
            base.OnExit(e);
        }
    }
}
