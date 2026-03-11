# NovaBackup Enterprise Installer
# Veeam-like installation wizard for Windows

param(
    [string]$InstallPath = "C:\Program Files\NovaBackup",
    [string]$DataPath = "C:\ProgramData\NovaBackup",
    [switch]$Silent = $false,
    [switch]$NoService = $false,
    [switch]$NoFirewall = $false
)

#Requires -RunAsAdministrator

$ErrorActionPreference = "Stop"

# Branding
$ProductName = "NovaBackup Enterprise"
$ProductVersion = "6.0.0"
$CompanyName = "NovaSoft"

# Installation components
$Components = @(
    @{ Name = "Core Server"; ID = "CORE"; Default = $true; Size = 150MB; Description = "Main backup server and engine" }
    @{ Name = "Management Console"; ID = "CONSOLE"; Default = $true; Size = 50MB; Description = "Web-based management interface" }
    @{ Name = "VMware Agent"; ID = "VMWARE"; Default = $true; Size = 80MB; Description = "VMware vSphere backup agent" }
    @{ Name = "Hyper-V Agent"; ID = "HYPERV"; Default = $true; Size = 60MB; Description = "Hyper-V backup agent" }
    @{ Name = "SQL Agent"; ID = "SQL"; Default = $false; Size = 40MB; Description = "SQL Server application-aware processing" }
    @{ Name = "Exchange Agent"; ID = "EXCHANGE"; Default = $false; Size = 35MB; Description = "Exchange Server backup agent" }
    @{ Name = "Cloud Gateway"; ID = "CLOUD"; Default = $false; Size = 45MB; Description = "AWS S3, Azure Blob, GCS integration" }
    @{ Name = "PowerShell Module"; ID = "PSMODULE"; Default = $true; Size = 10MB; Description = "PowerShell cmdlets for automation" }
)

function Show-Header {
    Clear-Host
    Write-Host @"
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║                    NovaBackup Enterprise                         ║
║                      Version $ProductVersion                            ║
║                                                                  ║
║          Enterprise Backup & Recovery Solution                   ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
"@ -ForegroundColor Cyan
    Write-Host ""
}

function Show-LicenseAgreement {
    Show-Header
    Write-Host "LICENSE AGREEMENT" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    Write-Host @"
END USER LICENSE AGREEMENT

This software is licensed, not sold. By installing NovaBackup Enterprise,
you agree to the following terms and conditions:

1. GRANT OF LICENSE
   NovaSoft grants you a non-exclusive license to use this software
   on the terms and conditions set forth herein.

2. RESTRICTIONS
   You may not reverse engineer, decompile, or disassemble this software.

3. BACKUP AND RECOVERY
   This software is provided for data backup and recovery purposes.
   You are responsible for ensuring adequate backup procedures.

4. SUPPORT
   Technical support is provided according to your support contract.

5. LIMITATION OF LIABILITY
   NovaSoft shall not be liable for any data loss or business interruption.

Press [Y] to accept and continue, [N] to decline and exit.
"@
    
    $response = Read-Host "Do you accept the license agreement? (Y/N)"
    if ($response -ne 'Y' -and $response -ne 'y') {
        Write-Host "`nInstallation cancelled by user." -ForegroundColor Red
        exit 1
    }
}

function Select-Components {
    Show-Header
    Write-Host "SELECT COMPONENTS" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    Write-Host "Choose the components you want to install:"
    Write-Host ""
    
    $selectedComponents = @()
    
    for ($i = 0; $i -lt $Components.Count; $i++) {
        $comp = $Components[$i]
        $defaultMarker = if ($comp.Default) { "[X]" } else { "[ ]" }
        $sizeStr = "{0:N1} MB" -f ($comp.Size / 1MB)
        
        Write-Host "$($i + 1). $defaultMarker $($comp.Name.PadRight(25)) $sizeStr" -NoNewline
        Write-Host " - $($comp.Description)" -ForegroundColor Gray
    }
    
    Write-Host ""
    Write-Host "Enter component numbers to toggle (comma-separated), or:"
    Write-Host "  [A] Select all    [N] Select none    [C] Continue with defaults"
    Write-Host ""
    
    $response = Read-Host "Your choice"
    
    switch ($response.ToUpper()) {
        'A' { $selectedComponents = $Components }
        'N' { $selectedComponents = @() }
        'C' { $selectedComponents = $Components | Where-Object { $_.Default } }
        default {
            $selectedIndices = $response.Split(',') | ForEach-Object { $_.Trim() }
            foreach ($idx in $selectedIndices) {
                if ([int]::TryParse($idx, [ref]$null)) {
                    $index = [int]$idx - 1
                    if ($index -ge 0 -and $index -lt $Components.Count) {
                        $selectedComponents += $Components[$index]
                    }
                }
            }
            if ($selectedComponents.Count -eq 0) {
                $selectedComponents = $Components | Where-Object { $_.Default }
            }
        }
    }
    
    return $selectedComponents
}

