# Backup Server Installer for Windows
# Version: 1.0.0

param(
    [string]$InstallDir = "C:\BackupServer",
    [string]$JwtKey = "",
    [string]$Port = "8000",
    [string]$AdminPassword = "admin123",
    [switch]$AutoStart,
    [switch]$Force,
    [switch]$Uninstall,
    [switch]$SkipBuild,
    [string]$LocalSource = ""
)

$ErrorActionPreference = "Stop"
$Version = "1.0.0"
$ServiceName = "BackupServer"
$PublishDir = Join-Path $InstallDir "publish"

function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] $Message"
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

function Show-Help {
@"
Backup Server Installer v$Version

Usage: .\install-server.ps1 [OPTIONS]

Options:
    -InstallDir DIR      Installation directory (default: C:\BackupServer)
    -JwtKey KEY          JWT secret key (auto-generated if not provided)
    -Port PORT           Server port (default: 8000)
    -AdminPassword PWD   Admin password (default: admin123)
    -AutoStart           Start service after installation
    -Force               Force reinstallation
    -SkipBuild           Skip build, use existing publish folder
    -LocalSource PATH    Use local source code instead of downloading
    -Uninstall           Uninstall server

Examples:
    .\install-server.ps1 -AutoStart
    .\install-server.ps1 -InstallDir "D:\Backup" -Port 9000 -AutoStart
    .\install-server.ps1 -Uninstall

"@
    exit 0
}

function Check-Admin {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        Write-Error "This script must be run as Administrator"
    }
}

function Install-DotNet {
    $dotnetOk = $false
    if (Get-Command dotnet -ErrorAction SilentlyContinue) {
        try {
            $sdkOutput = dotnet --list-sdks 2>&1
            if ($sdkOutput -match '\d+\.\d+') {
                $dotnetOk = $true
                Write-Log ".NET SDK already installed"
            }
        } catch {}
    }

    if ($dotnetOk) { return }

    Write-Log "Installing .NET SDK 8.0..."
    $scriptPath = Join-Path $env:TEMP "dotnet-install.ps1"
    if (-not (Test-Path $scriptPath)) {
        Write-Log "Downloading dotnet-install.ps1..."
        Invoke-WebRequest -Uri "https://dot.net/v1/dotnet-install.ps1" -OutFile $scriptPath -UseBasicParsing
    }

    & $scriptPath -Channel 8.0 -InstallDir "C:\Program Files\dotnet"
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    Write-Log ".NET SDK installed"
}

function Install-Git {
    if (Get-Command git -ErrorAction SilentlyContinue) {
        return
    }

    Write-Log "Installing Git..."
    $gitUrl = "https://github.com/git-for-windows/git/releases/download/v2.44.0.windows.1/Git-2.44.0-64-bit.exe"
    $gitInstaller = Join-Path $env:TEMP "git-installer.exe"

    Invoke-WebRequest -Uri $gitUrl -OutFile $gitInstaller -UseBasicParsing
    Start-Process -FilePath $gitInstaller -Args "/VERYSILENT /NORESTART" -Wait
    Remove-Item $gitInstaller -Force

    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    Write-Log "Git installed"
}

