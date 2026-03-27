# 🚀 NovaBackup Production Deployment Runbook

## Document Information

| Attribute | Value |
|-----------|-------|
| **Version** | 1.0 |
| **Last Updated** | March 27, 2026 |
| **Owner** | Operations Team |
| **Review Cycle** | Quarterly |
| **Classification** | Internal Use |

---

## Table of Contents

1. [Overview](#1-overview)
2. [Pre-Deployment Requirements](#2-pre-deployment-requirements)
3. [Deployment Scenarios](#3-deployment-scenarios)
4. [Post-Deployment Verification](#4-post-deployment-verification)
5. [Rollback Procedures](#5-rollback-procedures)
6. [Troubleshooting](#6-troubleshooting)
7. [Appendix](#7-appendix)

---

## 1. Overview

### 1.1 Purpose

This runbook provides step-by-step instructions for deploying NovaBackup to production environments. It covers preparation, deployment, verification, and rollback procedures.

### 1.2 Scope

- Standalone Linux deployment
- Standalone Windows deployment
- Docker Compose deployment
- Kubernetes deployment (future)

### 1.3 Definitions

| Term | Definition |
|------|------------|
| **RTO** | Recovery Time Objective - Maximum acceptable downtime |
| **RPO** | Recovery Point Objective - Maximum acceptable data loss |
| **SLA** | Service Level Agreement - Expected uptime guarantee |

---

## 2. Pre-Deployment Requirements

### 2.1 Infrastructure Requirements

#### Minimum Specifications

| Resource | Requirement | Notes |
|----------|-------------|-------|
| **CPU** | 2 cores | 4+ cores recommended |
| **RAM** | 2 GB | 4+ GB recommended |
| **Disk** | 10 GB + backup storage | SSD recommended |
| **Network** | 100 Mbps | 1 Gbps recommended |

#### Operating System Support

| OS | Version | Status |
|----|---------|--------|
| Ubuntu | 20.04, 22.04 | ✅ Supported |
| Debian | 10, 11 | ✅ Supported |
| RHEL | 8, 9 | ✅ Supported |
| CentOS | 8, 9 | ✅ Supported |
| Windows Server | 2019, 2022 | ✅ Supported |
| macOS | 12+ | ⚠️ Development only |

### 2.2 Software Dependencies

#### Required

| Software | Version | Purpose |
|----------|---------|---------|
| Python | 3.9 - 3.12 | Runtime |
| pip | 21.0+ | Package manager |
| Git | Any | Source control |

#### Optional (Database)

| Software | Version | Purpose |
|----------|---------|---------|
| PostgreSQL | 13+ | Production database |
| Docker | 20.10+ | Containerization |
| Docker Compose | 2.0+ | Multi-container orchestration |

### 2.3 Network Requirements

#### Ports

| Port | Protocol | Purpose | Direction |
|------|----------|---------|-----------|
| 8050 | TCP | HTTP API | Inbound |
| 8443 | TCP | HTTPS API | Inbound |
| 8080 | TCP | Web Dashboard | Inbound |
| 5432 | TCP | PostgreSQL | Local only |

#### Firewall Rules

```bash
# Linux (ufw)
sudo ufw allow 8050/tcp
sudo ufw allow 8443/tcp
sudo ufw allow 8080/tcp
sudo ufw enable

# Linux (firewalld)
sudo firewall-cmd --permanent --add-port=8050/tcp
sudo firewall-cmd --permanent --add-port=8443/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# Windows Firewall
netsh advfirewall firewall add rule name="NovaBackup API" dir=in action=allow protocol=TCP localport=8050
netsh advfirewall firewall add rule name="NovaBackup Dashboard" dir=in action=allow protocol=TCP localport=8080
```

### 2.4 Security Prerequisites

#### Secrets Generation

Generate all required secrets BEFORE deployment:

```bash
# Master Key (32+ characters)
openssl rand -hex 32

# JWT Secret (32+ characters)
openssl rand -hex 32

# API Key (16+ characters)
openssl rand -hex 16
```

#### SSL/TLS Certificates

For HTTPS production deployment:

```bash
# Let's Encrypt (Certbot)
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d backup.yourdomain.com

# Or use your organization's CA
# Place certificates at:
# /etc/ssl/certs/novabackup.crt
# /etc/ssl/private/novabackup.key
```

### 2.5 Database Setup

#### PostgreSQL Installation (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install postgresql postgresql-contrib postgresql-client

# Start and enable
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Verify status
sudo systemctl status postgresql
```

#### Database Creation

```bash
sudo -u postgres psql

CREATE DATABASE novabackup;
CREATE USER novabackup WITH PASSWORD 'STRONG_PASSWORD_HERE';
GRANT ALL PRIVILEGES ON DATABASE novabackup TO novabackup;
\q
```

#### Test Connection

```bash
psql -h localhost -U novabackup -d novabackup
```

---

## 3. Deployment Scenarios

### 3.1 Scenario A: Linux Standalone (Systemd)

#### Step 1: Prepare System

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install dependencies
sudo apt-get install -y python3 python3-pip python3-venv postgresql git curl

# Create application user
sudo useradd --system --no-create-home --shell /bin/bash novabackup

# Create directories
sudo mkdir -p /opt/novabackup
sudo mkdir -p /var/lib/novabackup
sudo mkdir -p /var/log/novabackup
sudo mkdir -p /var/backups/novabackup

# Set permissions
sudo chown -R novabackup:novabackup /opt/novabackup
sudo chown -R novabackup:novabackup /var/lib/novabackup
sudo chown -R novabackup:novabackup /var/log/novabackup
sudo chown -R novabackup:novabackup /var/backups/novabackup
```

#### Step 2: Install Application

```bash
# Clone repository (or copy files)
cd /opt/novabackup
sudo git clone https://github.com/ajjs1ajjs/Backup.git .
# OR copy from local source
# sudo cp -r /path/to/backup/* .

# Create virtual environment
sudo -u novabackup python3 -m venv venv

# Activate and install
sudo -u novabackup /opt/novabackup/venv/bin/pip install --upgrade pip
sudo -u novabackup /opt/novabackup/venv/bin/pip install -e ".[api,db]"
```

#### Step 3: Configure Environment

```bash
# Create .env file
sudo -u novabackup tee /opt/novabackup/.env > /dev/null <<EOF
NOVABACKUP_ENV=production
NOVABACKUP_DEBUG=false
NOVABACKUP_DATABASE_URL=postgresql://novabackup:PASSWORD@localhost:5432/novabackup
NOVABACKUP_CLOUD_PROVIDERS=AWS,AZURE,GCP
NOVABACKUP_HOST=0.0.0.0
NOVABACKUP_PORT=8050
NOVABACKUP_HTTPS=false
NOVABACKUP_BACKUP_DIR=/var/backups/novabackup
NOVABACKUP_DATA_DIR=/var/lib/novabackup
NOVABACKUP_LOGS_DIR=/var/log/novabackup
NOVABACKUP_MASTER_KEY=YOUR_MASTER_KEY_HERE
NOVABACKUP_JWT_SECRET=YOUR_JWT_SECRET_HERE
NOVABACKUP_API_KEY=YOUR_API_KEY_HERE
EOF

# Set .env permissions
sudo chmod 600 /opt/novabackup/.env
sudo chown novabackup:novabackup /opt/novabackup/.env
```

#### Step 4: Run Migrations

```bash
sudo -u novabackup /opt/novabackup/venv/bin/python -m novabackup.migrate
```

#### Step 5: Install Systemd Service

```bash
sudo tee /etc/systemd/system/novabackup.service > /dev/null <<EOF
[Unit]
Description=NovaBackup Service
Documentation=https://github.com/ajjs1ajjs/Backup
After=network.target postgresql.service

[Service]
Type=notify
User=novabackup
Group=novabackup
WorkingDirectory=/opt/novabackup
Environment="PATH=/opt/novabackup/venv/bin"
ExecStart=/opt/novabackup/venv/bin/uvicorn novabackup.api:get_app --host 0.0.0.0 --port 8050 --workers 4
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=novabackup

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/novabackup /var/log/novabackup /var/backups/novabackup

# Resource limits
LimitNOFILE=65535
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
sudo systemctl daemon-reload

# Enable and start
sudo systemctl enable novabackup
sudo systemctl start novabackup

# Check status
sudo systemctl status novabackup
```

#### Step 6: Configure Log Rotation

```bash
sudo tee /etc/logrotate.d/novabackup > /dev/null <<EOF
/var/log/novabackup/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0640 novabackup novabackup
    postrotate
        systemctl reload novabackup > /dev/null 2>&1 || true
    endscript
}
EOF
```

---

### 3.2 Scenario B: Windows Standalone (NSSM)

#### Step 1: Prepare System

```powershell
# Run PowerShell as Administrator

# Install Python (if not installed)
winget install Python.Python.3.12

# Install PostgreSQL (if not installed)
winget install PostgreSQL.PostgreSQL.15

# Create directories
New-Item -ItemType Directory -Force -Path "C:\NovaBackup"
New-Item -ItemType Directory -Force -Path "C:\NovaBackup\Data"
New-Item -ItemType Directory -Force -Path "C:\NovaBackup\Logs"
New-Item -ItemType Directory -Force -Path "C:\NovaBackup\Backups"
```

#### Step 2: Install Application

```powershell
# Navigate to directory
Set-Location C:\NovaBackup

# Clone repository
git clone https://github.com/ajjs1ajjs/Backup.git .

# Create virtual environment
python -m venv venv

# Activate and install
.\venv\Scripts\Activate.ps1
python -m pip install --upgrade pip
pip install -e ".[api,db]"
```

#### Step 3: Configure Environment

```powershell
# Create .env file
$envContent = @"
NOVABACKUP_ENV=production
NOVABACKUP_DEBUG=false
NOVABACKUP_DATABASE_URL=postgresql://novabackup:PASSWORD@localhost:5432/novabackup
NOVABACKUP_CLOUD_PROVIDERS=AWS,AZURE,GCP
NOVABACKUP_HOST=0.0.0.0
NOVABACKUP_PORT=8050
NOVABACKUP_BACKUP_DIR=C:\NovaBackup\Backups
NOVABACKUP_DATA_DIR=C:\NovaBackup\Data
NOVABACKUP_LOGS_DIR=C:\NovaBackup\Logs
NOVABACKUP_MASTER_KEY=YOUR_MASTER_KEY_HERE
NOVABACKUP_JWT_SECRET=YOUR_JWT_SECRET_HERE
NOVABACKUP_API_KEY=YOUR_API_KEY_HERE
"@

Set-Content -Path ".env" -Value $envContent
```

#### Step 4: Run Migrations

```powershell
.\venv\Scripts\Activate.ps1
python -m novabackup.migrate
```

#### Step 5: Install as Windows Service (NSSM)

```powershell
# Download NSSM
Invoke-WebRequest -Uri "https://nssm.cc/release/nssm-2.24.zip" -OutFile "$env:TEMP\nssm.zip"
Expand-Archive -Path "$env:TEMP\nssm.zip" -DestinationPath "$env:TEMP\nssm"

# Install service (run as Administrator)
& "$env:TEMP\nssm\nssm-2.24\win64\nssm.exe" install NovaBackup "C:\NovaBackup\venv\Scripts\python.exe" "-m", "uvicorn", "novabackup.api:get_app", "--host", "0.0.0.0", "--port", "8050", "--workers", "4"

# Set service directory
& "$env:TEMP\nssm\nssm-2.24\win64\nssm.exe" set NovaBackup AppDirectory "C:\NovaBackup"

# Set environment variables
& "$env:TEMP\nssm\nssm-2.24\win64\nssm.exe" set NovaBackup AppEnvironmentExtra "PATH=C:\NovaBackup\venv\Scripts;%PATH%"

# Start service
Start-Service NovaBackup

# Check status
Get-Service NovaBackup
```

#### Step 6: Configure Windows Firewall

```powershell
# Add firewall rule
New-NetFirewallRule -DisplayName "NovaBackup API" -Direction Inbound -Protocol TCP -LocalPort 8050 -Action Allow
New-NetFirewallRule -DisplayName "NovaBackup Dashboard" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow
```

---

### 3.3 Scenario C: Docker Compose Production

#### Step 1: Prepare Host

```bash
# Install Docker (Ubuntu/Debian)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
rm get-docker.sh

# Install Docker Compose
sudo apt-get install docker-compose-plugin

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify installation
docker --version
docker compose version
```

#### Step 2: Prepare Configuration

```bash
# Create application directory
mkdir -p ~/novabackup
cd ~/novabackup

# Copy docker-compose-prod.yml
cp /path/to/docker-compose-prod.yml .

# Create .env file
cat > .env <<EOF
# Database
POSTGRES_USER=novabackup
POSTGRES_PASSWORD=STRONG_PASSWORD_HERE
POSTGRES_DB=novabackup

# Application
NOVABACKUP_ENV=production
NOVABACKUP_DEBUG=false
NOVABACKUP_DATABASE_URL=postgresql://novabackup:STRONG_PASSWORD_HERE@db:5432/novabackup
NOVABACKUP_CLOUD_PROVIDERS=AWS,AZURE,GCP
NOVABACKUP_HOST=0.0.0.0
NOVABACKUP_PORT=8000
NOVABACKUP_MASTER_KEY=YOUR_MASTER_KEY_HERE
NOVABACKUP_JWT_SECRET=YOUR_JWT_SECRET_HERE
NOVABACKUP_API_KEY=YOUR_API_KEY_HERE

# Volumes
BACKUP_VOLUME_PATH=$HOME/novabackup/backups
DATA_VOLUME_PATH=$HOME/novabackup/data
LOG_VOLUME_PATH=$HOME/novabackup/logs
EOF

# Create directories
mkdir -p backups data logs
```

#### Step 3: Deploy

```bash
# Build and start
docker compose up -d --build

# Check status
docker compose ps

# View logs
docker compose logs -f api
docker compose logs -f db
```

#### Step 4: Configure Auto-Start

```bash
# Create systemd service for Docker Compose
sudo tee /etc/systemd/system/novabackup-docker.service > /dev/null <<EOF
[Unit]
Description=NovaBackup Docker Compose
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/home/$USER/novabackup
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

# Enable service
sudo systemctl daemon-reload
sudo systemctl enable novabackup-docker
sudo systemctl start novabackup-docker
```

---

## 4. Post-Deployment Verification

### 4.1 Health Checks

```bash
# API Health
curl -s http://localhost:8050/health | jq

# Expected response:
# {"status":"healthy","timestamp":"2026-03-27T..."}

# API Documentation
curl -s http://localhost:8050/docs | head -20

# Metrics endpoint
curl -s http://localhost:8050/metrics | head -20
```

### 4.2 Authentication Test

```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8050/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=alice&password=secret" | jq -r '.access_token')

echo "Token: $TOKEN"

# Test authenticated endpoint
curl -s http://localhost:8050/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq

# Expected: User information
```

### 4.3 Backup Test

```bash
# List VMs
curl -s http://localhost:8050/vms \
  -H "Authorization: Bearer $TOKEN" | jq

# Create backup
curl -s -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"vm_id":"vm-test-001","description":"Test backup"}' | jq

# List backups
curl -s http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 4.4 Database Verification

```bash
# PostgreSQL connection check
psql -h localhost -U novabackup -d novabackup -c "SELECT COUNT(*) FROM users;"

# Check tables
psql -h localhost -U novabackup -d novabackup -c "\dt"
```

### 4.5 Log Verification

```bash
# Check application logs
sudo journalctl -u novabackup -n 50 --no-pager

# Or for Docker
docker compose logs api | tail -50

# Check for errors
grep -i "error" /var/log/novabackup/*.log | tail -20
```

### 4.6 Resource Monitoring

```bash
# Check memory usage
ps aux | grep novabackup | grep -v grep

# Check disk usage
df -h /var/backups/novabackup
df -h /var/lib/novabackup

# Check open files
lsof -i :8050
```

---

## 5. Rollback Procedures

### 5.1 When to Rollback

Rollback should be considered when:

- Critical functionality is broken
- Performance degradation > 50%
- Security vulnerability discovered
- Data corruption detected
- SLA breach imminent

### 5.2 Rollback Steps

#### Linux Standalone

```bash
# Stop service
sudo systemctl stop novabackup

# Backup current version
sudo mv /opt/novabackup /opt/novabackup.failed

# Restore previous version
sudo mv /opt/novabackup.backup /opt/novabackup

# Restore database (if needed)
sudo -u postgres pg_restore -d novabackup /var/backups/novabackup/db_backup.sql

# Start service
sudo systemctl start novabackup

# Verify
sudo systemctl status novabackup
curl http://localhost:8050/health
```

#### Docker Compose

```bash
# Stop current deployment
cd ~/novabackup
docker compose down

# Restore previous docker-compose.yml
cp docker-compose-prod.yml.backup docker-compose-prod.yml

# Restore previous .env
cp .env.backup .env

# Redeploy
docker compose up -d

# Verify
docker compose ps
docker compose logs api
```

### 5.3 Database Rollback

```bash
# Stop application
sudo systemctl stop novabackup

# Drop current database
sudo -u postgres dropdb novabackup

# Recreate database
sudo -u postgres createdb novabackup
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE novabackup TO novabackup;"

# Restore from backup
sudo -u postgres psql novabackup < /var/backups/novabackup/db_backup_20260326.sql

# Restart application
sudo systemctl start novabackup
```

---

## 6. Troubleshooting

### 6.1 Common Issues

#### Service Won't Start

```bash
# Check logs
sudo journalctl -u novabackup -n 100 --no-pager

# Common causes:
# 1. Port already in use
netstat -tulpn | grep 8050

# 2. Database connection failed
psql -h localhost -U novabackup -d novabackup -c "SELECT 1"

# 3. Missing .env file
ls -la /opt/novabackup/.env

# 4. Permission issues
ls -la /var/lib/novabackup
ls -la /var/log/novabackup
```

#### Database Connection Errors

```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Check connection string
cat /opt/novabackup/.env | grep DATABASE_URL

# Test connection
psql -h localhost -U novabackup -d novabackup

# Check PostgreSQL logs
sudo tail -50 /var/log/postgresql/postgresql-*.log
```

#### High Memory Usage

```bash
# Check memory
ps aux | grep novabackup

# Reduce workers
# Edit .env: NOVABACKUP_WORKERS=2

# Restart service
sudo systemctl restart novabackup
```

#### Backup Failures

```bash
# Check disk space
df -h

# Check permissions
ls -la /var/backups/novabackup

# Check logs
tail -100 /var/log/novabackup/*.log | grep -i error

# Test backup manually
curl -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"vm_id":"test","description":"Manual test"}'
```

### 6.2 Emergency Contacts

| Role | Contact | Phone |
|------|---------|-------|
| On-Call Engineer | oncall@company.com | +1-XXX-XXX-XXXX |
| Database Admin | dba@company.com | +1-XXX-XXX-XXXX |
| Security Team | security@company.com | +1-XXX-XXX-XXXX |
| Management | ops-manager@company.com | +1-XXX-XXX-XXXX |

---

## 7. Appendix

### 7.1 Configuration Reference

See `.env.examples.complete` for all available configuration options.

### 7.2 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/token` | POST | Get JWT token |
| `/auth/me` | GET | Current user |
| `/vms` | GET | List VMs |
| `/backups` | GET/POST | List/Create backups |
| `/backups/{id}/restore` | POST | Restore backup |
| `/metrics` | GET | Prometheus metrics |
| `/audit/logs` | GET | Audit logs (admin) |

### 7.3 Default Credentials

| Username | Password | Role |
|----------|----------|------|
| alice | secret | admin |
| bob | secret | user |
| service | service-secret | service |

**⚠️ CHANGE THESE IMMEDIATELY AFTER DEPLOYMENT!**

### 7.4 Useful Commands

```bash
# Service management
sudo systemctl start|stop|restart|status novabackup

# Logs
sudo journalctl -u novabackup -f
sudo tail -f /var/log/novabackup/*.log

# Database backup
sudo -u postgres pg_dump novabackup > backup.sql

# Database restore
sudo -u postgres psql novabackup < backup.sql

# Docker commands
docker compose up -d
docker compose down
docker compose logs -f
docker compose ps
```

### 7.5 Performance Tuning

#### Database (PostgreSQL)

Add to `/etc/postgresql/*/main/postgresql.conf`:

```ini
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 16MB
maintenance_work_mem = 128MB
max_connections = 100
```

#### Application

In `.env`:

```ini
NOVABACKUP_WORKERS=4
NOVABACKUP_DB_POOL_SIZE=10
NOVABACKUP_DB_MAX_OVERFLOW=20
```

---

**END OF RUNBOOK**
