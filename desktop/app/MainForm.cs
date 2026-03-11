using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Data;
using System.Drawing;
using System.IO;
using System.Linq;
using System.Net.Http;
using System.ServiceProcess;
using System.Threading.Tasks;
using System.Windows.Forms;
using Newtonsoft.Json;
using NovaBackup.Desktop.Services;

namespace NovaBackup.Desktop
{
    public partial class MainForm : Form
    {
        private readonly NovaBackupService _backupService;
        private readonly WebApiService _webApiService;
        private readonly SystemTrayManager _trayManager;
        private readonly NotificationManager _notificationManager;
        
        private Timer _refreshTimer;
        private bool _isServiceRunning;
        private string _currentBackupStatus;

        public MainForm()
        {
            InitializeComponent();
            InitializeServices();
            InitializeUI();
            StartBackgroundMonitoring();
        }

        private void InitializeServices()
        {
            _backupService = new NovaBackupService();
            _webApiService = new WebApiService("http://localhost:8080");
            _trayManager = new SystemTrayManager();
            _notificationManager = new NotificationManager();
            
            // Setup system tray
            _trayManager.Initialize(this);
            _trayManager.OnTrayIconClick += TrayIcon_Click;
            _trayManager.OnExitRequested += ExitApplication;
        }

        private void InitializeUI()
        {
            this.Text = "NOVA Backup - Desktop Application";
            this.Size = new Size(1200, 800);
            this.StartPosition = FormStartPosition.CenterScreen;
            this.Icon = Properties.Resources.NovaBackupIcon;
            
            // Setup main layout
            SetupMainLayout();
            SetupStatusBar();
            SetupMenu();
        }

        private void SetupMainLayout()
        {
            var splitContainer = new SplitContainer
            {
                Dock = DockStyle.Fill,
                SplitterDistance = 250
            };

            // Left panel - Navigation
            var leftPanel = new Panel
            {
                Dock = DockStyle.Fill,
                BackColor = Color.FromArgb(240, 240, 240)
            };
            
            var navigationTree = new TreeView
            {
                Dock = DockStyle.Fill,
                Font = new Font("Segoe UI", 9F)
            };
            
            navigationTree.Nodes.Add("Dashboard", "📊 Dashboard");
            navigationTree.Nodes.Add("Backups", "💾 Backups");
            navigationTree.Nodes.Add("Schedules", "📅 Schedules");
            navigationTree.Nodes.Add("Storage", "🗄️ Storage");
            navigationTree.Nodes.Add("Reports", "📈 Reports");
            navigationTree.Nodes.Add("Settings", "⚙️ Settings");
            
            navigationTree.AfterSelect += NavigationTree_AfterSelect;
            leftPanel.Controls.Add(navigationTree);
            
            // Right panel - Content
            var rightPanel = new Panel
            {
                Dock = DockStyle.Fill
            };
            
            _contentPanel = new Panel
            {
                Dock = DockStyle.Fill
            };
            
            ShowDashboard();
            rightPanel.Controls.Add(_contentPanel);
            
            splitContainer.Panel1.Controls.Add(leftPanel);
            splitContainer.Panel2.Controls.Add(rightPanel);
            
            this.Controls.Add(splitContainer);
        }

        private void SetupStatusBar()
        {
            var statusBar = new StatusStrip();
            
            var statusLabel = new ToolStripStatusLabel
            {
                Text = "Ready",
                Spring = true
            };
            
            var serviceStatusLabel = new ToolStripStatusLabel
            {
                Text = "Service: Stopped",
                BackColor = Color.LightGray
            };
            
            var webStatusLabel = new ToolStripStatusLabel
            {
                Text = "Web: Disconnected",
                BackColor = Color.LightGray
            };
            
            statusBar.Items.Add(statusLabel);
            statusBar.Items.Add(serviceStatusLabel);
            statusBar.Items.Add(webStatusLabel);
            
            this.Controls.Add(statusBar);
            
            _statusLabel = statusLabel;
            _serviceStatusLabel = serviceStatusLabel;
            _webStatusLabel = webStatusLabel;
        }

