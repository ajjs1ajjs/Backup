# NovaBackup Enterprise - Production Deployment Guide

## Pre-Deployment Checklist

### System Requirements

**Minimum:**
- CPU: 4 cores
- RAM: 8 GB
- Disk: 100 GB + backup storage
- OS: Windows Server 2019+ or Ubuntu 20.04+

**Recommended:**
- CPU: 8+ cores
- RAM: 16-32 GB
- Disk: SSD for database + HDD for backups
- Network: 10 Gbps

### Network Requirements

| Port | Protocol | Purpose | Direction |
|------|----------|---------|-----------|
| 8050 | TCP | HTTP API | Inbound |
| 8443 | TCP | HTTPS API (optional) | Inbound |
| 445 | TCP | SMB/CIFS (if using NAS) | Outbound |
| 3306 | TCP | MySQL (if backing up) | Outbound |
| 5432 | TCP | PostgreSQL (if backing up) | Outbound |

### Firewall Configuration

**Windows (PowerShell):**
```powershell
# Allow HTTP
New-NetFirewallRule -DisplayName "NovaBackup HTTP" `
  -Direction Inbound -Protocol TCP -LocalPort 8050 -Action Allow

# Allow HTTPS (if enabled)
New-NetFirewallRule -DisplayName "NovaBackup HTTPS" `
  -Direction Inbound -Protocol TCP -LocalPort 8443 -Action Allow
```

**Linux (UFW):**
```bash
ufw allow 8050/tcp
ufw allow 8443/tcp  # If using HTTPS
ufw enable
```

---

## Installation

### Windows (Production)

**1. Download Installer:**
```powershell
Invoke-WebRequest -Uri "https://github.com/ajjs1ajjs/Backup/releases/latest/download/install.bat" `
  -OutFile "install.bat"
```

**2. Run as Administrator:**
```powershell
.\install.bat
```

**3. Set Master Key:**
```powershell
# Generate strong master key
$masterKey = -join ((65..90) + (97..122) + (48..57) + (33..47) | Get-Random -Count 32 | ForEach-Object {[char]$_})

# Set system-wide environment variable
setx /M NOVABACKUP_MASTER_KEY "$masterKey"

# Restart service
Restart-Service NovaBackup
```

**4. Verify Installation:**
```powershell
# Check service status
Get-Service NovaBackup

# Check logs
Get-EventLog -LogName Application -Source NovaBackup -Newest 10

# Test API
Invoke-RestMethod http://localhost:8050/api/health
```

### Linux (Production)

**1. Download Installer:**
```bash
curl -fsSL https://github.com/ajjs1ajjs/Backup/releases/latest/download/install.sh -o install.sh
chmod +x install.sh
```

**2. Run as Root:**
```bash
sudo ./install.sh
```

**3. Set Master Key:**
```bash
# Generate strong master key
MASTER_KEY=$(openssl rand -base64 32)

# Create environment file
echo "NOVABACKUP_MASTER_KEY=$MASTER_KEY" | sudo tee /etc/novabackup.env
sudo chmod 600 /etc/novabackup.env

# Update systemd service
sudo systemctl daemon-reload
sudo systemctl restart novabackup
```

**4. Verify Installation:**
```bash
# Check service status
systemctl status novabackup

# Check logs
journalctl -u novabackup -n 20

# Test API
curl http://localhost:8050/api/health
```

---

## Post-Installation Configuration

### 1. Change Default Password

**Web UI:**
1. Navigate to `http://localhost:8050`
2. Login with `admin` / `admin123`
3. Go to Settings → Change Password
4. Set strong password (12+ chars, special chars)

**API:**
```bash
curl -X POST http://localhost:8050/api/auth/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "admin123",
    "new_password": "NewSecureP@ss123!"
  }'
```

### 2. Configure Backup Storage

**Local Storage:**
```json
{
  "directories": {
    "backup_dir": "D:\\Backups",
    "data_dir": "C:\\ProgramData\\NovaBackup\\data",
    "logs_dir": "C:\\ProgramData\\NovaBackup\\logs"
  }
}
```

**Network Storage (SMB):**
```json
{
  "storage": {
    "type": "smb",
    "path": "\\\\nas\\backups",
    "credentials": {
      "username": "backup_user",
      "password": "secure_password"
    }
  }
}
```

**Cloud Storage (S3):**
```json
{
  "storage": {
    "type": "s3",
    "bucket": "company-backups",
    "region": "eu-west-1",
    "credentials": {
      "access_key": "AKIAXXXXXXXXXXXXXXXX",
      "secret_key": "your_secret_key"
    },
    "encryption": true
  }
}
```

