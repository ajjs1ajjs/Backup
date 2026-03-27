# 📋 NovaBackup Deployment Checklist

## Pre-Deployment Verification

### 1. Environment Requirements

- [ ] Python 3.9+ installed (`python --version`)
- [ ] Git installed (`git --version`)
- [ ] Docker 20.10+ (if using Docker deployment)
- [ ] Docker Compose 2.0+ (if using Docker deployment)
- [ ] Minimum 2GB RAM available
- [ ] Minimum 10GB disk space for application + backups

### 2. Source Code & Dependencies

- [ ] Clone repository: `git clone https://github.com/ajjs1ajjs/Backup.git`
- [ ] Navigate to project: `cd Backup`
- [ ] Create virtual environment: `python -m venv venv`
- [ ] Activate virtual environment:
  - Windows: `.\venv\Scripts\Activate.ps1`
  - Linux/macOS: `source venv/bin/activate`
- [ ] Upgrade pip: `pip install --upgrade pip`
- [ ] Install dependencies: `pip install -e ".[api,db,dev]"`

### 3. Configuration Files

- [ ] Copy environment template: `copy .env.example .env` (Windows) or `cp .env.example .env` (Linux)
- [ ] Generate security secrets:
  - Windows: `.\generate-secrets.ps1 -All`
  - Linux: `python3 generate-secrets.py --all`
- [ ] Verify `.env` file exists and contains:
  - [ ] `NOVABACKUP_MASTER_KEY` (32+ characters)
  - [ ] `NOVABACKUP_JWT_SECRET` (32+ characters)
  - [ ] `NOVABACKUP_API_KEY` (16+ characters)
  - [ ] `NOVABACKUP_DATABASE_URL`
  - [ ] `NOVABACKUP_HOST` and `NOVABACKUP_PORT`
  - [ ] `NOVABACKUP_CLOUD_PROVIDERS`

### 4. Security Secrets Verification

```bash
# Verify secret lengths (Python)
python -c "
import os
from dotenv import load_dotenv
load_dotenv()

master_key = os.getenv('NOVABACKUP_MASTER_KEY', '')
jwt_secret = os.getenv('NOVABACKUP_JWT_SECRET', '')
api_key = os.getenv('NOVABACKUP_API_KEY', '')

print(f'Master Key Length: {len(master_key)} (min 32)')
print(f'JWT Secret Length: {len(jwt_secret)} (min 32)')
print(f'API Key Length: {len(api_key)} (min 16)')

assert len(master_key) >= 32, 'Master key too short!'
assert len(jwt_secret) >= 32, 'JWT secret too short!'
assert len(api_key) >= 16, 'API key too short!'
print('✅ All secrets meet minimum length requirements')
"
```

- [ ] All secrets meet minimum length requirements
- [ ] Secrets are NOT committed to git (`.env` in `.gitignore`)

---

## Deployment Scenarios

### Scenario A: Development Deployment

- [ ] Create `.env` with development settings:
  ```ini
  NOVABACKUP_DATABASE_URL=sqlite:///./novabackup.db
  NOVABACKUP_CLOUD_PROVIDERS=MOCK
  NOVABACKUP_DEBUG=true
  NOVABACKUP_HOST=0.0.0.0
  NOVABACKUP_PORT=8000
  ```
- [ ] Run database migration: `python -m novabackup.migrate`
- [ ] Start development server: `python -m uvicorn novabackup.api:get_app --reload --port 8000`
- [ ] Verify API health: `curl http://localhost:8000/health`
- [ ] Access dashboard: `http://localhost:8000/static/index.html`
- [ ] Test login with default credentials (alice/secret)

---

### Scenario B: Production Deployment (Standalone)

#### B1. Linux Production

