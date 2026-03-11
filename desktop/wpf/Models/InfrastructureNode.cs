using CommunityToolkit.Mvvm.ComponentModel;
using System.Collections.ObjectModel;

namespace NovaBackup.GUI.Models
{
    public partial class InfrastructureNode : ObservableObject
    {
        [ObservableProperty]
        private string _id = string.Empty;

        [ObservableProperty]
        private string _name = string.Empty;

        [ObservableProperty]
        private string _nodeType = string.Empty;

        [ObservableProperty]
        private string _iconKind = "Folder";

        [ObservableProperty]
        private ObservableCollection<InfrastructureNode> _children = new();

        [ObservableProperty]
        private bool _isExpanded;

        [ObservableProperty]
        private bool _isSelected;
        
        public InfrastructureNode() { }

        public InfrastructureNode(string name, string iconKind = "Folder")
        {
            Name = name;
            IconKind = iconKind;
        }
    }
}
