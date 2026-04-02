# Тести проекту

## Структура тестів

### Unit Tests

#### Backup.Server.Tests
- `JobServiceTests.cs` - тести JobService

### Integration Tests

#### Backup.Server.IntegrationTests
- `AgentIntegrationTests.cs` - gRPC communication tests
  - AgentRegistrationTests - agent registration scenarios
  - AgentHeartbeatTests - bidirectional streaming
  - AgentCapabilitiesTests - capability discovery
- `BackupRestoreIntegrationTests.cs` - backup/restore workflows
  - BackupFlowIntegrationTests - complete backup scenarios
  - RestoreFlowIntegrationTests - restore operations
  - RepositoryIntegrationTests - repository management

### Stress Tests

#### StressTestService
- Parallel backup tests (100+ VMs)
- Network failure simulation
- Endurance testing (8+ hours)
- Scalability testing

### Performance Benchmarks

| Metric | Target | Status |
|--------|--------|--------|
| Speed | > 1 GB/s | ✅ |
| Latency | < 100ms | ✅ |
| Concurrency | 100+ jobs | ✅ |
| Success Rate | > 99.9% | ✅ |

## Запуск тестів

> Для серверних/API тестів використовуйте порт `8050`.
> Для першого входу після інсталяції використовується bootstrap-користувач (за замовчуванням `admin/admin123`) з обов'язковою зміною пароля.

### Вимоги
- .NET 8.0 SDK
- PostgreSQL 14+
- Node.js 20+ (для UI)
- Docker (optional, for integration tests)

### Quick Start

```bash
# 1. Start PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_DB=backup_test \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  postgres:14

# 2. Run all tests
cd src/server/Backup.Server.Tests
dotnet test

# 3. Run integration tests
cd src/server/Backup.Server.IntegrationTests
dotnet test
```

### Docker Compose (Recommended)

```bash
# Start complete test environment
docker-compose -f docker-compose.integration-tests.yml up -d

# View results
docker logs backup-integration-test-runner

# Stop environment
docker-compose -f docker-compose.integration-tests.yml down
```

### .NET тести
```bash
cd src/server/Backup.Server.Tests
dotnet test --logger "console;verbosity=detailed"
```

### Integration тести
```bash
# All integration tests
cd src/server/Backup.Server.IntegrationTests
dotnet test

# Specific test category
dotnet test --filter "FullyQualifiedName~AgentRegistrationTests"
dotnet test --filter "FullyQualifiedName~BackupFlowIntegrationTests"

# With code coverage
dotnet test --collect:"XPlat Code Coverage"
```

### UI тести
```bash
cd src/ui
npm install
npm test
```

### Стрес тести (REST API)

```bash
# Start server first
cd src/server/Backup.Server
dotnet run

# Run stress test for 100 VMs
curl -X POST http://localhost:8050/api/stresstest/run \
  -H "Content-Type: application/json" \
  -d '{"vmCount":100,"concurrentCount":50}'

# Scalability test
curl -X POST http://localhost:8050/api/stresstest/scalability \
  -H "Content-Type: application/json" \
  -d '{"vmCount":100,"startConcurrency":10,"maxConcurrency":100,"stepSize":10}'

# Endurance test (8 hours)
curl -X POST http://localhost:8050/api/stresstest/endurance \
  -H "Content-Type: application/json" \
  -d '{"vmCount":50,"concurrentCount":25,"durationMinutes":480}'
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

# 4. Run stress tests
curl -X POST http://localhost:8050/api/stresstest/run \
  -H "Content-Type: application/json" \
  -d '{"vmCount":100,"concurrentCount":100}'
```

## Перевірка компонентів

### 1. Protos
- [x] agent.proto
- [x] job.proto
- [x] backup.proto
- [x] restore.proto
- [x] repository.proto
- [x] transfer.proto
- [x] common.proto

### 2. Server Services
- [x] AgentService
- [x] JobService
- [x] BackupService
- [x] RestoreService
- [x] RepositoryService
- [x] DashboardService
- [x] FileTransferService
- [x] StressTestService ✨ NEW

### 3. Background Services
- [x] JobSchedulerService
- [x] AgentHealthCheckService
- [x] RetentionPolicyService
- [x] AgentCommunicationService

### 4. REST Controllers
- [x] JobsController
- [x] BackupsController
- [x] RestoreController
- [x] RepositoriesController
- [x] AgentsController
- [x] SettingsController
- [x] ReportsController
- [x] StressTestController ✨ NEW

### 5. C++ Agent
- [x] DataMover
- [x] HyperV Agent
- [x] VMware Agent
- [x] KVM Agent
- [x] Database Agents

### 6. UI Components
- [x] Dashboard
- [x] Jobs
- [x] Backups
- [x] Restore
- [x] Repositories
- [x] Agents
- [x] Settings
- [x] Reports

### 7. Integration Tests ✨ NEW
- [x] Agent Registration Tests
- [x] Agent Heartbeat Tests
- [x] Agent Capabilities Tests
- [x] Backup Flow Tests
- [x] Restore Flow Tests
- [x] Repository Tests
- [x] Stress Test API Tests

## Test Reports

### View Test Results

```bash
# HTML report (requires reportgenerator)
dotnet tool install -g dotnet-reportgenerator-globaltool
reportgenerator -reports:coverage.cobertura.xml -targetdir:coveragereport

# View in browser
start coveragereport/index.html
```

### Stress Test Results

Access stress test results via:
- REST API: `GET http://localhost:8050/api/stresstest/results/{sessionId}`
- Database: `SELECT * FROM stress_test_sessions ORDER BY started_at DESC`

## CI/CD Integration

### GitHub Actions

Tests run automatically on:
- Push to `main` or `develop`
- Pull requests

See `.github/workflows/integration-tests.yml`

## Troubleshooting

### Common Issues

**PostgreSQL connection failed:**
```bash
docker logs backup-integration-test-db
docker-compose -f docker-compose.integration-tests.yml restart postgres-test
```

**Tests timeout:**
Increase timeout in runsettings:
```xml
<TestTimeout>600000</TestTimeout>
```

**gRPC connection refused:**
```bash
curl http://localhost:8050/health
docker logs backup-integration-test-server
```

### Collect Diagnostics

```bash
# All container logs
docker-compose -f docker-compose.integration-tests.yml logs > test-logs.txt

# Export test results
docker cp backup-integration-test-runner:/app/TestResults ./test-results
```

## Documentation

- 📖 [Integration Testing Guide](INTEGRATION_TESTING.md) - Detailed testing guide
- 📖 [API Documentation](API_DOCS.md) - REST API reference
- 📖 [Requirements](requirements.md) - System requirements
