using System;

namespace NovaBackup.GUI.Services
{
    public delegate void ToastEvent(string message);
    public static class ToastService
    {
        public static event ToastEvent? OnToast;
        public static void Show(string message) => OnToast?.Invoke(message);
    }
}
