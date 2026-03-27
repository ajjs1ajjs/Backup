# NovaBackup Quick Start Script for Windows
# This script provides quick setup for different scenarios

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet('dev', 'prod', 'docker', 'test')]
    [string]$Scenario,
    
    [switch]$SkipSecrets,
    [switch]$NoPrompt
)

$ErrorActionPreference = "Stop"
$PROJECT_ROOT = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  NovaBackup Quick Start" -ForegroundColor Cyan
Write-Host "  Scenario: $Scenario" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Function to check prerequisites
function Check-Prerequisites {
    Write-Host "[1/6] Checking prerequisites..." -ForegroundColor Green
    
    # Check Python
    try {
        $pythonVersion = python --version 2>&1
        Write-Host "  ✓ Python found: $pythonVersion" -ForegroundColor Green
        
        # Check version >= 3.9
        $version = python -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')"
        $major, $minor = $version.Split('.')
        if ([int]$major -lt 3 -or ([int]$major -eq 3 -and [int]$minor -lt 9)) {
            throw "Python 3.9+ required, found $version"
        }
    } catch {
        Write-Host "  ✗ Python not found or version < 3.9" -ForegroundColor Red
        Write-Host "  Please install Python 3.9+ from https://python.org" -ForegroundColor Yellow
        if (!$NoPrompt) { Read-Host "Press Enter to exit" }
        exit 1
    }
    
    # Check Git
    try {
        $gitVersion = git --version 2>&1
        Write-Host "  ✓ Git found: $gitVersion" -ForegroundColor Green
    } catch {
        Write-Host "  ⚠ Git not found (optional for existing projects)" -ForegroundColor Yellow
    }
    
    # Check Docker (for docker scenario)
    if ($Scenario -eq 'docker') {
        try {
            $dockerVersion = docker --version 2>&1
            Write-Host "  ✓ Docker found: $dockerVersion" -ForegroundColor Green
            
            $composeVersion = docker compose version 2>&1
            Write-Host "  ✓ Docker Compose found: $composeVersion" -ForegroundColor Green
        } catch {
            Write-Host "  ✗ Docker not found" -ForegroundColor Red
            Write-Host "  Please install Docker Desktop from https://docker.com" -ForegroundColor Yellow
            if (!$NoPrompt) { Read-Host "Press Enter to exit" }
            exit 1
        }
    }
    
    Write-Host ""
}

# Function to create virtual environment
function Create-Venv {
    Write-Host "[2/6] Setting up virtual environment..." -ForegroundColor Green
    
    if (Test-Path "$PROJECT_ROOT\venv") {
        Write-Host "  ℹ Virtual environment already exists" -ForegroundColor Yellow
    } else {
        python -m venv venv
        Write-Host "  ✓ Virtual environment created" -ForegroundColor Green
    }
    
    # Activate
    & "$PROJECT_ROOT\venv\Scripts\Activate.ps1"
    Write-Host "  ✓ Virtual environment activated" -ForegroundColor Green
    Write-Host ""
}

# Function to install dependencies
function Install-Dependencies {
    Write-Host "[3/6] Installing dependencies..." -ForegroundColor Green
    
    # Upgrade pip
    python -m pip install --upgrade pip --quiet
    
    if ($Scenario -eq 'test') {
        pip install -e ".[api,db,dev]" --quiet
        Write-Host "  ✓ Dependencies installed (including dev tools)" -ForegroundColor Green
    } elseif ($Scenario -eq 'docker') {
        Write-Host "  ℹ Skipping pip install (using Docker)" -ForegroundColor Yellow
    } else {
        pip install -e ".[api,db]" --quiet
        Write-Host "  ✓ Dependencies installed" -ForegroundColor Green
    }
    
    Write-Host ""
}

