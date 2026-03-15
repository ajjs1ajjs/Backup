# 🏢 NovaBackup Enterprise v7.0 - Deployment Guide

## 📦 Enterprise Installation Options

### Option 1: Interactive Installation (Recommended for Single Server)

```powershell
# Download installer
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/enterprise-install.bat" -OutFile "$HOME\enterprise-install.bat"

# Run as Administrator
.\enterprise-install.bat
```

---

### Option 2: Silent Installation (For SCCM/Intune)

```powershell
# Silent install with default settings
.\enterprise-install.bat /silent

# Silent install with custom parameters
.\enterprise-install.bat /silent /installDir="D:\NovaBackup" /port=8080
```

**Silent Installation Parameters:**
- `/silent` - No user interaction
- `/installDir=PATH` - Custom installation directory
- `/port=PORT` - Custom web UI port
- `/noFirewall` - Skip firewall configuration
- `/noShortcut` - Don't create desktop shortcut

---

### Option 3: Group Policy Deployment (Active Directory)

#### Prerequisites:
- Windows Server with Group Policy Management
- Shared network location for installer
- Computer or User GPO permissions

#### Steps:

1. **Copy installer to network share:**
   ```
   \\domain.local\software\NovaBackup\enterprise-install.bat
   ```

2. **Create GPO:**
   - Open Group Policy Management
   - Create new GPO: "NovaBackup Deployment"
   - Edit → Computer Configuration → Policies → Windows Settings → Scripts (Startup/Shutdown)
   - Add startup script: `\\domain.local\software\NovaBackup\enterprise-install.bat /silent`

3. **Link GPO to OU:**
   - Link to target Organizational Unit
   - Force update: `gpupdate /force`

---

### Option 4: Microsoft SCCM/MECM Deployment

#### Application Model:

**Detection Rule:**
```
Path: C:\Program Files\NovaBackup\NovaBackup.exe
Setting: File or folder
Condition: File or folder exists
```

**Installation Program:**
```
cmd.exe /c "enterprise-install.bat /silent"
```

**Uninstall String:**
```
cmd.exe /c "C:\Program Files\NovaBackup\NovaBackup.exe remove"
```

---

### Option 5: Microsoft Intune (Cloud Deployment)

#### Win32 App Package:

1. **Create IntuneWin package:**
   ```powershell
   # Download Microsoft Win32 Content Prep Tool
   # Run: IntuneWinAppUtil.exe
   # Source folder: Path to enterprise-install.bat
   # Setup file: enterprise-install.bat
   # Output folder: For .intunewin file
   ```

2. **Add to Intune:**
   - Apps → All apps → Add → App type: Windows app (Win32)
   - Upload .intunewin file
   - Configure detection rules
   - Assign to groups

---

## 🔧 Configuration Management

### Centralized Configuration

**File:** `C:\ProgramData\NovaBackup\Config\config.json`

```json
{
  "server": {
    "ip": "0.0.0.0",
    "port": 8050,
    "https": true,
    "https_port": 8443
  },
  "backup": {
    "default_path": "D:\\Backups",
    "retention_days": 90,
    "compression": true,
    "encryption": true
  },
  "logging": {
    "level": "info",
    "file": "C:\\ProgramData\\NovaBackup\\Logs\\novabackup.log"
  },
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_server": "smtp.company.com",
      "smtp_port": 587,
      "from": "novabackup@company.com",
      "to": "it-admins@company.com"
    },
    "teams": {
      "enabled": true,
      "webhook_url": "https://outlook.office.com/webhook/..."
    }
  }
}
```

### Deploy Configuration via GPO:

1. **Create configuration file** on network share
2. **GPO:** Computer Configuration → Preferences → Windows Settings → Files
3. **Copy** `\\share\config.json` → `C:\ProgramData\NovaBackup\Config\config.json`

---

## 🔐 Security Hardening

### Service Account (Recommended for Production)

```powershell
# Create dedicated service account
net user /add novabackup_svc "StrongPassword123!" /passwordreq:yes /passwordchg:no

# Grant logon as service right
secedit /export /cfg secpol.cfg
# Edit secpol.cfg: add novabackup_svc to SeServiceLogonRight
secedit /configure /db secedit.sdb /cfg secpol.cfg

# Set service to run as dedicated account
sc config NovaBackup obj= ".\novabackup_svc" password= "StrongPassword123!"
```