### 3. Create First Backup Job

**Via Web UI:**
1. Go to Jobs → Create Job
2. Configure:
   - Name: Daily Files Backup
   - Type: File
   - Sources: `C:\Data`, `D:\Documents`
   - Destination: `E:\Backups`
   - Schedule: Daily at 02:00
   - Retention: 30 days
   - Encryption: Enabled

**Via API:**
```bash
curl -X POST http://localhost:8050/api/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Files Backup",
    "type": "file",
    "sources": ["C:\\Data", "D:\\Documents"],
    "destination": "E:\\Backups",
    "compression": true,
    "compression_level": 5,
    "encryption": true,
    "schedule": "daily",
    "schedule_time": "02:00",
    "retention_days": 30,
    "enabled": true
  }'
```

### 4. Configure Notifications

**Email:**
```json
{
  "notifications": {
    "email": {
      "smtp_server": "smtp.example.com",
      "smtp_port": 587,
      "username": "novabackup@example.com",
      "password": "smtp_password",
      "from": "novabackup@example.com",
      "to": ["admin@example.com", "backup-team@example.com"],
      "events": {
        "backup_success": true,
        "backup_failure": true,
        "storage_warning": true
      }
    }
  }
}
```

**Slack:**
```json
{
  "notifications": {
    "slack": {
      "webhook_url": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
      "channel": "#backup-alerts",
      "events": {
        "backup_failure": true,
        "storage_critical": true
      }
    }
  }
}
```

---

## High Availability Setup

### Active-Passive Cluster

**Primary Server:**
```json
{
  "cluster": {
    "mode": "active-passive",
    "node_id": "primary",
    "peer_address": "192.168.1.11:8050"
  }
}
```

**Secondary Server:**
```json
{
  "cluster": {
    "mode": "active-passive",
    "node_id": "secondary",
    "peer_address": "192.168.1.10:8050"
  }
}
```

### Load Balancer Configuration

**NGINX:**
```nginx
upstream novabackup {
    server 192.168.1.10:8050;
    server 192.168.1.11:8050;
}

server {
    listen 80;
    server_name backup.example.com;

    location / {
        proxy_pass http://novabackup;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

---

## Performance Tuning

### Database Optimization

**SQLite:**
```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -64000;  -- 64MB cache
PRAGMA temp_store = MEMORY;
```

### Backup Performance

**Parallel Threads:**
```json
{
  "max_threads": 8,
  "block_size": 1048576,  -- 1MB blocks
  "compression_level": 5
}
```

**Deduplication:**
```json
{
  "deduplication": true,
  "block_size": 4194304  -- 4MB blocks for dedup
}
```

### Storage Optimization

**SSD for Database:**
```
Database: /dev/nvme0n1 (SSD)
Backups: /dev/sda1 (HDD)
```

**RAID Configuration:**
- Database: RAID 1 (mirroring)
- Backups: RAID 6 (dual parity)

---

## Backup Strategy

### 3-2-1 Rule

- **3** copies of data
- **2** different media types
- **1** offsite copy

### GFS Retention

**Grandfather-Father-Son:**
```json
{
  "retention": {
    "gfs_daily": 7,      -- Keep 7 daily backups
    "gfs_weekly": 4,     -- Keep 4 weekly backups
    "gfs_monthly": 12,   -- Keep 12 monthly backups
    "gfs_yearly": 7      -- Keep 7 yearly backups
  }
}
```

### Backup Schedule

| Job Type | Frequency | Time | Retention |
|----------|-----------|------|-----------|
| Critical Files | Every 4 hours | 00:00, 04:00, ... | 7 days |
| Full System | Daily | 02:00 | 30 days |
| Database | Every hour | :00 | 7 days |
| Archive | Weekly | Sunday 03:00 | 1 year |

---

## Disaster Recovery

### Recovery Procedures

**1. Service Failure:**
```bash
# Windows
Restart-Service NovaBackup

# Linux
systemctl restart novabackup
```

**2. Database Corruption:**
```bash
# Stop service
systemctl stop novabackup

# Backup corrupted DB
mv novabackup.db novabackup.db.corrupted

# Restore from backup
cp /backups/novabackup.db.backup novabackup.db

# Start service
systemctl start novabackup
```

**3. Full System Recovery:**
```bash
# 1. Install NovaBackup on new server
# 2. Restore configuration
cp backup_config.json /etc/novabackup/

# 3. Restore database
novabackup restore database --from=/backups/db_backup.sql

# 4. Restore backup metadata
novabackup restore metadata --from=/backups/metadata

