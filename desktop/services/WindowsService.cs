using System;
using System.IO;
using System.ServiceProcess;
using System.Threading.Tasks;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using NovaBackup.Desktop.Services;

namespace NovaBackup.Desktop.Services
{
    public class NovaBackupWindowsService : ServiceBase
    {
        private IHost _host;
        private readonly ILogger<NovaBackupWindowsService> _logger;

        public NovaBackupWindowsService()
        {
            ServiceName = "NovaBackupService";
            _logger = LoggerFactory.Create(builder => builder.AddConsole()).CreateLogger<NovaBackupWindowsService>();
        }

        protected override void OnStart(string[] args)
        {
            try
            {
                _logger.LogInformation("Starting NOVA Backup Service...");

                _host = Host.CreateDefaultBuilder()
                    .ConfigureServices((context, services) =>
                    {
                        services.AddSingleton<NovaBackupService>();
                        services.AddSingleton<WebApiService>();
                        services.AddHostedService<BackgroundBackupService>();
                    })
                    .ConfigureLogging(logging =>
                    {
                        logging.AddConsole();
                        logging.AddEventLog();
                    })
                    .Build();

                Task.Run(async () => await _host.StartAsync());

                _logger.LogInformation("NOVA Backup Service started successfully");
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to start NOVA Backup Service");
                throw;
            }
        }

        protected override void OnStop()
        {
            try
            {
                _logger.LogInformation("Stopping NOVA Backup Service...");

                if (_host != null)
                {
                    _host.StopAsync().Wait();
                    _host.Dispose();
                }

                _logger.LogInformation("NOVA Backup Service stopped successfully");
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to stop NOVA Backup Service");
                throw;
            }
        }

        protected override void OnPause()
        {
            _logger.LogInformation("NOVA Backup Service paused");
            base.OnPause();
        }

        protected override void OnContinue()
        {
            _logger.LogInformation("NOVA Backup Service resumed");
            base.OnContinue();
        }

        protected override void OnShutdown()
        {
            _logger.LogInformation("NOVA Backup Service shutting down");
            base.OnShutdown();
        }
    }

    public class BackgroundBackupService : BackgroundService
    {
        private readonly NovaBackupService _backupService;
        private readonly ILogger<BackgroundBackupService> _logger;

        public BackgroundBackupService(NovaBackupService backupService, ILogger<BackgroundBackupService> logger)
        {
            _backupService = backupService;
            _logger = logger;
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            _logger.LogInformation("Background backup service started");

            try
            {
                _backupService.StartMonitoring();

                while (!stoppingToken.IsCancellationRequested)
                {
                    // Perform background tasks
                    await Task.Delay(TimeSpan.FromMinutes(1), stoppingToken);
                }
            }
            catch (OperationCanceledException)
            {
                // Normal shutdown
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error in background backup service");
            }
            finally
            {
                _backupService.StopMonitoring();
                _logger.LogInformation("Background backup service stopped");
            }
        }
    }

    public class ServiceInstaller
    {
        public static void Install()
        {
            try
            {
                var servicePath = System.Reflection.Assembly.GetExecutingAssembly().Location;
                var serviceName = "NovaBackupService";
                var displayName = "NOVA Backup Service";
                var description = "NOVA Backup - Enterprise backup solution";

                using (var serviceManager = new System.Management.ManagementClass("Win32_Service"))
                {
                    var inParams = serviceManager.GetMethodParameters("Create");
                    inParams["Name"] = serviceName;
                    inParams["DisplayName"] = displayName;
                    inParams["Description"] = description;
                    inParams["PathName"] = servicePath;
                    inParams["ServiceType"] = "Own Process";
                    inParams["StartMode"] = "Automatic";
                    inParams["ErrorControl"] = "Normal";
                    inParams["DesktopInteract"] = false;

                    var outParams = serviceManager.InvokeMethod("Create", inParams, null);
                    var result = Convert.ToInt32(outParams["ReturnValue"]);

                    if (result == 0)
                    {
                        Console.WriteLine("Service installed successfully");
                    }
                    else
                    {
                        Console.WriteLine($"Service installation failed with error code: {result}");
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error installing service: {ex.Message}");
            }
        }

        public static void Uninstall()
        {
            try
            {
                var serviceName = "NovaBackupService";

                using (var serviceManager = new System.Management.ManagementClass("Win32_Service"))
                {
                    var services = serviceManager.GetInstances();
                    
                    foreach (var service in services)
                    {
                        if (service["Name"].ToString() == serviceName)
                        {
                            var result = service.Delete();
                            
                            if (Convert.ToInt32(result) == 0)
                            {
                                Console.WriteLine("Service uninstalled successfully");
                            }
                            else
                            {
                                Console.WriteLine($"Service uninstallation failed with error code: {result}");
                            }
                            break;
                        }
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error uninstalling service: {ex.Message}");
            }
        }

        public static bool IsInstalled()
        {
            try
            {
                var serviceName = "NovaBackupService";
                
                using (var serviceController = new ServiceController(serviceName))
                {
                    return true;
                }
            }
            catch
            {
                return false;
            }
        }

        public static void Start()
        {
            try
            {
                var serviceName = "NovaBackupService";
                
                using (var serviceController = new ServiceController(serviceName))
                {
                    if (serviceController.Status == ServiceControllerStatus.Stopped)
                    {
                        serviceController.Start();
                        serviceController.WaitForStatus(ServiceControllerStatus.Running, TimeSpan.FromSeconds(30));
                        Console.WriteLine("Service started successfully");
                    }
                    else
                    {
                        Console.WriteLine($"Service is already running (Status: {serviceController.Status})");
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error starting service: {ex.Message}");
            }
        }

        public static void Stop()
        {
            try
            {
                var serviceName = "NovaBackupService";
                
                using (var serviceController = new ServiceController(serviceName))
                {
                    if (serviceController.Status == ServiceControllerStatus.Running)
                    {
                        serviceController.Stop();
                        serviceController.WaitForStatus(ServiceControllerStatus.Stopped, TimeSpan.FromSeconds(30));
                        Console.WriteLine("Service stopped successfully");
                    }
                    else
                    {
                        Console.WriteLine($"Service is already stopped (Status: {serviceController.Status})");
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error stopping service: {ex.Message}");
            }
        }
    }

    // Program entry point for service
    public class Program
    {
        public static void Main(string[] args)
        {
            if (args.Length > 0)
            {
                switch (args[0].ToLower())
                {
                    case "install":
                        ServiceInstaller.Install();
                        return;
                    case "uninstall":
                        ServiceInstaller.Uninstall();
                        return;
                    case "start":
                        ServiceInstaller.Start();
                        return;
                    case "stop":
                        ServiceInstaller.Stop();
                        return;
                }
            }

            // Run as service
            ServiceBase.Run(new NovaBackupWindowsService());
        }
    }
}
