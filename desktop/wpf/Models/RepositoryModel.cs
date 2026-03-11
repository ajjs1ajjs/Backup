using System;

namespace NovaBackup.GUI.Models
{
    public class RepositoryModel
    {
        public string ID { get; set; } = Guid.NewGuid().ToString();
        public string Name { get; set; } = string.Empty;
        public string Type { get; set; } = "local"; // local, s3, sobr
        public string Path { get; set; } = string.Empty;
        public string Endpoint { get; set; } = string.Empty;
        public string Bucket { get; set; } = string.Empty;
        public string Region { get; set; } = string.Empty;
        public string AccessKey { get; set; } = string.Empty;
        public string SecretKey { get; set; } = string.Empty;
        public bool IsSOBR { get; set; }
        public string Tier { get; set; } = "performance"; // performance, capacity
        
        public double CapacityGB { get; set; }
        public double FreeGB { get; set; }
        public string Status { get; set; } = "Online";
        public double UsedGB => CapacityGB - FreeGB;
    }
}
