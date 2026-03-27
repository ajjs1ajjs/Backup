# 💾 NovaBackup Backup & Restore Procedures

## Document Information

| Attribute | Value |
|-----------|-------|
| **Version** | 1.0 |
| **Last Updated** | March 27, 2026 |
| **Owner** | Operations Team |
| **Review Cycle** | Quarterly |

---

## Table of Contents

1. [Overview](#1-overview)
2. [Backup Procedures](#2-backup-procedures)
3. [Restore Procedures](#3-restore-procedures)
4. [Cloud Provider Backups](#4-cloud-provider-backups)
5. [Scheduled Backups](#5-scheduled-backups)
6. [Backup Verification](#6-backup-verification)
7. [Troubleshooting](#7-troubleshooting)

---

## 1. Overview

### 1.1 Purpose

This document describes procedures for creating, managing, and restoring backups using NovaBackup. It covers local backups, cloud backups, and automated scheduling.

### 1.2 Backup Types

| Type | Description | Use Case |
|------|-------------|----------|
| **Full Backup** | Complete copy of all data | Weekly baseline |
| **Incremental** | Only changed data since last backup | Daily operations |
| **Differential** | Changed data since last full backup | Mid-week protection |
| **Snapshot** | Point-in-time copy (cloud) | VM protection |

### 1.3 Retention Policies

| Policy | Description | Example |
|--------|-------------|---------|
| **Days** | Keep backups for N days | Keep 30 days |
| **Copies** | Keep N most recent copies | Keep 10 copies |
| **Grandfather-Father-Son** | Hierarchical retention | Daily/Weekly/Monthly |

### 1.4 Storage Locations

| Location | Type | Performance | Cost |
|----------|------|-------------|------|
| **Local Disk** | Direct attached | Fast | Low |
| **Network Share** | NFS/SMB | Medium | Medium |
| **AWS S3** | Object storage | Medium | Pay-per-use |
| **Azure Blob** | Object storage | Medium | Pay-per-use |
| **GCP Cloud Storage** | Object storage | Medium | Pay-per-use |

---

## 2. Backup Procedures

### 2.1 Creating Backups via Web Dashboard

#### Step-by-Step

1. **Access Dashboard**
   - Open browser: `http://your-server:8080`
   - Login with credentials

2. **Navigate to Backup Page**
   - Click "Backups" in sidebar
   - Click "Create Backup" button

3. **Configure Backup**
   ```
   VM ID: vm-001
   Description: Daily backup
   Destination Type: Local
   Destination Path: /var/backups/novabackup
   Compression: gzip
   Encryption: Enabled
   ```

4. **Start Backup**
   - Click "Create"
   - Monitor progress in real-time
   - Wait for completion notification

5. **Verify Backup**
   - Check backup appears in list
   - Verify status = "completed"
   - Note backup ID for future reference

### 2.2 Creating Backups via API

#### Single Backup

```bash
# Get authentication token
TOKEN=$(curl -s -X POST http://localhost:8050/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin&password=secret" | jq -r '.access_token')

# Create backup
curl -s -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vm_id": "vm-001",
    "description": "Manual backup via API",
    "destination_type": "local",
    "destination_path": "/var/backups/novabackup",
    "compression": "gzip",
    "encryption": true
  }' | jq
```

#### Expected Response

```json
{
  "backup_id": "backup-20260327-001",
  "vm_id": "vm-001",
  "status": "pending",
  "created_at": "2026-03-27T10:00:00Z",
  "size_bytes": null,
  "destination": "/var/backups/novabackup/backup-20260327-001"
}
```

### 2.3 Creating Backups via CLI

```bash
# Activate virtual environment (if needed)
source venv/bin/activate

# Create backup
novabackup create-backup \
  --vm-id vm-001 \
  --description "Manual backup via CLI" \
  --destination /var/backups/novabackup \
  --compression gzip \
  --encrypt

# Check status
novabackup list-backups --vm-id vm-001
```

### 2.4 Listing Backups

```bash
# Via API
curl -s http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" | jq

# Via CLI
novabackup list-backups

# Filter by VM
novabackup list-backups --vm-id vm-001

# Filter by status
novabackup list-backups --status completed
```

### 2.5 Deleting Backups

```bash
# Via API
BACKUP_ID="backup-20260327-001"
curl -s -X DELETE http://localhost:8050/backups/$BACKUP_ID \
  -H "Authorization: Bearer $TOKEN"

# Via CLI
novabackup delete-backup --id $BACKUP_ID

# Via Dashboard
# 1. Navigate to Backups page
# 2. Find backup in list
# 3. Click delete icon
# 4. Confirm deletion
```

---

## 3. Restore Procedures

### 3.1 Pre-Restore Checklist

Before restoring, verify:

- [ ] Backup exists and is valid
- [ ] Sufficient disk space for restore
- [ ] Target VM is powered off (for VM restores)
- [ ] Network connectivity to backup storage
- [ ] Appropriate permissions for restore operation
- [ ] Restore destination is accessible

### 3.2 Restoring via Web Dashboard

#### Step-by-Step

1. **Access Dashboard**
   - Open browser: `http://your-server:8080`
   - Login with credentials

2. **Navigate to Backups**
   - Click "Backups" in sidebar
   - Find the backup to restore

3. **Initiate Restore**
   - Click "Restore" button next to backup
   - Confirm backup details

4. **Configure Restore**
   ```
   Restore Type: Full
   Destination VM: vm-001 (or new VM)
   Restore Location: /var/restore/vm-001
   Overwrite Existing: No (unless intended)
   ```

5. **Start Restore**
   - Click "Restore"
   - Monitor progress
   - Wait for completion

6. **Post-Restore Verification**
   - Verify restored data integrity
   - Power on VM (if applicable)
   - Test application functionality

### 3.3 Restoring via API

```bash
# Get backup ID
BACKUP_ID=$(curl -s http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" | jq -r '.[0].backup_id')

# Initiate restore
curl -s -X POST http://localhost:8050/backups/$BACKUP_ID/restore \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "destination_vm": "vm-001",
    "destination_path": "/var/restore/vm-001",
    "restore_type": "full",
    "overwrite": false
  }' | jq
```

#### Expected Response

```json
{
  "restore_id": "restore-20260327-001",
  "backup_id": "backup-20260327-001",
  "status": "pending",
  "started_at": "2026-03-27T11:00:00Z",
  "estimated_completion": "2026-03-27T11:30:00Z"
}
```

### 3.4 Restoring via CLI

```bash
# Restore latest backup
novabackup restore \
  --backup-id backup-20260327-001 \
  --destination-vm vm-001 \
  --destination-path /var/restore/vm-001

# Restore to new location
novabackup restore \
  --backup-id backup-20260327-001 \
  --destination-path /var/restore/new-location \
  --overwrite

# Monitor restore progress
novabackup restore-status --restore-id restore-20260327-001
```

### 3.5 Point-in-Time Restore

```bash
# List backups for specific VM
curl -s "http://localhost:8050/backups?vm_id=vm-001" \
  -H "Authorization: Bearer $TOKEN" | jq

# Find backup closest to desired time
# Note the backup_id

# Restore that specific backup
curl -s -X POST http://localhost:8050/backups/BACKUP_ID/restore \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "destination_vm": "vm-001",
    "restore_type": "point-in-time"
  }'
```

### 3.6 File-Level Restore

For restoring individual files from a backup:

```bash
# Extract backup archive
cd /var/backups/novabackup
tar -xzf backup-20260327-001.tar.gz -C /tmp/restore

# Find specific file
find /tmp/restore -name "important-file.txt"

# Copy to destination
cp /tmp/restore/path/to/file /destination/path

# Cleanup
rm -rf /tmp/restore
```

---

## 4. Cloud Provider Backups

### 4.1 AWS Backup

#### Prerequisites

```ini
# In .env file
NOVABACKUP_CLOUD_PROVIDERS=AWS
AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY
AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY
AWS_DEFAULT_REGION=us-east-1
AWS_S3_BUCKET=novabackup-bucket
```

#### Create AWS Backup

```bash
# Via API
curl -s -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vm_id": "i-1234567890abcdef0",
    "destination_type": "cloud",
    "cloud_provider": "AWS",
    "cloud_region": "us-east-1",
    "cloud_destination": "s3://novabackup-bucket/backups/"
  }'

# Via CLI
novabackup create-backup \
  --vm-id i-1234567890abcdef0 \
  --destination-type cloud \
  --cloud-provider AWS \
  --cloud-region us-east-1 \
  --cloud-dest s3://novabackup-bucket/backups/
```

#### Create EBS Snapshot

```bash
# List EC2 instances
curl -s http://localhost:8050/vms \
  -H "Authorization: Bearer $TOKEN" | jq

# Create snapshot for specific instance
curl -s -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vm_id": "i-1234567890abcdef0",
    "backup_type": "snapshot"
  }'
```

#### Restore from AWS

```bash
# List available backups
curl -s "http://localhost:8050/backups?cloud_provider=AWS" \
  -H "Authorization: Bearer $TOKEN" | jq

# Restore from S3
curl -s -X POST http://localhost:8050/backups/BACKUP_ID/restore \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "destination_type": "cloud",
    "cloud_provider": "AWS",
    "cloud_region": "us-east-1"
  }'

# Restore from EBS snapshot
# This will create new EBS volume from snapshot
curl -s -X POST http://localhost:8050/backups/BACKUP_ID/restore \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "destination_type": "ec2",
    "instance_type": "t3.medium",
    "availability_zone": "us-east-1a"
  }'
```

### 4.2 Azure Backup

#### Prerequisites

```ini
# In .env file
NOVABACKUP_CLOUD_PROVIDERS=AZURE
AZURE_TENANT_ID=YOUR_TENANT_ID
AZURE_CLIENT_ID=YOUR_CLIENT_ID
AZURE_CLIENT_SECRET=YOUR_CLIENT_SECRET
AZURE_SUBSCRIPTION_ID=YOUR_SUBSCRIPTION_ID
AZURE_RESOURCE_GROUP=novabackup-rg
```

#### Create Azure Backup

```bash
# List Azure VMs
curl -s http://localhost:8050/vms \
  -H "Authorization: Bearer $TOKEN" | jq

# Create backup
curl -s -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vm_id": "/subscriptions/xxx/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1",
    "destination_type": "cloud",
    "cloud_provider": "AZURE",
    "cloud_region": "eastus",
    "cloud_destination": "azure://novabackup-rg/backups/"
  }'
```

#### Create Azure Snapshot

```bash
# Create OS disk snapshot
curl -s -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vm_id": "/subscriptions/xxx/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1",
    "backup_type": "snapshot",
    "snapshot_type": "os_disk"
  }'

# Create data disk snapshot
curl -s -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vm_id": "/subscriptions/xxx/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1",
    "backup_type": "snapshot",
    "snapshot_type": "data_disk",
    "disk_lun": 0
  }'
```

#### Restore from Azure

```bash
# Restore VM from snapshot
curl -s -X POST http://localhost:8050/backups/BACKUP_ID/restore \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "destination_type": "azure_vm",
    "resource_group": "production-rg",
    "location": "eastus",
    "vm_name": "restored-vm"
  }'
```

### 4.3 GCP Backup

#### Prerequisites

```ini
# In .env file
NOVABACKUP_CLOUD_PROVIDERS=GCP
GOOGLE_APPLICATION_CREDENTIALS=/etc/novabackup/gcp-service-account.json
GOOGLE_CLOUD_PROJECT=your-project-id
GCS_BUCKET=novabackup-bucket
```

#### Create GCP Backup

```bash
# List GCP instances
curl -s http://localhost:8050/vms \
  -H "Authorization: Bearer $TOKEN" | jq

# Create backup
curl -s -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vm_id": "my-instance-1",
    "destination_type": "cloud",
    "cloud_provider": "GCP",
    "cloud_region": "us-central1-a",
    "cloud_destination": "gcp://novabackup-bucket/backups/"
  }'
```

#### Create Persistent Disk Snapshot

```bash
curl -s -X POST http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "vm_id": "my-instance-1",
    "backup_type": "snapshot",
    "disk_name": "my-instance-1-boot"
  }'
```

#### Restore from GCP

```bash
# Restore from snapshot
curl -s -X POST http://localhost:8050/backups/BACKUP_ID/restore \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "destination_type": "gcp_instance",
    "project": "your-project-id",
    "zone": "us-central1-a",
    "instance_name": "restored-instance"
  }'
```

---

## 5. Scheduled Backups

### 5.1 Creating Scheduled Jobs via API

```bash
# Create daily backup job (2 AM UTC)
curl -s -X POST http://localhost:8050/scheduler/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily VM-001 Backup",
    "vm_id": "vm-001",
    "schedule": "0 2 * * *",
    "destination_type": "local",
    "destination_path": "/var/backups/novabackup",
    "retention_days": 30,
    "enabled": true,
    "notification_on_success": true,
    "notification_on_failure": true
  }'

# Create weekly backup job (Sunday 3 AM)
curl -s -X POST http://localhost:8050/scheduler/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Weekly Full Backup",
    "vm_id": "vm-001",
    "schedule": "0 3 * * 0",
    "backup_type": "full",
    "retention_days": 90,
    "enabled": true
  }'

# Create interval-based job (every 6 hours)
curl -s -X POST http://localhost:8050/scheduler/jobs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Frequent Incremental",
    "vm_id": "vm-001",
    "schedule_type": "interval",
    "interval_seconds": 21600,
    "backup_type": "incremental",
    "retention_days": 7,
    "enabled": true
  }'
```

### 5.2 Managing Scheduled Jobs

```bash
# List all jobs
curl -s http://localhost:8050/scheduler/jobs \
  -H "Authorization: Bearer $TOKEN" | jq

# Get specific job
curl -s http://localhost:8050/scheduler/jobs/JOB_ID \
  -H "Authorization: Bearer $TOKEN" | jq

# Enable job
curl -s -X POST http://localhost:8050/scheduler/jobs/JOB_ID/enable \
  -H "Authorization: Bearer $TOKEN"

# Disable job
curl -s -X POST http://localhost:8050/scheduler/jobs/JOB_ID/disable \
  -H "Authorization: Bearer $TOKEN"

# Delete job
curl -s -X DELETE http://localhost:8050/scheduler/jobs/JOB_ID \
  -H "Authorization: Bearer $TOKEN"
```

### 5.3 Viewing Job Execution History

```bash
# Get job executions
curl -s "http://localhost:8050/scheduler/jobs/JOB_ID/executions" \
  -H "Authorization: Bearer $TOKEN" | jq

# Filter by status
curl -s "http://localhost:8050/scheduler/jobs/JOB_ID/executions?status=failed" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 5.4 Cron Schedule Examples

| Description | Cron Expression |
|-------------|-----------------|
| Every hour | `0 * * * *` |
| Every day at 2 AM | `0 2 * * *` |
| Every Monday at 3 AM | `0 3 * * 1` |
| Every Sunday at 4 AM | `0 4 * * 0` |
| 1st of every month at 5 AM | `0 5 1 * *` |
| Every 6 hours | `0 */6 * * *` |
| Every 15 minutes | `*/15 * * * *` |
| Business hours (9 AM - 5 PM) | `0 9-17 * * 1-5` |

---

## 6. Backup Verification

### 6.1 Integrity Checks

```bash
# Verify backup checksum
cd /var/backups/novabackup
sha256sum backup-20260327-001.tar.gz

# Compare with stored checksum
cat backup-20260327-001.sha256

# Verify archive integrity
tar -tzf backup-20260327-001.tar.gz > /dev/null && echo "Archive OK"
```

### 6.2 Test Restore Procedure

Perform test restores monthly:

```bash
# 1. Select random backup
BACKUP_ID=$(curl -s http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" | jq -r '.[0].backup_id')

# 2. Restore to test location
curl -s -X POST http://localhost:8050/backups/$BACKUP_ID/restore \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "destination_path": "/var/restore/test-'"$(date +%Y%m%d)"'",
    "verify_only": true
  }'

# 3. Verify restored data
# 4. Document results
# 5. Cleanup test restore
```

### 6.3 Backup Reports

```bash
# Generate backup report
curl -s http://localhost:8050/backups/report \
  -H "Authorization: Bearer $TOKEN" | jq

# Export to CSV
curl -s http://localhost:8050/backups/export?format=csv \
  -H "Authorization: Bearer $TOKEN" > backups.csv
```

### 6.4 Monitoring Backup Health

```bash
# Check last backup time
curl -s "http://localhost:8050/backups?limit=1" \
  -H "Authorization: Bearer $TOKEN" | jq '.[0].created_at'

# Check backup size trend
curl -s http://localhost:8050/backups \
  -H "Authorization: Bearer $TOKEN" | jq '[.[] | {date: .created_at, size: .size_bytes}]'

# Check failed backups
curl -s "http://localhost:8050/backups?status=failed" \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## 7. Troubleshooting

### 7.1 Backup Failures

#### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| **Insufficient space** | Disk full | Free space or expand storage |
| **Permission denied** | Wrong permissions | `chown novabackup:novabackup` |
| **Network timeout** | Network issues | Check connectivity, increase timeout |
| **Lock conflict** | VM running | Stop VM or use snapshot |
| **Cloud credentials invalid** | Expired keys | Update credentials in .env |

#### Diagnostic Commands

```bash
# Check disk space
df -h /var/backups

# Check permissions
ls -la /var/backups/novabackup

# Check logs
tail -100 /var/log/novabackup/backup.log | grep -i error

# Check network connectivity
ping storage-server
curl -I http://cloud-provider.com
```

### 7.2 Restore Failures

#### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| **Backup corrupted** | Archive damaged | Use different backup, verify checksums |
| **Insufficient space** | Target disk full | Free space or use different target |
| **Version mismatch** | Software version | Update NovaBackup to same version |
| **Missing dependencies** | Required packages | Install required packages |

#### Diagnostic Commands

```bash
# Verify backup before restore
tar -tzf backup-file.tar.gz > /dev/null && echo "OK"

# Check target space
df -h /var/restore

# Check restore logs
tail -100 /var/log/novabackup/restore.log
```

### 7.3 Cloud Provider Issues

#### AWS

```bash
# Check AWS credentials
aws sts get-caller-identity

# Check S3 bucket access
aws s3 ls s3://novabackup-bucket/

# Check EC2 permissions
aws ec2 describe-instances
```

#### Azure

```bash
# Check Azure credentials
az account show

# Check storage access
az storage container list --account-name novabackup

# Check VM permissions
az vm list
```

#### GCP

```bash
# Check GCP credentials
gcloud auth list

# Check GCS bucket access
gsutil ls gs://novabackup-bucket/

# Check compute permissions
gcloud compute instances list
```

---

## Appendix A: Backup Size Estimation

| VM Size | Estimated Backup Size | Compression Ratio |
|---------|----------------------|-------------------|
| 50 GB | 15-25 GB | 2:1 - 3:1 |
| 100 GB | 30-50 GB | 2:1 - 3:1 |
| 500 GB | 150-250 GB | 2:1 - 3:1 |
| 1 TB | 300-500 GB | 2:1 - 3:1 |

**Formula:** `Estimated Size = Used Space / Compression Ratio`

---

## Appendix B: RTO/RPO Guidelines

| Backup Type | RTO | RPO |
|-------------|-----|-----|
| Local Full | 30 min | 24 hours |
| Local Incremental | 15 min | 6 hours |
| Cloud Full | 2 hours | 24 hours |
| Cloud Incremental | 1 hour | 6 hours |
| Snapshot | 5 min | 1 hour |

---

**END OF DOCUMENT**
