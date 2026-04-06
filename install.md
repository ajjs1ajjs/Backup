# Backup System — Installation Guide

## Quick Install

### Windows (PowerShell — Administrator)

```powershell
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install-server.ps1 | iex -AutoStart
```

### Linux (bash — root)

```bash
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash -s -- --auto-start
```

---

## What Gets Installed

| Component | Windows | Linux |
|-----------|---------|-------|
| .NET SDK 8.0 | ✅ (if missing) | ✅ (if missing) |
| Node.js 18 | ✅ (if missing) | ✅ (if missing) |
| Git | ✅ (if missing) | ✅ (if missing) |
| Nginx | ❌ | ✅ (if missing) |
| PostgreSQL | ❌ (uses SQLite) | ❌ (uses SQLite) |
| Backup Server (port 8000) | ✅ Windows Service | ✅ systemd service |
| Backup UI (port 80) | ✅ (wwwroot) | ✅ (Nginx) |

---

## Detailed Installation

### Windows

```powershell
# Download script
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install-server.ps1 -OutFile install-server.ps1

# Run with defaults
.\install-server.ps1 -AutoStart

# Custom options
.\install-server.ps1 -InstallDir "D:\Backup" -Port 9000 -AdminPassword "MyPass123" -AutoStart

# Use local source (skip download)
.\install-server.ps1 -LocalSource "C:\Projects\Backup\src\server\Backup.Server" -AutoStart

# Uninstall
.\install-server.ps1 -Uninstall
```

### Linux

```bash
# Download script
curl -fsSL -o install.sh https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh
sudo chmod +x install.sh

# Run with defaults
sudo ./install.sh --auto-start

# Custom JWT key
sudo ./install.sh --jwt-key "my-secret-key" --auto-start

# Show help
./install.sh --help
```

---

## Access After Installation

| Service | URL |
|---------|-----|
| UI | http://localhost |
| API | http://localhost:8000 |
| Swagger | http://localhost:8000/swagger |

**Login:** `admin` / `admin123` (change on first login)

---

## Service Management

### Windows

```powershell
# Status
Get-Service -Name BackupServer

# Start / Stop / Restart
Start-Service -Name BackupServer
Stop-Service -Name BackupServer
Restart-Service -Name BackupServer

# Logs (daily rolling files)
Get-Content "C:\BackupServer\publish\logs\backup-server-$(Get-Date -Format 'yyyyMMdd').log" -Tail 50
```

### Linux

```bash
# Status
sudo systemctl status backup-server

# Start / Stop / Restart
sudo systemctl start backup-server
sudo systemctl stop backup-server
sudo systemctl restart backup-server

# Logs
sudo journalctl -u backup-server -f
```

---

## Uninstall

### Windows

```powershell
.\install-server.ps1 -Uninstall
```

### Linux

```bash
sudo systemctl stop backup-server
sudo systemctl disable backup-server
sudo rm -rf /opt/backup
sudo rm -f /etc/systemd/system/backup-server.service
sudo systemctl daemon-reload
```
