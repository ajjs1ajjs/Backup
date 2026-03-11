using System;
using System.Threading.Tasks;
using System.Windows.Forms;
using NovaBackup.Desktop.Services;

namespace NovaBackup.Desktop.Controls
{
    public partial class DashboardControl : UserControl, IRefreshable
    {
        private readonly NovaBackupService _backupService;

        public DashboardControl(NovaBackupService backupService)
        {
            _backupService = backupService;
            InitializeComponent();
        }

        private void InitializeComponent()
        {
            this.Size = new System.Drawing.Size(800, 600);
            this.Dock = DockStyle.Fill;
        }

        public async Task Refresh()
        {
            // Implementation
        }
    }
}
