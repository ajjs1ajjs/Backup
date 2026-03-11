using CommunityToolkit.Mvvm.ComponentModel;
using System;

namespace NovaBackup.GUI.Models
{
    public partial class JobModel : ObservableObject
    {
        [ObservableProperty]
        private string _id = Guid.NewGuid().ToString();

        [ObservableProperty]
        private string _name = string.Empty;

        [ObservableProperty]
        private string _platform = string.Empty; // e.g., VMware, Hyper-V, Windows Agent

        [ObservableProperty]
        private string _status = "Stopped"; // Running, Success, Warning, Failed

        [ObservableProperty]
        private string _lastResult = "None"; // Success, Warning, Failed

        [ObservableProperty]
        private DateTime? _lastRun;

        [ObservableProperty]
        private DateTime? _nextRun;

        [ObservableProperty]
        private double _sizeGB;

        [ObservableProperty]
        private int _retentionDays = 30;

        [ObservableProperty]
        private bool _enableGuestProcessing;

        [ObservableProperty]
        private string _guestCredentialsId = string.Empty;
    }
}