function Select-InstallationType {
    Show-Header
    Write-Host "INSTALLATION TYPE" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    Write-Host "Select installation type:"
    Write-Host ""
    Write-Host "1. Complete Installation" -ForegroundColor Green
    Write-Host "   Install all components on this server"
    Write-Host ""
    Write-Host "2. Backup Server Only" -ForegroundColor Green
    Write-Host "   Install backup engine and web console only"
    Write-Host ""
    Write-Host "3. Remote Agent" -ForegroundColor Green
    Write-Host "   Install agent for backup from remote server"
    Write-Host ""
    Write-Host "4. Custom Installation" -ForegroundColor Green
    Write-Host "   Select individual components to install"
    Write-Host ""
    
    $choice = Read-Host "Enter your choice (1-4)"
    
    switch ($choice) {
        '1' { return $Components }
        '2' { return $Components | Where-Object { $_.ID -in @('CORE', 'CONSOLE', 'PSMODULE') } }
        '3' { return $Components | Where-Object { $_.ID -in @('VMWARE', 'HYPERV', 'SQL', 'EXCHANGE', 'PSMODULE') } }
        '4' { return Select-Components }
        default { return $Components | Where-Object { $_.Default } }
    }
}

function Set-InstallationPaths {
    Show-Header
    Write-Host "INSTALLATION PATHS" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    
    Write-Host "Default installation path: $InstallPath"
    $customPath = Read-Host "Enter custom path or press Enter to accept default"
    if ($customPath) {
        $script:InstallPath = $customPath
    }
    
    Write-Host ""
    Write-Host "Default data path: $DataPath"
    $customDataPath = Read-Host "Enter custom data path or press Enter to accept default"
    if ($customDataPath) {
        $script:DataPath = $customDataPath
    }
    
    # Create directories
    try {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
        New-Item -ItemType Directory -Path $DataPath -Force | Out-Null
        New-Item -ItemType Directory -Path "$DataPath\Logs" -Force | Out-Null
        New-Item -ItemType Directory -Path "$DataPath\Backups" -Force | Out-Null
        New-Item -ItemType Directory -Path "$DataPath\Config" -Force | Out-Null
    }
    catch {
        Write-Error "Failed to create installation directories: $_"
    }
}

function Install-Prerequisites {
    Show-Header
    Write-Host "CHECKING PREREQUISITES" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    
    # Check Windows version
    $osInfo = Get-CimInstance Win32_OperatingSystem
    Write-Host "Operating System: $($osInfo.Caption)" -NoNewline
    if ($osInfo.Version -ge "10.0") {
        Write-Host " [OK]" -ForegroundColor Green
    } else {
        Write-Host " [FAILED]" -ForegroundColor Red
        Write-Error "Windows 10 or Server 2016 or later is required"
    }
    
    # Check PowerShell version
    Write-Host "PowerShell Version: $($PSVersionTable.PSVersion)" -NoNewline
    if ($PSVersionTable.PSVersion.Major -ge 5) {
        Write-Host " [OK]" -ForegroundColor Green
    } else {
        Write-Host " [FAILED]" -ForegroundColor Red
        Write-Error "PowerShell 5.0 or later is required"
    }
    
    # Check .NET Framework
    Write-Host ".NET Framework: Checking..." -NoNewline
    $dotnet = Get-ItemProperty "HKLM:SOFTWARE\Microsoft\NET Framework Setup\NDP\v4\Full\" -ErrorAction SilentlyContinue
    if ($dotnet.Release -ge 461808) {
        Write-Host " [OK]" -ForegroundColor Green
    } else {
        Write-Host " [WARNING]" -ForegroundColor Yellow
        Write-Host ".NET Framework 4.7.2 or later is recommended"
    }
    
    # Check disk space
    $drive = (Get-Item $InstallPath).PSDrive
    $freeSpace = $drive.Free
    $requiredSpace = 500MB
    
    Write-Host "Disk Space Available: $([math]::Round($freeSpace / 1GB, 2)) GB" -NoNewline
    if ($freeSpace -gt $requiredSpace) {
        Write-Host " [OK]" -ForegroundColor Green
    } else {
        Write-Host " [FAILED]" -ForegroundColor Red
        Write-Error "At least 500 MB of free disk space is required"
    }
    
    # Check if ports are available
    Write-Host "Port 8080 (HTTP): Checking..." -NoNewline
    $port8080 = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue
    if (-not $port8080) {
        Write-Host " [OK]" -ForegroundColor Green
    } else {
        Write-Host " [WARNING - In Use]" -ForegroundColor Yellow
    }
    
    Write-Host ""
    Write-Host "All prerequisites checked. Press Enter to continue..."
    Read-Host
}

