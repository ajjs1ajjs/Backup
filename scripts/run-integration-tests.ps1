#!/usr/bin/env pwsh
# Integration Test Runner Script for Backup System
# Usage: .\run-integration-tests.ps1 [-TestType <All|Unit|Integration|Stress>] [-Verbose]

param(
    [ValidateSet("All", "Unit", "Integration", "Stress")]
    [string]$TestType = "All",
    
    [switch]$Verbose,
    
    [switch]$SkipDocker,
    
    [string]$PostgresHost = "localhost",
    
    [int]$PostgresPort = 5432
)

$ErrorActionPreference = "Stop"
$RootPath = Split-Path $PSScriptRoot -Parent
$ServerPath = Join-Path $RootPath "server\Backup.Server"
$TestProjectPath = Join-Path $RootPath "server\Backup.Server.IntegrationTests"
$UnitTestPath = Join-Path $RootPath "server\Backup.Server.Tests"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Backup System Integration Test Runner" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

function Write-Step {
    param([string]$Message)
    Write-Host ""
    Write-Host "► $Message" -ForegroundColor Yellow
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

# Check prerequisites
Write-Step "Checking prerequisites..."

$dotnetVersion = dotnet --version 2>$null
if (-not $dotnetVersion) {
    Write-Error-Custom ".NET SDK not found. Please install .NET 8.0 SDK."
    exit 1
}
Write-Success ".NET SDK $dotnetVersion found"

if (-not $SkipDocker) {
    $dockerVersion = docker --version 2>$null
    if (-not $dockerVersion) {
        Write-Host "⚠ Docker not found. Some tests may be skipped." -ForegroundColor Orange
    } else {
        Write-Success "Docker found"
    }
}

# Start PostgreSQL if using Docker
if (-not $SkipDocker) {
    Write-Step "Starting PostgreSQL test container..."
    
    $existingContainer = docker ps -a --filter "name=backup-integration-test-db" --format "{{.Names}}"
    
    if ($existingContainer) {
        Write-Host "Container already exists. Starting..." -ForegroundColor Gray
        docker start backup-integration-test-db 2>$null
    } else {
        docker run -d `
            --name backup-integration-test-db `
            -p ${PostgresPort}:5432 `
            -e POSTGRES_DB=backup_test `
            -e POSTGRES_USER=postgres `
            -e POSTGRES_PASSWORD=postgres `
            postgres:14
        
        Write-Success "PostgreSQL container started"
    }
    
    Write-Host "Waiting for PostgreSQL to be ready..." -ForegroundColor Gray
    Start-Sleep -Seconds 5
    
    # Check if PostgreSQL is ready
    $maxRetries = 10
    $retryCount = 0
    while ($retryCount -lt $maxRetries) {
        try {
            $result = docker exec backup-integration-test-db pg_isready -U postgres 2>$null
            if ($result -like "*accepting connections*") {
                Write-Success "PostgreSQL is ready"
                break
            }
        } catch {
            Write-Host "Waiting for PostgreSQL... ($($retryCount + 1)/$maxRetries)" -ForegroundColor Gray
            Start-Sleep -Seconds 2
            $retryCount++
        }
    }
    
    if ($retryCount -eq $maxRetries) {
        Write-Error-Custom "PostgreSQL failed to start. Tests may fail."
    }
}

# Run Unit Tests
if ($TestType -eq "All" -or $TestType -eq "Unit") {
    Write-Step "Running Unit Tests..."
    
    Set-Location $UnitTestPath
    
    $unitArgs = @(
        "test"
        "--logger", "console;verbosity=normal"
        "--results-directory", (Join-Path $RootPath "test-results\unit")
    )
    
    if ($Verbose) {
        $unitArgs += "--logger", "console;verbosity=detailed"
    }
    
    & dotnet $unitArgs
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Unit Tests passed"
    } else {
        Write-Error-Custom "Unit Tests failed"
    }
}

# Run Integration Tests
if ($TestType -eq "All" -or $TestType -eq "Integration") {
    Write-Step "Running Integration Tests..."
    
    Set-Location $TestProjectPath
    
    $integrationArgs = @(
        "test"
        "--logger", "console;verbosity=normal"
        "--results-directory", (Join-Path $RootPath "test-results\integration")
    )
    
    if ($Verbose) {
        $integrationArgs += "--logger", "console;verbosity=detailed"
    }
    
    & dotnet $integrationArgs
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Integration Tests passed"
    } else {
        Write-Error-Custom "Integration Tests failed"
    }
}

# Run Stress Tests (via API)
if ($TestType -eq "All" -or $TestType -eq "Stress") {
    Write-Step "Running Stress Tests..."
    
    # Check if server is running
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 5 -ErrorAction Stop
        Write-Success "Backup server is running"
    } catch {
        Write-Host "⚠ Backup server not running. Starting..." -ForegroundColor Orange
        Write-Host "   Please run: cd $ServerPath && dotnet run" -ForegroundColor Gray
        Write-Host "   Skipping stress tests..." -ForegroundColor Gray
        $runStressTests = $false
    }
    
    if ($runStressTests) {
        Write-Host "Running stress test for 100 VMs..." -ForegroundColor Gray
        
        $stressBody = @{
            vmCount = 100
            concurrentCount = 50
        } | ConvertTo-Json
        
        try {
            $result = Invoke-RestMethod `
                -Uri "http://localhost:8080/api/stresstest/run" `
                -Method POST `
                -ContentType "application/json" `
                -Body $stressBody
            
            Write-Success "Stress Test completed"
            Write-Host "  Total Backups: $($result.totalBackups)" -ForegroundColor Cyan
            Write-Host "  Successful: $($result.successfulBackups)" -ForegroundColor Green
            Write-Host "  Failed: $($result.failedBackups)" -ForegroundColor $(if ($result.failedBackups -gt 0) { "Red" } else { "Gray" })
            Write-Host "  Average Duration: $([math]::Round($result.averageDurationMs, 2)) ms" -ForegroundColor Cyan
            Write-Host "  P95 Duration: $([math]::Round($result.percentile95DurationMs, 2)) ms" -ForegroundColor Cyan
        } catch {
            Write-Error-Custom "Stress Test failed: $($_.Exception.Message)"
        }
    }
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Test Run Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

if (Test-Path (Join-Path $RootPath "test-results")) {
    Write-Host ""
    Write-Host "Test results saved to:" -ForegroundColor Cyan
    Write-Host "  $(Join-Path $RootPath "test-results")" -ForegroundColor Gray
}

Write-Host ""
Write-Host "To view detailed results:" -ForegroundColor Cyan
Write-Host "  dotnet test --logger "trx" --results-directory test-results" -ForegroundColor Gray

Write-Host ""
Write-Host "To clean up Docker container:" -ForegroundColor Cyan
Write-Host "  docker stop backup-integration-test-db" -ForegroundColor Gray
Write-Host "  docker rm backup-integration-test-db" -ForegroundColor Gray

Write-Host ""
Write-Success "Test run completed!"
Write-Host ""
