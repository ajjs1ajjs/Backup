# 💾 Backup System

<p align="center">
  <img src="https://img.shields.io/badge/.NET-8.0-blueviolet?style=for-the-badge&logo=.net" alt=".NET 8">
  <img src="https://img.shields.io/badge/React-18-blue?style=for-the-badge&logo=react" alt="React 18">
  <img src="https://img.shields.io/badge/C%2B%2B-20-orange?style=for-the-badge&logo=c%2B%2B" alt="C++ 20">
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-lightgrey?style=for-the-badge" alt="Platform">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="MIT">
  <img src="https://img.shields.io/badge/Version-1.0.0-orange?style=for-the-badge" alt="v1.0.0">
</p>

> **Enterprise-grade backup and restore system** with centralized management, web-based UI, and distributed agents supporting multiple hypervisors and databases.

---

## 📋 Table of Contents

- [✨ Features](#-features)
- [🏗️ Architecture](#️-architecture)
- [🚀 Quick Start](#-quick-start)
  - [Windows Installation](#windows-powershell--administrator)
  - [Linux Installation](#linux-bash--root)
  - [Manual Installation](#-manual-installation-from-source)
- [📖 Post-Installation](#-post-installation)
- [🛠️ Technology Stack](#️-technology-stack)
- [📁 Project Structure](#-project-structure)
- [🔧 Development Setup](#-development-setup)
- [📚 Documentation](#-documentation)
- [🧪 Testing](#-testing)
- [📝 License](#-license)

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🖥️ **Virtual Machine Support** | Inventory and manage Hyper-V, VMware, KVM virtual machines |
| 💾 **Database Backup** | PostgreSQL, Microsoft SQL Server, Oracle database backup |
| 📦 **Multi-Target Repositories** | Local storage, NFS, SMB, S3, Azure Blob, Google Cloud Storage |
| 📅 **Job Scheduling** | Flexible scheduling with cron expressions, full/incremental/differential backups |
| 🔄 **Point-in-Time Recovery** | Instant restore from any backup point with granular file recovery |
| 📊 **Reports & Analytics** | Activity dashboards, success rates, storage utilization, SLA tracking |
| 🔐 **Enterprise Security** | JWT authentication, RBAC, 2FA, audit logging, encryption at rest |
| 🌐 **Web-Based Management** | Modern React UI with real-time monitoring and alerts |
| 🔌 **Distributed Agents** | C++ agents with VSS support for consistent backups |
| 📈 **Scalable Architecture** | Designed for enterprises with thousands of protected workloads |

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                      Web Browser                             │
│                   (React UI - Port 80)                       │
└──────────────────────┬───────────────────────────────────────┘
                       │ HTTP/HTTPS
┌──────────────────────▼───────────────────────────────────────┐
│                  Backup Server (.NET 8)                      │
│              REST API + Swagger (Port 8000)                  │
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │  Jobs    │  │  Agents  │  │  Repos   │  │  Reports │    │
│  │ Engine   │  │ Manager  │  │ Manager  │  │ Engine   │    │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘    │
│                                                              │
│                      SQLite Database                         │
└──────────────────────┬───────────────────────────────────────┘
                       │ gRPC + Protocol Buffers
          ┌─────────────┼─────────────┐
          │             │             │
   ┌──────▼──────┐ ┌───▼──────┐ ┌───▼──────┐
   │ Hyper-V     │ │ VMware   │ │ MSSQL    │
   │ Agent       │ │ Agent    │ │ Agent    │
   └─────────────┘ └──────────┘ └──────────┘
```

---

## 🚀 Quick Start

### Windows (PowerShell - Administrator)

**🚀 Recommended: One-Command Installation**
Copy and paste this single command in an Administrator PowerShell window to clean up, install, and start the backup server:

```powershell
& {
    Write-Host "Cleaning up old installation..." -ForegroundColor Yellow
    sc.exe delete BackupServer 2>$null
    Remove-Item -Path "C:\BackupServer" -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item -Path "$env:TEMP\Backup-latest" -Recurse -Force -ErrorAction SilentlyContinue

    Write-Host "Downloading latest installer..." -ForegroundColor Yellow
    $installerUrl = "https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install-server.ps1"
    Invoke-WebRequest -Uri $installerUrl -OutFile "$env:TEMP\install-server.ps1"

    Write-Host "Running installation..." -ForegroundColor Yellow
    & "$env:TEMP\install-server.ps1" -AutoStart
}
```

**After installation:**
- 🌐 **Web UI**: http://localhost
- 🔌 **API**: http://localhost:8000
- 📖 **Swagger**: available in Development or when `Swagger:Enabled=true`
- 👤 **Bootstrap login**: `Admin`
- 🔑 **Bootstrap password**: check the server console output on first run (randomly generated)
- ✅ **Background Service**: Server runs in background; you can close the console window.

> ⚠️ **Important**: Look for the "Bootstrap admin user created" warning in your console to find your temporary password. This must be changed immediately after first login.

---

**Manual Step-by-Step (Alternative):**
If you prefer to run steps manually:

```powershell
# 1. Download installer
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install-server.ps1 -OutFile install-server.ps1

# 2. Run installer
.\install-server.ps1 -AutoStart
```

---

### Linux (bash - root)

**One-command installation:**

```bash
# Download and run the installer
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash -s -- --auto-start
```

**After installation:**
- 🌐 **Web UI**: http://localhost
- 🔌 **API**: http://localhost:8000
- 📖 **Swagger**: available in Development or when `Swagger:Enabled=true`

---

### 🛠️ Manual Installation from Source

<details>
<summary><strong>Click to expand detailed manual installation instructions</strong></summary>

#### Prerequisites

| Dependency | Version | Required |
|------------|---------|----------|
| .NET SDK | 8.0+ | ✅ Yes |
| Node.js | 18+ | ✅ Yes (for UI) |
| Git | Latest | ✅ Yes |
| CMake | 3.20+ | ⚠️ For agents only |
| MSVC Build Tools | 2022+ | ⚠️ For Windows agents |

#### 1. Clone the repository

```bash
git clone https://github.com/ajjs1ajjs/Backup.git
cd Backup
```

#### 2. Build and run the server

```bash
# Restore dependencies
dotnet restore src/server/Backup.Server/Backup.Server.csproj

# Build and run
dotnet run --project src/server/Backup.Server/Backup.Server.csproj
```

The server will start on http://localhost:8000

#### 3. Build the UI (optional)

```bash
cd src/ui
npm install
npm run build

# Copy built files to server's wwwroot
cp -r build/* ../server/Backup.Server/wwwroot/
```

#### 4. Build the Agent (optional)

**Windows:**
```bash
cd src/agent/Backup.Agent
mkdir build && cd build
cmake .. -G "Visual Studio 17 2022" -A x64
cmake --build . --config Release
```

**Linux:**
```bash
cd src/agent/Backup.Agent
mkdir build && cd build
cmake ..
cmake --build . --config Release
```

</details>

---

## 📖 Post-Installation

### First Login

1. Open http://localhost in your browser
2. Login with:
   - **Username**: `Admin`
   - **Password**: `Lkmo291263@`

> ⚠️ **Important**: Change the bootstrap password immediately after first login!

### Initial Setup

1. **Navigate to Settings** → Configure your backup infrastructure:
   - 🏢 Add hypervisors (Hyper-V, VMware, KVM)
   - 💾 Configure backup repositories (Local, S3, Azure, etc.)
   - 🔑 Set up encryption keys (optional but recommended)
   - 📧 Configure email notifications

2. **Create your first backup job**:
   - Go to **Jobs** → **Create Job**
   - Select source VMs or databases
   - Choose destination repository
   - Set schedule (manual or automated)

3. **Deploy agents** to protected hosts:
   ```powershell
   # Windows agent
   iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.ps1 -OutFile install.ps1
   .\install.ps1 -Server "your-server:8000" -Token "your-token" -AgentType "hyperv" -AutoStart
   ```

---

## 🛠️ Technology Stack

### Server
- **Framework**: .NET 8 (ASP.NET Core)
- **Database**: SQLite with Entity Framework Core
- **Authentication**: JWT with ASP.NET Identity
- **API**: RESTful + Swagger documentation
- **Communication**: gRPC with Protocol Buffers

### Client
- **Framework**: React 18 with TypeScript
- **UI Library**: Material UI (MUI)
- **State Management**: Redux Toolkit
- **Charts**: Chart.js / Recharts
- **HTTP Client**: Axios

### Agent
- **Language**: C++20
- **Build**: CMake
- **Communication**: gRPC (C++)
- **Windows**: VSS (Volume Shadow Copy Service)
- **Libraries**: OpenSSL, cURL, libxml2, Zstandard

---

## 📁 Project Structure

```
Backup/
├── 📄 README.md                          # This file
├── 📄 LICENSE                            # MIT License
├── 📄 RELEASE_NOTES.md                   # Version history
├── 📄 roadmap.md                         # Development roadmap
│
├── 📁 src/
│   ├── 📁 server/Backup.Server/          # .NET 8 API Server
│   │   ├── 📁 Controllers/               # REST API endpoints
│   │   ├── 📁 Services/                  # Business logic
│   │   ├── 📁 Database/                  # EF Core configuration
│   │   │   ├── 📁 Entities/              # Database models
│   │   │   └── 📁 Repositories/          # Data access layer
│   │   ├── 📁 Auth/                      # Authentication & authorization
│   │   ├── 📁 Grpc/                      # gRPC service implementations
│   │   └── 📄 appsettings.json           # Server configuration
│   │
│   ├── 📁 ui/                            # React web application
│   │   ├── 📁 src/
│   │   │   ├── 📁 components/            # React components
│   │   │   ├── 📁 pages/                 # Page components
│   │   │   ├── 📁 store/                 # Redux store
│   │   │   └── 📁 services/              # API clients
│   │   ├── 📄 package.json
│   │   └── 📄 tsconfig.json
│   │
│   ├── 📁 agent/Backup.Agent/            # C++ backup agent
│   │   ├── 📁 src/                       # Agent source code
│   │   ├── 📁 providers/                 # Backup providers (Hyper-V, VMware, etc.)
│   │   └── 📄 CMakeLists.txt
│   │
│   └── 📁 protos/                        # Protocol Buffers contracts
│       ├── agent.proto                   # Agent-server communication
│       └── shared.proto                  # Shared message types
│
├── 📁 scripts/
│   ├── 📄 install-server.ps1             # Windows server installer
│   ├── 📄 install.ps1                    # Windows agent installer
│   ├── 📄 install.sh                     # Linux server/agent installer
│   └── 📄 dotnet-install.ps1             # .NET SDK installer
│
└── 📁 docs/
    ├── 📄 API_DOCS.md                    # API documentation
    ├── 📄 TESTING.md                     # Testing guide
    ├── 📄 INTEGRATION_TESTING.md         # Integration tests
    ├── 📄 requirements.md                # System requirements
    └── 📄 install.md                     # Detailed installation guide
```

---

## 🔧 Development Setup

### Server Development

```bash
# Navigate to server directory
cd src/server/Backup.Server

# Watch mode (auto-reload on changes)
dotnet watch run

# Run with specific environment
ASPNETCORE_ENVIRONMENT=Development dotnet run
```

### UI Development

```bash
cd src/ui

# Install dependencies
npm install

# Start development server with hot reload
npm start

# Run tests
npm test

# Build for production
npm run build
```

### Agent Development

```bash
cd src/agent/Backup.Agent

# Create build directory
mkdir build && cd build

# Configure (Windows)
cmake .. -G "Visual Studio 17 2022" -A x64 -DCMAKE_BUILD_TYPE=Debug

# Configure (Linux)
cmake .. -DCMAKE_BUILD_TYPE=Debug

# Build
cmake --build . --config Debug
```

---

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [📖 Installation Guide](install.md) | Detailed installation instructions |
| [📡 API Documentation](API_DOCS.md) | REST API reference with examples |
| [🧪 Testing Guide](TESTING.md) | Unit and integration testing |
| [📋 Requirements](requirements.md) | System requirements and compatibility |
| [🗺️ Roadmap](roadmap.md) | Development roadmap and planned features |
| [📝 Release Notes](RELEASE_NOTES.md) | Version history and changelog |

---

## 🧪 Testing

### Server Tests

```bash
cd src/server/Backup.Server.Tests
dotnet test
```

### Integration Tests

```bash
# Start the test environment
docker-compose -f docker-compose.integration-tests.yml up -d

# Run integration tests
cd src/server/Backup.Server.IntegrationTests
dotnet test
```

---

## 🔒 Security

- 🔐 All passwords are hashed using ASP.NET Identity PasswordHasher
- 🎟️ JWT tokens expire and must be refreshed
- 🔑 API keys for agent authentication
- 📝 Comprehensive audit logging
- 🔒 HTTPS support for all communications
- 💾 Optional encryption at rest for backups

---

## 🤝 Contributing

We welcome contributions! Please see our [Roadmap](roadmap.md) for planned features.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📝 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

## 📞 Support

- 📧 **Email**: support@backupsystem.com
- 💬 **Issues**: [GitHub Issues](https://github.com/ajjs1ajjs/Backup/issues)
- 📖 **Documentation**: [Wiki](https://github.com/ajjs1ajjs/Backup/wiki)

---

<p align="center">
  Made with ❤️ by the Backup Team
</p>