# 5. Verify backups
novabackup verify --all
```

### RTO/RPO Targets

| Scenario | RTO | RPO |
|----------|-----|-----|
| Service Restart | 5 min | 0 |
| Database Recovery | 30 min | 1 hour |
| Full System Recovery | 4 hours | 24 hours |

---

## Monitoring & Alerts

### Health Checks

**Prometheus:**
```yaml
scrape_configs:
  - job_name: 'novabackup'
    static_configs:
      - targets: ['localhost:8050']
    metrics_path: '/api/metrics'
```

### Alert Thresholds

| Metric | Warning | Critical |
|--------|---------|----------|
| Backup Failure Rate | >5% | >10% |
| Storage Usage | >80% | >95% |
| API Latency (p95) | >500ms | >2000ms |
| Failed Logins (15min) | >5 | >10 |

### Notification Channels

Configure in `config.json`:
```json
{
  "notifications": {
    "email": {
      "to": ["admin@example.com"]
    },
    "slack": {
      "webhook": "https://hooks.slack.com/..."
    },
    "pagerduty": {
      "routing_key": "your-pagerduty-key"
    }
  }
}
```

---

## Security Hardening

### Network Segmentation

```
[Internet]
    |
[Firewall]
    |
[DMZ] - Web Server
    |
[Internal Firewall]
    |
[Internal Network] - NovaBackup (port 8050)
    |
[Storage Network] - NAS/SAN
```

### Access Control

**IP Whitelist:**
```json
{
  "security": {
    "allowed_ips": [
      "10.0.0.0/8",
      "192.168.1.0/24"
    ]
  }
}
```

**API Rate Limiting:**
```json
{
  "security": {
    "rate_limit": {
      "requests_per_minute": 100,
      "burst": 20
    }
  }
}
```

---

## Maintenance

### Regular Tasks

**Daily:**
- [ ] Check backup success rate
- [ ] Review failed backups
- [ ] Check storage usage

**Weekly:**
- [ ] Test restore from backup
- [ ] Review audit logs
- [ ] Check for updates

**Monthly:**
- [ ] Full disaster recovery test
- [ ] Security patch review
- [ ] Capacity planning

### Log Rotation

**Windows:**
```powershell
# Scheduled Task to rotate logs weekly
$action = New-ScheduledTaskAction -Execute "novabackup.exe" -Argument "rotate-logs"
$trigger = New-ScheduledTaskTrigger -Weekly -DaysOfWeek Sunday -At 3am
Register-ScheduledTask -TaskName "NovaBackup Log Rotation" -Action $action -Trigger $trigger
```

**Linux:**
```bash
# /etc/logrotate.d/novabackup
/var/lib/novabackup/logs/*.log {
    weekly
    rotate 12
    compress
    delaycompress
    missingok
    notifempty
    create 0640 novabackup novabackup
}
```

---

## Troubleshooting

### Common Issues

**Service Won't Start:**
```bash
# Check logs
journalctl -u novabackup -n 50

# Check port is free
netstat -tlnp | grep 8050

# Check permissions
ls -la /var/lib/novabackup/
```

**Backups Failing:**
```bash
# Check storage space
df -h

# Check network connectivity
ping nas.example.com

# Check credentials
novabackup test-storage --id=storage_id
```

**High CPU Usage:**
```bash
# Check running jobs
curl http://localhost:8050/api/jobs | jq '.[] | select(.running == true)'

# Reduce parallel threads
# Edit config.json: "max_threads": 4
```

---

## Support

**Contact Information:**
- Email: support@example.com
- Phone: +1-800-BACKUP
- Slack: #backup-support

**Escalation:**
1. Level 1: Help Desk (helpdesk@example.com)
2. Level 2: Backup Team (backup-team@example.com)
3. Level 3: Vendor Support (vendor-support@example.com)

---

## Deployment Verification

After deployment, verify:

```bash
# 1. Health check
curl http://localhost:8050/api/health

# 2. Login
curl -X POST http://localhost:8050/api/auth/login \
  -d '{"username":"admin","password":"<new_password>"}'

# 3. List jobs
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8050/api/jobs

# 4. Run test backup
curl -X POST http://localhost:8050/api/jobs/<job_id>/run \
  -H "Authorization: Bearer $TOKEN"

# 5. Check audit logs
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8050/api/audit/logs?limit=5"
```

**Expected Results:**
- ✅ Health check returns 200
- ✅ Login successful
- ✅ Jobs listed
- ✅ Backup started
- ✅ Audit log entry created

---

## Sign-Off

**Deployed By:** ________________  
**Date:** ________________  
**Verified By:** ________________  
**Date:** ________________