- [ ] Create production `.env`:
  ```ini
  NOVABACKUP_DATABASE_URL=postgresql://novabackup:password@localhost:5432/novabackup
  NOVABACKUP_CLOUD_PROVIDERS=AWS,AZURE,GCP
  NOVABACKUP_DEBUG=false
  NOVABACKUP_HTTPS=true
  NOVABACKUP_HOST=0.0.0.0
  NOVABACKUP_PORT=8050
  ```
- [ ] Install PostgreSQL: `sudo apt-get install postgresql postgresql-contrib`
- [ ] Create database and user:
  ```sql
  CREATE DATABASE novabackup;
  CREATE USER novabackup WITH PASSWORD 'strong_password';
  GRANT ALL PRIVILEGES ON DATABASE novabackup TO novabackup;
  ```
- [ ] Run database migration: `python -m novabackup.migrate`
- [ ] Install systemd service:
  ```bash
  sudo cp deploy/systemd/novabackup.service /etc/systemd/system/
  sudo systemctl daemon-reload
  sudo systemctl enable novabackup
  ```
- [ ] Start service: `sudo systemctl start novabackup`
- [ ] Verify status: `sudo systemctl status novabackup`
- [ ] Check logs: `sudo journalctl -u novabackup -f`
- [ ] Configure firewall (if needed): `sudo ufw allow 8050/tcp`
- [ ] Set up reverse proxy (nginx/traefik) for HTTPS

#### B2. Windows Production

- [ ] Create production `.env` (same as Linux)
- [ ] Install PostgreSQL for Windows (or use SQLite for small deployments)
- [ ] Create database and user (via pgAdmin or psql)
- [ ] Run database migration: `python -m novabackup.migrate`
- [ ] Run production installer: `.\deploy\windows_production_install.ps1`
- [ ] Install as Windows Service (using NSSM or similar)
- [ ] Configure Windows Firewall: `netsh advfirewall firewall add rule name="NovaBackup" dir=in action=allow protocol=TCP localport=8050`
- [ ] Verify service is running: `Get-Service novabackup`
- [ ] Check event logs for errors

---

### Scenario C: Docker Production Deployment

- [ ] Ensure Docker and Docker Compose are installed
- [ ] Create `.env` with production settings
- [ ] Verify `docker-compose-prod.yml` configuration:
  - [ ] Database credentials are set
  - [ ] Volume mounts are configured
  - [ ] Resource limits are appropriate
- [ ] Build and start services:
  ```bash
  docker-compose -f docker-compose-prod.yml up -d --build
  ```
- [ ] Verify all containers are running:
  ```bash
  docker-compose -f docker-compose-prod.yml ps
  ```
- [ ] Check API logs: `docker-compose -f docker-compose-prod.yml logs api`
- [ ] Check database logs: `docker-compose -f docker-compose-prod.yml logs db`
- [ ] Verify health checks: `curl http://localhost:8000/health`
- [ ] Access dashboard: `http://localhost:8080`
- [ ] Set up auto-start on boot:
  ```bash
  sudo systemctl enable docker
  ```

---

## Post-Deployment Verification

### 1. Health Checks

- [ ] API Health: `curl http://localhost:8000/health`
  - Expected: `{"status": "healthy"}`
- [ ] Database Connection: `curl http://localhost:8000/docs` (Swagger UI should load)
- [ ] Dashboard Access: Open `http://localhost:8000/static/index.html` in browser

### 2. Authentication Tests

- [ ] Obtain token:
  ```bash
  curl -X POST http://localhost:8000/token \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=alice&password=secret"
  ```
- [ ] Verify token works:
  ```bash
  curl -X GET http://localhost:8000/auth/me \
    -H "Authorization: Bearer YOUR_TOKEN"
  ```
- [ ] Test token refresh endpoint
- [ ] Test logout endpoint

### 3. RBAC Verification

- [ ] Login as admin (alice/secret)
  - [ ] Access `/audit/logs` (should work)
  - [ ] Create backup (should work)
  - [ ] Delete backup (should work)