# Function to configure environment
function Configure-Environment {
    Write-Host "[4/6] Configuring environment..." -ForegroundColor Green
    
    $envFile = "$PROJECT_ROOT\.env"
    
    if (Test-Path $envFile) {
        Write-Host "  ℹ .env file already exists" -ForegroundColor Yellow
        if (!$NoPrompt) {
            $overwrite = Read-Host "Overwrite existing .env? (y/n)"
            if ($overwrite -ne 'y') {
                Write-Host "  ℹ Keeping existing .env" -ForegroundColor Yellow
                Write-Host ""
                return
            }
        }
    }
    
    # Copy template
    Copy-Item "$PROJECT_ROOT\.env.example" $envFile
    Write-Host "  ✓ Created .env from template" -ForegroundColor Green
    
    # Generate secrets
    if ($SkipSecrets) {
        Write-Host "  ⚠ Skipping secret generation (use .\generate-secrets.ps1 -All manually)" -ForegroundColor Yellow
    } else {
        if (Test-Path "$PROJECT_ROOT\generate-secrets.ps1") {
            & "$PROJECT_ROOT\generate-secrets.ps1" -All
            Write-Host "  ✓ Secrets generated" -ForegroundColor Green
        } else {
            Write-Host "  ⚠ generate-secrets.ps1 not found, please generate secrets manually" -ForegroundColor Yellow
        }
    }
    
    # Configure based on scenario
    Write-Host "  Configuring for $scenario scenario..." -ForegroundColor Green
    
    $envContent = Get-Content $envFile -Raw
    
    switch ($Scenario) {
        'dev' {
            $envContent = $envContent -replace 'NOVABACKUP_DATABASE_URL=.*', 'NOVABACKUP_DATABASE_URL=sqlite:///./novabackup.db'
            $envContent = $envContent -replace 'NOVABACKUP_CLOUD_PROVIDERS=.*', 'NOVABACKUP_CLOUD_PROVIDERS=MOCK'
            $envContent = $envContent -replace 'NOVABACKUP_DEBUG=.*', 'NOVABACKUP_DEBUG=true'
            $envContent = $envContent -replace 'NOVABACKUP_PORT=.*', 'NOVABACKUP_PORT=8000'
        }
        'prod' {
            $envContent = $envContent -replace 'NOVABACKUP_DEBUG=.*', 'NOVABACKUP_DEBUG=false'
            $envContent = $envContent -replace 'NOVABACKUP_PORT=.*', 'NOVABACKUP_PORT=8050'
            Write-Host "  ⚠ Production config created - UPDATE SECRETS AND DATABASE URL!" -ForegroundColor Red
        }
        'test' {
            $envContent = $envContent -replace 'NOVABACKUP_DATABASE_URL=.*', 'NOVABACKUP_DATABASE_URL=sqlite:///:memory:'
            $envContent = $envContent -replace 'NOVABACKUP_CLOUD_PROVIDERS=.*', 'NOVABACKUP_CLOUD_PROVIDERS=MOCK'
            $envContent = $envContent -replace 'NOVABACKUP_DEBUG=.*', 'NOVABACKUP_DEBUG=true'
            $envContent = $envContent -replace 'NOVABACKUP_TESTING_MODE=.*', 'NOVABACKUP_TESTING_MODE=true'
        }
    }
    
    Set-Content $envFile $envContent -NoNewline
    Write-Host "  ✓ Environment configured for $Scenario" -ForegroundColor Green
    Write-Host ""
}

# Function to run database migrations
function Run-Migrations {
    Write-Host "[5/6] Running database migrations..." -ForegroundColor Green
    
    if ($Scenario -eq 'docker') {
        Write-Host "  ℹ Migrations will run automatically in Docker" -ForegroundColor Yellow
    } else {
        try {
            python -m novabackup.migrate
            Write-Host "  ✓ Database migrations completed" -ForegroundColor Green
        } catch {
            Write-Host "  ⚠ Migration script not found or failed" -ForegroundColor Yellow
            Write-Host "  ℹ Database will initialize on first run" -ForegroundColor Yellow
        }
    }
    
    Write-Host ""
}

# Function to start the application
function Start-Application {
    Write-Host "[6/6] Starting NovaBackup..." -ForegroundColor Green
    
    if ($Scenario -eq 'docker') {
        Write-Host "  Starting Docker Compose..." -ForegroundColor Green
        docker-compose -f docker-compose-prod.yml up -d
        Write-Host "  ✓ Docker containers started" -ForegroundColor Green
        Write-Host ""
        Write-Host "  Access points:" -ForegroundColor Cyan
        Write-Host "  - API:       http://localhost:8000" -ForegroundColor White
        Write-Host "  - Dashboard: http://localhost:8080" -ForegroundColor White
        Write-Host "  - Docs:      http://localhost:8000/docs" -ForegroundColor White
        Write-Host ""
        Write-Host "  View logs: docker-compose logs -f" -ForegroundColor Yellow
        Write-Host "  Stop:      docker-compose down" -ForegroundColor Yellow
    } elseif ($Scenario -eq 'test') {
        Write-Host "  Running tests..." -ForegroundColor Green
        pytest tests/ -v
        Write-Host "  ✓ Tests completed" -ForegroundColor Green
    } else {
        Write-Host "  Starting development server..." -ForegroundColor Green
        Write-Host ""
        Write-Host "  Access points:" -ForegroundColor Cyan
        Write-Host "  - API:       http://localhost:8000" -ForegroundColor White
        Write-Host "  - Dashboard: http://localhost:8000/static/index.html" -ForegroundColor White
        Write-Host "  - Docs:      http://localhost:8000/docs" -ForegroundColor White
        Write-Host ""
        Write-Host "  Default credentials:" -ForegroundColor Yellow
        Write-Host "  - Username: alice" -ForegroundColor White
        Write-Host "  - Password: secret" -ForegroundColor White
        Write-Host ""
        Write-Host "  Press Ctrl+C to stop the server" -ForegroundColor Yellow
        Write-Host ""
        
        python -m uvicorn novabackup.api:get_app --reload --host 0.0.0.0 --port 8000
    }
}

# Main execution
Check-Prerequisites

if ($Scenario -ne 'docker') {
    Create-Venv
    Install-Dependencies
    Configure-Environment
    Run-Migrations
}

Start-Application

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Setup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

if (!$NoPrompt -and $Scenario -ne 'test') {
    Read-Host "Press Enter to exit"
}
