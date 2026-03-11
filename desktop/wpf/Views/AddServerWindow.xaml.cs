using System.Windows;
using NovaBackup.GUI.ViewModels;

namespace NovaBackup.GUI.Views
{
    public partial class AddServerWindow : Window
    {
        public AddServerWindow(AddServerViewModel viewModel)
        {
            InitializeComponent();
            DataContext = viewModel;
        }
    }
}
