using System;

namespace NovaBackup.GUI.Models
{
    public class RecoverySessionModel
    {
        public string SessionID { get; set; } = string.Empty;
        public string VMName { get; set; } = string.Empty;
        public string RestorePointID { get; set; } = string.Empty;
        public string NFSExport { get; set; } = string.Empty;
        public string Status { get; set; } = string.Empty;
        public DateTime StartTime { get; set; }
        public string TargetHost { get; set; } = string.Empty;
        public string Platform { get; set; } = "Hyper-V"; // Default for now
        public double Progress { get; set; }
        public long TotalSize { get; set; }
        public string FormattedProgress => $"{Progress:P0}";
    }
}
