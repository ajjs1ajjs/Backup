using System;
using System.Drawing;
using System.Windows.Forms;
using System.Runtime.InteropServices;

namespace NovaBackup.Desktop.Services
{
    public class SystemTrayManager : IDisposable
    {
        private NotifyIcon _notifyIcon;
        private ContextMenuStrip _contextMenu;
        private MainForm _mainForm;
        private bool _disposed = false;

        public event EventHandler OnTrayIconClick;
        public event EventHandler OnExitRequested;

        public SystemTrayManager()
        {
            InitializeTrayIcon();
            InitializeContextMenu();
        }

        public void Initialize(MainForm mainForm)
        {
            _mainForm = mainForm;
        }

        private void InitializeTrayIcon()
        {
            _notifyIcon = new NotifyIcon
            {
                Icon = Properties.Resources.NovaBackupIcon,
                Text = "NOVA Backup",
                Visible = true
            };

            _notifyIcon.Click += NotifyIcon_Click;
            _notifyIcon.DoubleClick += NotifyIcon_DoubleClick;
        }

        private void InitializeContextMenu()
        {
            _contextMenu = new ContextMenuStrip();

            // Show/Hide
            var showItem = new ToolStripMenuItem("Show", null, (s, e) => ShowApplication());
            var hideItem = new ToolStripMenuItem("Hide", null, (s, e) => HideApplication());

            // Separator
            _contextMenu.Items.Add(new ToolStripSeparator());

            // Quick Actions
            var quickBackupItem = new ToolStripMenuItem("Quick Backup", null, (s, e) => QuickBackup());
            var restoreItem = new ToolStripMenuItem("Restore", null, (s, e) => RestoreBackup());
            var settingsItem = new ToolStripMenuItem("Settings", null, (s, e) => OpenSettings());

            // Separator
            _contextMenu.Items.Add(new ToolStripSeparator());

            // Web Console
            var webConsoleItem = new ToolStripMenuItem("Web Console", null, (s, e) => OpenWebConsole());

            // Separator
            _contextMenu.Items.Add(new ToolStripSeparator());

            // Exit
            var exitItem = new ToolStripMenuItem("Exit", null, (s, e) => OnExitRequested?.Invoke(this, EventArgs.Empty));

            _contextMenu.Items.AddRange(new ToolStripItem[]
            {
                showItem,
                hideItem,
                new ToolStripSeparator(),
                quickBackupItem,
                restoreItem,
                settingsItem,
                new ToolStripSeparator(),
                webConsoleItem,
                new ToolStripSeparator(),
                exitItem
            });

            _notifyIcon.ContextMenuStrip = _contextMenu;
        }

        private void NotifyIcon_Click(object sender, EventArgs e)
        {
            var mouseArgs = e as MouseEventArgs;
            if (mouseArgs?.Button == MouseButtons.Left)
            {
                OnTrayIconClick?.Invoke(this, EventArgs.Empty);
            }
        }

        private void NotifyIcon_DoubleClick(object sender, EventArgs e)
        {
            ShowApplication();
        }

        private void ShowApplication()
        {
            if (_mainForm != null)
            {
                _mainForm.Show();
                _mainForm.WindowState = FormWindowState.Normal;
                _mainForm.BringToFront();
            }
        }

        private void HideApplication()
        {
            if (_mainForm != null)
            {
                _mainForm.Hide();
            }
        }

        private void QuickBackup()
        {
            if (_mainForm != null)
            {
                // Trigger quick backup through main form
                _mainForm.Invoke(new Action(() => {
                    // This would call the quick backup method
                    ShowBalloonTip("Quick Backup", "Starting quick backup...", ToolTipIcon.Info);
                }));
            }
        }

        private void RestoreBackup()
        {
            if (_mainForm != null)
            {
                _mainForm.Invoke(new Action(() => {
                    // This would open restore dialog
                    ShowBalloonTip("Restore", "Opening restore dialog...", ToolTipIcon.Info);
                }));
            }
        }

        private void OpenSettings()
        {
            if (_mainForm != null)
            {
                _mainForm.Invoke(new Action(() => {
                    // This would open settings
                    ShowBalloonTip("Settings", "Opening settings...", ToolTipIcon.Info);
                }));
            }
        }

        private void OpenWebConsole()
        {
            try
            {
                System.Diagnostics.Process.Start("http://localhost:8080");
            }
            catch (Exception ex)
            {
                ShowBalloonTip("Error", $"Unable to open web console: {ex.Message}", ToolTipIcon.Error);
            }
        }

        public void ShowBalloonTip(string title, string text, ToolTipIcon icon = ToolTipIcon.Info)
        {
            _notifyIcon.ShowBalloonTip(title, text, icon, 5000);
        }

        public void UpdateIcon(Icon icon)
        {
            _notifyIcon.Icon = icon;
        }

        public void UpdateText(string text)
        {
            _notifyIcon.Text = text;
        }

        public void Dispose()
        {
            if (!_disposed)
            {
                _notifyIcon?.Dispose();
                _contextMenu?.Dispose();
                _disposed = true;
            }
        }
    }

    public class NotificationManager
    {
        private readonly string _appName = "NOVA Backup";

        public void ShowNotification(string title, string message, NotificationType type = NotificationType.Info)
        {
            try
            {
                // Try to show Windows 10/11 toast notification
                if (Environment.OSVersion.Version >= new Version(10, 0))
                {
                    ShowWindowsNotification(title, message, type);
                }
                else
                {
                    // Fallback to older notification method
                    ShowLegacyNotification(title, message, type);
                }
            }
            catch (Exception ex)
            {
                // Ultimate fallback
                MessageBox.Show($"{title}: {message}", _appName, MessageBoxButtons.OK, MessageBoxIcon.Information);
            }
        }

        private void ShowWindowsNotification(string title, string message, NotificationType type)
        {
            try
            {
                // Use Windows Runtime API for toast notifications
                // This would require additional NuGet packages and COM interop
                // For now, use system tray notification
                ShowSystemTrayNotification(title, message, type);
            }
            catch
            {
                ShowLegacyNotification(title, message, type);
            }
        }

        private void ShowLegacyNotification(string title, string message, NotificationType type)
        {
            var icon = type switch
            {
                NotificationType.Success => MessageBoxIcon.Information,
                NotificationType.Error => MessageBoxIcon.Error,
                NotificationType.Warning => MessageBoxIcon.Warning,
                _ => MessageBoxIcon.Information
            };

            MessageBox.Show(message, title, MessageBoxButtons.OK, icon);
        }

        private void ShowSystemTrayNotification(string title, string message, NotificationType type)
        {
            var icon = type switch
            {
                NotificationType.Success => ToolTipIcon.Info,
                NotificationType.Error => ToolTipIcon.Error,
                NotificationType.Warning => ToolTipIcon.Warning,
                _ => ToolTipIcon.Info
            };

            // This would be called through the main form's tray manager
            // For now, just log it
            Console.WriteLine($"[{title}] {message}");
        }
    }

    public enum NotificationType
    {
        Info,
        Success,
        Warning,
        Error
    }

    // Windows 10/11 Toast Notification Helper (would require additional setup)
    public class ToastNotificationHelper
    {
        [DllImport("user32.dll")]
        private static extern int MessageBox(IntPtr hWnd, string lpText, string lpCaption, uint uType);

        public static void ShowToast(string title, string message)
        {
            // This would implement Windows Runtime toast notifications
            // Requires Windows Runtime reference and proper manifest setup
            MessageBox(IntPtr.Zero, $"{title}: {message}", "NOVA Backup", 0);
        }
    }
}
