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
            Application.SetHighDpiMode(HighDpiMode.SystemAware);
            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);

            var mainForm = new TestMainForm();
            Application.Run(mainForm);
        }
    }

    public class TestMainForm : Form
    {
        private Label _titleLabel;
        private Label _statusLabel;
        private Button _startButton;
        private Button _exitButton;

        public TestMainForm()
        {
            InitializeComponent();
        }

        private void InitializeComponent()
        {
            this.Text = "NOVA Backup - Test Version";
            this.Size = new Size(500, 300);
            this.StartPosition = FormStartPosition.CenterScreen;
            this.Icon = SystemIcons.Application;

            // Title label
            _titleLabel = new Label
            {
                Text = "NOVA BACKUP - TEST VERSION",
                Font = new Font("Segoe UI", 16, FontStyle.Bold),
                ForeColor = Color.FromArgb(0, 120, 215),
                Location = new Point(20, 20),
                AutoSize = true
            };

            // Status label
            _statusLabel = new Label
            {
                Text = "This is a test version of NOVA Backup.\n\nTo build the full version:\n1. Run build-final.ps1 in PowerShell\n2. Or use build-simple.bat\n\n3. The final executable will be created in installer/ directory",
                Font = new Font("Segoe UI", 9),
                Location = new Point(20, 60),
                Size = new Size(440, 100),
                BackColor = Color.FromArgb(240, 240, 240)
            };

            // Start button
            _startButton = new Button
            {
                Text = "Build Full Version",
                Size = new Size(150, 35),
                Location = new Point(20, 180),
                BackColor = Color.FromArgb(0, 120, 215),
                ForeColor = Color.White,
                FlatStyle = FlatStyle.Flat,
                Font = new Font("Segoe UI", 10, FontStyle.Bold)
            };

            _startButton.Click += (s, e) => {
                MessageBox.Show("To build the full version:\n\n1. Run build-final.ps1 in PowerShell\n2. Or use build-simple.bat\n\n3. The final executable will be created in installer/ directory", 
                    "Build Instructions", MessageBoxButtons.OK, MessageBoxIcon.Information);
            };

            // Exit button
            _exitButton = new Button
            {
                Text = "Exit",
                Size = new Size(100, 35),
                Location = new Point(190, 180),
                BackColor = Color.FromArgb(220, 53, 69),
                ForeColor = Color.White,
                FlatStyle = FlatStyle.Flat,
                Font = new Font("Segoe UI", 10, FontStyle.Bold)
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
    }
}
