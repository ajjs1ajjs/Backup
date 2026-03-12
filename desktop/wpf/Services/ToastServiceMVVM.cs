namespace NovaBackup.GUI.Services
{
    public static class ToastServiceMVVM
    {
        public static event System.Action<string>? OnToast;
        public static void Show(string message) => OnToast?.Invoke(message);
    }
}