function Install-Components {
    param($SelectedComponents)
    
    Show-Header
    Write-Host "INSTALLING COMPONENTS" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    
    $totalSize = ($SelectedComponents | Measure-Object -Property Size -Sum).Sum
    Write-Host "Total installation size: $([math]::Round($totalSize / 1MB, 2)) MB"
    Write-Host "Installation path: $InstallPath"
    Write-Host ""
    
    $progress = 0
    $stepSize = 100 / $SelectedComponents.Count
    
    foreach ($comp in $SelectedComponents) {
        $progress += $stepSize
        Write-Progress -Activity "Installing $ProductName" -Status "Installing $($comp.Name)..." -PercentComplete $progress
        
        Write-Host "Installing $($comp.Name)..." -NoNewline
        
        # Simulate installation
        Start-Sleep -Milliseconds 500
        
        # Create component directory
        $compPath = Join-Path $InstallPath $comp.ID
        New-Item -ItemType Directory -Path $compPath -Force | Out-Null
        
        # Create marker file
        "$ProductVersion" | Out-File -FilePath (Join-Path $compPath "version.txt") -Force
        
        Write-Host " [DONE]" -ForegroundColor Green
    }
    
    Write-Progress -Activity "Installing $ProductName" -Completed
    Write-Host ""
    Write-Host "All components installed successfully!" -ForegroundColor Green
}

function Install-Service {
    if ($NoService) { return }
    
    Show-Header
    Write-Host "INSTALLING WINDOWS SERVICE" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    
    $serviceName = "NovaBackupService"
    $serviceDisplayName = "NovaBackup Enterprise Service"
    $servicePath = Join-Path $InstallPath "nova-service.exe"
    
    # Check if service already exists
    $existingService = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
    if ($existingService) {
        Write-Host "Service already exists. Stopping and removing..." -NoNewline
        Stop-Service -Name $serviceName -Force -ErrorAction SilentlyContinue
        sc.exe delete $serviceName | Out-Null
        Write-Host " [DONE]" -ForegroundColor Green
    }
    
    # Create service
    Write-Host "Creating Windows service..." -NoNewline
    New-Service -Name $serviceName -BinaryPathName "$servicePath --service" -DisplayName $serviceDisplayName -StartupType Automatic | Out-Null
    Write-Host " [DONE]" -ForegroundColor Green
    
    # Configure service
    Write-Host "Configuring service..." -NoNewline
    sc.exe failure $serviceName reset= 86400 actions= restart/5000/restart/5000/restart/5000 | Out-Null
    Write-Host " [DONE]" -ForegroundColor Green
    
    # Start service
    Write-Host "Starting service..." -NoNewline
    Start-Service -Name $serviceName
    Write-Host " [DONE]" -ForegroundColor Green
}

