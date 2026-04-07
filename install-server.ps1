# Backup Server Installer for Windows
# Version: 1.0.0

param(
    [string]$InstallDir = "C:\BackupServer",
    [int]$Port = 8000,
    [string]$AdminPassword = "",
    [string]$LocalSource = "",
    [switch]$AutoStart,
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"
$Version = "1.0.0"

function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $color = switch ($Level) {
        "ERROR" { "Red" }
        "WARN" { "Yellow" }
        "SUCCESS" { "Green" }
        default { "White" }
    }
    Write-Host "[$timestamp] $Message" -ForegroundColor $color
}

function Write-Error {
    param([string]$Message)
    Write-Log $Message "ERROR"
    exit 1
}

function Check-Admin {
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        Write-Error "This script requires Administrator privileges. Please run as Administrator."
    }
}

function Check-Dependencies {
    Write-Log "Checking dependencies..."
    
    # Check .NET SDK
    $dotnet = Get-Command dotnet -ErrorAction SilentlyContinue
    if (-not $dotnet) {
        Write-Log ".NET SDK not found. Installing..."
        $dotnetInstallUrl = "https://dot.net/v1/dotnet-install.ps1"
        $tempScript = Join-Path $env:TEMP "dotnet-install.ps1"
        try {
            Invoke-WebRequest -Uri $dotnetInstallUrl -OutFile $tempScript -UseBasicParsing
            & $tempScript -Channel 8.0 -InstallDir "$env:LOCALAPPDATA\Microsoft\dotnet"
            Remove-Item $tempScript -Force
            Write-Log ".NET 8.0 installed" "SUCCESS"
        } catch {
            Write-Error "Failed to install .NET SDK: $_"
        }
    } else {
        $dotnetVersion = & dotnet --version 2>$null
        Write-Log ".NET SDK found: $dotnetVersion" "SUCCESS"
    }
    
    # Check Node.js
    $node = Get-Command node -ErrorAction SilentlyContinue
    if (-not $node) {
        Write-Log "Node.js not found. UI build may fail." "WARN"
    } else {
        $nodeVersion = & node --version 2>$null
        Write-Log "Node.js found: $nodeVersion" "SUCCESS"
    }
    
    # Check Git
    $git = Get-Command git -ErrorAction SilentlyContinue
    if (-not $git) {
        Write-Log "Git not found. Downloading source..." "WARN"
    } else {
        Write-Log "Git found" "SUCCESS"
    }
}

