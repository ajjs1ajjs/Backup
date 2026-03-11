# NOVA Backup Desktop Application

## Overview

NOVA Backup Desktop is a comprehensive backup solution for Windows that provides both a desktop application and web-based management interface. It's designed for enterprise environments with robust features including scheduled backups, encryption, compression, and multi-tenant support.

## Features

### Desktop Application
- **Modern WinForms Interface**: Clean, intuitive user interface with navigation tree and status bar
- **System Tray Integration**: Minimize to system tray with quick access to common operations
- **Real-time Monitoring**: Live backup status and progress tracking
- **Notification System**: Toast notifications for backup events
- **Multi-language Support**: English and Ukrainian language support

### Web Console
- **Responsive Web Interface**: Modern web console accessible via browser
- **Remote Access**: Accessible from any device on the network
- **Authentication**: Secure login for remote access (localhost access without auth)
- **RESTful API**: Complete API for programmatic access
- **Real-time Dashboard**: Overview of backup status and system health
- **Backup Management**: Create, edit, and manage backup jobs
- **Schedule Management**: Configure automated backup schedules
- **Storage Monitoring**: Track storage usage and drive health
- **Reporting**: Generate detailed backup reports

**Access URLs:**
- **Local Access**: `http://localhost:8080` (no authentication required)
- **Remote Access**: `http://[IP_ADDRESS]:8080` (authentication required)
- **Default Credentials**: `admin / admin`

### Windows Service
- **Background Processing**: Runs as Windows Service for automated operations
- **Service Management**: Install, start, stop, and uninstall service
- **Configuration Management**: Persistent configuration storage
- **Logging**: Comprehensive logging for troubleshooting

## Installation

### Prerequisites
- Windows 10 or later / Windows Server 2016 or later
- .NET 6.0 Runtime or later
- Administrator privileges for installation

### Installation Steps
1. Download `NovaBackupSetup.exe` from the releases page
2. Run the installer as Administrator
3. Follow the installation wizard:
   - Choose installation directory (default: `C:\Program Files\NovaBackup`)
   - Select components to install (Desktop App, Windows Service, Web Console)
   - Configure installation options (desktop icon, service startup)
4. Complete the installation
5. Launch NOVA Backup from Start Menu or desktop icon

### Silent Installation
```cmd
NovaBackupSetup.exe /VERYSILENT /DIR="C:\NovaBackup" /COMPONENTS="main,service,webconsole"
```

## Usage

### Desktop Application
1. **Launch**: Start NOVA Backup from Start Menu
2. **Dashboard**: View overview of backup status and system health
3. **Backups**: Create and manage backup jobs
4. **Schedules**: Configure automated backup schedules
5. **Storage**: Monitor storage usage and drive health
6. **Reports**: Generate and view backup reports
7. **Settings**: Configure application settings

### Web Console
1. **Access**: Open `http://localhost:8080` in your browser for local access
2. **Remote Access**: Open `http://[COMPUTER_IP]:8080` from any device on the network
3. **Authentication**: Login with `admin / admin` for remote access (no auth required for localhost)
4. **Navigate**: Use the sidebar to access different sections
5. **Manage**: Perform all backup operations via web interface

### Remote Access Setup
1. **Install**: Install NOVA Backup with Web Console component
2. **Firewall**: Installer automatically configures Windows Firewall for port 8080
3. **Network**: Ensure the computer is accessible from the network
4. **Access**: Use `http://[IP_ADDRESS]:8080` from any device
5. **Login**: Use credentials `admin / admin` for first-time access

### Security Considerations
- **Local Access**: No authentication required for localhost
- **Remote Access**: Basic authentication required for non-localhost connections
- **Default Credentials**: Change default username/password after installation
- **Network Security**: Consider VPN or secure network for remote access
- **Firewall**: Port 8080 must be open for remote access

### Windows Service
The Windows Service runs automatically in the background and handles:
- Scheduled backup execution
- Background monitoring
- Web API hosting
- Configuration management

## Configuration

### Default Configuration File Location
```
C:\ProgramData\NovaBackup\config\backup-config.json
```

### Key Configuration Options
```json
{
  "Settings": {
    "DefaultBackupPath": "C:\\NovaBackup\\Backups",
    "MaxConcurrentBackups": 3,
    "CompressionLevel": "normal",
    "EnableEncryption": true,
    "EnableNotifications": true,
    "WebConsoleEnabled": true,
    "WebConsolePort": 8080
  }
}
```

### Log Files Location
```
C:\ProgramData\NovaBackup\logs\
```

## API Documentation

### Base URL
```
http://localhost:8080/api
```

### Endpoints

#### Status
- `GET /api/status` - Get current backup status

#### Backups
- `GET /api/backups` - List all backup jobs
- `POST /api/backups` - Create new backup job
- `GET /api/backups/{id}` - Get specific backup job
- `PUT /api/backups/{id}` - Update backup job
- `DELETE /api/backups/{id}` - Delete backup job

#### Schedules
- `GET /api/schedules` - List all schedules
- `POST /api/schedules` - Create new schedule
- `GET /api/schedules/{id}` - Get specific schedule
- `PUT /api/schedules/{id}` - Update schedule
- `DELETE /api/schedules/{id}` - Delete schedule

