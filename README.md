# 🛡️ Backup System

> Enterprise-grade backup solution for virtual machines and databases

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![.NET](https://img.shields.io/badge/.NET-8.0-purple.svg)](https://dotnet.microsoft.com/)
[![React](https://img.shields.io/badge/React-18.2-blue.svg)](https://reactjs.org/)
[![C++](https://img.shields.io/badge/C++-20-yellow.svg)](https://isocpp.org/)

Modern backup system with hybrid architecture (C# server + C++ agents) supporting Hyper-V, VMware, KVM and databases (MS SQL, PostgreSQL, Oracle).

## ✨ Features

### Backup Capabilities
- 🌐 **Multi-Hypervisor** - Hyper-V, VMware, KVM support
- 🗄️ **Database Backup** - MS SQL, PostgreSQL, Oracle
- 📦 **Compression** - Zstd, LZ4, Gzip
- 🔄 **Incremental** - CBT (Changed Block Tracking)
- 💾 **Deduplication** - Source & target side
- ☁️ **Cloud Storage** - AWS S3, Azure Blob, GCS

### Management
- 📊 **Web UI** - React + Material UI
- 🔌 **REST API** - Full CRUD operations
- 📅 **Scheduler** - Cron, GFS rotation
- 🔔 **Notifications** - Email, Telegram, Slack, Webhooks
- 🔒 **Security** - TLS, RBAC, Audit logging

### Recovery
- 🔄 **Full VM Restore** - All hypervisors
- ⚡ **Instant Restore** - Mount backups
- 📁 **File-Level Recovery** - Extract specific files
- ⏱️ **Point-in-Time** - Database recovery

## 🏗️ Architecture

- **Web UI**: `src/ui` (React + Material UI)
- **Management Server**: `src/server/Backup.Server` (.NET 8, REST + gRPC)
- **Agent Runtime**: `src/agent/Backup.Agent` (C++20)
- **Shared Contracts**: `src/protos` (Protocol Buffers)
- **Storage Targets**: Local, NFS/SMB, S3-compatible, Azure Blob, GCS

## 🚀 Quick Start

### 1. Start Server

```bash
# Using Docker
docker run -d \
  -p 8000:8000 \
  -e POSTGRES_HOST=localhost \
  -e Jwt__Key="CHANGE_ME_TO_A_STRONG_SECRET" \
  backupsystem/server:latest

# Or from source
cd src/server/Backup.Server
dotnet run
```

### 2. Install Agent

**Linux:**
```bash
# One-line installation (recommended)
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash -s -- --server 10.0.0.1:8000 --token "YOUR_TOKEN" --agent-type hyperv --auto-start

# Download script first, then run
curl -fsSL -o install.sh https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh
sudo chmod +x install.sh && sudo ./install.sh --server 10.0.0.1:8000 --token "YOUR_TOKEN" --agent-type hyperv --auto-start
```

**Windows (PowerShell):**
```powershell
# One-line installation (recommended)
irm https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.ps1 | iex -Server "10.0.0.1:8000" -Token "YOUR_TOKEN" -AgentType hyperv -AutoStart

# Download script first, then run
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.ps1" -OutFile install.ps1
.\install.ps1 -Server "10.0.0.1:8000" -Token "YOUR_TOKEN" -AgentType hyperv -AutoStart
```

### 3. Access UI

Open `http://localhost:8000/swagger` for API and configure UI API URL (see Configuration section).

## 📁 Project Structure

```
src/
├── protos/                    # gRPC Protocol Buffers
│   ├── agent.proto            # Agent communication
│   ├── job.proto              # Job definitions
│   ├── backup.proto           # Backup operations
│   ├── restore.proto          # Restore operations
│   ├── repository.proto       # Storage repositories
│   └── transfer.proto         # File transfer
│
├── server/Backup.Server/      # .NET 8 Server
│   ├── Services/               # Business logic
│   ├── Controllers/            # REST API
│   ├── BackgroundServices/     # Scheduled tasks
│   └── Database/               # EF Core + PostgreSQL
│
├── agent/Backup.Agent/         # C++ Agent
│   ├── core/                  # DataMover, Compression
│   ├── hyperv/                # Hyper-V integration
│   ├── vmware/                # VMware VDDK
│   ├── kvm/                   # libvirt
│   └── database/              # DB agents
│
└── ui/                        # React Application
    ├── src/
    │   ├── components/        # Reusable UI components
    │   ├── pages/              # Page components
    │   ├── services/           # API client
    │   └── store/             # State management
    └── public/
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_HOST` | PostgreSQL host | localhost |
| `POSTGRES_PORT` | PostgreSQL port | 5432 |
| `POSTGRES_DB` | Database name | backup |
| `POSTGRES_USER` | Database user | postgres |
| `POSTGRES_PASSWORD` | Database password | postgres |
| `SERVER_PORT` | Server port | 8000 |
| `Server__PublicUrl` | Public URL for agents/install scripts | auto-detected as `http://<local-ip>:8000` |
| `Jwt__Key` | JWT signing key (required) | no default, server won't start without it |
| `BootstrapAdmin__Username` | First admin username | admin |
| `BootstrapAdmin__Email` | First admin email | admin@backupsystem.com |
| `BootstrapAdmin__Password` | First admin temporary password | admin123 |

### Configuration File

Create `appsettings.json`:

```json
{
  "ConnectionStrings": {
    "DefaultConnection": "Host=localhost;Database=backup;Username=postgres;Password=postgres"
  },
  "Jwt": {
    "Key": "CHANGE_ME_TO_A_STRONG_SECRET",
    "Issuer": "BackupServer",
    "Audience": "BackupClients"
  },
  "Server": {
    "PublicUrl": "http://10.0.0.10:8000"
  },
  "BootstrapAdmin": {
    "Username": "admin",
    "Email": "admin@backupsystem.com",
    "Password": "admin123"
  },
  "Smtp": {
    "Host": "smtp.example.com",
    "Port": 587,
    "EnableSsl": true,
    "FromAddress": "noreply@backupsystem.com"
  },
  "Telegram": {
    "BotToken": "YOUR_BOT_TOKEN",
    "ChatId": "YOUR_CHAT_ID"
  },
  "Slack": {
    "WebhookUrl": "https://hooks.slack.com/YOUR_WEBHOOK"
  }
}
```

## 🔐 Security Defaults

- `Jwt:Key` is mandatory and must be set before server startup.
- On first installation, bootstrap admin credentials are created from `BootstrapAdmin:*`.
- First login with bootstrap admin enforces password change before issuing a token.
- Update `server.public_url` (Settings API) after installation if external/public endpoint changes.

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [API Documentation](API_DOCS.md) | REST API reference |
| [Installation Guide](install.md) | Server & agent installation |
| [Requirements](requirements.md) | System requirements |
| [Roadmap](roadmap.md) | Development roadmap |
| [Testing](TESTING.md) | Testing guide |
| [Release Notes](RELEASE_NOTES.md) | Version history |

## 🔌 API Endpoints

### Jobs
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/jobs` | List all jobs |
| POST | `/api/jobs` | Create new job |
| GET | `/api/jobs/{id}` | Get job details |
| POST | `/api/jobs/{id}/run` | Run job immediately |
| POST | `/api/jobs/{id}/stop` | Stop running job |

### Backups
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/backups` | List all backups |
| GET | `/api/backups/{id}` | Get backup details |
| DELETE | `/api/backups/{id}` | Delete backup |
| POST | `/api/backups/{id}/verify` | Verify backup |

### Restore
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/restore` | Start restore |
| GET | `/api/restore/{id}` | Get restore progress |
| POST | `/api/restore/{id}/cancel` | Cancel restore |

### Repositories
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/repositories` | List repositories |
| POST | `/api/repositories` | Add repository |
| POST | `/api/repositories/{id}/test` | Test connection |

### Reports
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/reports/summary` | Dashboard summary |
| GET | `/api/reports/activity` | Activity log |
| GET | `/api/reports/storage` | Storage usage |

## 🧪 Testing

```bash
# PostgreSQL
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:14

# Run tests
cd src/server/Backup.Server.Tests
dotnet test

# UI tests
cd src/ui
npm test
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the [MIT License](LICENSE).

## 🙏 Acknowledgments

- [gRPC](https://grpc.io/)
- [Entity Framework Core](https://docs.microsoft.com/en-us/ef/)
- [React](https://reactjs.org/)
- [Material UI](https://mui.com/)
- [Quartz.NET](https://www.quartz-scheduler.net/)

---

<p align="center">
  <strong>Made with ❤️ for enterprise backup solutions</strong>
</p>
