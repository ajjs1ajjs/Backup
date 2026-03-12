using System;

namespace NovaBackup.GUI.Models
{
    public class RecoverySessionModel
    {
        public string SessionID { get; set; } = string.Empty;
        public string VMName { get; set; } = string.Empty;
        public string Status { get; set; } = string.Empty;
        public double Progress { get; set; } = 0; // 0.0 - 1.0
        public string FormattedProgress => $"{(int)(Progress * 100)}%";
    }
}