        private void SetupMenu()
        {
            var menuStrip = new MenuStrip();
            
            // File menu
            var fileMenu = new ToolStripMenuItem("File");
            fileMenu.DropDownItems.Add("New Backup", null, (s, e) => CreateNewBackup());
            fileMenu.DropDownItems.Add("Restore", null, (s, e) => RestoreBackup());
            fileMenu.DropDownItems.Add(new ToolStripSeparator());
            fileMenu.DropDownItems.Add("Exit", null, (s, e) => ExitApplication());
            
            // Tools menu
            var toolsMenu = new ToolStripMenuItem("Tools");
            toolsMenu.DropDownItems.Add("Service Manager", null, (s, e) => OpenServiceManager());
            toolsMenu.DropDownItems.Add("Web Console", null, (s, e) => OpenWebConsole());
            toolsMenu.DropDownItems.Add("Settings", null, (s, e) => OpenSettings());
            
            // Help menu
            var helpMenu = new ToolStripMenuItem("Help");
            helpMenu.DropDownItems.Add("Documentation", null, (s, e) => OpenDocumentation());
            helpMenu.DropDownItems.Add("About", null, (s, e) => ShowAbout());
            
            menuStrip.Items.Add(fileMenu);
            menuStrip.Items.Add(toolsMenu);
            menuStrip.Items.Add(helpMenu);
            
            this.Controls.Add(menuStrip);
        }

        private void StartBackgroundMonitoring()
        {
            _refreshTimer = new Timer
            {
                Interval = 5000 // 5 seconds
            };
            _refreshTimer.Tick += RefreshTimer_Tick;
            _refreshTimer.Start();
            
            // Check service status
            Task.Run(async () => await CheckServiceStatus());
        }

        private async void RefreshTimer_Tick(object sender, EventArgs e)
        {
            await CheckServiceStatus();
            await RefreshCurrentView();
        }

        private async Task CheckServiceStatus()
        {
            try
            {
                // Check if Nova Backup service is running
                _isServiceRunning = _backupService.IsServiceRunning();
                
                if (_isServiceRunning)
                {
                    _serviceStatusLabel.Text = "Service: Running";
                    _serviceStatusLabel.BackColor = Color.LightGreen;
                    
                    // Get backup status
                    var status = await _backupService.GetBackupStatus();
                    _currentBackupStatus = status.Status;
                    _statusLabel.Text = $"Status: {status.Status}";
                }
                else
                {
                    _serviceStatusLabel.Text = "Service: Stopped";
                    _serviceStatusLabel.BackColor = Color.LightGray;
                    _statusLabel.Text = "Service is not running";
                }
                
                // Check web API connectivity
                var webConnected = await _webApiService.CheckConnection();
                if (webConnected)
                {
                    _webStatusLabel.Text = "Web: Connected";
                    _webStatusLabel.BackColor = Color.LightGreen;
                }
                else
                {
                    _webStatusLabel.Text = "Web: Disconnected";
                    _webStatusLabel.BackColor = Color.LightGray;
                }
            }
            catch (Exception ex)
            {
                _statusLabel.Text = $"Error: {ex.Message}";
            }
        }

        private void NavigationTree_AfterSelect(object sender, TreeViewEventArgs e)
        {
            switch (e.Node.Name)
            {
                case "Dashboard":
                    ShowDashboard();
                    break;
                case "Backups":
                    ShowBackups();
                    break;
                case "Schedules":
                    ShowSchedules();
                    break;
                case "Storage":
                    ShowStorage();
                    break;
                case "Reports":
                    ShowReports();
                    break;
                case "Settings":
                    ShowSettings();
                    break;
            }
        }

        private void ShowDashboard()
        {
            _contentPanel.Controls.Clear();
            
            var dashboard = new DashboardControl();
            dashboard.Dock = DockStyle.Fill;
            _contentPanel.Controls.Add(dashboard);
        }

