# 🚀 NovaBackup v6.0: Enterprise Backup & Disaster Recovery Platform

> **Enterprise-grade backup solution** — повноцінна платформа резервного копіювання корпоративного рівня, здатна конкурувати з Veeam Backup & Replication.

---

## 📋 1. Візія та Огляд Проєкту

**NovaBackup v6.0** — це повноцінна платформа резервного копіювання корпоративного рівня для захисту:
- Віртуальних машин (VMware, Hyper-V, KVM)
- Фізичних серверів (Windows/Linux Agents)
- Контейнерів (Kubernetes)
- Баз даних (MySQL, PostgreSQL, SQLite, MSSQL)
- Хмарних робочих навантажень (AWS, Azure, GCP)

**Ключові можливості:**
- ✅ Безперервний захист даних (CDP) з RPO ≈ 0
- ✅ Глобальна блокова дедуплікація + стиснення Zstd
- ✅ Шифрування AES-256 (at-rest & in-transit)
- ✅ Instant VM Recovery та гранулярне відновлення
- ✅ SureBackup (авто-верифікація в Sandbox)
- ✅ Immutable backups (WORM/S3 Object Lock) проти ransomware
- ✅ Scale-out архітектура сховищ
- ✅ Оркестрація Disaster Recovery (Failover/Failback)

---

## 🎯 2. Цілі Системи (System Goals)

| Категорія | Можливості |
|-----------|-----------|
| **Backup** | VM, physical servers, files, databases, Kubernetes |
| **Data Protection** | CDP, global deduplication, compression (zstd), encryption (AES-256) |
| **Recovery** | Instant VM Recovery, file-level, bare-metal, granular DB restore |
| **Disaster Recovery** | Cross-site replication, failover orchestration, failback automation |
| **Verification** | SureBackup auto-testing in isolated sandbox |
| **Storage** | Scale-out repos, immutable/WORM, S3 tiering, garbage collection |
| **Security** | RBAC, audit logging, ransomware protection, TLS everywhere |

---

## 🌐 3. Підтримувана Інфраструктура

