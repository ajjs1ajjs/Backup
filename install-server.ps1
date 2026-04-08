# Backup Server Installer - Enterprise Edition
# Version: 1.0.0
param(
    [switch]$AutoStart,
    [switch]$SkipSSL,
    [string]$InstallDir = "C:\BackupServer",
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"
$Version = "1.0.0"
$ServerDir = Join-Path $InstallDir "server"
$PublishDir = Join-Path $ServerDir "publish"

function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] $Message" -ForegroundColor Cyan
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

function Show-Help {
    @"
Backup Server Installer v$Version

Usage: .\install-server.ps1 [OPTIONS]

Options:
    -InstallDir DIR   Installation directory (default: C:\BackupServer)
    -AutoStart        Start server after installation
    -SkipSSL          Skip SSL certificate verification (insecure)
    -Uninstall        Uninstall server

Examples:
    .\install-server.ps1 -AutoStart
    .\install-server.ps1 -InstallDir "D:\BackupServer" -AutoStart
    .\install-server.ps1 -Uninstall

"@
    exit 0
}

function Check-Admin {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        Write-Error-Custom "This script must be run as Administrator"
    }
}

function Install-DotNet {
    if (Get-Command dotnet -ErrorAction SilentlyContinue) {
        $dotnetVersion = dotnet --version
        Write-Log ".NET SDK already installed: $dotnetVersion"
        return
    }

    Write-Log "Installing .NET SDK 8.0..."
    $scriptUrl = "https://dot.net/v1/dotnet-install.ps1"
    $scriptPath = Join-Path $env:TEMP "dotnet-install.ps1"

    try {
        Invoke-WebRequest -Uri $scriptUrl -OutFile $scriptPath -UseBasicParsing -SkipCertificateCheck:$SkipSSL
        & $scriptPath -Channel 8.0 -InstallDir "C:\Program Files\dotnet"
        Write-Log ".NET SDK 8.0 installed successfully"
        Remove-Item $scriptPath -Force -ErrorAction SilentlyContinue
    }
    catch {
        Write-Error-Custom "Failed to install .NET SDK. Please install manually from https://dotnet.microsoft.com/download"
    }
}

function Install-NodeJS {
    if (Get-Command node -ErrorAction SilentlyContinue) {
        $nodeVersion = node --version
        Write-Log "Node.js already installed: $nodeVersion"
        return
    }

    Write-Log "Installing Node.js..."
    winget install --id OpenJS.NodeJS --silent --accept-package-agreements
    Write-Log "Node.js installed"
}

function Clone-OrUpdate-Repo {
    $repoUrl = "https://github.com/ajjs1ajjs/Backup.git"
    $cloneDir = Join-Path $env:TEMP "Backup-latest"

    if (Test-Path (Join-Path $cloneDir ".git")) {
        Write-Log "Updating repository..."
        Push-Location $cloneDir
        try {
            # Reset any local changes to avoid merge conflicts
            git checkout -- . 2>&1 | Out-Null
            git clean -fd 2>&1 | Out-Null
            git pull 2>&1 | Out-Null
        }
        finally {
            Pop-Location
        }
    }
    else {
        Write-Log "Cloning repository..."
        if (Test-Path $cloneDir) { Remove-Item $cloneDir -Recurse -Force }
        git clone $repoUrl $cloneDir | Out-Null
    }

    return $cloneDir.Trim()
}

function Build-Server {
    Write-Log "Building Backup Server..."

    $repoDir = Clone-OrUpdate-Repo
    $serverProject = Join-Path $repoDir "src\server\Backup.Server\Backup.Server.csproj"

    if (-not (Test-Path $serverProject)) {
        Write-Error-Custom "Server project not found at $serverProject"
    }

    # Clean build artifacts
    Remove-Item (Join-Path $repoDir "src\server\Backup.Server\obj") -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item $PublishDir -Recurse -Force -ErrorAction SilentlyContinue

    # Restore and publish
    Write-Log "Restoring NuGet packages..."
    dotnet restore $serverProject

    Write-Log "Publishing server..."
    dotnet publish $serverProject -c Release -o $PublishDir

    # Build UI if available
    $uiDir = Join-Path $repoDir "src\ui"
    if (Test-Path (Join-Path $uiDir "package.json")) {
        Write-Log "Building UI..."
        Set-Location $uiDir
        npm install
        npm run build

        $wwwroot = Join-Path $PublishDir "wwwroot"
        if (Test-Path (Join-Path $uiDir "build")) {
            Copy-Item (Join-Path $uiDir "build\*") -Destination $wwwroot -Recurse -Force
            Write-Log "UI built and copied to wwwroot"
        }
    }

    Set-Location $PSScriptRoot
    Write-Log "Server built successfully"
}

