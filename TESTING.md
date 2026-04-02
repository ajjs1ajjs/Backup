# Тести проекту

## Структура тестів

### Unit Tests

#### Backup.Server.Tests
- `JobServiceTests.cs` - тести JobService

### Integration Tests

#### Тести БД
- Підключення до PostgreSQL
- CRUD операції для Jobs, Backups, Repositories
- Міграції

#### Тести API
- REST endpoints
- gRPC services
- Authentication

#### Тести агентів
- Hyper-V agent
- VMware agent  
- KVM agent
- Database agents

## Запуск тестів

### Вимоги
- .NET 8.0 SDK
- PostgreSQL 14+
- Node.js 20+ (для UI)

### .NET тести
```bash
cd src/server/Backup.Server.Tests
dotnet test
```

### UI тести
```bash
cd src/ui
npm install
npm test
```

### Комплексний тест
```bash
# 1. Запуск БД
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:14

# 2. Запуск сервера
cd src/server/Backup.Server
dotnet run

# 3. Запуск UI
cd src/ui
npm start
```

## Перевірка компонентів

### 1. Protos
- [ ] agent.proto
- [ ] job.proto
- [ ] backup.proto
- [ ] restore.proto
- [ ] repository.proto
- [ ] transfer.proto
- [ ] common.proto

### 2. Server Services
- [ ] AgentService
- [ ] JobService
- [ ] BackupService
- [ ] RestoreService
- [ ] RepositoryService
- [ ] DashboardService
- [ ] FileTransferService

### 3. Background Services
- [ ] JobSchedulerService
- [ ] AgentHealthCheckService
- [ ] RetentionPolicyService

### 4. REST Controllers
- [ ] JobsController
- [ ] BackupsController
- [ ] RestoreController
- [ ] RepositoriesController
- [ ] AgentsController
- [ ] SettingsController
- [ ] ReportsController

### 5. C++ Agent
- [ ] DataMover
- [ ] HyperV Agent
- [ ] VMware Agent
- [ ] KVM Agent
- [ ] Database Agents

### 6. UI Components
- [ ] Dashboard
- [ ] Jobs
- [ ] Backups
- [ ] Restore
- [ ] Repositories
- [ ] Agents
- [ ] Settings
- [ ] Reports
