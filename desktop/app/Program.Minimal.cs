using System;
using System.IO;
using System.Threading.Tasks;
using System.Windows.Forms;
using Microsoft.AspNetCore.Builder;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

namespace NovaBackup.Desktop
{
    internal class Program
    {
        [STAThread]
        static async Task Main(string[] args)
        {
            try
            {
                // Create host builder
                var builder = Host.CreateDefaultBuilder(args);

                // Configure services
                builder.ConfigureServices((context, services) =>
                {
                    services.AddLogging();
                    services.AddSingleton<BackupService>();
                    services.AddSingleton<WebApiService>();
                    services.AddSingleton<SystemTrayManager>();
                    services.AddSingleton<NotificationManager>();
                });

                // Build the host
                var host = builder.Build();

                // Create and run the main form
                Application.SetHighDpiMode(HighDpiMode.SystemAware);
                Application.EnableVisualStyles();
                Application.SetCompatibleTextRenderingDefault(false);

                var mainForm = new MainForm(host.Services);
                Application.Run(mainForm);
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Fatal error: {ex.Message}", 
                    "NOVA Backup - Fatal Error", 
                    MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }
    }

    // Minimal service implementations
    public class BackupService
    {
        public Task StartAllBackups() => Task.CompletedTask;
        public Task StopAllBackups() => Task.CompletedTask;
        public void StartMonitoring() { }
        public void StopMonitoring() { }
        public event Action<object> OnStatusChanged;
    }

    public class WebApiService
    {
        public async Task<bool> StartWebServer()
        {
            // Start web server
            var builder = WebApplication.CreateBuilder();
            
            builder.Services.AddControllers();
            builder.Services.AddCors();
            
            var app = builder.Build();
            
            // Configure web API
            app.UseRouting();
            app.UseCors(cors => cors.AllowAnyOrigin().AllowAnyMethod().AllowAnyHeader());
            
            // API endpoints
            app.MapGet("/api/status", () => new { status = "Running" });
            app.MapGet("/api/backups", () => new List<object>());
            app.MapGet("/api/schedules", () => new List<object>());
            app.MapGet("/api/storage", () => new { totalSpace = 1000000000000L });
            
            // Static files
            app.UseStaticFiles();
            app.MapFallbackToFile("index.html");
            
            await app.StartAsync();
            return true;
        }
        
        public void Dispose() { }
    }

    public class SystemTrayManager
    {
        public void Initialize(Form mainForm) { }
        public void ShowBalloonTip(string title, string text) { }
        public event EventHandler OnTrayIconClick;
    }

    public class NotificationManager
    {
        public void ShowNotification(string message, string type = "info") { }
    }
}
