# 📂 Backup System - Directory Structure

```
.
├── 📄 README.md                 # Main documentation
├── 📄 LICENSE                    # MIT License
├── 📄 .gitignore                # Git ignore rules
├── 📄 .gitattributes            # Git line ending settings
│
├── 📂 .github/
│   └── workflows/
│       └── build.yml            # CI/CD pipeline
│
├── 📂 src/
│   │
│   ├── 📂 protos/              # gRPC Protocol Buffers
│   │   ├── agent.proto          # Agent registration & commands
│   │   ├── job.proto            # Job definitions
│   │   ├── backup.proto         # Backup operations
│   │   ├── restore.proto        # Restore operations
│   │   ├── repository.proto     # Storage repositories
│   │   ├── transfer.proto       # File transfer
│   │   └── common.proto         # Shared messages
│   │
│   ├── 📂 server/
│   │   │
│   │   ├── 📂 Backup.Server/    # .NET 8 Server
│   │   │   │
│   │   │   ├── 📂 Services/      # Business logic (gRPC + internal)
│   │   │   │   ├── AgentServiceImpl.cs
│   │   │   │   ├── JobServiceImpl.cs
│   │   │   │   ├── BackupServiceImpl.cs
│   │   │   │   ├── RestoreServiceImpl.cs
│   │   │   │   ├── RepositoryServiceImpl.cs
│   │   │   │   ├── DashboardServiceImpl.cs
│   │   │   │   ├── FileTransferServiceImpl.cs
│   │   │   │   ├── AgentCommunicationService.cs
│   │   │   │   ├── AgentDeploymentService.cs
│   │   │   │   ├── FastCloneAndRestoreServices.cs
│   │   │   │   ├── FileLevelRecoveryService.cs
│   │   │   │   ├── EmailNotificationService.cs
│   │   │   │   ├── TelegramSlackWebhookService.cs
│   │   │   │   ├── PdfReportService.cs
│   │   │   │   ├── SchedulerAndRepositoryServices.cs
│   │   │   │   └── StressTestService.cs
│   │   │   │
│   │   │   ├── 📂 Controllers/   # REST API Controllers
│   │   │   │   ├── MainControllers.cs
│   │   │   │   └── ExtendedControllers.cs
│   │   │   │
│   │   │   ├── 📂 BackgroundServices/  # Scheduled tasks
│   │   │   │   └── JobSchedulerService.cs
│   │   │   │
│   │   │   ├── 📂 Database/      # Entity Framework Core
│   │   │   │   ├── BackupDbContext.cs
│   │   │   │   ├── Entities/Entities.cs
│   │   │   │   └── schema.sql
│   │   │   │
│   │   │   ├── 📂 Program.cs
│   │   │   └── 📂 Backup.Server.csproj
│   │   │
│   │   └── 📂 Backup.Server.Tests/  # Unit tests
│   │       ├── JobServiceTests.cs
│   │       └── Backup.Server.Tests.csproj
│   │
│   ├── 📂 agent/
│   │   │
│   │   └── 📂 Backup.Agent/     # C++ Agent
│   │       ├── 📂 core/         # Core functionality
│   │       │   ├── data_mover.h/cpp    # File transfer
│   │       │   ├── compression.h/cpp   # Zstd/LZ4/Gzip
│   │       │   └── cbt.h/cpp           # Changed Block Tracking
│   │       │
│   │       ├── 📂 hyperv/       # Hyper-V integration
│   │       │   ├── hyperv_agent.h
│   │       │   └── hyperv_agent.cpp
│   │       │
│   │       ├── 📂 vmware/       # VMware VDDK integration
│   │       │   ├── vmware_agent.h
│   │       │   └── vmware_agent.cpp
│   │       │
│   │       ├── 📂 kvm/          # KVM/libvirt integration
│   │       │   ├── kvm_agent.h
│   │       │   └── kvm_agent.cpp
│   │       │
│   │       ├── 📂 database/     # Database agents
│   │       │   ├── database_agent.h
│   │       │   └── database_agent.cpp
│   │       │
│   │       ├── 📂 main.cpp      # Entry point
│   │       ├── 📂 CMakeLists.txt
│   │       ├── 📂 Makefile
│   │       └── 📂 Dockerfile
│   │
│   └── 📂 ui/                   # React Frontend
│       ├── 📂 src/
│       │   ├── 📂 components/
│       │   │   └── Layout.js
│       │   │
│       │   ├── 📂 pages/
│       │   │   ├── Dashboard.js
│       │   │   ├── Jobs.js
│       │   │   ├── Backups.js
│       │   │   ├── Restore.js
│       │   │   ├── Repositories.js
│       │   │   ├── Agents.js
│       │   │   ├── Settings.js
│       │   │   ├── Reports.js
│       │   │   └── Login.js
│       │   │
│       │   ├── 📂 services/
│       │   │   └── ApiContext.js
│       │   │
│       │   ├── 📂 store/
│       │   │   └── authStore.js
│       │   │
│       │   └── 📂 App.js
│       │
│       ├── 📂 public/
│       ├── 📂 package.json
│       └── 📂 Dockerfile
│
├── 📂 docs/
│   └── 📂 README.md             # Documentation index
│
└── 📂 (root files)
    ├── 📄 roadmap.md             # Original roadmap
    ├── 📄 roadmap_recommendations.md
    ├── 📄 requirements.md       # System requirements
    ├── 📄 install.md           # Installation guide
    ├── 📄 PLAN_FACT.md          # Task tracking
    ├── 📄 API_DOCS.md          # API documentation
    ├── 📄 RELEASE_NOTES.md      # Version history
    ├── 📄 TESTING.md           # Testing guide
    └── 📄 VALIDATION.md         # Validation report
```

---

## 📊 Statistics

| Category | Files |
|----------|-------|
| C# Source | ~20 |
| C++ Source | ~10 |
| Protos | 7 |
| React Components | ~15 |
| CI/CD | 1 |
| Documentation | ~10 |
| **Total** | **~66** |

## 🛠️ Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | .NET 8, ASP.NET Core |
| Agents | C++20 |
| UI | React 18, Material UI |
| Database | PostgreSQL, Entity Framework Core |
| Communication | gRPC, REST |
| Scheduling | Background Services |
| Compression | Zstd, LZ4 |
| Storage | S3, Azure Blob, GCS |
