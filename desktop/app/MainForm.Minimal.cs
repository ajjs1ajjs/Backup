using System;
using System.Drawing;
using System.Threading.Tasks;
using System.Windows.Forms;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;

namespace NovaBackup.Desktop
{
    public partial class MainForm : Form
    {
        private readonly IServiceProvider _services;
        private readonly ILogger<MainForm> _logger;
        private readonly BackupService _backupService;
        private readonly WebApiService _webApiService;
        private readonly SystemTrayManager _trayManager;
        private readonly NotificationManager _notificationManager;

        private Panel _mainPanel;

        public MainForm(IServiceProvider services)
        {
            _services = services;
            _logger = services.GetRequiredService<ILogger<MainForm>>();
            _backupService = services.GetRequiredService<BackupService>();
            _webApiService = services.GetRequiredService<WebApiService>();
            _trayManager = services.GetRequiredService<SystemTrayManager>();
            _notificationManager = services.GetRequiredService<NotificationManager>();
            
            InitializeComponent();
            SetupUI();
        }

        private void InitializeComponent()
        {
            this.Text = "NOVA Backup - All-in-One";
            this.Size = new Size(1000, 700);
            this.StartPosition = FormStartPosition.CenterScreen;
            this.Icon = SystemIcons.Application;
            this.WindowState = FormWindowState.Maximized;

            // Create menu strip
            var menuStrip = new MenuStrip();
            var fileMenu = new ToolStripMenuItem("File");
            var toolsMenu = new ToolStripMenuItem("Tools");
            var helpMenu = new ToolStripMenuItem("Help");

            // File menu
            fileMenu.DropDownItems.Add("Exit", null, (s, e) => Application.Exit());

            // Tools menu
            toolsMenu.DropDownItems.Add("Open Web Console", null, async (s, e) => await OpenWebConsole());
            toolsMenu.DropDownItems.Add("About", null, (s, e) => ShowAbout());

            menuStrip.Items.AddRange(new ToolStripItem[] { fileMenu, toolsMenu, helpMenu });
            this.MainMenuStrip = menuStrip;

            // Create main panel
            _mainPanel = new Panel
            {
                Dock = DockStyle.Fill,
                BackColor = Color.White,
                Padding = new Padding(20)
            };

            // Add controls
            this.Controls.AddRange(new Control[] { menuStrip, _mainPanel });
        }

        private void SetupUI()
        {
            ShowWelcome();
        }

        private void ShowWelcome()
        {
            _mainPanel.Controls.Clear();
            
            var titleLabel = new Label
            {
                Text = "NOVA BACKUP - ALL-IN-ONE",
                Font = new Font("Segoe UI", 24, FontStyle.Bold),
                ForeColor = Color.FromArgb(0, 120, 215),
                Location = new Point(0, 50),
                AutoSize = true,
                TextAlign = ContentAlignment.MiddleCenter
            };

            var subtitleLabel = new Label
            {
                Text = "Complete backup solution in a single executable",
                Font = new Font("Segoe UI", 14),
                ForeColor = Color.FromArgb(100, 100, 100),
                Location = new Point(0, 100),
                AutoSize = true,
                TextAlign = ContentAlignment.MiddleCenter
            };

            var featuresPanel = new Panel
            {
                Location = new Point(0, 150),
                Size = new Size(1000, 400),
                BorderStyle = BorderStyle.None
            };

            var features = new[]
            {
                "✅ Desktop Application with Full GUI",
                "✅ Web Console (localhost:8080)",
                "✅ Remote Access (http://[IP]:8080)",
                "✅ Windows Service Support",
                "✅ All Dependencies Included",
                "✅ No Installation Required",
                "✅ Works on Windows 10/11"
            };

            var yPos = 20;
            foreach (var feature in features)
            {
                var featureLabel = new Label
                {
                    Text = feature,
                    Font = new Font("Segoe UI", 12),
                    Location = new Point(50, yPos),
                    AutoSize = true
                };
                featuresPanel.Controls.Add(featureLabel);
                yPos += 30;
            }

            var startPanel = new Panel
            {
                Location = new Point(0, 570),
                Size = new Size(1000, 80),
                BackColor = Color.FromArgb(240, 240, 240)
            };

            var startButton = new Button
            {
                Text = "Start Web Console",
                Size = new Size(200, 40),
                Location = new Point(400, 20),
                BackColor = Color.FromArgb(0, 120, 215),
                ForeColor = Color.White,
                FlatStyle = FlatStyle.Flat,
                Font = new Font("Segoe UI", 12, FontStyle.Bold)
            };

            startButton.Click += async (s, e) => await StartWebServer();

            var statusLabel = new Label
            {
                Text = "Ready to start backup operations...",
                Font = new Font("Segoe UI", 10),
                ForeColor = Color.FromArgb(100, 100, 100),
                Location = new Point(50, 20),
                AutoSize = true
            };

            startPanel.Controls.AddRange(new Control[] { startButton, statusLabel });

            _mainPanel.Controls.AddRange(new Control[] { titleLabel, subtitleLabel, featuresPanel, startPanel });
        }

        private async Task StartWebServer()
        {
            try
            {
                statusLabel.Text = "Starting web server...";
                var success = await _webApiService.StartWebServer();
                
                if (success)
                {
                    statusLabel.Text = "✅ Web server running on http://localhost:8080";
                    _notificationManager.ShowNotification("Web console started successfully!", "success");
                    
                    // Open web console
                    System.Diagnostics.Process.Start("http://localhost:8080");
                }
                else
                {
                    statusLabel.Text = "❌ Failed to start web server";
                    _notificationManager.ShowNotification("Failed to start web server", "error");
                }
            }
            catch (Exception ex)
            {
                statusLabel.Text = "❌ Error starting web server";
                MessageBox.Show($"Error: {ex.Message}", "Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
        }

        private void ShowAbout()
        {
            var aboutText = $"NOVA BACKUP - ALL-IN-ONE\n\n" +
                           $"Version: 1.0.0.0\n" +
                           $"Build: {System.DateTime.Now:yyyy-MM-dd}\n\n" +
                           $"Complete backup solution:\n" +
                           $"• Single executable file\n" +
                           $"• Desktop GUI interface\n" +
                           $"• Web console with remote access\n" +
                           $"• All dependencies included\n" +
                           $"• Windows 10/11 compatible\n\n" +
                           $"No installation required!\n\n" +
                           $"© 2024 NOVA Backup";

            MessageBox.Show(aboutText, "About NOVA Backup", 
                MessageBoxButtons.OK, MessageBoxIcon.Information);
        }

        protected override void OnResize(EventArgs e)
        {
            if (this.WindowState == FormWindowState.Minimized)
            {
                this.Hide();
                _trayManager.ShowBalloonTip("NOVA Backup", "Minimized to system tray");
            }
            base.OnResize(e);
        }

        protected override void OnFormClosing(FormClosingEventArgs e)
        {
            if (e.CloseReason == CloseReason.UserClosing)
            {
                var result = MessageBox.Show("Are you sure you want to exit NOVA Backup?", 
                    "Confirm Exit", MessageBoxButtons.YesNo, MessageBoxIcon.Question);
                
                if (result != DialogResult.Yes)
                {
                    e.Cancel = true;
                    return;
                }
            }

            base.OnFormClosing(e);
        }
    }
}
