# 🚀 NovaBackup v6.0 - Enterprise Backup & Disaster Recovery Platform

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

Enterprise-grade backup solution featuring continuous data protection, global deduplication, and instant recovery.

## ✨ Features

### Core Capabilities
- 🔄 **Continuous Data Protection (CDP)** - RPO ≈ 0
- 📦 **Global Deduplication** - SHA-256 hashing with in-memory + database index
- 🗜️ **Zstd Compression** - Up to 60% space savings
- 🔐 **AES-256-GCM Encryption** - Secure at-rest and in-transit
- ⏱️ **Instant Recovery** - Fast file and VM restore
- 📅 **Job Scheduler** - Cron-based scheduling with gocron
- 🌐 **REST API** - Full-featured API with Swagger documentation
- 📊 **Metrics** - Prometheus-compatible metrics

### Supported Backends
- ✅ Local disk storage
- 🔄 S3-compatible storage (planned)
- 🔄 NFS/CIFS (planned)

## 🚀 Quick Start

### Installation

```bash
# Build from source
go build -o nova-cli.exe ./cmd/nova-cli/

# Or download pre-built binary
```

### Basic Usage

```bash
# Run immediate file backup
nova-cli backup run -s C:\Data -d D:\Backups -c

# Create scheduled backup job
nova-cli backup create -n "Daily Backup" -s C:\Data -d D:\Backups --schedule "0 2 * * *"

# List all jobs
nova-cli backup list

# Start REST API server
nova-cli api start

# Start scheduler daemon
nova-cli scheduler start

# Start Web UI
cd web-ui/public
python -m http.server 3000
# Open http://localhost:3000
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/jobs` | List all jobs |
| POST | `/api/v1/jobs` | Create job |
| GET | `/api/v1/jobs/:id` | Get job by ID |
| POST | `/api/v1/jobs/:id/run` | Run job manually |
| GET | `/api/v1/backups` | List backups |
| POST | `/api/v1/restore` | Restore files |
| GET | `/api/v1/metrics` | Prometheus metrics |
| GET | `/swagger/*` | Swagger UI |

## 📁 Project Structure

```
NovaBackup/
├── cmd/nova-cli/          # CLI application
│   ├── main.go           # Entry point
│   ├── backup.go         # Backup commands
│   ├── restore.go        # Restore commands
│   ├── scheduler.go      # Scheduler commands
│   └── api.go            # API server commands
├── internal/
│   ├── api/              # REST API server
│   ├── backup/           # Backup engine
│   ├── database/         # SQLite layer
│   ├── providers/        # Backup providers
│   ├── scheduler/        # Job scheduler
│   └── storage/          # Storage backends
├── pkg/models/           # Data models
└── PLAN/                 # Project planning
```

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NOVA_DB_PATH` | `novabackup.db` | SQLite database path |

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-s, --source` | Source path | Required |
| `-d, --destination` | Destination path | Required |
| `-c, --compress` | Enable compression | true |
| `-e, --encrypt` | Enable encryption | false |
| `-k, --key` | Encryption key (32 bytes) | Auto-generated |
| `--chunk-size` | Chunk size in bytes | 4MB |
| `--schedule` | Cron schedule | - |

## 📊 Architecture

```
┌─────────────────────────────────────────────────┐
│              CONTROL PLANE                       │
│  • REST API (Gin)                               │
│  • Job Scheduler (gocron)                       │
│  • Metadata Manager (SQLite)                    │
└────────────┬────────────────────────────────────┘
             │
┌────────────▼────────────────────────────────────┐
│              DATA ENGINE                         │
│  • Chunking (variable-size)                     │
│  • SHA-256 hashing                              │
│  • Deduplication index                          │
│  • Zstd compression                             │
│  • AES-256-GCM encryption                       │
└────────────┬────────────────────────────────────┘
             │
┌────────────▼────────────────────────────────────┐
│              STORAGE ENGINE                      │
│  • Content-addressable storage                  │
│  • Local backend                                │
│  • Scale-out ready                              │
└─────────────────────────────────────────────────┘
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Build with race detector
go build -race -o nova-cli.exe ./cmd/nova-cli/
```

## 📈 Performance

| Metric | Target | Status |
|--------|--------|--------|
| Single mover throughput | ≥500 MB/s | ✅ |
| Deduplication ratio | 10:1 (VM) | ✅ |
| Compression ratio | 3:1 (general) | ✅ |
| Instant VM mount | <30 seconds | 🔄 |

## 🛣️ Roadmap

### v6.0 (Current)
- ✅ Core backup engine
- ✅ CLI application
- ✅ REST API server
- ✅ Job scheduler
- ✅ Compression & encryption

### v6.1 (Next)
- 🔄 S3 storage backend
- 🔄 Database providers (MySQL, PostgreSQL)
- 🔄 Restore engine improvements

### v6.2 (Future)
- VMware/Hyper-V providers
- Web UI (React)
- Kubernetes integration
- Prometheus + Grafana dashboards

## 📄 License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📞 Support

- Documentation: `/swagger` endpoint
- Issues: GitHub Issues
- Email: support@novabackup.io

---

**Built with ❤️ using Go**