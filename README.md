# 💾 Backup System

<p align="center">
  <img src="https://img.shields.io/badge/.NET-8.0-blueviolet?style=for-the-badge&logo=.net" alt=".NET 8">
  <img src="https://img.shields.io/badge/React-18-blue?style=for-the-badge&logo=react" alt="React 18">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="MIT">
  <img src="https://img.shields.io/badge/Version-1.0.0-orange?style=for-the-badge" alt="v1.0.0">
</p>

Enterprise backup management system with a .NET 8 server, React web UI, and C++ agent.

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🖥️ **Virtual Machines** | Inventory and manage Hyper-V VMs |
| 📦 **Repositories** | Configure backup storage repositories |
| 📅 **Job Scheduling** | Automated backup jobs with manual trigger |
| 🔄 **Restore** | Point-in-time recovery operations |
| 📊 **Reports** | Activity summaries and exportable reports |
| 🔒 **Security** | JWT auth with role-based access control |

## 🚀 Quick Install

### Windows (PowerShell — Administrator)

```powershell
iwr -useb https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install-server.ps1 -OutFile install-server.ps1
.\install-server.ps1 -AutoStart
```

### Linux (bash — root)

```bash
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash -s -- --auto-start
```

After installation:
- **UI**: http://localhost
- **API**: http://localhost:8000
- **Swagger**: http://localhost:8000/swagger

## 🛠️ Technology Stack

- **Server**: .NET 8 (ASP.NET Core)
- **UI**: React 18 + Material UI
- **Database**: SQLite (EF Core)
- **Agent**: C++
- **Authentication**: JWT

## 📁 Project Structure

```
src/
├── server/Backup.Server/     # .NET 8 API server
├── ui/                       # React web application
├── agent/Backup.Agent/       # C++ backup agent
└── protos/                   # Protocol Buffers contracts
```

## 🔧 Development

```bash
# Restore and run server
dotnet restore src/server/Backup.Server/Backup.Server.csproj
dotnet run --project src/server/Backup.Server/Backup.Server.csproj

# Build UI
cd src/ui && npm install && npm run build
```

## 📄 Documentation

- [📖 Installation Guide](install.md)
- [📚 API Documentation](API_DOCS.md)
- [🧪 Testing Guide](TESTING.md)
- [📋 Requirements](requirements.md)

## 📝 License

MIT License — see [LICENSE](LICENSE) for details.