function Configure-Firewall {
    if ($NoFirewall) { return }
    
    Show-Header
    Write-Host "CONFIGURING WINDOWS FIREWALL" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    
    $rules = @(
        @{ Name = "NovaBackup-HTTP"; Port = 8080; Protocol = "TCP" }
        @{ Name = "NovaBackup-HTTPS"; Port = 8443; Protocol = "TCP" }
    )
    
    foreach ($rule in $rules) {
        # Remove existing rule
        Remove-NetFirewallRule -DisplayName $rule.Name -ErrorAction SilentlyContinue
        
        Write-Host "Creating firewall rule: $($rule.Name) (Port $($rule.Port))..." -NoNewline
        
        New-NetFirewallRule -DisplayName $rule.Name -Direction Inbound -Protocol $rule.Protocol -LocalPort $rule.Port -Action Allow -Profile Any | Out-Null
        
        Write-Host " [DONE]" -ForegroundColor Green
    }
}

function Register-Application {
    Show-Header
    Write-Host "REGISTERING APPLICATION" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    
    # Create Start Menu shortcut
    $startMenuPath = Join-Path $env:ProgramData "Microsoft\Windows\Start Menu\Programs\NovaBackup"
    New-Item -ItemType Directory -Path $startMenuPath -Force | Out-Null
    
    $WshShell = New-Object -ComObject WScript.Shell
    
    # Management Console shortcut
    $consoleShortcut = Join-Path $startMenuPath "NovaBackup Management Console.lnk"
    $consoleTarget = "http://localhost:8080"
    $consoleSC = $WshShell.CreateShortcut($consoleShortcut)
    $consoleSC.TargetPath = $consoleTarget
    $consoleSC.IconLocation = Join-Path $InstallPath "gui\icon.ico"
    $consoleSC.Save()
    
    Write-Host "Start Menu shortcuts created" -ForegroundColor Green
    
    # Create Desktop shortcut
    $desktopPath = [Environment]::GetFolderPath("Desktop")
    $desktopShortcut = Join-Path $desktopPath "NovaBackup Console.url"
    "[InternetShortcut]`nURL=http://localhost:8080`nIconFile=$InstallPath\gui\icon.ico" | Out-File -FilePath $desktopShortcut -Force
    
    Write-Host "Desktop shortcut created" -ForegroundColor Green
    
    # Register with Windows
    $regPath = "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\NovaBackup"
    New-Item -Path $regPath -Force | Out-Null
    Set-ItemProperty -Path $regPath -Name "DisplayName" -Value $ProductName
    Set-ItemProperty -Path $regPath -Name "DisplayVersion" -Value $ProductVersion
    Set-ItemProperty -Path $regPath -Name "Publisher" -Value $CompanyName
    Set-ItemProperty -Path $regPath -Name "InstallLocation" -Value $InstallPath
    Set-ItemProperty -Path $regPath -Name "UninstallString" -Value "PowerShell.exe -ExecutionPolicy Bypass -File `"$InstallPath\uninstall.ps1`""
    
    Write-Host "Application registered with Windows" -ForegroundColor Green
}

function Initialize-Configuration {
    Show-Header
    Write-Host "INITIALIZING CONFIGURATION" -ForegroundColor Yellow
    Write-Host ("=" * 70)
    Write-Host ""
    
    $configPath = Join-Path $DataPath "Config\config.json"
    
    $config = @{
        server = @{
            http_port = 8080
            https_port = 8443
            bind_address = "0.0.0.0"
        }
        logging = @{
            level = "info"
            file = Join-Path $DataPath "Logs\novabackup.log"
            max_size = "100MB"
            max_backups = 10
        }
        backup = @{
            default_path = Join-Path $DataPath "Backups"
            retention_days = 30
        }
        database = @{
            path = Join-Path $DataPath "novabackup.db"
        }
        security = @{
            jwt_secret = [Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 } | ForEach-Object { [byte]$_ }))
            session_timeout = 24
        }
        installed_components = @()
        installed_date = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        version = $ProductVersion
    }
    
    $config | ConvertTo-Json -Depth 4 | Out-File -FilePath $configPath -Force
    
    Write-Host "Configuration file created: $configPath" -ForegroundColor Green
}

function Show-Summary {
    Show-Header
    Write-Host "INSTALLATION COMPLETE" -ForegroundColor Green
    Write-Host ("=" * 70)
    Write-Host ""
    
    Write-Host "$ProductName v$ProductVersion has been successfully installed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Installation Summary:"
    Write-Host "  Installation Path: $InstallPath"
    Write-Host "  Data Path: $DataPath"
    Write-Host "  Web Console: http://localhost:8080"
    Write-Host "  Service: NovaBackupService"
    Write-Host ""
    Write-Host "Default Credentials:"
    Write-Host "  Username: admin"
    Write-Host "  Password: admin123"
    Write-Host "  (Please change the default password after first login)"
    Write-Host ""
    Write-Host "Getting Started:"
    Write-Host "  1. Open http://localhost:8080 in your web browser"
    Write-Host "  2. Log in with default credentials"
    Write-Host "  3. Add backup jobs and configure storage"
    Write-Host "  4. Start protecting your data!"
    Write-Host ""
    Write-Host "Documentation: https://docs.novabackup.local"
    Write-Host "Support: support@novabackup.local"
    Write-Host ""
    Write-Host "Thank you for choosing $ProductName!" -ForegroundColor Cyan
    Write-Host ""
    
    if (-not $Silent) {
        Write-Host "Press Enter to exit..."
        Read-Host
    }
}

function New-UninstallScript {
    $uninstallScript = @'
# NovaBackup Uninstaller
#Requires -RunAsAdministrator

param([switch]$KeepData = $false)

$ErrorActionPreference = "Stop"

Write-Host "NovaBackup Enterprise Uninstaller" -ForegroundColor Yellow
Write-Host ("=" * 50)
Write-Host ""

$InstallPath = "C:\Program Files\NovaBackup"
$DataPath = "C:\ProgramData\NovaBackup"

# Stop and remove service
Write-Host "Stopping service..." -NoNewline
Stop-Service -Name "NovaBackupService" -Force -ErrorAction SilentlyContinue
sc.exe delete "NovaBackupService" | Out-Null
Write-Host " [DONE]" -ForegroundColor Green

# Remove firewall rules
Write-Host "Removing firewall rules..." -NoNewline
Remove-NetFirewallRule -DisplayName "NovaBackup-*" -ErrorAction SilentlyContinue
Write-Host " [DONE]" -ForegroundColor Green

# Remove shortcuts
Write-Host "Removing shortcuts..." -NoNewline
$startMenuPath = Join-Path $env:ProgramData "Microsoft\Windows\Start Menu\Programs\NovaBackup"
Remove-Item -Path $startMenuPath -Recurse -Force -ErrorAction SilentlyContinue
$desktopPath = [Environment]::GetFolderPath("Desktop")
Remove-Item -Path (Join-Path $desktopPath "NovaBackup Console.url") -Force -ErrorAction SilentlyContinue
Write-Host " [DONE]" -ForegroundColor Green

# Remove registry entries
Write-Host "Removing registry entries..." -NoNewline
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\NovaBackup" -Force -ErrorAction SilentlyContinue
Write-Host " [DONE]" -ForegroundColor Green

# Remove installation files
Write-Host "Removing installation files..." -NoNewline
Remove-Item -Path $InstallPath -Recurse -Force -ErrorAction SilentlyContinue
Write-Host " [DONE]" -ForegroundColor Green

# Remove data (optional)
if (-not $KeepData) {
    Write-Host "Removing data files..." -NoNewline
    Remove-Item -Path $DataPath -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host " [DONE]" -ForegroundColor Green
} else {
    Write-Host "Data files preserved at: $DataPath" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "NovaBackup has been uninstalled." -ForegroundColor Green
'@
    
    $uninstallPath = Join-Path $InstallPath "uninstall.ps1"
    $uninstallScript | Out-File -FilePath $uninstallPath -Force
}

# Main installation flow
function Start-Installation {
    try {
        if (-not $Silent) {
            Show-LicenseAgreement
            $selectedComponents = Select-InstallationType
        } else {
            $selectedComponents = $Components | Where-Object { $_.Default }
        }
        
        Set-InstallationPaths
        Install-Prerequisites
        Install-Components -SelectedComponents $selectedComponents
        Install-Service
        Configure-Firewall
        Initialize-Configuration
        Register-Application
        New-UninstallScript
        Show-Summary
        
        exit 0
    }
    catch {
        Write-Host "`nInstallation failed: $_" -ForegroundColor Red
        Write-Host $_.ScriptStackTrace
        exit 1
    }
}

# Start installation
Start-Installation
