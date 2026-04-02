# Backup System MVP - Release Notes

## Version 1.0.0

### Release Date
_________________

### Features

#### Backup
- Full VM backup for Hyper-V, VMware, KVM
- Incremental backup with CBT (Changed Block Tracking)
- Differential backup
- Synthetic full backup
- Database backup (MS SQL, PostgreSQL, Oracle)
- Compression (Zstd, LZ4, Gzip)
- Deduplication

#### Restore
- Full VM recovery
- Instant restore (mount backup)
- File-level recovery (FLR)
- Point-in-time recovery for databases

#### Management
- Web UI (React + Material UI)
- REST API
- gRPC communication
- Job scheduling (cron, GFS)
- Retention policies
- Multi-repository support
- Cloud storage integration (S3, Azure Blob, GCS)

#### Notifications
- Email notifications
- Telegram notifications
- Slack integration
- Webhooks

#### Security
- TLS encryption
- RBAC
- Audit logging

### System Requirements

#### Server
- OS: Windows Server 2019+ / Linux (Ubuntu 22.04+)
- .NET 8.0
- PostgreSQL 14+
- 4+ CPU cores
- 8+ GB RAM
- 100+ GB storage

#### Agent
- OS: Windows Server 2019+ / Linux (Ubuntu 22.04+)
- 2+ CPU cores
- 4+ GB RAM

### Quick Start

1. Install server:
```bash
docker run -d -p 8050:8050 -p 8080:8080 backupsystem/server:latest
```

2. Access UI: http://localhost:8080

3. Install agent:
```bash
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash
```

### Known Limitations
- VMware CBT requires VDDK 7.0+
- Oracle RMAN requires Oracle 12c+
- File-level recovery for encrypted backups not supported

### Support
- Email: support@backupsystem.com
- Documentation: docs.backupsystem.com

---

## Changelog

### v1.0.0 (Initial Release)
- Initial MVP release
- Core backup/restore functionality
- Web UI
- REST API
- Basic notifications