### Firewall Rules (Enterprise)

```powershell
# Restrict to management subnet only
netsh advfirewall firewall add rule name="NovaBackup Web UI" dir=in action=allow protocol=TCP localport=8050 remoteip=10.0.0.0/24

# Block external access
netsh advfirewall firewall add rule name="NovaBackup Block External" dir=in action=block protocol=TCP localport=8050 remoteip=any
```

### SSL/TLS Configuration

```json
{
  "server": {
    "https": true,
    "https_port": 8443,
    "ssl_cert_path": "C:\\ProgramData\\NovaBackup\\Config\\server.crt",
    "ssl_key_path": "C:\\ProgramData\\NovaBackup\\Config\\server.key"
  }
}
```

---

## 📊 Monitoring & Alerting

### Windows Event Log Integration

NovaBackup logs to Windows Event Log automatically.

**View logs:**
```powershell
Get-EventLog -LogName Application -Source NovaBackup -Newest 50
```

### SCOM Management Pack

Import NovaBackup Management Pack for System Center Operations Manager:
```
NovaBackup.Enterprise.MP.msi
```

### Health Checks

**API Health Endpoint:**
```
GET http://localhost:8050/api/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "7.0.0",
  "time": "2024-03-15T20:00:00Z"
}
```

---

## 🔄 Update Management

### WSUS/SCCM Update Deployment

1. **Download update package:**
   ```
   \\share\software\NovaBackup\NovaBackup-7.0.0-Update.zip
   ```

2. **Create update deployment:**
   - Application: NovaBackup Update
   - Detection: Version check in config.json
   - Deadline: Set maintenance window

3. **Deploy with PowerShell:**
   ```powershell
   Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/update.bat" -OutFile "$HOME\update.bat"
   .\update.bat /silent
   ```

---

## 📁 File Locations

| Component | Path |
|-----------|------|
| Executable | `C:\Program Files\NovaBackup\NovaBackup.exe` |
| Configuration | `C:\ProgramData\NovaBackup\Config\config.json` |
| Database | `C:\ProgramData\NovaBackup\novabackup.db` |
| Logs | `C:\ProgramData\NovaBackup\Logs\` |
| Backups | `C:\ProgramData\NovaBackup\Backups\` |
| Service | `NovaBackup` (Windows Service) |

---

## 🛠️ Troubleshooting

### Service Won't Start

```powershell
# Check service status
Get-Service NovaBackup

# View service logs
Get-Content "C:\ProgramData\NovaBackup\Logs\novabackup.log" -Tail 100

# Check Windows Event Log
Get-EventLog -LogName Application -Source NovaBackup -Newest 20

# Restart service
Restart-Service NovaBackup -Force
```

### Port Already in Use

```powershell
# Find process using port 8050
Get-NetTCPConnection -LocalPort 8050 | Select-Object OwningProcess

# Kill process
Stop-Process -Id <PID> -Force

# Change port in config
# Edit: C:\ProgramData\NovaBackup\Config\config.json
# Set: "port": 8051
```

### Backup Jobs Failing

1. **Check logs:**
   ```
   C:\ProgramData\NovaBackup\Logs\novabackup.log
   ```

2. **Verify storage permissions:**
   ```powershell
   icacls "D:\Backups" /grant "NT SERVICE\NovaBackup:(OI)(CI)F"
   ```

3. **Test storage connectivity:**
   ```powershell
   Test-Path "D:\Backups"
   ```

---

## 📞 Enterprise Support

| Level | Contact | Response Time |
|-------|---------|---------------|
| L1 Support | support@novabackup.local | 24 hours |
| L2 Support | enterprise@novabackup.local | 4 hours |
| L3 Support | +380 44 123 4567 | 1 hour |
| Critical | critical@novabackup.local | 15 minutes |

**Support Portal:** https://support.novabackup.local

---

## 📄 License

Enterprise License - See LICENSE.ENTERPRISE file

---

<div align="center">

**NovaBackup Enterprise v7.0**

Production-Ready Backup Solution for Windows Server

[Download](https://github.com/ajjs1ajjs/Backup/releases) • [Documentation](https://github.com/ajjs1ajjs/Backup/wiki) • [Support](https://support.novabackup.local)

🇺🇦 Made in Ukraine

</div>
