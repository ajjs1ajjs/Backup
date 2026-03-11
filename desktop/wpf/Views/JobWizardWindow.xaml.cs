using System.Windows;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class JobWizardWindow : Window
    {
        public JobWizardWindow(JobWizardViewModel viewModel)
        {
            InitializeComponent();
            DataContext = viewModel;
        }
    }
}
