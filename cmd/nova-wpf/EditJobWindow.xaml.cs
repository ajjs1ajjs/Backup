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

            txtName.Text = job.Name;
            txtDestination.Text = job.Destination;
            chkCompression.IsChecked = job.Compression;
            chkEncryption.IsChecked = job.Encryption;
            cmbSchedule.SelectedIndex = 0;
            txtTime.Text = "02:00";
        }

        private void BtnCancel_Click(object sender, RoutedEventArgs e)
        {
            DialogResult = false;
            Close();
        }

        private void BtnSave_Click(object sender, RoutedEventArgs e)
        {
            _job.Name = txtName.Text;
            _job.Destination = txtDestination.Text;
            _job.Compression = chkCompression.IsChecked == true;
            _job.Encryption = chkEncryption.IsChecked == true;

            JobManager.SaveJobs(new System.Collections.ObjectModel.ObservableCollection<BackupJob> { _job });
            MessageBox.Show("Завдання оновлено!", "Успіх", MessageBoxButton.OK, MessageBoxImage.Information);
            DialogResult = true;
            Close();
        }
    }
}
