# NovaBackup Enterprise v7.0

Production-ready backup & recovery platform for Windows Server.

[![License: Enterprise](https://img.shields.io/badge/License-Enterprise-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Windows%20Server%202019%2B-lightgrey)](https://microsoft.com)
[![Ukraine](https://img.shields.io/badge/Made%20in-%F0%9F%87%BA%F0%9F%87%A6-blue)](https://ukraine.ua)

---

## Highlights
- File/Folder, DB, Hyper-V VM backups
- Incremental + deduplication + compression
- Fast restore: files, DB, VM
- Web UI + RBAC + audit logs

---

## Quick Start (Windows, PowerShell as Administrator)

**Install**
```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.bat" -OutFile "install.bat"; .\install.bat
```

**Update**
```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.bat" -OutFile "update.bat"; .\update.bat
```

**Access Web UI**
```
URL: http://localhost:8050
Login: admin
Password: admin123
```

Change the default password after first login.

---

## Documentation
- [README.md](README.md)
- [INSTALL.md](INSTALL.md)
- [ENTERPRISE_DEPLOYMENT.md](ENTERPRISE_DEPLOYMENT.md)
- [Releases](https://github.com/ajjs1ajjs/Backup/releases)
- [Wiki](https://github.com/ajjs1ajjs/Backup/wiki)

---

## Build From Source (Git)
```powershell
git clone https://github.com/ajjs1ajjs/Backup.git
cd Backup
go build -o novabackup.exe .\cmd\novabackup
.\novabackup.exe server
```

**Update from Git**
```powershell
git pull
go build -o novabackup.exe .\cmd\novabackup

# If running as a service
.\novabackup.exe stop
.\novabackup.exe start
```

Note: The Web UI is served from the `web/` folder next to `novabackup.exe`.

---

## System Requirements
| Component | Minimum | Recommended |
|-----------|---------|-------------|
| OS | Windows Server 2019 | Windows Server 2022 |
| CPU | 2 cores | 4+ cores |
| RAM | 4 GB | 8+ GB |
| Disk | 1 GB + backup storage | SSD for database |
| Network | 1 Gbps | 10 Gbps |

---

## Support
Support portal: https://support.novabackup.local

---

## License
Enterprise License - see [LICENSE](LICENSE)
