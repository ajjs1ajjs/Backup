# 🛡️ NovaBackup Enterprise v7.0

**Modern Web-Based Backup Platform**

---

## 🚀 Quick Start

### Windows
```powershell
# Download and run
.\novabackup.exe server

# Access Web UI
http://localhost:8050
```

### Linux
```bash
# Run
./novabackup server

# Access Web UI
http://localhost:8050
```

**Default Login:** `admin` / `admin123`

---

## 📋 Features

### ✅ Backup Jobs
- Files and Folders
- Databases (MySQL, PostgreSQL, SQLite)
- Cloud Storage (S3, Azure, Google Drive)
- Incremental Backups
- Compression + Encryption
- Flexible Scheduling (Cron)

### ✅ Restore
- File-level restore
- Full backup restore
- Database restore
- Point-in-time recovery

### ✅ Storage
- Local repositories
- SMB/CIFS shares
- S3-compatible (AWS, MinIO, Wasabi)
- Azure Blob Storage
- Google Cloud Storage

### ✅ Monitoring
- Real-time dashboard
- Session history
- Email notifications
- Telegram alerts
- Webhook support

---

## 🏗️ Architecture

```
┌──────────────┐     ┌──────────────┐     ┌─────────────┐
│  Web UI      │────▶│  REST API    │────▶│  Backup     │
│  (HTML/JS)   │◀────│  (Go + Gin)  │◀────│  Engine     │
└──────────────┘     └──────────────┘     └─────────────┘
                            │
                            ▼
                     ┌──────────────┐
                     │  SQLite DB   │
                     └──────────────┘
```

---

## 🔧 API Endpoints

### Authentication
```bash
POST /api/auth/login
POST /api/auth/logout
```

### Jobs
```bash
GET    /api/jobs
POST   /api/jobs
PUT    /api/jobs/:id
DELETE /api/jobs/:id
POST   /api/jobs/:id/run
```

### Backup
```bash
POST /api/backup/run
GET  /api/backup/sessions
```

### Restore
```bash
GET  /api/restore/points
POST /api/restore/files
POST /api/restore/database
```

---

## 📦 Installation

### From Source
```bash
go build -o novabackup ./cmd/novabackup/
./novabackup server
```

### Binary Releases
Download from [Releases](https://github.com/ajjs1ajjs/Backup/releases)

---

## 🎨 Web UI

The web interface is available at `http://localhost:8050` with:

- **Dark Theme** - Easy on the eyes
- **Responsive Design** - Works on all devices
- **Real-time Updates** - Live status updates
- **Intuitive Navigation** - Veeam-inspired UX

---

## 🔐 Security

- User authentication
- Role-based access control
- Encrypted backups (AES-256)
- Secure API with CORS

---

## 📊 System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| OS        | Windows 10 / Linux | Windows Server 2019+ / Ubuntu 22.04+ |
| CPU       | 2 cores | 4+ cores |
| RAM       | 2 GB    | 4+ GB      |
| Storage   | 1 GB    | As needed  |

---

## 🛠️ Development

### Project Structure
```
Backup/
├── cmd/
│   └── novabackup/     # Main application
├── internal/
│   ├── api/            # REST API handlers
│   ├── database/       # SQLite database layer
│   ├── scheduler/      # Job scheduler
│   └── backup/         # Backup engine (TODO)
├── web/
│   └── index.html      # Web UI
└── go.mod              # Go module
```

### Build
```bash
go mod tidy
go build -o novabackup ./cmd/novabackup/
```

### Test
```bash
go test ./...
```

---

## 📝 License

MIT License - See LICENSE file

---

## 🙏 Acknowledgments

- Inspired by [Veeam Backup & Replication](https://www.veeam.com)
- Built with [Go](https://go.dev) and [Gin](https://gin-gonic.com)
- UI inspired by modern dashboard designs

---

<div align="center">

**Made with ❤️ by NovaBackup Team**

[Report Issue](https://github.com/ajjs1ajjs/Backup/issues) • [Documentation](https://github.com/ajjs1ajjs/Backup/wiki)

</div>
