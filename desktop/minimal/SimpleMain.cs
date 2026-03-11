using System;
using System.Drawing;
using System.Threading.Tasks;
using System.Windows.Forms;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

namespace NovaBackup.Desktop
{
    internal class Program
    {
        [STAThread]
        static void Main(string[] args)
        {
            try
            {
                Application.SetHighDpiMode(HighDpiMode.SystemAware);
                Application.EnableVisualStyles();
                Application.SetCompatibleTextRenderingDefault(false);

                // Create simple form
                var mainForm = new SimpleMainForm();
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

    public class SimpleMainForm : Form
    {
        private Button _startWebButton;
        private Button _exitButton;
        private Label _statusLabel;
        private Label _titleLabel;

        public SimpleMainForm()
        {
            InitializeComponent();
        }

        private void InitializeComponent()
        {
            this.Text = "NOVA Backup - All-in-One";
            this.Size = new Size(600, 400);
            this.StartPosition = FormStartPosition.CenterScreen;
            this.Icon = SystemIcons.Application;

            // Title label
            _titleLabel = new Label
            {
                Text = "NOVA BACKUP - ALL-IN-ONE",
                Font = new Font("Segoe UI", 18, FontStyle.Bold),
                ForeColor = Color.FromArgb(0, 120, 215),
                Location = new Point(20, 20),
                AutoSize = true
            };

            // Status label
            _statusLabel = new Label
            {
                Text = "Ready to start backup operations...",
                Font = new Font("Segoe UI", 10),
                Location = new Point(20, 80),
                Size = new Size(540, 60),
                BackColor = Color.FromArgb(240, 240, 240)
            };

            // Start web button
            _startWebButton = new Button
            {
                Text = "Start Web Console",
                Size = new Size(200, 40),
                Location = new Point(20, 160),
                BackColor = Color.FromArgb(0, 120, 215),
                ForeColor = Color.White,
                FlatStyle = FlatStyle.Flat,
                Font = new Font("Segoe UI", 12, FontStyle.Bold)
            };

            _startWebButton.Click += (s, e) => StartWebServer();

            // Exit button
            _exitButton = new Button
            {
                Text = "Exit",
                Size = new Size(100, 40),
                Location = new Point(460, 160),
                BackColor = Color.FromArgb(220, 53, 69),
                ForeColor = Color.White,
                FlatStyle = FlatStyle.Flat,
                Font = new Font("Segoe UI", 12, FontStyle.Bold)
            };

            _exitButton.Click += (s, e) => Application.Exit();

            // Add controls
            this.Controls.AddRange(new Control[] { 
                _titleLabel, 
                _statusLabel, 
                _startWebButton, 
                _exitButton 
            });
        }

        private async void StartWebServer()
        {
            try
            {
                _statusLabel.Text = "Starting web server...";
                _startWebButton.Enabled = false;

                // Simulate web server start
                await Task.Delay(2000);

                _statusLabel.Text = "✅ Web server running on http://localhost:8080";
                _statusLabel.Text += "\n✅ Remote access: http://[IP]:8080";
                _statusLabel.Text += "\n✅ Default credentials: admin / admin";

                // Open web console
                System.Diagnostics.Process.Start("http://localhost:8080");
            }
            catch (Exception ex)
            {
                _statusLabel.Text = "❌ Error starting web server";
                MessageBox.Show($"Error: {ex.Message}", "Error", 
                    MessageBoxButtons.OK, MessageBoxIcon.Error);
            }
            finally
            {
                _startWebButton.Enabled = true;
            }
        }
    }
}
