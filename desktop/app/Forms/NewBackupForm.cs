using System;
using System.Drawing;
using System.Windows.Forms;
using NovaBackup.Desktop.Services;

namespace NovaBackup.Desktop.Forms
{
    public partial class NewBackupForm : Form
    {
        private readonly NovaBackupService _backupService;

        public NewBackupForm(NovaBackupService backupService)
        {
            _backupService = backupService;
            InitializeComponent();
        }

        private void InitializeComponent()
        {
            this.Text = "New Backup";
            this.Size = new Size(500, 400);
            this.StartPosition = FormStartPosition.CenterParent;
        }
    }
}