#### Storage
- `GET /api/storage` - Get storage information

#### Reports
- `GET /api/reports?from=YYYY-MM-DD&to=YYYY-MM-DD` - Generate report

### Example API Usage
```bash
# Get backup status
curl http://localhost:8080/api/status

# Create backup job
curl -X POST http://localhost:8080/api/backups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Backup",
    "sourcePath": "C:\\Data",
    "destinationPath": "D:\\Backups",
    "backupType": "full"
  }'
```

## Command Line Options

### Service Management
```cmd
# Install Windows Service
NovaBackup.exe install

# Uninstall Windows Service
NovaBackup.exe uninstall

# Start Service
NovaBackup.exe start

# Stop Service
NovaBackup.exe stop
```

### Console Mode
```cmd
# Run in console mode (for debugging)
NovaBackup.exe console
```

## Troubleshooting

### Common Issues

#### Service Won't Start
1. Check if .NET 6.0 Runtime is installed
2. Verify administrator privileges
3. Check Windows Event Log for errors
4. Ensure port 8080 is not in use

#### Web Console Not Accessible
1. Verify web console is enabled in settings
2. Check if service is running
3. Verify firewall allows port 8080
4. Check if another application is using port 8080

#### Backup Failures
1. Check source and destination paths
2. Verify disk space availability
3. Check file permissions
4. Review log files for specific error messages

### Log Analysis
Log files are located in `C:\ProgramData\NovaBackup\logs\`:
- `backup-YYYY-MM-DD.log` - Backup operation logs
- `error-YYYY-MM-DD.log` - Error logs
- `service-YYYY-MM-DD.log` - Service logs

## Security

### Encryption
- AES-256 encryption for backup data
- Configurable encryption keys
- Secure key storage

### Access Control
- Windows Service runs with appropriate privileges
- Web console access limited to localhost by default
- Configuration files protected by Windows file permissions

### Best Practices
- Use strong encryption for sensitive data
- Regularly update encryption keys
- Monitor log files for security events
- Restrict web console access in production environments

## Performance Optimization

### Backup Performance
- Use incremental backups for large datasets
- Configure appropriate compression levels
- Schedule backups during off-peak hours
- Use SSD storage for better performance

### System Performance
- Limit concurrent backup operations
- Monitor system resource usage
- Configure appropriate retention policies
- Regularly clean up old backup files

## Integration

### Third-party Tools
- Compatible with Windows Task Scheduler
- Integrates with Windows Event Log
- Supports PowerShell automation
- REST API for custom integrations

### PowerShell Examples
```powershell
# Get backup status
$response = Invoke-RestMethod -Uri "http://localhost:8080/api/status"
$response | ConvertTo-Json

# Create backup job
$backup = @{
    name = "PowerShell Backup"
    sourcePath = "C:\Data"
    destinationPath = "D:\Backups"
    backupType = "incremental"
}
Invoke-RestMethod -Uri "http://localhost:8080/api/backups" -Method POST -Body ($backup | ConvertTo-Json) -ContentType "application/json"
```

## Support

### Documentation
- Online documentation: https://novabackup.com/docs
- API reference: https://novabackup.com/api
- Community forums: https://novabackup.com/community

### Contact Support
- Email: support@novabackup.com
- Phone: +1-800-NOVA-BACK
- Live chat: Available on website

## License

NOVA Backup is licensed under the MIT License. See LICENSE file for details.

## Version History

### Version 1.0.0 (2024-03-11)
- Initial release
- Desktop application with WinForms interface
- Web console with responsive design
- Windows Service integration
- REST API for programmatic access
- Comprehensive backup and scheduling features
- Multi-language support (English, Ukrainian)

## System Requirements

### Minimum Requirements
- Windows 10 (Version 1903) or later
- Windows Server 2016 or later
- 4 GB RAM
- 2 GB available disk space
- .NET 6.0 Runtime

### Recommended Requirements
- Windows 11 or Windows Server 2022
- 8 GB RAM or more
- 10 GB available disk space
- SSD storage for better performance
- Network connection for web console access

## Development

### Building from Source
```bash
# Clone repository
git clone https://github.com/novabackup/desktop.git
cd desktop

# Build application
dotnet build NovaBackup.Desktop.csproj

# Run application
dotnet run --project NovaBackup.Desktop.csproj

# Publish for distribution
dotnet publish NovaBackup.Desktop.csproj -c Release -r win-x64 --self-contained
```

### Project Structure
```
desktop/
├── app/                    # Desktop application
│   ├── MainForm.cs         # Main application window
│   ├── Program.cs          # Application entry point
│   └── NovaBackup.Desktop.csproj
├── services/               # Service implementations
│   ├── NovaBackupService.cs
│   ├── SystemTrayManager.cs
│   ├── WebApiService.cs
│   └── WindowsService.cs
├── web-ui/                 # Web console files
│   ├── index.html
│   └── assets/
└── installer/              # Installation scripts
    └── NovaBackup.iss
```

## Contributing

We welcome contributions to NOVA Backup! Please see our contributing guidelines for more information.

### Contributing Guidelines
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

### Code Style
- Follow C# coding conventions
- Use meaningful variable names
- Add comments for complex logic
- Include XML documentation for public APIs
