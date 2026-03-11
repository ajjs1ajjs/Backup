using System;
using System.Drawing;
using System.Windows.Forms;
using NovaBackup.Desktop.Services;

namespace NovaBackup.Desktop.Forms
{
    public partial class RestoreBackupForm : Form
    {
        private readonly NovaBackupService _backupService;

        public RestoreBackupForm(NovaBackupService backupService)
        {
            _backupService = backupService;
            InitializeComponent();
        }

        private void InitializeComponent()
        {
            this.Text = "Restore Backup";
            this.Size = new Size(500, 400);
            this.StartPosition = FormStartPosition.CenterParent;
        }
    }
}
