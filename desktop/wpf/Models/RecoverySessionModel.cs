namespace NovaBackup.GUI.Models
{
    public class RecoverySessionModel
    {
        public string SessionID { get; set; } = string.Empty;
        public string VMName { get; set; } = string.Empty;
        public string Status { get; set; } = string.Empty;
        public double Progress { get; set; }
        public string FormattedProgress => $"{(int)(Progress * 100)}%";
    }
}
