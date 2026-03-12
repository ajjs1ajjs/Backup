using System.Collections.ObjectModel;
using System.Threading.Tasks;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using NovaBackup.GUI.Services;

namespace NovaBackup.GUI.ViewModels
{
    public partial class ReportsViewModel : ObservableObject
    {
        private readonly IApiClient _apiClient;

        [ObservableProperty]
        private ObservableCollection<ReportModel> _reports = new();

        [ObservableProperty]
        private ReportModel? _selectedReport;

        [ObservableProperty]
        private bool _isLoading;

        [ObservableProperty]
        private string _statusMessage = string.Empty;

        public ReportsViewModel(IApiClient apiClient)
        {
            _apiClient = apiClient;
        }

        [RelayCommand]
        private async Task LoadDataAsync()
        {
            IsLoading = true;
            StatusMessage = "Loading reports...";
            try
            {
                var reports = await _apiClient.GetReportsAsync();
                Reports.Clear();
                foreach (var r in reports) Reports.Add(r);
                StatusMessage = $"Loaded {Reports.Count} reports";
            }
            catch { StatusMessage = "Error loading data"; }
            finally { IsLoading = false; }
        }

        [RelayCommand]
        private async Task GenerateReportAsync()
        {
            var request = new ReportRequest
            {
                Name = "Backup Summary",
                Type = "summary",
                From = System.DateTime.Now.AddDays(-30).ToString("yyyy-MM-dd"),
                To = System.DateTime.Now.ToString("yyyy-MM-dd"),
                Format = "pdf"
            };
            
            var success = await _apiClient.GenerateReportAsync(request);
            StatusMessage = success ? "Report generated successfully" : "Error generating report";
            if (success) await LoadDataAsync();
        }
    }
}