- [ ] Login as user (bob/secret)
  - [ ] Access `/audit/logs` (should fail - 403)
  - [ ] Create backup (should work)
  - [ ] Delete backup (should fail - 403)
- [ ] Login as service (service/service-secret)
  - [ ] Create backup (should work)
  - [ ] Restore backup (should work)

### 4. Backup Operations

- [ ] List VMs: `curl -H "Authorization: Bearer TOKEN" http://localhost:8000/vms`
- [ ] Create backup via API or dashboard
- [ ] Verify backup appears in list: `curl -H "Authorization: Bearer TOKEN" http://localhost:8000/backups`
- [ ] Test restore operation
- [ ] Verify backup files exist in `./data/backups`

### 5. Cloud Provider Tests (if configured)

- [ ] AWS: Test VM listing with real credentials
- [ ] Azure: Test VM listing with real credentials
- [ ] GCP: Test VM listing with real credentials
- [ ] Create cloud backup to each configured provider
- [ ] Restore from cloud backup

### 6. Notification Tests

- [ ] Send test email notification
- [ ] Send test Telegram notification
- [ ] Send test webhook notification
- [ ] Verify notification history: `curl http://localhost:8000/notifications/history`

### 7. Scheduler Tests

- [ ] Create scheduled job via API
- [ ] Verify job appears: `curl http://localhost:8000/scheduler/jobs`
- [ ] Wait for scheduled execution
- [ ] Verify backup was created
- [ ] Test enable/disable job endpoints

### 8. Audit Logging

- [ ] Perform various actions (login, backup, restore, delete)
- [ ] Check audit logs: `curl -H "Authorization: Bearer ADMIN_TOKEN" http://localhost:8000/audit/logs`
- [ ] Verify actions are logged

### 9. Monitoring & Metrics

- [ ] Access Prometheus metrics: `curl http://localhost:8000/metrics`
- [ ] Verify metrics include:
  - [ ] `novabackup_backups_total`
  - [ ] `novabackup_backup_duration_seconds`
  - [ ] `novabackup_api_requests_total`
  - [ ] `novabackup_storage_usage_bytes`

### 10. Performance Checks

- [ ] API response time < 500ms for most endpoints
- [ ] Backup creation completes within expected time
- [ ] Restore operation completes successfully
- [ ] No memory leaks (monitor over 24 hours)
- [ ] CPU usage within acceptable limits

---

## Security Hardening

### Network Security

- [ ] Configure firewall to allow only required ports (8000/8050, 8080)
- [ ] Use reverse proxy (nginx/traefik) for HTTPS termination
- [ ] Enable HTTPS in production: `NOVABACKUP_HTTPS=true`
- [ ] Restrict database access to localhost only
- [ ] Use private networks for Docker containers

### Application Security

- [ ] Change all default passwords
- [ ] Use strong secrets (32+ characters)
- [ ] Enable rate limiting (configured by default)
- [ ] Review and restrict CORS settings
- [ ] Enable audit logging
- [ ] Regularly update dependencies: `pip install --upgrade -r requirements.txt`

### Database Security

- [ ] Use strong database passwords
- [ ] Restrict database user privileges (principle of least privilege)
- [ ] Enable database encryption at rest (if supported)
- [ ] Regular database backups
- [ ] Monitor database connections

### Secret Management

- [ ] Never commit `.env` file to git
- [ ] Use environment variables or secret management tools (Vault, AWS Secrets Manager)
- [ ] Rotate secrets regularly (every 90 days recommended)
- [ ] Use different secrets for dev/staging/production

---

## Backup & Recovery Procedures

### Application Backup

- [ ] Backup `.env` file (store securely)
- [ ] Backup database:
  - SQLite: Copy `novabackup.db` file
  - PostgreSQL: `pg_dump novabackup > backup.sql`
- [ ] Backup configuration files
- [ ] Backup scheduled jobs configuration
- [ ] Store backups in secure off-site location

### Application Recovery

