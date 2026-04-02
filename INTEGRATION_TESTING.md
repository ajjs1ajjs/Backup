# Integration Testing Guide

## Overview

This guide covers integration testing between the C# server and C++ agents, as well as stress testing for 100+ parallel VMs.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Integration Test Suite](#integration-test-suite)
3. [Stress Testing](#stress-testing)
4. [Performance Benchmarks](#performance-benchmarks)
5. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Run All Tests with Docker Compose

```bash
# Start the integration testing environment
docker-compose -f docker-compose.integration-tests.yml up -d

# View test results
docker logs backup-integration-test-runner

# Stop environment
docker-compose -f docker-compose.integration-tests.yml down
```

### Run Tests Locally

```bash
# 1. Start PostgreSQL
docker run -d -p 5432:5432 \
  -e POSTGRES_DB=backup_test \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -v $(pwd)/database/init-test.sql:/docker-entrypoint-initdb.d/init.sql \
  postgres:14

# 2. Run integration tests
cd src/server/Backup.Server.IntegrationTests
dotnet restore
dotnet test --logger "console;verbosity=detailed"

# 3. Run with coverage
dotnet test --collect:"XPlat Code Coverage"
```

---

## Integration Test Suite

### Test Categories

#### 1. Agent Registration Tests (`AgentIntegrationTests.cs`)

Tests gRPC communication between server and agents.

```bash
# Run specific test class
dotnet test --filter "FullyQualifiedName~AgentRegistrationTests"

# Test scenarios:
# - RegisterAgent_ShouldReturnSuccess
# - RegisterAgent_WithDifferentTypes_ShouldSucceed
# - RegisterAgent_WithInvalidData_ShouldReturnError
# - RegisterAgent_MultipleTimes_ShouldHandleCorrectly
```

#### 2. Agent Heartbeat Tests

Tests bidirectional streaming for agent heartbeats.

```bash
dotnet test --filter "FullyQualifiedName~AgentHeartbeatTests"
```

#### 3. Backup Flow Tests (`BackupRestoreIntegrationTests.cs`)

Tests complete backup/restore workflows.

```bash
# Run backup flow tests
dotnet test --filter "FullyQualifiedName~BackupFlowIntegrationTests"

# Test scenarios:
# - CreateBackupJob_ShouldSucceed
# - CreateAndRunBackupJob_FullFlow
# - CreateIncrementalBackupJob_ShouldSucceed
# - ListJobs_ShouldReturnCreatedJobs
# - StopRunningJob_ShouldSucceed
# - DeleteJob_ShouldSucceed
```

#### 4. Restore Flow Tests

Tests restore operations.

```bash
dotnet test --filter "FullyQualifiedName~RestoreFlowIntegrationTests"
```

#### 5. Repository Tests

Tests repository management.

```bash
dotnet test --filter "FullyQualifiedName~RepositoryIntegrationTests"
```

---

## Stress Testing

### Stress Test API Endpoints

The `StressTestController` provides REST APIs for stress testing:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/stresstest/run` | POST | Run parallel backup test |
| `/api/stresstest/run-with-failures` | POST | Test with network failures |
| `/api/stresstest/scalability` | POST | Scalability test |
| `/api/stresstest/endurance` | POST | Long-running endurance test |
| `/api/stresstest/performance/{vmId}` | GET | Get performance metrics |
| `/api/stresstest/simulate-network-failure` | POST | Simulate network failure |

### Example: Run Stress Test for 100 VMs

```bash
# Using curl
curl -X POST http://localhost:8080/api/stresstest/run \
  -H "Content-Type: application/json" \
  -d '{
    "vmCount": 100,
    "concurrentCount": 50
  }'

# Using PowerShell
Invoke-RestMethod -Uri "http://localhost:8080/api/stresstest/run" \
  -Method POST \
  -ContentType "application/json" \
  -Body '{"vmCount":100,"concurrentCount":50}'
```

### Example: Stress Test with Network Failures

```bash
curl -X POST http://localhost:8080/api/stresstest/run-with-failures \
  -H "Content-Type: application/json" \
  -d '{
    "vmCount": 100,
    "concurrentCount": 50,
    "failureRatePercent": 10,
    "failureDurationMs": 5000
  }'
```

### Example: Scalability Test

```bash
curl -X POST http://localhost:8080/api/stresstest/scalability \
  -H "Content-Type: application/json" \
  -d '{
    "vmCount": 100,
    "startConcurrency": 10,
    "maxConcurrency": 100,
    "stepSize": 10
  }'
