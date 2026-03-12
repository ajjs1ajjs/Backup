namespace NovaBackup.GUI.Services
{
    public static class Logger
    {
        public static void Log(string message)
        {
            // Simple console logger; replace with proper logging in the future
            System.Console.WriteLine($"[NovaBackup] {message}");
        }
    }
}