- [ ] Install NovaBackup on new server
- [ ] Restore `.env` configuration
- [ ] Restore database from backup
- [ ] Run migrations if needed
- [ ] Start service
- [ ] Verify all functionality

### Data Backup Verification

- [ ] Schedule regular test restores (monthly recommended)
- [ ] Verify backup integrity
- [ ] Document recovery time objectives (RTO)
- [ ] Document recovery point objectives (RPO)

---

## Monitoring Setup

### Prometheus Integration

- [ ] Configure Prometheus to scrape `/metrics` endpoint
- [ ] Import Grafana dashboard (if available)
- [ ] Set up alert rules for:
  - Backup failure rate > 10%
  - No backups in last 24 hours
  - Storage usage > 80%
  - API error rate > 5%
  - Service down

### Log Aggregation

- [ ] Configure log rotation
- [ ] Set up centralized logging (ELK, Loki, etc.)
- [ ] Create log alerts for:
  - Critical errors
  - Authentication failures
  - Permission denied events

### Health Check Monitoring

- [ ] Monitor `/health` endpoint uptime
- [ ] Alert on health check failures
- [ ] Monitor database connection pool

---

## Documentation & Training

### Administrator Documentation

- [ ] Document installation procedure
- [ ] Document configuration options
- [ ] Document backup/restore procedures
- [ ] Document troubleshooting steps
- [ ] Document escalation procedures

### User Documentation

- [ ] Create user guide for dashboard
- [ ] Document how to create backups
- [ ] Document how to restore from backups
- [ ] Document password reset procedure

### Training

- [ ] Train administrators on system management
- [ ] Train users on self-service features
- [ ] Conduct backup/restore drill
- [ ] Schedule regular refresher training

---

## Go/No-Go Decision

### Must Pass (All Required)

- [ ] All health checks passing
- [ ] Authentication working correctly
- [ ] RBAC enforced on all endpoints
- [ ] Backup creation successful
- [ ] Restore operation successful
- [ ] Audit logging functional
- [ ] Security secrets properly configured
- [ ] No critical errors in logs

### Should Pass (Recommended)

- [ ] Cloud provider integration tested
- [ ] Notifications configured and tested
- [ ] Scheduled jobs running
- [ ] Monitoring integrated
- [ ] Performance benchmarks met
- [ ] Documentation complete

### Nice to Have (Optional)

- [ ] High availability configured
- [ ] Load balancing configured
- [ ] Advanced monitoring dashboards
- [ ] Automated testing in CI/CD
- [ ] Performance optimization completed

---

## Sign-Off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Project Manager | | | |
| System Administrator | | | |
| Security Officer | | | |
| Operations Lead | | | |

---

## Appendix: Quick Reference Commands

### Service Management

```bash
# Linux
sudo systemctl start novabackup
sudo systemctl stop novabackup
sudo systemctl restart novabackup
sudo systemctl status novabackup
sudo journalctl -u novabackup -f

# Windows
Start-Service novabackup
Stop-Service novabackup
Restart-Service novabackup
Get-Service novabackup

# Docker
docker-compose -f docker-compose-prod.yml up -d
docker-compose -f docker-compose-prod.yml down
docker-compose -f docker-compose-prod.yml logs -f
```

### Database Operations

```bash
# PostgreSQL backup
pg_dump novabackup > backup_$(date +%Y%m%d).sql

# PostgreSQL restore
psql novabackup < backup_20260327.sql

# SQLite backup
cp novabackup.db novabackup.db.backup
```

### Troubleshooting

```bash
# Check if port is in use
netstat -tulpn | grep 8000  # Linux
netstat -ano | findstr 8000  # Windows

# Test API endpoint
curl -v http://localhost:8000/health

# View application logs
tail -f data/logs/novabackup.log

# Check Docker containers
docker ps -a
docker logs novabackup_api
```

---

**Document Version:** 1.0  
**Last Updated:** March 27, 2026  
**Next Review:** June 27, 2026