```

Response:
```json
{
  "success": true,
  "metrics": [
    {
      "concurrency": 10,
      "totalBackups": 100,
      "successfulBackups": 100,
      "averageDurationMs": 5234.5,
      "percentile95DurationMs": 8921.3,
      "throughputPerSecond": 19.1
    },
    {
      "concurrency": 20,
      "totalBackups": 100,
      "successfulBackups": 99,
      "averageDurationMs": 5456.2,
      "percentile95DurationMs": 9123.7,
      "throughputPerSecond": 18.3
    }
  ]
}
```

### Example: Endurance Test (8 hours)

```bash
curl -X POST http://localhost:8080/api/stresstest/endurance \
  -H "Content-Type: application/json" \
  -d '{
    "vmCount": 50,
    "concurrentCount": 25,
    "durationMinutes": 480
  }'
```

---

## Performance Benchmarks

### Target Metrics (per requirements.md)

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Speed** | > 1 GB/s | Disk throughput |
| **Latency** | < 100ms | Metadata operations |
| **Concurrency** | 100+ | Parallel backup jobs |
| **Success Rate** | > 99.9% | Backup completion |
| **RTO** | < 1 hour | Recovery time objective |

### Running Benchmarks

```bash
# Run scalability benchmark
curl -X POST http://localhost:8080/api/stresstest/scalability \
  -H "Content-Type: application/json" \
  -d '{
    "vmCount": 100,
    "startConcurrency": 10,
    "maxConcurrency": 100,
    "stepSize": 10
  }' | jq '.metrics[] | {concurrency, throughputPerSecond, averageDurationMs}'
```

### Expected Results

```
Concurrency: 10  | Throughput: ~20/s  | Avg: 5000ms  | P95: 8500ms
Concurrency: 20  | Throughput: ~19/s  | Avg: 5200ms  | P95: 8800ms
Concurrency: 50  | Throughput: ~17/s  | Avg: 5500ms  | P95: 9200ms
Concurrency: 100 | Throughput: ~15/s  | Avg: 6000ms  | P95: 9800ms
```

---

## Test Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TEST_PostgresHost` | PostgreSQL host | localhost |
| `TEST_PostgresPort` | PostgreSQL port | 5432 |
| `TEST_PostgresDb` | Database name | backup_test |
| `TEST_PostgresUser` | Database user | postgres |
| `TEST_PostgresPassword` | Database password | postgres |
| `TEST_ServerUrl` | Server URL | http://localhost:8080 |

### Stress Test Configuration

```json
{
  "maxConcurrentBackups": 100,
  "testDurationMinutes": 60,
  "failureRatePercent": 5,
  "enableNetworkFailureSimulation": true,
  "enableResourceMonitoring": true,
  "monitoringIntervalSeconds": 30
}
```

---

## Troubleshooting

### Common Issues

#### 1. PostgreSQL Connection Failed

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# View logs
docker logs backup-integration-test-db

# Restart container
docker-compose -f docker-compose.integration-tests.yml restart postgres-test
```

#### 2. Tests Timeout

Increase timeout in test settings:
```xml
<!-- In .csproj or runsettings -->
<TestTimeout>600000</TestTimeout> <!-- 10 minutes -->
```

#### 3. gRPC Connection Refused

Ensure server is running:
```bash
# Check server health
curl http://localhost:8080/health

# View server logs
docker logs backup-integration-test-server
```

#### 4. Insufficient Resources

For 100+ VM stress tests, ensure:
- **RAM**: 8+ GB
- **CPU**: 4+ cores
- **Disk**: 50+ GB free

### Collecting Diagnostics

```bash
# Collect all container logs
docker-compose -f docker-compose.integration-tests.yml logs > test-logs.txt

# Export test results
docker cp backup-integration-test-runner:/app/TestResults ./test-results

# Database diagnostics
docker exec -it backup-integration-test-db psql -U postgres -d backup_test -c "SELECT * FROM stress_test_sessions ORDER BY started_at DESC LIMIT 10;"
```

---

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Integration Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_DB: backup_test
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3
    
    - name: Setup .NET
      uses: actions/setup-dotnet@v3
      with:
        dotnet-version: 8.0.x
    
    - name: Run integration tests
      run: |
        cd src/server/Backup.Server.IntegrationTests
        dotnet test --logger trx --results-directory ../TestResults
    
    - name: Upload test results
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: test-results
        path: src/TestResults/
```

---

## Next Steps

1. **Run Integration Tests**: Start with `dotnet test` in the IntegrationTests project
2. **Execute Stress Test**: Use the REST API to run 100 VM stress test
3. **Analyze Results**: Check metrics and identify bottlenecks
4. **Optimize**: Address performance issues based on test results
5. **Automate**: Add tests to CI/CD pipeline

---

## Support

- 📧 Email: support@backupsystem.com
- 📖 Documentation: [docs/](./docs/)
- 🐛 Issues: [GitHub Issues](https://github.com/ajjs1ajjs/Backup/issues)
