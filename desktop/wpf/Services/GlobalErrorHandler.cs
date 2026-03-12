using System;
using System.IO;
using System.Windows;
using System.Windows.Threading;
using System.Threading.Tasks;

namespace NovaBackup.GUI.Services
{
    public static class GlobalErrorHandler
    {
        private static string _logPath = "";

        public static void Init(string logPath = "")
        {
            _logPath = logPath;
            
            AppDomain.CurrentDomain.UnhandledException += OnUnhandledException;
            Application.Current.DispatcherUnhandledException += OnDispatcherUnhandledException;
            TaskScheduler.UnobservedTaskException += OnUnobservedTaskException;
        }

        private static void Log(string message, Exception? ex = null)
        {
            var logMessage = $"[{DateTime.Now:yyyy-MM-dd HH:mm:ss}] {message}";
            if (ex != null) logMessage += $"\n{ex}";
            
            try
            {
                if (!string.IsNullOrEmpty(_logPath))
                {
                    File.AppendAllText(_logPath, logMessage + "\n");
                }
            }
            catch { }
        }

        private static void OnUnhandledException(object sender, UnhandledExceptionEventArgs e)
        {
            var ex = e.ExceptionObject as Exception;
            Log($"Unhandled domain exception: {ex?.Message}", ex);
            
            MessageBox.Show(
                $"A critical error occurred: {ex?.Message}\n\nThe application will now close.",
                "Critical Error",
                MessageBoxButton.OK,
                MessageBoxImage.Error);
            
            Environment.Exit(1);
        }

        private static void OnDispatcherUnhandledException(object sender, DispatcherUnhandledExceptionEventArgs e)
        {
            Log($"Dispatcher exception: {e.Exception.Message}", e.Exception);

            MessageBox.Show(
                $"An error occurred: {e.Exception.Message}",
                "Error",
                MessageBoxButton.OK,
                MessageBoxImage.Error);

            e.Handled = true;
        }

        private static void OnUnobservedTaskException(object? sender, UnobservedTaskExceptionEventArgs e)
        {
            Log($"Unobserved task exception: {e.Exception.Message}", e.Exception);
            e.SetObserved();
        }
    }
}
