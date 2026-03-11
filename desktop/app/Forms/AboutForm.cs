using System;
using System.Drawing;
using System.Windows.Forms;

namespace NovaBackup.Desktop.Forms
{
    public partial class AboutForm : Form
    {
        public AboutForm()
        {
            InitializeComponent();
        }

        private void InitializeComponent()
        {
            this.Text = "About NOVA Backup";
            this.Size = new Size(400, 300);
            this.StartPosition = FormStartPosition.CenterParent;
        }
    }
}
