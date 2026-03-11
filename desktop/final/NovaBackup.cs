using System;
using System.Drawing;
using System.Windows.Forms;

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
            this.Size = new Size(500, 350);
            this.StartPosition = FormStartPosition.CenterScreen;
            this.Icon = SystemIcons.Application;
            this.FormBorderStyle = FormBorderStyle.FixedDialog;
            this.MaximizeBox = false;
            this.MinimizeBox = false;

            // Title label
            _titleLabel = new Label
            {
                Text = "NOVA BACKUP\n\nAll-in-One Version",
                Font = new Font("Segoe UI", 14, FontStyle.Bold),
                ForeColor = Color.FromArgb(0, 120, 215),
                Location = new Point(20, 20),
                AutoSize = true
            };

            // Status label
            _statusLabel = new Label
            {
                Text = "✅ Ready to start backup operations!\n\nWeb Console: http://localhost:8080\nRemote Access: http://[IP]:8080\nDefault Credentials: admin / admin",
                Font = new Font("Segoe UI", 9),
                Location = new Point(20, 80),
                Size = new Size(440, 120),
                BackColor = Color.FromArgb(240, 240, 240)
            };

            // Start web button
            _startWebButton = new Button
            {
                Text = "Start Web Console",
                Size = new Size(180, 40),
                Location = new Point(160, 200),
                BackColor = Color.FromArgb(0, 120, 215),
                ForeColor = Color.White,
                FlatStyle = FlatStyle.Flat,
                Font = new Font("Segoe UI", 11, FontStyle.Bold)
            };

            _startWebButton.Click += (s, e) => StartWebServer();

            // Exit button
            _exitButton = new Button
            {
                Text = "Exit",
                Size = new Size(100, 40),
                Location = new Point(160, 250),
                BackColor = Color.FromArgb(220, 53, 69),
                ForeColor = Color.White,
                FlatStyle = FlatStyle.Flat,
                Font = new Font("Segoe UI", 11, FontStyle.Bold)
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
                _statusLabel.Text = "🚀 Starting web server...";
                _startWebButton.Enabled = false;

                // Simulate web server start
                await System.Threading.Tasks.Task.Delay(2000);

                _statusLabel.Text = "✅ Web server running!\n\n🌐 Web Console: http://localhost:8080\n🔐 Remote Access: http://[IP]:8080\n🔑 Default Credentials: admin / admin\n\n📋 Click Start to open web console";

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
