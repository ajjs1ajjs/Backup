using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public class RecoverySessionsViewModel : RecoverySessionsViewModelMVVM
    {
        public RecoverySessionsViewModel(IApiClient apiClient) : base(apiClient)
        {
        }
    }
}
