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
- **Database**: SQLite (file-based, zero configuration)

## 🚀 Quick Install

### Windows (PowerShell — Administrator)

```powershell
# One command to install everything
& ([scriptblock]::Create((iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install-server.ps1).Content)) -AutoStart
```

Or download and run:
```powershell
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install-server.ps1 -OutFile install-server.ps1
.\install-server.ps1 -AutoStart
```

### Linux (bash — root)

```bash
# One command to install everything
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash -s -- --auto-start
```

Or save script first:
```bash
curl -fsSL -o install.sh https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh
sudo chmod +x install.sh && sudo ./install.sh --auto-start
```

## Access After Installation

- **UI**: http://localhost
- **API**: http://localhost:8000
- **Swagger**: http://localhost:8000/swagger

## Login

- Username: `admin`
- Password: `admin123`

## 📋 Installation Options

### Windows

| Option | Description | Default |
|--------|-------------|---------|
| `-InstallDir DIR` | Installation directory | `C:\BackupServer` |
| `-JwtKey KEY` | JWT secret key | auto-generated |
| `-Port PORT` | Server port | `8000` |
| `-AdminPassword PWD` | Admin password | `admin123` |
| `-AutoStart` | Start service after install | off |
| `-Force` | Force reinstallation | off |
| `-SkipBuild` | Use existing publish folder | off |
| `-LocalSource PATH` | Use local source code | download from GitHub |
| `-Uninstall` | Uninstall server | off |

**Examples:**
```powershell
# Install with custom port and password
.\install-server.ps1 -Port 9000 -AdminPassword "MySecurePass!" -AutoStart

# Install from local source
.\install-server.ps1 -LocalSource "C:\Projects\Backup\src\server\Backup.Server" -AutoStart

# Uninstall
.\install-server.ps1 -Uninstall
```

### Linux

| Option | Description | Default |
|--------|-------------|---------|
| `--auto-start` | Start services after install | on |
| `--jwt-key KEY` | JWT secret key | auto-generated |
| `-h, --help` | Show help | - |

**Examples:**
```bash
# Install with custom JWT key
sudo ./install.sh --jwt-key "my-secret-key" --auto-start

# Show help
./install.sh --help
```

## 🔧 Service Management

### Windows

```powershell
# Check status
Get-Service -Name BackupServer

# Start
Start-Service -Name BackupServer

# Stop
Stop-Service -Name BackupServer

# Restart
Restart-Service -Name BackupServer

# View logs
Get-Content "C:\BackupServer\publish\logs\backup-server-$(Get-Date -Format 'yyyyMMdd').log" -Tail 50
```

### Linux

```bash
# Check status
sudo systemctl status backup-server

# Start
sudo systemctl start backup-server

# Stop
sudo systemctl stop backup-server

# Restart
sudo systemctl restart backup-server

# View logs
sudo journalctl -u backup-server -f
# or
tail -f /var/log/backup-server.log
```

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
│   ├── Migrations/             # EF Core migrations
│   └── Database/               # EF Core + SQLite
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

### Configuration File

Edit `appsettings.json` in the publish directory:

```json
{
  "ConnectionStrings": {
    "DefaultConnection": "Data Source=backup.db"
  },
  "Jwt": {
    "Key": "",
    "Issuer": "BackupServer",
    "Audience": "BackupClients"
  },
  "Server": {
    "PublicUrl": "http://localhost:8000"
  },
  "BootstrapAdmin": {
    "Username": "admin",
    "Email": "admin@backupsystem.com",
    "Password": "admin123"
  },
  "AllowedOrigins": [],
  "Encryption": {
    "KeyFilePath": ""
  },
  "Serilog": {
    "MinimumLevel": "Information"
  }
}
```

> **Note:** If `Jwt:Key` is empty, a secure key is auto-generated and saved to `jwt.key` on first startup.

## 🔐 Security

- JWT authentication (auto-generated key if not provided)
- Role-Based Access Control: `Admin`, `Operator`, `Viewer`
- Bootstrap admin with enforced password change on first login
- AES-256 encryption for hypervisor credentials
- Audit logging of all operations
- Configurable CORS origins

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

## 🧪 Development

### Build from source

```bash
# Server
dotnet restore src/server/Backup.Server
dotnet publish src/server/Backup.Server -c Release -r win-x64 --self-contained true \
    -p:PublishSingleFile=true -o ./publish

# UI
cd src/ui
npm install
npm run build
```

### Run tests

```bash
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