function Install-Server {
    Write-Log "Installing Backup Server to $InstallDir..."
    
    $publishDir = Join-Path $InstallDir "publish"
    $srcDir = Join-Path $InstallDir "src"
    
    # Create directories
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    New-Item -ItemType Directory -Path $publishDir -Force | Out-Null
    
    if ($LocalSource -and (Test-Path $LocalSource)) {
        Write-Log "Using local source: $LocalSource"
        $projectRoot = $LocalSource
    } else {
        # Clone repository
        $projectRoot = Join-Path $env:TEMP "Backup-$(Get-Random)"
        Write-Log "Cloning repository..."
        git clone --depth 1 https://github.com/ajjs1ajjs/Backup.git $projectRoot
    }
    
    $serverProject = Join-Path $projectRoot "src\server\Backup.Server\Backup.Server.csproj"
    
    if (-not (Test-Path $serverProject)) {
        Write-Error "Server project not found at $serverProject"
    }
    
    # Build server using absolute paths
    Write-Log "Building server..."
    
    $env:DOTNET_CLI_TELEMETRY_OPTOUT = "1"
    & dotnet restore $serverProject --verbosity quiet 2>$null
    & dotnet publish $serverProject -c Release -o $publishDir --self-contained false --verbosity quiet 2>$null
    
    # Copy wwwroot
    $wwwrootSrc = Join-Path $projectRoot "src\ui\build"
    $wwwrootDest = Join-Path $publishDir "wwwroot"
    
    if (Test-Path $wwwrootSrc) {
        Write-Log "Copying UI files..."
        Copy-Item -Path $wwwrootSrc -Destination $wwwrootDest -Recurse -Force
    }
    
    # Configure appsettings
    $appsettingsPath = Join-Path $publishDir "appsettings.json"
    $appsettings = Get-Content $appsettingsPath -Raw | ConvertFrom-Json
    
    if ($AdminPassword) {
        $appsettings.BootstrapAdmin.Password = $AdminPassword
    }
    
    $appsettings.Server.PublicUrl = "http://localhost:$Port"
    $appsettings | ConvertTo-Json -Depth 10 | Set-Content $appsettingsPath -Encoding UTF8
    
    Write-Log "Server installed successfully" "SUCCESS"
    
    # Cleanup temp
    if (-not $LocalSource) {
        Remove-Item $projectRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
    
    return $publishDir
}

function New-WindowsService {
    param([string]$ExePath)
    
    Write-Log "Creating Windows service..."
    
    $serviceName = "BackupServer"
    
    # Check if service exists
    $service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
    
    if ($service) {
        Write-Log "Service already exists, stopping..."
        Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
        sc.exe delete $serviceName | Out-Null
        Start-Sleep -Seconds 2
    }
    
    # Create service
    $exeFullPath = Join-Path $ExePath "Backup.Server.exe"
    $binPath = "`"$exeFullPath`" --urls http://localhost:$Port"
    sc.exe create $serviceName binPath= $binPath start= demand DisplayName= "Backup Server"
    Start-Sleep -Seconds 1
    sc.exe description $serviceName "Backup System Server Service"
    sc.exe config $serviceName failurecnt= 3 failure= "restart/60000/restart/60000/restart/60000"
    
    Write-Log "Service created: $serviceName" "SUCCESS"
    
    Start-Sleep -Seconds 3
}

function Start-ServerService {
    param([string]$ExePath)
    
    Write-Log "Starting Backup Server..."
    Start-Sleep -Seconds 2
    
    $service = Get-Service -Name "BackupServer" -ErrorAction SilentlyContinue
    if ($service) {
        Start-Service -Name "BackupServer" -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 3
    }
    
    if (-not $service -or $service.Status -ne "Running") {
        Write-Log "Service not found or failed. Attempting to start manually..."
        $exeFullPath = Join-Path $ExePath "Backup.Server.exe"
        if (Test-Path $exeFullPath) {
            Start-Process -FilePath $exeFullPath -ArgumentList "--urls http://localhost:$Port" -WindowStyle Hidden
            Start-Sleep -Seconds 3
            Write-Log "Server started manually on http://localhost:$Port" "SUCCESS"
        } else {
            Write-Log "Server executable not found at $exeFullPath" "ERROR"
            return
        }
    }
    
    $service = Get-Service -Name "BackupServer" -ErrorAction SilentlyContinue
    if ($service -and $service.Status -eq "Running") {
        Write-Log "Server is running on http://localhost:$Port" "SUCCESS"
    }
}

function Uninstall-Server {
    Write-Log "Uninstalling Backup Server..."
    
    $serviceName = "BackupServer"
    
    $service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
    if ($service) {
        Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
        sc.exe delete $serviceName | Out-Null
        Write-Log "Service removed" "SUCCESS"
    }
    
    if (Test-Path $InstallDir) {
        Remove-Item $InstallDir -Recurse -Force
        Write-Log "Installation directory removed" "SUCCESS"
    }
}

# Main
Check-Admin

if ($Uninstall) {
    Uninstall-Server
    exit 0
}

Check-Dependencies
$publishDir = Install-Server

if ($AutoStart) {
    New-WindowsService -ExePath $publishDir
    Start-ServerService -ExePath $publishDir
}

Write-Log ""
Write-Log "======================================" "SUCCESS"
Write-Log "Backup Server installed successfully!" "SUCCESS"
Write-Log "======================================" "SUCCESS"
Write-Log ""
Write-Log "Access UI: http://localhost:$Port"
Write-Log "Login: admin / $AdminPassword"
Write-Log ""