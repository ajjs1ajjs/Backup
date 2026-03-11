using System;
using System.Drawing;
using System.Windows.Forms;
using NovaBackup.Desktop.Services;

namespace NovaBackup.Desktop.Controls
{
    public partial class SchedulesControl : UserControl, IRefreshable
    {
        private readonly NovaBackupService _backupService;

        public SchedulesControl(NovaBackupService backupService)
        {
            _backupService = backupService;
            InitializeComponent();
        }

        private void InitializeComponent()
        {
            this.Size = new Size(800, 600);
            this.Dock = DockStyle.Fill;
        }

        public async Task Refresh()
        {
            // Implementation
        }
    }
}
