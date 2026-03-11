# NOVA Backup - All-in-One Version

## Overview

This is the **All-in-One** version of NOVA Backup - a single executable file that contains everything you need to run the complete backup solution on any Windows machine.

## Features Included

✅ **Complete Desktop Application**
- Modern WinForms interface with full GUI
- System tray integration
- Real-time backup monitoring
- Multi-language support (English & Ukrainian)

✅ **Web Console**
- Built-in web server (port 8080)
- Remote access from any device
- Secure authentication for remote connections
- Responsive web interface

✅ **Windows Service**
- Background processing
- Scheduled backup execution
- Service management capabilities

✅ **All Dependencies**
- .NET 6.0 runtime included
- All libraries and frameworks embedded
- No additional installation required

## Quick Start

### 1. Build the All-in-One Executable

Run the build script:
```cmd
build-all-in-one.bat
```

This will create:
- `installer\NovaBackup.exe` - The complete all-in-one executable

### 2. Install and Run

#### Method 1: Simple Installation
1. Run `install-all-in-one.bat` as Administrator
2. Follow the installation prompts
3. Launch from desktop or Start Menu

#### Method 2: Portable Mode
1. Simply run `NovaBackup.exe` directly
2. No installation required
3. All data stored in local folders

## Access Methods

### Desktop Application
- **Launch**: Double-click `NovaBackup.exe` or use desktop shortcut
- **Features**: Complete GUI with all backup management features

### Web Console
- **Local Access**: `http://localhost:8080` (no authentication)
- **Remote Access**: `http://[IP_ADDRESS]:8080` (admin/admin credentials)
- **Features**: Web-based management interface

### Command Line
```cmd
# Run as console application
NovaBackup.exe console

# Install as Windows Service
NovaBackup.exe install

# Start service
NovaBackup.exe start

# Stop service
NovaBackup.exe stop

# Uninstall service
NovaBackup.exe uninstall
```

## File Structure

After installation, the following structure is created:

```
C:\Program Files\NovaBackup\
├── NovaBackup.exe              # Main executable
└── uninstall.bat               # Uninstallation script

C:\ProgramData\NovaBackup\
├── logs\                        # Application logs
├── config\                      # Configuration files
│   └── backup-config.json       # Main configuration
└── backups\                     # Default backup storage
```

## Configuration

The main configuration file is located at:
```
C:\ProgramData\NovaBackup\config\backup-config.json
```

Default settings:
```json
{
  "Settings": {
    "DefaultBackupPath": "C:\\ProgramData\\NovaBackup\\backups",
    "MaxConcurrentBackups": 3,
    "CompressionLevel": "normal",
    "EnableEncryption": true,
    "EnableNotifications": true,
    "WebConsoleEnabled": true,
    "WebConsolePort": 8080
  },
  "BackupJobs": [],
  "Schedules": []
}
```

## Security

### Local Access
- No authentication required for localhost access
- Full functionality available immediately

### Remote Access
- Basic authentication required for non-localhost connections
- Default credentials: `admin / admin`
- **Important**: Change default credentials after first use

### Firewall
- Port 8080 automatically configured during installation
- Web console accessible from network devices

## System Requirements

### Minimum Requirements
- Windows 10 (Version 1903) or later
- Windows Server 2016 or later
- 4 GB RAM
- 2 GB available disk space

### Recommended Requirements
- Windows 11 or Windows Server 2022
- 8 GB RAM or more
- 10 GB available disk space
- Network connection for remote access

## Troubleshooting

### Common Issues

#### Application Won't Start
1. Check Windows Event Viewer for errors
2. Ensure .NET 6.0 is compatible with your system
3. Run as Administrator if permission issues

#### Web Console Not Accessible
1. Check if port 8080 is blocked by firewall
2. Verify web console is enabled in settings
3. Check if another application uses port 8080

#### Remote Access Issues
1. Verify firewall allows port 8080
2. Check network connectivity
3. Use correct credentials (admin/admin)

#### Service Issues
1. Run as Administrator for service operations
2. Check Windows Services console
3. Review event logs for service errors

### Log Files
Log files are located at:
```
C:\ProgramData\NovaBackup\logs\
```

- `backup-YYYY-MM-DD.log` - Backup operations
- `error-YYYY-MM-DD.log` - Error logs
- `service-YYYY-MM-DD.log` - Service logs

## Uninstallation

### Method 1: Use Uninstaller
1. Go to Start Menu → Programs → NovaBackup
2. Click "Uninstall"
3. Follow the prompts

### Method 2: Manual Uninstall
1. Run `C:\Program Files\NovaBackup\uninstall.bat`
2. Follow the prompts

### Method 3: Portable Mode Cleanup
For portable mode, simply:
1. Stop any running NovaBackup processes
2. Delete the executable file
3. Remove data folders if desired

## Performance Tips

### Optimize Backup Performance
- Use SSD storage for better performance
- Configure appropriate compression levels
- Schedule backups during off-peak hours
- Limit concurrent backup operations

### System Performance
- Monitor system resource usage
- Configure appropriate retention policies
- Regularly clean up old backup files
- Use network storage for large backups

## Advanced Configuration

### Custom Port
To change the web console port:
1. Edit `backup-config.json`
2. Change `WebConsolePort` value
3. Restart the application
4. Update firewall rules

### Custom Storage Location
To change default backup path:
1. Edit `backup-config.json`
2. Update `DefaultBackupPath`
3. Ensure the new path exists and has proper permissions

### Authentication
To change remote access credentials:
1. Edit `WebApiService.cs` (in source code)
2. Update `IsValidAuth` method
3. Rebuild the executable

## Support

### Documentation
- Full documentation: Available in the application
- Web console help: Built-in help system
- Command line help: `NovaBackup.exe --help`

### Getting Help
- Check log files for error details
- Review system event logs
- Test with different configurations
- Contact support if issues persist

## Version Information

- **Version**: 1.0.0
- **Build Date**: Current build
- **Framework**: .NET 6.0
- **Platform**: Windows x64
- **Deployment**: Self-contained single file

## License

NOVA Backup is licensed under the MIT License. See LICENSE file for details.

---

**Enjoy using NOVA Backup - Your Complete All-in-One Backup Solution!** 🚀
