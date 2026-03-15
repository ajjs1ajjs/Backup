using System.Windows;

namespace NovaBackup.WPF
{
    public partial class EditJobWindow : Window
    {
        private readonly BackupJob _job;

        public EditJobWindow(BackupJob job)
        {
            InitializeComponent();
            _job = job;

            // Load job data
            txtName.Text = job.Name;
            txtDescription.Text = job.Description;
            txtDestination.Text = job.Destination;
            chkCompression.IsChecked = job.Compression;
            chkEncryption.IsChecked = job.Encryption;
            cmbSchedule.SelectedIndex = job.ScheduleType switch
            {
                "Daily" => 0,
                "Weekly" => 1,
                "Monthly" => 2,
                _ => 0
            };
            txtTime.Text = job.ScheduleTime;
        }

        private void BtnCancel_Click(object sender, RoutedEventArgs e)
        {
            DialogResult = false;
            Close();
        }

        private void BtnSave_Click(object sender, RoutedEventArgs e)
        {
            // Update job
            _job.Name = txtName.Text;
            _job.Description = txtDescription.Text;
            _job.Destination = txtDestination.Text;
            _job.Compression = chkCompression.IsChecked == true;
            _job.Encryption = chkEncryption.IsChecked == true;
            _job.ScheduleType = cmbSchedule.SelectedIndex switch
            {
                0 => "Daily",
                1 => "Weekly",
                2 => "Monthly",
                _ => "Daily"
            };
            _job.ScheduleTime = txtTime.Text;

            // Save
            JobManager.UpdateJob(_job);

            MessageBox.Show("Job updated successfully!", "Success",
                MessageBoxButton.OK, MessageBoxImage.Information);

            DialogResult = true;
            Close();
        }
    }
}
