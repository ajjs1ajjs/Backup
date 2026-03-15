using System;
using System.Collections.Generic;

namespace NovaBackup.WPF
{
    public class BackupJob
    {
        public string Id { get; set; } = Guid.NewGuid().ToString();
        public string Name { get; set; } = "";
        public string Type { get; set; } = "";
        public List<string> Sources { get; set; } = new List<string>();
        public string Destination { get; set; } = "";
        public bool Compression { get; set; }
        public bool Encryption { get; set; }
        public string ScheduleType { get; set; } = "";
        public string ScheduleTime { get; set; } = "";
        public bool Enabled { get; set; } = true;

        // Display properties
        public string StatusIcon { get; set; } = "✅";
        public string Status { get; set; } = "Активно";
        public string LastRun { get; set; } = "";
        public string NextRun { get; set; } = "";
        public string Schedule { get; set; } = "";
    }
}