function New-Config {
    $configFile = Join-Path $PublishDir "appsettings.json"

    Write-Log "Creating configuration file..."

    $jwtKey = [Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))

    $config = @{
        ConnectionStrings = @{
            DefaultConnection = "Data Source=$InstallDir\backup.db"
        }
        Jwt               = @{
            Key      = $jwtKey
            Issuer   = "BackupServer"
            Audience = "BackupClients"
        }
        Server            = @{
            PublicUrl = "http://localhost:8000"
        }
        BootstrapAdmin    = @{
            Username = "admin"
            Email    = "admin@backupsystem.com"
            Password = ""
        }
        AllowedOrigins    = @()
        Encryption        = @{
            KeyFilePath = ""
        }
        Serilog           = @{
            MinimumLevel = "Information"
        }
    } | ConvertTo-Json -Depth 10

    $config | Out-File -FilePath $configFile -Encoding UTF8 -Force
    Write-Log "Configuration created at $configFile"
}

function New-WindowsService {
    Write-Log "Creating Windows service..."

    $exePath = Join-Path $PublishDir "Backup.Server.exe"
    $serviceName = "BackupServer"

    # Check if service exists
    $service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
    if ($service) {
        Write-Log "Service already exists, stopping and removing..."
        Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
        sc.exe delete $serviceName | Out-Null
        Start-Sleep -Seconds 2
    }

    # Create service
    $binPath = "`"$exePath`" --urls=http://0.0.0.0:8000"
    sc.exe create $serviceName binPath= $binPath start= auto DisplayName= "Backup Server"
    sc.exe description $serviceName "Enterprise Backup Management Server"
    sc.exe config $serviceName failure= "restart/60000/restart/60000/restart/60000"

    Write-Log "Service created: $serviceName"
}

function Start-ServerService {
    Write-Log "Starting Backup Server..."

    try {
        Start-Service -Name "BackupServer" -ErrorAction Stop
        Start-Sleep -Seconds 5

        $service = Get-Service -Name "BackupServer"
        if ($service.Status -eq "Running") {
            Write-Log "Backup Server started successfully" -ForegroundColor Green
        }
        else {
            Write-Log "Warning: Service may not be running properly. Checking logs..." -ForegroundColor Yellow
        }
    }
    catch {
        Write-Log "Warning: Service could not start automatically. This is normal on first installation." -ForegroundColor Yellow
        Write-Log "The service will start when you manually start it or restart the computer." -ForegroundColor Yellow
        Write-Log "To start manually: Start-Service -Name BackupServer" -ForegroundColor Yellow
        Write-Log ""
        Write-Log "If the service fails to start, check the logs at:" -ForegroundColor Yellow
        Write-Log "  C:\BackupServer\server\publish\logs\" -ForegroundColor White
    }
}

function Uninstall-Server {
    Write-Log "Uninstalling Backup Server..."

    $serviceName = "BackupServer"
    $service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue

    if ($service) {
        Write-Log "Stopping service..."
        Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
        sc.exe delete $serviceName | Out-Null
        Write-Log "Service removed"
    }

    if (Test-Path $InstallDir) {
        Remove-Item $InstallDir -Recurse -Force
        Write-Log "Installation directory removed"
    }

    Write-Log "Backup Server uninstalled successfully" -ForegroundColor Green
}

# Main
if ($Uninstall) {
    Check-Admin
    Uninstall-Server
    exit 0
}

Check-Admin

Write-Log "========================================="
Write-Log "Installing Backup Server v$Version..."
Write-Log "========================================="

Install-DotNet
Install-NodeJS
Build-Server
New-Config
New-WindowsService

if ($AutoStart) {
    Start-ServerService
}

Write-Log ""
Write-Log "=========================================" -ForegroundColor Green
Write-Log "Installation Complete!" -ForegroundColor Green
Write-Log "=========================================" -ForegroundColor Green
Write-Log ""
Write-Log "Access the application:" -ForegroundColor Yellow
Write-Log "  UI: http://localhost" -ForegroundColor White
Write-Log "  API: http://localhost:8000" -ForegroundColor White
Write-Log "  Swagger: http://localhost:8000/swagger" -ForegroundColor White
Write-Log ""
Write-Log "Login credentials:" -ForegroundColor Yellow
Write-Log "  Username: admin" -ForegroundColor White
Write-Log "  Password: Check server logs for bootstrap password" -ForegroundColor White
Write-Log ""
Write-Log "To start server manually:" -ForegroundColor Yellow
Write-Log "  Start-Service -Name BackupServer" -ForegroundColor White