        private void ShowBackups()
        {
            _contentPanel.Controls.Clear();
            
            var backups = new BackupsControl(_backupService);
            backups.Dock = DockStyle.Fill;
            _contentPanel.Controls.Add(backups);
        }

        private void ShowSchedules()
        {
            _contentPanel.Controls.Clear();
            
            var schedules = new SchedulesControl(_backupService);
            schedules.Dock = DockStyle.Fill;
            _contentPanel.Controls.Add(schedules);
        }

        private void ShowStorage()
        {
            _contentPanel.Controls.Clear();
            
            var storage = new StorageControl(_backupService);
            storage.Dock = DockStyle.Fill;
            _contentPanel.Controls.Add(storage);
        }

        private void ShowReports()
        {
            _contentPanel.Controls.Clear();
            
            var reports = new ReportsControl(_backupService);
            reports.Dock = DockStyle.Fill;
            _contentPanel.Controls.Add(reports);
        }

        private void ShowSettings()
        {
            _contentPanel.Controls.Clear();
            
            var settings = new SettingsControl(_backupService);
            settings.Dock = DockStyle.Fill;
            _contentPanel.Controls.Add(settings);
        }

        private async Task RefreshCurrentView()
        {
            if (_contentPanel.Controls.Count > 0)
            {
                var control = _contentPanel.Controls[0];
                if (control is IRefreshable refreshable)
                {
                    await refreshable.Refresh();
                }
            }
        }

        private void CreateNewBackup()
        {
            var form = new NewBackupForm(_backupService);
            var result = form.ShowDialog(this);
            
            if (result == DialogResult.OK)
            {
                _notificationManager.ShowNotification("Backup started successfully", "Success");
            }
        }

        private void RestoreBackup()
        {
            var form = new RestoreBackupForm(_backupService);
            var result = form.ShowDialog(this);
            
            if (result == DialogResult.OK)
            {
                _notificationManager.ShowNotification("Restore started successfully", "Success");
            }
        }

        private void OpenServiceManager()
        {
            var form = new ServiceManagerForm(_backupService);
            form.ShowDialog(this);
        }

        private void OpenWebConsole()
        {
            try
            {
                System.Diagnostics.Process.Start("http://localhost:8080");
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Unable to open web console: {ex.Message}", "Error", 
                    MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void OpenSettings()
        {
            ShowSettings();
        }

        private void OpenDocumentation()
        {
            try
            {
                System.Diagnostics.Process.Start("https://novabackup.com/docs");
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Unable to open documentation: {ex.Message}", "Error", 
                    MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void ShowAbout()
        {
            var aboutForm = new AboutForm();
            aboutForm.ShowDialog(this);
        }

        private void TrayIcon_Click(object sender, EventArgs e)
        {
            this.Show();
            this.WindowState = FormWindowState.Normal;
            this.BringToFront();
        }

        private void ExitApplication()
        {
            if (MessageBox.Show("Are you sure you want to exit NOVA Backup?", "Exit", 
                MessageBoxButtons.YesNo, MessageBoxIcon.Question) == DialogResult.Yes)
            {
                _trayManager.Dispose();
                Application.Exit();
            }
        }

        protected override void OnResize(EventArgs e)
        {
            if (WindowState == FormWindowState.Minimized)
            {
                Hide();
                _trayManager.ShowBalloonTip("NOVA Backup", "Application minimized to system tray", ToolTipIcon.Info);
            }
            base.OnResize(e);
        }

        protected override void OnFormClosing(FormClosingEventArgs e)
        {
            if (e.CloseReason == CloseReason.UserClosing)
            {
                e.Cancel = true;
                this.WindowState = FormWindowState.Minimized;
                Hide();
            }
            base.OnFormClosing(e);
        }

        // Controls
        private Panel _contentPanel;
        private ToolStripStatusLabel _statusLabel;
        private ToolStripStatusLabel _serviceStatusLabel;
        private ToolStripStatusLabel _webStatusLabel;
    }

    // Interface for refreshable controls
    public interface IRefreshable
    {
        Task Refresh();
    }
}
