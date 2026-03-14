# 🛡️ NovaBackup - Modern Backup System for Windows

A comprehensive enterprise-grade backup solution built with .NET 8 and WPF, featuring a Veeam-inspired modern interface.

![.NET](https://img.shields.io/badge/.NET-8-purple)
![C#](https://img.shields.io/badge/C%23-latest-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![GUI](https://img.shields.io/badge/GUI-Veeam--style-orange)

---

## 🎯 What's New in v1.0.0

### ✨ Modern Veeam-style Interface
- **Ribbon Menu** - Familiar tabbed interface (Home, Jobs, Infrastructure, Monitoring)
- **Navigation Tree** - Hierarchical view of backup infrastructure
- **Properties Panel** - Detailed object information
- **Status Bar** - Real-time status updates
- **DataGrid Views** - Professional tables for jobs, repositories, sessions

### 🔧 Technical Updates
- Migrated to .NET 8.0 (from .NET 10)
- Fixed MSI installer compilation
- Improved WPF data binding
- Enhanced error handling

---

## 📋 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Requirements](#-requirements)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Usage Guide](#-usage-guide)
- [Configuration](#-configuration)
- [API Reference](#-api-reference)
- [Development](#-development)
- [Troubleshooting](#-troubleshooting)
- [Contributing](#-contributing)
- [License](#-license)

---

## ✨ Features

### Core Backup
- ✅ **Full & Incremental Backups** - Efficient backup strategies
- ✅ **File & Folder Backup** - Selective or complete directory backup
- ✅ **AES-256 Encryption** - Military-grade data protection
- ✅ **GZip Compression** - Reduce storage requirements
- ✅ **Multi-threaded Processing** - Fast backup operations
- ✅ **Large File Support** - Handle files of any size

### Storage Options
- 📁 **Local Storage** - Direct disk backup
- 🌐 **SMB/CIFS Shares** - Network attached storage
- ☁️ **S3 Compatible** - AWS S3, MinIO, Wasabi

### Advanced Features
- 🔄 **Block-level Deduplication** - Save storage space
- 📦 **Repository Replication** - Off-site backup copies
- 🖥️ **Hyper-V VM Backup** - Virtual machine protection with VSS snapshots
- 📅 **Job Scheduler** - Automated daily/weekly/monthly backups
- 🔔 **Notifications** - Email, Telegram, Webhook alerts

### User Interface
- 🎨 **Modern WPF GUI** - Clean, intuitive interface
- 🌙 **Dark Theme** - Easy on the eyes
- 📊 **Real-time Progress** - Live backup status
- 📈 **Dashboard Statistics** - Storage usage, success rates

---

## 🏗️ Architecture

```
NovaBackup/
├── NovaBackup.Core          # Core models, interfaces, enums
├── NovaBackup.Common        # Notification services
├── NovaBackup.Engine        # Backup/Restore engine
├── NovaBackup.Storage       # Storage providers (Local, SMB, S3)
├── NovaBackup.Scheduler     # Job scheduling service
├── NovaBackup.Agent         # Background agent service
├── NovaBackup.Deduplication # Block-level deduplication
├── NovaBackup.Replication   # Repository replication
├── NovaBackup.Virtualization# Hyper-V VM backup
├── NovaBackup.Cloud         # Cloud storage (S3)
├── NovaBackup.GUI           # WPF user interface
└── installer                # WiX MSI installer
```

### Data Flow

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Source     │────▶│  Backup      │────▶│  Storage    │
│  Files/VM   │     │  Engine      │     │  Provider   │
└─────────────┘     └──────────────┘     └─────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │  Metadata    │
                    │  (SQLite)    │
                    └──────────────┘
```

---

## 📦 Requirements

### Minimum Requirements
- **OS:** Windows 10/11 or Windows Server 2019+
- **.NET:** .NET 10 Runtime
- **RAM:** 4 GB minimum (8 GB recommended)
- **Disk:** 500 MB for installation + backup storage

### Development Requirements
- **Visual Studio 2022** or **VS Code**
- **.NET 10 SDK**
- **WiX Toolset 3.14+** (for building installer)

---

## 📥 Installation

### Option 1: MSI Installer (Recommended)

1. **Download the latest release:**
   ```powershell
   # Visit https://github.com/ajjs1ajjs/Backup/releases
   ```

2. **Run the installer:**
   ```
   NovaBackup-1.0.0.msi
   ```

3. **Follow the installation wizard:**
   - Accept license agreement
   - Choose installation directory
   - Configure default backup repository
   - Create shortcuts

### Option 2: Manual Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/ajjs1ajjs/Backup.git
   cd Backup
   ```

2. **Build from source:**
   ```bash
   dotnet build NovaBackup.sln -c Release
   ```

3. **Publish:**
   ```bash
   dotnet publish NovaBackup.sln -c Release -o ./publish
   ```

4. **Run the application:**
   ```bash
   ./publish/NovaBackup.GUI/NovaBackup.GUI.exe
   ```

### Option 3: Dotnet Global Tool

```bash
dotnet tool install -g NovaBackup.CLI
novabackup --version
```

---

## 🚀 Quick Start

### 1. Build and Run

```bash
# Build the solution
dotnet build NovaBackup.sln

# Run the GUI application
dotnet run --project NovaBackup.GUI/NovaBackup.GUI.csproj
```

### 2. Create Your First Backup Job

**Using the Ribbon (Home tab):**
1. Click **Home** → **Backup Job** (💾)
2. Enter job name (e.g., "Daily Backup")
3. Add files or folders to backup
4. Select destination repository
5. Configure schedule (Daily/Weekly)
6. Click **Save**

**Using Navigation Tree:**
1. Expand **Backup Jobs** → Click **Active Jobs**
2. Click **New Job** in Quick Access Toolbar
3. Fill in the job configuration dialog

### 3. Run a Backup

1. Select the job in the data grid
2. Click **Jobs** tab → **Run** button (▶)
3. Monitor progress on Dashboard

### 4. View Results

- **Dashboard** - Statistics and recent sessions
- **Monitoring** → **Sessions** - Full session history
- **Properties Panel** (right) - Details of selected object

---

## 📖 Usage Guide

### Creating Backup Jobs

#### Via Ribbon
```
Home tab → Backup Job
or
Jobs tab → New Job
```

#### Via CLI (when available)
```bash
novabackup job create \
  --name "Daily Backup" \
  --source "C:\Data" \
  --destination "D:\Backups" \
  --schedule "daily 02:00" \
  --encryption "AES256"
```

### Backup Types

| Type | Description | Use Case |
|------|-------------|----------|
| **Full** | Complete backup of all files | First backup, monthly |
| **Incremental** | Only changed files since last backup | Daily backups |
| **Differential** | Changes since last full backup | Weekly backups |

### Scheduling Options

| Schedule | Configuration |
|----------|---------------|
| **Daily** | Set time (e.g., 02:00) |
| **Weekly** | Select days + time |
| **Monthly** | Day of month + time |
| **Manual** | Run on demand only |

### Restore Operations

1. **Navigate to Restore tab**
2. **Select restore point** from the list
3. **Choose files** to restore
4. **Specify destination** path
5. **Click Restore**

#### Restore Types
- **Single File** - Restore individual files
- **Folder** - Restore entire folders
- **Full Backup** - Restore complete backup set
- **Point-in-Time** - Restore to specific date

### Managing Repositories

1. **Go to Repositories**
2. **Add Repository:**
   - Local: `D:\Backups`
   - SMB: `\\server\share`
   - S3: Configure bucket credentials
3. **Set as Default** for new jobs

---

## ⚙️ Configuration

### Configuration Files

#### appsettings.json
```json
{
  "Backup": {
    "DefaultRepository": "D:\\Backups",
    "MaxConcurrentJobs": 4,
    "CompressionLevel": "Normal",
    "EncryptionLevel": "AES256"
  },
  "Scheduler": {
    "Enabled": true,
    "CheckIntervalMinutes": 1
  },
  "Notifications": {
    "Email": {
      "Enabled": false,
      "SmtpServer": "smtp.example.com",
      "SmtpPort": 587,
      "From": "backup@example.com",
      "To": "admin@example.com"
    },
    "Telegram": {
      "Enabled": false,
      "BotToken": "",
      "ChatId": ""
    }
  }
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NOVABACKUP_REPO` | Default repository path | `C:\ProgramData\NovaBackup\Backups` |
| `NOVABACKUP_DB` | Database location | `%APPDATA%\NovaBackup` |
| `NOVABACKUP_LOG` | Log file path | `%PROGRAMDATA%\NovaBackup\Logs` |

### Database Schema

NovaBackup uses SQLite for metadata storage:

- **repositories.db** - Repository configurations
- **dedup.db** - Deduplication block index
- **jobs.db** - Job definitions and history

---

## 🔌 API Reference

### Core Interfaces

```csharp
// Backup Engine
public interface IBackupEngine
{
    Task<BackupSession> ExecuteBackupAsync(BackupJob job, CancellationToken ct);
    Task<bool> ValidateBackupAsync(Guid sessionId);
    IAsyncEnumerable<BackupFileInfo> ScanFilesAsync(BackupJob job, CancellationToken ct);
}

// Restore Engine
public interface IRestoreEngine
{
    Task RestoreFilesAsync(RestoreRequest request, CancellationToken ct);
    Task<BackupMetadata> LoadMetadataAsync(Guid sessionId, CancellationToken ct);
    IAsyncEnumerable<string> GetRestorePointsAsync(Guid jobId, CancellationToken ct);
}

// Storage Provider
public interface IStorageProvider
{
    StorageType Type { get; }
    Task<Stream> OpenWriteAsync(string path, CancellationToken ct);
    Task<Stream> OpenReadAsync(string path, CancellationToken ct);
    Task<bool> ExistsAsync(string path, CancellationToken ct);
    Task DeleteAsync(string path, CancellationToken ct);
    Task<long> GetSizeAsync(string path, CancellationToken ct);
    Task<IEnumerable<string>> ListFilesAsync(string path, CancellationToken ct);
}
```

### Example: Create Backup Job Programmatically

```csharp
using NovaBackup.Core.Models;
using NovaBackup.Core.Enums;

var job = new BackupJob
{
    Name = "My Backup Job",
    Description = "Daily backup of important files",
    BackupType = BackupType.Incremental,
    ScheduleType = ScheduleType.Daily,
    SourcePaths = new List<string> { "C:\\Data" },
    DestinationRepositoryId = repositoryId,
    CompressionLevel = CompressionLevel.Normal,
    EncryptionLevel = EncryptionLevel.AES256,
    EncryptionPassword = "SecurePassword123",
    ScheduleSettings = new ScheduleSettings
    {
        RunTime = TimeSpan.FromHours(2),
        RetryCount = 3,
        RetryDelayMinutes = 5
    }
};
```

---

## 👨‍💻 Development

### Project Structure

```
Backup/
├── *.sln                 # Solution file
├── Directory.Build.props # Global build settings
├── README.md             # This file
├── NovaBackup.Core/      # Core library
├── NovaBackup.Engine/    # Backup engine
├── NovaBackup.GUI/       # WPF application
└── installer/            # WiX installer
```

### Build Commands

```bash
# Restore dependencies
dotnet restore

# Build Debug
dotnet build -c Debug

# Build Release
dotnet build -c Release

# Run tests
dotnet test

# Publish
dotnet publish -c Release -o ./publish

# Build installer
.\installer\BuildInstaller.ps1 -Version 1.0.0.0
```

### Adding New Storage Provider

1. **Create new class implementing `IStorageProvider`:**

```csharp
public class AzureBlobStorageProvider : IStorageProvider
{
    public StorageType Type => StorageType.AzureBlob;
    
    public Task<Stream> OpenReadAsync(string path, CancellationToken ct)
    {
        // Implementation
    }
    
    // ... other methods
}
```

2. **Register in DI container:**

```csharp
services.AddSingleton<IStorageProvider, AzureBlobStorageProvider>();
```

### Code Style

- **C# Language:** Latest version
- **Nullable Reference Types:** Enabled
- **Indentation:** 4 spaces
- **Braces:** Required for all blocks

---

## 🐛 Troubleshooting

### Common Issues

#### "Access Denied" During Backup

**Solution:**
1. Run as Administrator
2. Check file permissions
3. Add service account to backup operators group

```powershell
# Add user to Backup Operators
net localgroup "Backup Operators" /add username
```

#### SMB Share Connection Fails

**Solution:**
```powershell
# Test connection
Test-Path \\server\share

# Map network drive
net use Z: \\server\share /user:username password
```

#### Deduplication Database Locked

**Solution:**
1. Stop NovaBackup service
2. Delete `dedup.db-shm` and `dedup.db-wal` files
3. Restart service

#### High Memory Usage

**Solution:**
1. Reduce concurrent jobs in settings
2. Enable streaming mode for large files
3. Increase GC pressure threshold

### Log Files

Location: `%PROGRAMDATA%\NovaBackup\Logs\`

- **NovaBackup.log** - Main application log
- **Backup.log** - Backup operations
- **Scheduler.log** - Job scheduling
- **Error.log** - Error details

### Enable Debug Logging

```json
// appsettings.json
{
  "Logging": {
    "LogLevel": {
      "Default": "Debug",
      "NovaBackup": "Trace"
    }
  }
}
```

---

## 🤝 Contributing

We welcome contributions! Here's how to help:

### How to Contribute

1. **Fork the repository**
2. **Create a feature branch:**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**
4. **Run tests:**
   ```bash
   dotnet test
   ```
5. **Commit changes:**
   ```bash
   git commit -m "Add amazing feature"
   ```
6. **Push to branch:**
   ```bash
   git push origin feature/amazing-feature
   ```
7. **Open a Pull Request**

### Code Review Process

1. PR reviewed by maintainer
2. Automated tests must pass
3. Code style checked
4. Documentation updated

### Reporting Issues

- Use GitHub Issues
- Include steps to reproduce
- Attach log files
- Specify version and OS

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

```
Copyright (c) 2024 NovaBackup Team

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
```

---

## 📞 Support

- **Documentation:** https://github.com/ajjs1ajjs/Backup/wiki
- **Issues:** https://github.com/ajjs1ajjs/Backup/issues
- **Discussions:** https://github.com/ajjs1ajjs/Backup/discussions
- **Email:** support@novabackup.local

---

## 🙏 Acknowledgments

- Inspired by [Veeam Backup & Replication](https://www.veeam.com)
- Built with [.NET](https://dotnet.microsoft.com)
- GUI powered by [WPF](https://docs.microsoft.com/dotnet/desktop/wpf)
- Installer by [WiX Toolset](https://wixtoolset.org)

---

<div align="center">

**Made with ❤️ by the NovaBackup Team**

[⬆ Back to Top](#-novabackup---modern-backup-system-for-windows)

</div>
