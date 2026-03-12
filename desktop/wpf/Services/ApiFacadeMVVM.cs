using NovaBackup.GUI.Models;
using NovaBackup.GUI.Services;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace NovaBackup.GUI.Services
{
    public class ApiFacadeMVVM : ApiFacade
    {
        public ApiFacadeMVVM(IApiClient client) : base(client) { }
        // Inherit and extend for MVVM usage
    }
}
