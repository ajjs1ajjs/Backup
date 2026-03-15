# 🛡️ NovaBackup Enterprise v7.0

**Production-Ready Backup & Recovery Platform for Windows Server**

[![License: Enterprise](https://img.shields.io/badge/License-Enterprise-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Windows%20Server%202019%2B-lightgrey)](https://microsoft.com)
[![Ukraine](https://img.shields.io/badge/Made%20in-%F0%9F%87%BA%F0%9F%87%A6-blue)](https://ukraine.ua)

---

## 🚀 Quick Start

### Install (Run as Administrator)

```powershell
# Download and run installer
cd $HOME
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/enterprise-install.bat" -OutFile "install.bat"
.\install.bat
```

### One-Line Install / Update

**Windows (PowerShell as Administrator)**
```powershell
# Install
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat" -OutFile "install.bat"; .\install.bat

# Update
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.bat" -OutFile "update.bat"; .\update.bat
```

**Windows (Single Setup Script)**
```powershell
# Download once and run (interactive install/update/remove)
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/novabackup-setup.bat" -OutFile "novabackup-setup.bat"; .\novabackup-setup.bat
```

Note: If you do not use GitHub Releases, keep `novabackup.exe` updated in the repository root.
If the raw download returns a tiny file (for example "Not Found"), update the binary in the repo or use Releases.

**Linux**
```bash
# Install
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash

# Update
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.sh | sudo bash
```

### Access Web UI

```
URL: http://localhost:8050
Login: admin
Password: admin123
```

⚠️ **IMPORTANT:** Change default password after first login!

---

## 📦 Enterprise Features

### ✅ Backup Capabilities
- 📁 **File & Folder Backup** - Incremental with deduplication
- 🗄️ **Database Backup** - MySQL, PostgreSQL, SQLite, MSSQL
- 🖥️ **Hyper-V VM Backup** - With VSS integration
- ☁️ **Cloud Storage** - S3, Azure Blob, Google Cloud

### ✅ Recovery Options
- ♻️ **File-Level Restore** - Granular file recovery
- 🗄️ **Database Restore** - Point-in-time recovery
- 🖥️ **VM Restore** - Full VM recovery
- ⚡ **Instant Recovery** - Run VM from backup

### ✅ Enterprise Security
- 🔐 **AES-256 Encryption** - Military-grade encryption
- 🔒 **Role-Based Access** - 4 predefined roles
- 📝 **Audit Logging** - Complete activity trail
- 🔑 **Password Policy** - Enforce complexity

### ✅ Management
- 🌐 **Web UI** - Modern dark theme interface
- 📊 **Dashboard** - Real-time monitoring
- 📈 **Reports** - Backup success rates, storage usage
- 🔔 **Notifications** - Email, Teams, Slack, Webhook

---

## 🏢 Deployment Options

| Method | Best For | Documentation |
|--------|----------|---------------|
| **Interactive Install** | Single server | [Guide](#quick-start) |
| **Silent Install** | SCCM/Intune | [ENTERPRISE_DEPLOYMENT.md](ENTERPRISE_DEPLOYMENT.md) |
| **Group Policy** | Active Directory | [ENTERPRISE_DEPLOYMENT.md](ENTERPRISE_DEPLOYMENT.md) |
| **SCCM/MECM** | Enterprise | [ENTERPRISE_DEPLOYMENT.md](ENTERPRISE_DEPLOYMENT.md) |
| **Intune** | Cloud-managed | [ENTERPRISE_DEPLOYMENT.md](ENTERPRISE_DEPLOYMENT.md) |

---

## 📋 System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **OS** | Windows Server 2019 | Windows Server 2022 |
| **CPU** | 2 cores | 4+ cores |
| **RAM** | 4 GB | 8+ GB |
| **Disk** | 1 GB + backup storage | SSD for database |
| **Network** | 1 Gbps | 10 Gbps |

---

## 🔧 Configuration

### Default Configuration File

**Location:** `C:\ProgramData\NovaBackup\Config\config.json`

```json
{
  "server": {
    "ip": "0.0.0.0",
    "port": 8050,
    "https": false
  },
  "backup": {
    "default_path": "C:\\ProgramData\\NovaBackup\\Backups",
    "retention_days": 30,
    "compression": true,
    "encryption": false
  },
  "logging": {
    "level": "info",
    "file": "C:\\ProgramData\\NovaBackup\\Logs\\novabackup.log"
  }
}
```

---

## 🔐 Security

### Default Credentials

| User | Password | Role |
|------|----------|------|
| admin | admin123 | Administrator |

⚠️ **CHANGE IMMEDIATELY AFTER FIRST LOGIN!**

### Firewall Rules

Automatically configured during installation:
- **Port 8050** - Web UI (HTTP)
- **Port 8443** - Web UI (HTTPS)

### Service Account

For production, create dedicated service account:
```powershell
net user /add novabackup_svc "StrongPassword123!"
sc config NovaBackup obj= ".\novabackup_svc" password= "StrongPassword123!"
```

---

## 📊 API Reference

### Health Check
```bash
GET http://localhost:8050/api/health
```

### Authentication
```bash
POST http://localhost:8050/api/auth/login
{
  "username": "admin",
  "password": "admin123"
}
```

### Backup Jobs
```bash
# List jobs
GET http://localhost:8050/api/jobs

# Create job
POST http://localhost:8050/api/jobs
{
  "name": "Daily Backup",
  "type": "file",
  "sources": ["C:\\Data"],
  "destination": "D:\\Backups",
  "compression": true,
  "retention_days": 30
}

# Run job
POST http://localhost:8050/api/jobs/{id}/run
```

### Sessions
```bash
# List sessions
GET http://localhost:8050/api/backup/sessions
```

---

## 🛠️ Troubleshooting

### Service Won't Start
```powershell
# Check status
Get-Service NovaBackup

# View logs
Get-Content "C:\ProgramData\NovaBackup\Logs\novabackup.log" -Tail 100

# Restart
Restart-Service NovaBackup -Force
```

### Port Already in Use
```powershell
# Find process
Get-NetTCPConnection -LocalPort 8050

# Change port in config.json
# Edit: "port": 8051
```

### Web UI Not Loading
1. Check service: `Get-Service NovaBackup`
2. Check firewall: `netsh advfirewall firewall show rule name="NovaBackup"`
3. Check logs: `Get-Content "C:\ProgramData\NovaBackup\Logs\novabackup.log"`

---

## 📞 Support

| Level | Contact | Hours |
|-------|---------|-------|
| **L1 Support** | support@novabackup.local | 24/7 |
| **L2 Support** | enterprise@novabackup.local | 8x5 |
| **Critical** | +380 44 123 4567 | 24/7 |

**Support Portal:** https://support.novabackup.local

**Documentation:** https://github.com/ajjs1ajjs/Backup/wiki

---

## 📄 License

Enterprise License - See [LICENSE](LICENSE) file

---

## 🙏 Acknowledgments

- Inspired by [Veeam Backup & Replication](https://www.veeam.com)
- Built with [Go](https://go.dev) and [Gin](https://gin-gonic.com)
- UI inspired by modern enterprise dashboards

---

<div align="center">

**NovaBackup Enterprise v7.0**

Production-Ready Backup Solution for Windows Server

[Download](https://github.com/ajjs1ajjs/Backup/releases) • [Documentation](ENTERPRISE_DEPLOYMENT.md) • [Support](https://support.novabackup.local)

🇺🇦 Made in Ukraine

</div>
