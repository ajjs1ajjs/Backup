using System.Collections.Generic;

namespace NovaBackup.GUI.Models
{
    public class InfrastructureObject
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string ObjType { get; set; } = "VM";
        public string ExternalId { get; set; } = string.Empty;
        public string Metadata { get; set; } = string.Empty;
        public string IconKind => ObjType == "VM" ? "DesktopClassic" : "Database";
    }
}
