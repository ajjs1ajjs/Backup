using System;
using CommunityToolkit.Mvvm.ComponentModel;

namespace NovaBackup.GUI.Models
{
    public partial class RestorePointModel : ObservableObject
    {
        [ObservableProperty]
        private string _id = string.Empty;

        [ObservableProperty]
        private string _jobId = string.Empty;

        [ObservableProperty]
        private DateTime _pointTime;

        [ObservableProperty]
        private string _status = "Completed";

        [ObservableProperty]
        private long _totalBytes;

        public string DisplayName => $"{PointTime:yyyy-MM-dd HH:mm} ({Status})";
    }
}
