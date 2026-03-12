using System;

namespace NovaBackup.GUI.Services
{
    public static class GlobalErrorHandler
    {
        public static void Init()
        {
            AppDomain.CurrentDomain.UnhandledException += (s, e) => ToastService.Show("Unhandled error: " + (e.ExceptionObject?.ToString() ?? "Unknown"));
        }
    }
}
