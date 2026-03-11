using System;
using System.IO;
using System.Threading.Tasks;
using System.Windows.Forms;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using NovaBackup.Desktop.Services;

namespace NovaBackup.Desktop
{
    internal class Program
    {
        private static IHost _host;
        private static ILogger<Program> _logger;

        [STAThread]
        static async Task Main(string[] args)
        {
            try
            {
                // Handle command line arguments for service operations
                if (args.Length > 0)
                {
                    await HandleServiceArguments(args);
                    return;
                }

                // Initialize application
                Application.EnableVisualStyles();
                Application.SetCompatibleTextRenderingDefault(false);
                Application.SetHighDpiMode(HighDpiMode.SystemAware);

                // Setup dependency injection
                await SetupHost();

                // Check if another instance is running
                if (IsApplicationRunning())
                {
                    MessageBox.Show("NOVA Backup is already running. Check the system tray.", 
                        "NOVA Backup", MessageBoxButtons.OK, MessageBoxIcon.Information);
                    return;
                }

                // Create and run main form
                using (var scope = _host.Services.CreateScope())
                {
                    var services = scope.ServiceProvider;
                    var mainForm = services.GetRequiredService<MainForm>();
                    
                    Application.Run(mainForm);
                }
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Fatal error: {ex.Message}", 
                    "NOVA Backup - Fatal Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
            finally
            {
                await CleanupHost();
            }
        }

        private static async Task SetupHost()
        {
            _host = Host.CreateDefaultBuilder()
                .ConfigureServices((context, services) =>
                {
                    // Add services
                    services.AddSingleton<NovaBackupService>();
                    services.AddSingleton<WebApiService>();
                    services.AddSingleton<SystemTrayManager>();
                    services.AddSingleton<NotificationManager>();
                    
                    // Add forms
                    services.AddTransient<MainForm>();
                    services.AddTransient<NewBackupForm>();
                    services.AddTransient<RestoreBackupForm>();
                    services.AddTransient<ServiceManagerForm>();
                    services.AddTransient<AboutForm>();
                    
                    // Add controls
                    services.AddTransient<DashboardControl>();
                    services.AddTransient<BackupsControl>();
                    services.AddTransient<SchedulesControl>();
                    services.AddTransient<StorageControl>();
                    services.AddTransient<ReportsControl>();
                    services.AddTransient<SettingsControl>();
                })
                .ConfigureLogging(logging =>
                {
                    logging.ClearProviders();
                    logging.AddConsole();
                    logging.AddDebug();
                    logging.SetMinimumLevel(LogLevel.Information);
                })
                .Build();

            _logger = _host.Services.GetRequiredService<ILogger<Program>>();
            
            // Start background services
            await _host.StartAsync();
            
            _logger.LogInformation("NOVA Backup application started");
        }

        private static async Task CleanupHost()
        {
            if (_host != null)
            {
                _logger?.LogInformation("NOVA Backup application shutting down");
                await _host.StopAsync();
                _host.Dispose();
            }
        }

        private static async Task HandleServiceArguments(string[] args)
        {
            var serviceManager = new ServiceInstaller();
            
            switch (args[0].ToLower())
            {
                case "install":
                    serviceManager.Install();
                    Console.WriteLine("Service installed successfully");
                    break;
                    
                case "uninstall":
                    serviceManager.Uninstall();
                    Console.WriteLine("Service uninstalled successfully");
                    break;
                    
                case "start":
                    serviceManager.Start();
                    Console.WriteLine("Service started successfully");
                    break;
                    
                case "stop":
                    serviceManager.Stop();
                    Console.WriteLine("Service stopped successfully");
                    break;
                    
                case "service":
                    // Run as Windows Service
                    ServiceBase.Run(new NovaBackupWindowsService());
                    break;
                    
                case "console":
                    // Run in console mode
                    await RunConsoleMode();
                    break;
                    
                default:
                    Console.WriteLine("Usage: NovaBackup.exe [install|uninstall|start|stop|service|console]");
                    break;
            }
        }

        private static async Task RunConsoleMode()
        {
            Console.WriteLine("NOVA Backup - Console Mode");
            Console.WriteLine("Press Ctrl+C to exit");
            
            var backupService = new NovaBackupService();
            var webApiService = new WebApiService("http://localhost:8080");
            
            try
            {
                // Start web server
                await webApiService.StartWebServer();
                Console.WriteLine("Web API started on http://localhost:8080");
                
                // Start backup monitoring
                backupService.StartMonitoring();
                Console.WriteLine("Backup monitoring started");
                
                // Keep running
                var cts = new System.Threading.CancellationTokenSource();
                Console.CancelKeyPress += (sender, e) => {
                    e.Cancel = true;
                    cts.Cancel();
                };
                
                try
                {
                    await Task.Delay(Timeout.Infinite, cts.Token);
                }
                catch (OperationCanceledException)
                {
                    Console.WriteLine("Shutting down...");
                }
                
                // Cleanup
                backupService.StopMonitoring();
                await webApiService.StopWebServer();
                Console.WriteLine("Shutdown complete");
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error: {ex.Message}");
            }
        }

        private static bool IsApplicationRunning()
        {
            try
            {
                var currentProcess = System.Diagnostics.Process.GetCurrentProcess();
                var processes = System.Diagnostics.Process.GetProcessesByName(currentProcess.ProcessName);
                
                foreach (var process in processes)
                {
                    if (process.Id != currentProcess.Id)
                    {
                        process.Close();
                        return true;
                    }
                    process.Close();
                }
                
                return false;
            }
            catch
            {
                return false;
            }
        }

        private static void EnsureDirectories()
        {
            var appDataPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData), "NovaBackup");
            var directories = new[]
            {
                appDataPath,
                Path.Combine(appDataPath, "logs"),
                Path.Combine(appDataPath, "config"),
                Path.Combine(appDataPath, "backups"),
                Path.Combine(appDataPath, "temp")
            };

            foreach (var dir in directories)
            {
                if (!Directory.Exists(dir))
                {
                    Directory.CreateDirectory(dir);
                }
            }
        }

        private static void SetupUnhandledExceptions()
        {
            // Handle unhandled exceptions
            Application.ThreadException += (sender, e) => {
                LogException(e.Exception);
                MessageBox.Show($"An unexpected error occurred: {e.Exception.Message}", 
                    "NOVA Backup - Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
            };

            AppDomain.CurrentDomain.UnhandledException += (sender, e) => {
                LogException(e.ExceptionObject as Exception);
                MessageBox.Show($"A fatal error occurred: {e.ExceptionObject}", 
                    "NOVA Backup - Fatal Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
            };
        }

        private static void LogException(Exception ex)
        {
            try
            {
                var logPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData), 
                    "NovaBackup", "logs");
                
                if (!Directory.Exists(logPath))
                {
                    Directory.CreateDirectory(logPath);
                }

                var logFile = Path.Combine(logPath, $"error-{DateTime.Now:yyyy-MM-dd}.log");
                var logEntry = $"[{DateTime.Now:yyyy-MM-dd HH:mm:ss}] [ERROR] {ex}\n{ex.StackTrace}\n\n";
                
                File.AppendAllText(logFile, logEntry);
            }
            catch
            {
                // If logging fails, we can't do much about it
            }
        }
    }
}