```yaml
Virtualization:
  - VMware vSphere (govmomi)
  - Microsoft Hyper-V (WMI/PowerShell)
  - KVM/QEMU
  - Proxmox VE (опціонально)

Physical:
  - Windows Server 2016+ (Agent)
  - Linux RHEL/Ubuntu/Debian (Agent)

Cloud:
  - AWS EC2 + S3
  - Azure VMs + Blob Storage
  - Google Cloud Platform

Containers:
  - Kubernetes (Velero-compatible API)
  - Docker volumes

Storage Backends:
  - Local disk / LVM
  - NFS v3/v4, SMB/CIFS
  - S3-compatible (MinIO, AWS, Ceph RGW)
  - Ceph RBD (опціонально)
  
  ┌─────────────────────────────────────────────────┐
  │              CONTROL PLANE                       │
  │  • API Server (REST/gRPC)                       │
  │  • Job Scheduler (gocron)                       │
  │  • Metadata Manager (PostgreSQL)                │
  │  • RBAC & Auth (JWT/OAuth2)                     │
  │  • Tenant Management                            │
  └────────────┬────────────────────────────────────┘
               │ NATS/Kafka Message Bus
               ▼
  ┌─────────────────────────────────────────────────┐
  │              DATA MOVERS (Proxies)              │
  │  • Snapshot orchestration                       │
  │  • Changed Block Tracking (CBT)                 │
  │  • Chunking (1-4MB variable)                    │
  │  • SHA-256 hashing + dedupe index check         │
  │  • Zstd compression + AES-256 encryption        │
  │  • Parallel streams + throttling                │
  │  • Resumable transfers                          │
  └────────────┬────────────────────────────────────┘
               │
               ▼
  ┌─────────────────────────────────────────────────┐
  │              STORAGE ENGINE                      │
  │  • Chunk-based content-addressable storage      │
  │  • Dedupe index (PostgreSQL + in-memory cache)  │
  │  • Garbage collection daemon                    │
  │  • Immutable/WORM support                       │
  │  • Scale-out repository logic                   │
  │  • S3 multipart upload support                  │
  └─────────────────────────────────────────────────┘
  
  graph TD
      Admin[Admin UI/CLI] --> API[Control Plane API]
      API --> DB[(PostgreSQL Metadata)]
      API --> Bus[Message Bus: NATS/Kafka]
      
      Bus --> Mover1[Data Mover #1]
      Bus --> Mover2[Data Mover #2]
      Bus --> MoverN[Data Mover #N]
      
      Mover1 --> Src1[VMware/Hyper-V/Agent]
      Mover2 --> Src2[K8s/Cloud/Physical]
      
      Mover1 --> Repo1[Repository: Local/S3/NFS]
      Mover2 --> Repo2[Repository: Scale-out Pool]
      
      DB --> Monitor[Prometheus Exporter]
      Monitor --> Grafana[Grafana Dashboards]
      
      API --> Restore[Restore/Replication Engine]
      Restore --> Sandbox[SureBackup Sandbox]
      
      D:\DOWNLOADS\table-fd86353a-9b81-4246-ae9a-3b745b35f690.csv
      
      1. SNAPSHOT & CBT
         ├─ Запит сніпшота через гіпервізор API
         ├─ Отримання Changed Block Tracking map
         └─ Блокування запису на час сніпшота (minimal I/O impact)
      
      2. BLOCK DISCOVERY
         ├─ Читання потоку даних (parallel streams)
         ├─ Фільтрація службових/тимчасових файлів
         └─ Підтримка application-aware (VSS для Windows)
      
      3. CHUNKING
         ├─ Variable-size chunking (1MB - 4MB)
         ├─ Content-defined boundaries (Rabin fingerprint)
         └─ Metadata extraction (file names, timestamps)
      
      4. HASHING
         ├─ SHA-256 для кожного чанку
         ├─ Parallel hash computation (CPU-optimized)
         └─ Bloom filter pre-check for quick dedupe skip
      
      5. DEDUPLICATION CHECK
         ├─ Query PostgreSQL: SELECT 1 FROM chunks WHERE hash = ?
         ├─ In-memory LRU cache for hot chunks
         └─ If exists: increment ref_count, skip storage
      
      6. PROCESSING (new chunks only)
         ├─ Zstandard compression (level 3-6 balance)
         ├─ AES-256-GCM encryption (per-tenant keys)
         └─ Optional: erasure coding for geo-redundancy
      
      7. TRANSFER
         ├─ Multipart upload for S3 (5MB+ chunks)
         ├─ Retry logic with exponential backoff
         └─ Bandwidth throttling per policy
      
      8. METADATA UPDATE
         ├─ INSERT INTO restore_point_chunks (...)
         ├─ UPDATE chunks SET ref_count = ref_count + 1
         └─ Async job completion notification via NATS
         
         -- Jobs configuration
         CREATE TABLE backup_jobs (
             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
             name VARCHAR(255) NOT NULL,
             tenant_id UUID REFERENCES tenants(id),
             source_type VARCHAR(50) NOT NULL,
             source_config JSONB NOT NULL,
             schedule_cron VARCHAR(100),
             retention_days INT DEFAULT 7,
             dedupe_enabled BOOLEAN DEFAULT true,
             compression_level INT DEFAULT 3,
             encryption_enabled BOOLEAN DEFAULT true,
             created_at TIMESTAMPTZ DEFAULT NOW(),
             updated_at TIMESTAMPTZ DEFAULT NOW()
         );
         
         -- Global chunk index (deduplication core)
         CREATE TABLE chunks (
             hash CHAR(64) PRIMARY KEY,
             size_bytes INT NOT NULL,
             compressed_size INT,
             storage_path VARCHAR(500) NOT NULL,
             repository_id UUID REFERENCES repositories(id),
             ref_count INT DEFAULT 1,
             first_seen TIMESTAMPTZ DEFAULT NOW(),
             last_accessed TIMESTAMPTZ DEFAULT NOW()
         );
         
         -- Restore points (backup instances)
         CREATE TABLE restore_points (
             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
             job_id UUID REFERENCES backup_jobs(id),
             point_time TIMESTAMPTZ NOT NULL,
             status VARCHAR(20) DEFAULT 'pending',
             total_bytes BIGINT,
             processed_bytes BIGINT,
             duration_seconds INT,
             metadata JSONB
         );
         
         -- Mapping: restore point -> chunks (with order)
         CREATE TABLE restore_point_chunks (
             restore_point_id UUID REFERENCES restore_points(id) ON DELETE CASCADE,
             chunk_hash CHAR(64) REFERENCES chunks(hash),
             sequence_order INT NOT NULL,
             original_path VARCHAR(1000),
             PRIMARY KEY (restore_point_id, sequence_order)
         );
         
         -- Repositories configuration
         CREATE TABLE repositories (
             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
             name VARCHAR(255) NOT NULL,
             type VARCHAR(50) NOT NULL,
             config JSONB NOT NULL,
             capacity_bytes BIGINT,
             used_bytes BIGINT DEFAULT 0,
             immutable_enabled BOOLEAN DEFAULT false,
             created_at TIMESTAMPTZ DEFAULT NOW()
         );
         
         -- Indexes for performance
         CREATE INDEX idx_chunks_hash ON chunks(hash);
         CREATE INDEX idx_chunks_repo ON chunks(repository_id);
         CREATE INDEX idx_restore_points_job ON restore_points(job_id, point_time DESC);
         CREATE INDEX idx_rpc_lookup ON restore_point_chunks(chunk_hash);
         CREATE INDEX idx_jobs_tenant ON backup_jobs(tenant_id, created_at);
         
NovaBackup/
├── .github/
│   └── workflows/           # CI/CD: test, build, release
├── ARCHITECTURE.md          # Цей документ
├── ROADMAP.md               # План релізів
├── README.md                # Quick start
├── LICENSE                  # Apache 2.0 / Enterprise
│
├── control-plane/           # Go: API, Scheduler, Metadata
│   ├── cmd/nova-api/main.go
│   ├── internal/api/
│   ├── internal/scheduler/
│   ├── internal/auth/
│   └── internal/metadata/
│
├── data-mover/              # Go/Rust: Data processing
│   ├── internal/chunker/
│   ├── internal/hasher/
│   ├── internal/dedupe/
│   ├── internal/compress/
│   └── internal/encrypt/
│
├── storage-engine/          # Repository abstractions
│   ├── local.go
│   ├── nfs.go
│   ├── s3.go
│   ├── scaleout.go
│   └── immutable.go
│
├── restore-engine/          # Recovery logic
│   ├── instant_vm.go
│   ├── file_restore.go
│   ├── baremetal.go
│   └── surebackup.go
│
├── replication-engine/      # DR orchestration
│   ├── replicator.go
│   ├── failover.go
│   └── failback.go
│
├── agents/                  # Host agents
│   ├── windows/ (VSS, service)
│   └── linux/ (systemd, LVM)
│
├── web-ui/                  # React + TypeScript
│   ├── src/components/
│   ├── src/pages/
│   └── src/api/
│
├── python/
│   ├── gui/ (PyQt6 Desktop)
│   └── ai_analytics/ (ML models)
│
├── deployments/             # Docker, K8s, Prometheus
├── docs/                    # API, user-guide, developer
├── tests/                   # unit, integration, e2e
└── Makefile

:root {
  --bg-primary: #1a1a2e;
  --bg-secondary: #16213e;
  --bg-tertiary: #0f3460;
  --accent: #e94560;
  --success: #00d26a;
  --warning: #ffc107;
  --error: #dc3545;
  --text-primary: #ffffff;
  --text-secondary: #a0a0a0;
  --border-radius: 8px;
  --shadow: 0 4px 12px rgba(0,0,0,0.3);
}

[data-theme="light"] {
  --bg-primary: #ffffff;
  --bg-secondary: #f5f5f5;
  --accent: #d63031;
  --success: #00b894;
  --text-primary: #2d3436;
}

┌─────────────────────────────────────────────────────────┐
│ ☰ NovaBackup v6.0        🔔3 ⚙️ 👤Admin ❌ │
├─────────────────────────────────────────────────────────┤
│ 📊 OVERVIEW                                              │
│ ┌───────┬────────┬─────────┬────────┐                  │
│ │📦Jobs │✅Success│⚠️Warnings│❌Failed│                  │
│ │  24   │   18   │    4    │   2    │                  │
│ └───────┴────────┴─────────┴────────┘                  │
│                                                         │
│ 📈 BACKUP STATUS (24h)                                 │
│ [████████████████░░░░░░░░░░████████] 68%              │
│  00:00  04:00  08:00  12:00  16:00  20:00             │
│                                                         │
│ 📋 RECENT JOBS                                          │
│ Job Name       │ Status │ Progress │ Next Run        │
│ Daily Backup   │ ✅Done │ 100%    │ Tomorrow 02:00   │
│ VM Replication │ 🔄Run  │ 68%     │ —                │
├─────────────────────────────────────────────────────────┤
│ 🏠Home 📦Jobs 🖥️Infra 📊Reports ⚙️Settings │
└─────────────────────────────────────────────────────────┘

D:\DOWNLOADS\table-fd86353a-9b81-4246-ae9a-3b745b35f690 (1).csv

POST /api/v1/jobs              # Create backup job
GET  /api/v1/jobs              # List jobs
POST /api/v1/jobs/:id/run      # Trigger manual run
GET  /api/v1/backups           # List restore points
POST /api/v1/restore           # Initiate restore
GET  /api/v1/metrics           # Prometheus metrics
GET  /swagger                  # OpenAPI documentation

# File backup
./nova-cli backup --source /data --dest /backup --dedupe --compress

# Database backup
./nova-cli db-backup --type mysql --host localhost --database mydb --dest /backup

# VM backup (VMware)
./nova-cli vm backup --type vmware --name "vm-db-01" \
  --vcenter https://vc.company.com --destination /backups/vmware --cbt

# Scheduler
./nova-cli scheduler start
./nova-cli scheduler run-now --job-id <uuid>

# Restore
./nova-cli restore files --backup-id <uuid> --path "/data" --dest /restored
./nova-cli restore instant-vm --backup-id <uuid> --target-esxi esxi01

// internal/backup/engine.go
type BackupEngine struct {
    db *database.Connection
    storage *storage.Engine
    bus messaging.Client
}

func (e *BackupEngine) PerformBackup(ctx context.Context, job BackupJob) (*BackupResult, error) {
    // 1. Snapshot & CBT
    // 2. Chunking & Hashing
    // 3. Dedupe check via PostgreSQL
    // 4. Compress (zstd) + Encrypt (AES-256-GCM)
    // 5. Write to storage + update metadata
    return result, nil
}

Безпека та Моніторинг
Security
RBAC: admin, operator, auditor, tenant roles
Encryption: AES-256-GCM at-rest, TLS 1.3 in-transit
Immutable: S3 Object Lock, WORM filesystem
Audit: Full request logging with retention

novabackup_jobs_total{status="success|failed"}
novabackup_bytes_processed_total{type="source|stored"}
novabackup_dedupe_ratio
novabackup_repository_capacity_bytes{repo="name"}

Alerting
Job failure → critical → PagerDuty
Repository >85% → warning → Email/Slack
Dedupe ratio drop >20% → possible ransomware → Security team

D:\DOWNLOADS\table-fd86353a-9b81-4246-ae9a-3b745b35f690 (2).csv

Примітки до Розробки
Принципи
Сумісність: Міграція з v5.x через JSON export/import
Мінімальні залежності: Go static binaries, Python тільки для GUI/AI
Тестування: go test -race, pytest, integration tests in Docker
CI/CD: GitHub Actions: lint → test → build → security scan → release
Документація: Swagger для API, MkDocs для user guide

Backup Throughput:
  - Single mover: ≥500 MB/s (10GbE, NVMe)
  - Scale-out: linear to 10+ movers

Deduplication:
  - Index lookup: <5ms p99 (PostgreSQL + Redis cache)
  - Ratio target: 10:1 for VM workloads, 3:1 for encrypted data

Recovery Time:
  - Instant VM mount: <30 seconds to boot
  - File restore: <5 seconds for metadata

Resource Usage:
  - Mover memory: ≤2GB per 1Gbps stream
  - API server: ≤512MB baseline, auto-scale under load
  
  
  Додаткові Ресурси
  Go Documentation
  PostgreSQL Documentation
  NATS Documentation
  React Documentation
  Prometheus + Grafana Guide
  S3 API Reference
  VMware vSphere API
  📄 License: Apache License 2.0 (core), Enterprise License for advanced features
  🏢 Organization: NovaBackup Team
  🗓️ Last Updated: March 2026