function Install-Node {
    if (Get-Command node -ErrorAction SilentlyContinue) {
        return
    }

    Write-Log "Installing Node.js 18..."
    $nodeUrl = "https://nodejs.org/dist/v18.20.4/node-v18.20.4-x64.msi"
    $nodeInstaller = Join-Path $env:TEMP "node-installer.msi"

    Invoke-WebRequest -Uri $nodeUrl -OutFile $nodeInstaller -UseBasicParsing
    Start-Process -FilePath "msiexec.exe" -Args "/i `"$nodeInstaller`" /quiet /norestart" -Wait
    Remove-Item $nodeInstaller -Force

    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    Write-Log "Node.js installed"
}

function Generate-JwtKey {
    if ($JwtKey) { return }

    $bytes = New-Object byte[] 64
    [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
    $script:JwtKey = [Convert]::ToBase64String($bytes)
    Write-Log "Generated JWT key"
}

function Build-Server {
    if ($SkipBuild) {
        if (-not (Test-Path (Join-Path $PublishDir "Backup.Server.exe"))) {
            Write-Error "SkipBuild specified but publish folder not found"
        }
        return
    }

    Write-Log "Building self-contained server..."

    $sourceDir = if ($LocalSource) {
        $LocalSource
    } else {
        $buildDir = Join-Path $env:TEMP "backup-server-build"
        if (Test-Path $buildDir) { Remove-Item $buildDir -Recurse -Force }

        Write-Log "Cloning repository..."
        git clone https://github.com/ajjs1ajjs/Backup.git $buildDir
        Join-Path $buildDir (Join-Path "src" (Join-Path "server" "Backup.Server"))
    }

    if (-not (Test-Path $sourceDir)) {
        Write-Error "Source directory not found: $sourceDir"
    }

    Push-Location $sourceDir
    try {
        dotnet restore
        dotnet publish -c Release -r win-x64 --self-contained true `
            -p:PublishSingleFile=true `
            -p:EnableCompressionInSingleFile=true `
            -p:IncludeNativeLibrariesForSelfExtract=true `
            -o $PublishDir
    } finally {
        Pop-Location
    }

    Write-Log "Server built: $(Join-Path $PublishDir 'Backup.Server.exe')"
}

function Build-UI {
    Write-Log "Building UI..."

    $uiSource = if ($LocalSource) {
        Join-Path (Split-Path $LocalSource -Parent) "ui"
    } else {
        $buildDir = Join-Path $env:TEMP "backup-server-build"
        Join-Path $buildDir (Join-Path "src" "ui")
    }

    if (-not (Test-Path $uiSource)) {
        Write-Log "UI source not found, skipping UI build"
        return
    }

    Push-Location $uiSource
    try {
        npm install --production
        npm run build

        $wwwroot = Join-Path $PublishDir "wwwroot"
        if (Test-Path (Join-Path $uiSource "build")) {
            if (Test-Path $wwwroot) { Remove-Item $wwwroot -Recurse -Force }
            Copy-Item (Join-Path $uiSource "build") $wwwroot -Recurse -Force
        }
    } finally {
        Pop-Location
    }

    Write-Log "UI built"
}

function New-Config {
    Write-Log "Creating configuration..."

    $dbPath = Join-Path $InstallDir "backup.db"
    $publicUrl = "http://$((Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.InterfaceAlias -notmatch 'Loopback' } | Select-Object -First 1).IPAddress):$Port"

    $config = @{
        ConnectionStrings = @{
            DefaultConnection = "Data Source=$dbPath"
        }
        Jwt = @{
            Key = $JwtKey
            Issuer = "BackupServer"
            Audience = "BackupClients"
        }
        Server = @{
            PublicUrl = $publicUrl
        }
        BootstrapAdmin = @{
            Username = "admin"
            Email = "admin@backupsystem.com"
            Password = $AdminPassword
        }
        AllowedOrigins = @()
        Encryption = @{
            KeyFilePath = ""
        }
        Serilog = @{
            MinimumLevel = "Information"
        }
    } | ConvertTo-Json -Depth 4

    $configPath = Join-Path $PublishDir "appsettings.json"
    $config | Out-File -FilePath $configPath -Encoding UTF8
    Write-Log "Configuration saved to $configPath"
}

function Install-Service {
    Write-Log "Installing Windows Service..."

    $existing = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($existing) {
        Write-Log "Stopping existing service..."
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
        sc.exe delete $ServiceName | Out-Null
        Start-Sleep -Seconds 1
    }

    $exePath = Join-Path $PublishDir "Backup.Server.exe"
    if (-not (Test-Path $exePath)) {
        Write-Error "Server executable not found: $exePath"
    }

    sc.exe create $ServiceName binPath= "`"$exePath`" --service" start= auto DisplayName= "Backup Server"
    sc.exe description $ServiceName "Enterprise Backup Management Server"
    sc.exe failure $ServiceName reset= 0 actions= restart/5000/restart/5000/restart/5000

    Write-Log "Service created: $ServiceName"
}

function Start-Service-Custom {
    Write-Log "Starting service..."
    Start-Service -Name $ServiceName -ErrorAction Stop
    Start-Sleep -Seconds 3

    $service = Get-Service -Name $ServiceName
    if ($service.Status -eq "Running") {
        Write-Log "Service started successfully"
    } else {
        Write-Error "Failed to start service. Check logs at: $(Join-Path $PublishDir 'logs')"
    }
}

function Uninstall-Server {
    Write-Log "Uninstalling Backup Server..."

    $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($service) {
        Write-Log "Stopping service..."
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
        sc.exe delete $ServiceName | Out-Null
        Write-Log "Service removed"
    }

    if (Test-Path $InstallDir) {
        Remove-Item $InstallDir -Recurse -Force
        Write-Log "Installation directory removed: $InstallDir"
    }

    Write-Log "Server uninstalled successfully"
}

# Main
if ($Uninstall) {
    Check-Admin
    Uninstall-Server
    exit 0
}

Check-Admin

Write-Log ""
Write-Log "========================================="
Write-Log "Installing Backup Server v$Version..."
Write-Log "========================================="

Generate-JwtKey
Install-DotNet
Install-Git
Install-Node
Build-Server
Build-UI
New-Config
Install-Service

if ($AutoStart) {
    Start-Service-Custom
}

Write-Log ""
Write-Log "========================================="
Write-Log "Installation Complete!"
Write-Log "========================================="
Write-Log ""
Write-Log "Access the application:"
Write-Log "  API: http://localhost:$Port"
Write-Log "  Swagger: http://localhost:$Port/swagger"
Write-Log ""
Write-Log "Login credentials:"
Write-Log "  Username: admin"
Write-Log "  Password: $AdminPassword"
Write-Log ""
Write-Log "Service management:"
Write-Log "  Start:  Start-Service -Name $ServiceName"
Write-Log "  Stop:   Stop-Service -Name $ServiceName"
Write-Log "  Status: Get-Service -Name $ServiceName"
Write-Log "  Logs:   $(Join-Path $PublishDir 'logs')"
Write-Log ""
Write-Log "IMPORTANT: Change password on first login!"
Write-Log ""
