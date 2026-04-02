# Backup Agent Installer for Windows
# Version: 1.0.0

param(
    [string]$Server = "",
    [string]$Token = "",
    [string]$AgentType = "hyperv",
    [string]$InstallDir = "C:\Program Files\BackupAgent",
    [switch]$AutoStart,
    [switch]$Force,
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"
$Version = "1.0.0"

$BinDir = Join-Path $InstallDir "bin"
$ConfigDir = Join-Path $InstallDir "config"
$LogDir = Join-Path $InstallDir "log"
$DataDir = Join-Path $InstallDir "data"

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
Backup Agent Installer v$Version

Usage: .\install.ps1 [OPTIONS]

Options:
    -Server ADDR         Management server address (host:port)
    -Token TOKEN         Agent registration token
    -AgentType TYPE      Agent type: hyperv, vmware, kvm, mssql, postgres, oracle
    -InstallDir DIR      Installation directory (default: $InstallDir)
    -AutoStart           Start agent after installation
    -Force              Force reinstallation

Examples:
    .\install.ps1 -Server "10.0.0.1:50051" -Token "ABCD-1234" -AgentType "hyperv" -AutoStart
    iwr -useb https://get.backupsystem.com/agent/install.ps1 | iex -Server "10.0.0.1:50051" -Token "ABCD"

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

function Check-Dependencies {
    Write-Log "Checking dependencies..."
    
    $missing = @()
    
    # Check Visual Studio Build Tools / MSVC
    $msvcPaths = @(
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio",
        "${env:ProgramFiles}\Microsoft Visual Studio",
        "${env:ProgramFiles(x86)}\Microsoft Visual Studio\2022"
    )
    $msvcFound = $false
    foreach ($path in $msvcPaths) {
        if (Test-Path $path) {
            $msvcFound = $true
            break
        }
    }
    
    # Check CMake
    if (-not (Get-Command cmake -ErrorAction SilentlyContinue)) {
        $missing += "CMake"
    }
    
    # Check required libraries
    $libs = @("libssl", "libcurl", "libxml2", "libzstd")
    foreach ($lib in $libs) {
        $found = Get-ChildItem -Path "C:\Windows\System32" -Filter "$lib*.dll" -ErrorAction SilentlyContinue
        if (-not $found) {
            $missing += $lib
        }
    }
    
    if ($missing.Count -gt 0) {
        Write-Log "Missing dependencies: $($missing -join ', ')"
        Write-Log "Please install Visual Studio Build Tools with C++ workload"
        Write-Log "Download from: https://visualstudio.microsoft.com/visual-cpp-build-tools/"
    }
    
    Write-Log "Dependency check complete"
}

function New-Directories {
    Write-Log "Creating directories..."
    
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    foreach ($dir in @($BinDir, $ConfigDir, $LogDir, $DataDir)) {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
    }
    
    Write-Log "Directories created"
}

function Build-Agent {
    Write-Log "Building agent..."
    
    $srcDir = Join-Path $env:TEMP "backup-agent-build"
    
    if (Test-Path (Join-Path $BinDir "backup-agent.exe")) {
        if ($Force) {
            Write-Log "Force reinstall - rebuilding..."
        } else {
            Write-Log "Agent already installed. Use -Force to reinstall"
            return
        }
    }
    
    # Check if source exists
    $projectRoot = Split-Path -Parent (Split-Path -Parent (Get-Location))
    $agentSrc = Join-Path $projectRoot "src\agent\Backup.Agent"
    
    if (-not (Test-Path $agentSrc)) {
        Write-Log "Source not found at $agentSrc"
        Write-Log "Please ensure source code is available"
        return
    }
    
    # Build
    $buildDir = Join-Path $srcDir "build"
    Remove-Item $buildDir -Recurse -Force -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Path $buildDir -Force | Out-Null
    
    Write-Log "Compiling agent..."
    Set-Location $buildDir
    
    cmake .. -G "Visual Studio 17 2022" -A x64 -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=$InstallDir
    cmake --build . --config Release
    
    if (Test-Path (Join-Path $buildDir "backup-agent.exe")) {
        Copy-Item (Join-Path $buildDir "backup-agent.exe") $BinDir -Force
        Write-Log "Agent built successfully"
    } else {
        Write-Log "Build may have failed - checking for alternative..."
    }
    
    Set-Location $projectRoot
}

function New-Config {
    $configFile = Join-Path $ConfigDir "agent.conf"
    
    @"
# Backup Agent Configuration
server=$Server
token=$Token
agent_type=$AgentType
log_dir=$LogDir
data_dir=$DataDir
log_level=info
"@ | Out-File -FilePath $configFile -Encoding UTF8
    
    Write-Log "Configuration generated at $configFile"
}

function New-WindowsService {
    Write-Log "Creating Windows service..."
    
    $exePath = Join-Path $BinDir "backup-agent.exe"
    $configPath = Join-Path $ConfigDir "agent.conf"
    
    # Check if service exists
    $service = Get-Service -Name "BackupAgent" -ErrorAction SilentlyContinue
    
    if ($service) {
        Write-Log "Service already exists, stopping..."
        Stop-Service -Name "BackupAgent" -Force -ErrorAction SilentlyContinue
        sc.exe delete "BackupAgent" | Out-Null
        Start-Sleep -Seconds 2
    }
    
    # Create service
    $binPath = "`"$exePath`" --config `"$configPath`""
    sc.exe create "BackupAgent" binPath= $binPath start= demand DisplayName= "Backup Agent"
    sc.exe description "BackupAgent" "Backup Agent Service for backup operations"
    sc.exe config "BackupAgent" failurecnt= 3 failure= "restart/60000/restart/60000/restart/60000"
    
    Write-Log "Service created: BackupAgent"
}

function Test-Installation {
    Write-Log "Verifying installation..."
    
    $exePath = Join-Path $BinDir "backup-agent.exe"
    
    if (-not (Test-Path $exePath)) {
        Write-Error "Binary not found at $exePath"
    }
    
    Write-Log "Installation verified"
}

function Start-Agent {
    Write-Log "Starting agent..."
    
    Start-Service -Name "BackupAgent" -ErrorAction Stop
    
    Start-Sleep -Seconds 2
    
    $service = Get-Service -Name "BackupAgent"
    if ($service.Status -eq "Running") {
        Write-Log "Agent started successfully"
        $service | Format-List
    } else {
        Write-Error "Failed to start agent"
    }
}

function Uninstall-Agent {
    Write-Log "Uninstalling agent..."
    
    $service = Get-Service -Name "BackupAgent" -ErrorAction SilentlyContinue
    if ($service) {
        Write-Log "Stopping service..."
        Stop-Service -Name "BackupAgent" -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
        sc.exe delete "BackupAgent" | Out-Null
        Write-Log "Service removed"
    }
    
    if (Test-Path $InstallDir) {
        Remove-Item $InstallDir -Recurse -Force
        Write-Log "Installation directory removed"
    }
    
    Write-Log "Agent uninstalled successfully"
}

# Main
if ($Uninstall) {
    Check-Admin
    Uninstall-Agent
    exit 0
}

if ($Server -eq "" -or $Token -eq "") {
    Write-Log "Server and Token are required"
    Show-Help
}

Check-Admin
Check-Dependencies
New-Directories
Build-Agent
New-Config
New-WindowsService
Test-Installation

if ($AutoStart) {
    Start-Agent
} else {
    Write-Log "Installation complete. To start agent manually:"
    Write-Log "  Start-Service -Name BackupAgent"
}

Write-Log "Installation completed!"
Write-Log "Agent installed at: $InstallDir"
Write-Log "Config: $ConfigDir\agent.conf"
