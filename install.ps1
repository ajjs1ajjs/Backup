# Backup Agent Installer for Windows
# Version: 1.0.0

param(
    [string]$Server = "",
    [string]$Token = "",
    [string]$AgentType = "hyperv",
    [string]$InstallDir = "C:\Program Files\BackupAgent",
    [string]$Mode = "server",
    [switch]$AutoStart,
    [switch]$Force,
    [switch]$Uninstall,
    [switch]$SkipSSL,
    [string]$SourceUrl = "",
    [string]$LocalSource = ""
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
    -Mode MODE           Installation mode: agent, server, all (default: agent)
    -AutoStart           Start agent after installation
    -Force              Force reinstallation
    -SkipSSL            Skip SSL certificate verification (insecure)
    -SourceUrl URL      Alternative URL for install script
    -LocalSource PATH   Use local source code instead of downloading
    -Uninstall          Uninstall agent

Examples:
    .\install.ps1 -Server "10.0.0.1:8000" -Token "ABCD-1234" -AgentType "hyperv" -AutoStart
    .\install.ps1 -Mode server -InstallDir "C:\BackupServer"
    .\install.ps1 -SkipSSL -Server "10.0.0.1:8000" -Token "ABCD"
    .\install.ps1 -LocalSource "C:\Projects\Backup\src\agent" -Server "10.0.0.1:8000" -Token "ABCD"
    iwr -useb https://get.backupsystem.com/agent/install.ps1 | iex -Server "10.0.0.1:8000" -Token "ABCD"
    iwr -useb -SkipCertificateCheck https://get.backupsystem.com/agent/install.ps1 | iex -SkipSSL -Server "10.0.0.1:8000" -Token "ABCD"

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
    
    # Check VDDK for VMware
    if ($AgentType -eq "vmware") {
        $vddkPath = "${env:ProgramFiles}\VMware\VDDK"
        if (-not (Test-Path $vddkPath)) {
            Write-Log "Warning: VDDK not found. VMware backups may not work."
        }
    }
    
    # Check Hyper-V PowerShell modules for Hyper-V agent
    if ($AgentType -eq "hyperv") {
        try {
            $hypervModule = Get-Module -ListAvailable -Name Hyper-V -ErrorAction SilentlyContinue
            if (-not $hypervModule) {
                Write-Log "Warning: Hyper-V PowerShell module not found. Please install Hyper-V management tools."
            }
        } catch {
            Write-Log "Warning: Could not verify Hyper-V module: $_"
        }
    }
    
    # Check libvirt for KVM
    if ($AgentType -eq "kvm") {
        $libvirt = Get-ChildItem -Path "C:\Windows\System32" -Filter "libvirt.dll" -ErrorAction SilentlyContinue
        if (-not $libvirt) {
            Write-Log "Warning: libvirt not found. KVM backups may not work."
        }
    }
    
    # Check .NET for Server
    if (-not (Get-Command dotnet -ErrorAction SilentlyContinue)) {
        Write-Log "Note: .NET SDK not found - server components require .NET 8.0"
    }
    
    # Check Node.js for UI
    if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
        Write-Log "Note: Node.js not found - UI build requires Node.js 18+"
    }
    
    if ($missing.Count -gt 0) {
        Write-Log "Missing dependencies: $($missing -join ', ')"
        Write-Log "Please install Visual Studio Build Tools with C++ workload"
        Write-Log "Download from: https://visualstudio.microsoft.com/visual-cpp-build-tools/"
        
        # Try to offer automatic install of vcredist
        $vcredistUrl = "https://aka.ms/vs/17/release/vc_redist.x64.exe"
        $vcredistPath = Join-Path $env:TEMP "vc_redist.x64.exe"
        
        Write-Log "Attempting to install Visual C++ Redistributable..."
        try {
            Invoke-WebRequest -Uri $vcredistUrl -OutFile $vcredistPath -UseBasicParsing
            Start-Process -FilePath $vcredistPath -Args "/quiet /norestart" -Wait
            Write-Log "Visual C++ Redistributable installed"
            Remove-Item $vcredistPath -Force -ErrorAction SilentlyContinue
        } catch {
            Write-Log "Warning: Could not auto-install vcredist. Please install manually."
        }
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

function Clone-OrUpdate-Repo {
    $repoUrl = "https://github.com/ajjs1ajjs/Backup.git"
    $cloneRoot = Split-Path -Parent $InstallDir
    $repoDir = Join-Path $cloneRoot "Backup"
    
    if (Test-Path (Join-Path $repoDir ".git")) {
        Write-Log "Updating repository..."
        Set-Location $repoDir
        git pull
    } else {
        Write-Log "Cloning repository..."
        if (Test-Path $repoDir) { Remove-Item $repoDir -Recurse -Force }
        git clone $repoUrl $repoDir
        Set-Location $repoDir
    }
    
    return $repoDir
}

function Install-Server {
    Write-Log "Installing Backup Server..."
    
    $serverDir = Join-Path $InstallDir "server"
    $uiDir = Join-Path $serverDir "ui"
    
    # Clone or update repository
    $projectRoot = Clone-OrUpdate-Repo
    
    # Check and install .NET SDK
    $dotnetPath = $null
    if (Get-Command dotnet -ErrorAction SilentlyContinue) {
        $dotnetPath = (Get-Command dotnet).Source
    } else {
        Write-Log "Installing .NET SDK 8.0..."
        $dotnetInstallScript = Join-Path $env:TEMP "dotnet-install.ps1"
        try {
            Invoke-WebRequest -Uri "https://dot.net/v1/dotnet-install.ps1" -OutFile $dotnetInstallScript -UseBasicParsing
            & $dotnetInstallScript -Channel 8.0 -InstallDir "C:\Program Files\dotnet"
            $dotnetPath = "C:\Program Files\dotnet\dotnet.exe"
        } catch {
            Write-Log "Warning: Could not install .NET SDK. Please install manually from https://dotnet.microsoft.com/download"
        }
    }
    
    # Ensure server directory exists
    if (-not (Test-Path $serverDir)) {
        New-Item -ItemType Directory -Path $serverDir -Force | Out-Null
    }
    
    $serverProject = Join-Path $projectRoot "src\server\Backup.Server\Backup.Server.csproj"
    
    if (Test-Path $serverProject) {
        Write-Log "Building server..."
        if ($dotnetPath) {
            & $dotnetPath restore $serverProject
            & $dotnetPath publish $serverProject -c Release -o (Join-Path $serverDir "publish")
        }
    } else {
        Write-Log "Warning: Server source not found, skipping server build"
    }
    
    # Build UI if Node.js is available
    if (Get-Command node -ErrorAction SilentlyContinue) {
        $uiProject = Join-Path $projectRoot "src\ui\package.json"
        if (Test-Path $uiProject) {
            Write-Log "Building UI..."
            Set-Location (Join-Path $projectRoot "src\ui")
            npm install
            npm run build
            if (Test-Path "build") {
                $publishDir = Join-Path $serverDir "publish"
                $wwwroot = Join-Path $publishDir "wwwroot"
                Copy-Item "build/*" -Destination $wwwroot -Recurse -Force
            }
            Set-Location $projectRoot
        }
    } else {
        Write-Log "Warning: Node.js not found, skipping UI build"
    }
    
    # Restart service
    Write-Log "Restarting service..."
    if (Get-Service -Name "BackupServer" -ErrorAction SilentlyContinue) {
        Restart-Service -Name "BackupServer" -Force
    }
    
    Write-Log "Server installation complete"
}

# Main
if ($Uninstall) {
    Check-Admin
    Uninstall-Agent
    exit 0
}

# Handle mode parameter
if ($Mode -eq "server" -or $Mode -eq "all") {
    Check-Admin
    Check-Dependencies
    Install-Server
    Write-Log "Server installation completed!"
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
