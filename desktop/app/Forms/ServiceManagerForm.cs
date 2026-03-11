using System;
using System.Drawing;
using System.Windows.Forms;
using NovaBackup.Desktop.Services;

namespace NovaBackup.Desktop.Forms
{
    public partial class ServiceManagerForm : Form
    {
        private readonly NovaBackupService _backupService;

        public ServiceManagerForm(NovaBackupService backupService)
        {
            _backupService = backupService;
            InitializeComponent();
        }

        private void InitializeComponent()
        {
            this.Text = "Service Manager";
            this.Size = new Size(600, 500);
            this.StartPosition = FormStartPosition.CenterParent;
        }
    }
}
