#!/usr/bin/env pwsh
<#
.SYNOPSIS
    NovaBackup Universal Process Manager
    
.DESCRIPTION
    Universal script for managing NovaBackup processes on Windows.
    Supports kill, restart, start, stop, and status operations.
    
.PARAMETER Action
    The action to perform: kill, restart, start, stop, status
    
.PARAMETER Port
    The port NovaBackup is running on (default: 8050)
    
.PARAMETER ProcessName
    The process name to manage (default: nova-backup, novabackup)
    
.PARAMETER WorkingDirectory
    The directory where NovaBackup is installed
    
.EXAMPLE
    .\nova-manager.ps1 -Action kill
    .\nova-manager.ps1 -Action restart -Port 8050
    .\nova-manager.ps1 -Action status
    
.NOTES
    This script replaces multiple legacy scripts:
    - force_kill.ps1, kill_all.bat, kill_server.vbs, kill_pid.vbs
    - restart_server.bat, restart.bat, restart.ps1
    - stop_all.vbs, force_stop_all.vbs, kill_all_nova.vbs
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet('kill', 'restart', 'start', 'stop', 'status', 'install', 'uninstall')]
    [string]$Action,
    
    [int]$Port = 8050,
    
    [string[]]$ProcessNames = @('nova-backup', 'novabackup', 'NovaBackup'),
    
    [string]$WorkingDirectory,
    
    [switch]$Wait,
    
    [int]$TimeoutSeconds = 30
)

# Set working directory
if (-not $WorkingDirectory) {
    $WorkingDirectory = Split-Path -Parent $MyInvocation.MyCommand.Path
}
Set-Location -Path $WorkingDirectory -ErrorAction SilentlyContinue

# Colors for output
function Write-Success { param($Message) Write-Host $Message -ForegroundColor Green }
function Write-Error2 { param($Message) Write-Host $Message -ForegroundColor Red }
function Write-Warning2 { param($Message) Write-Host $Message -ForegroundColor Yellow }
function Write-Info { param($Message) Write-Host $Message -ForegroundColor Cyan }

# Get process by name
function Get-NovaProcess {
    foreach ($name in $ProcessNames) {
        $process = Get-Process -Name $name -ErrorAction SilentlyContinue
        if ($process) {
            return $process
        }
    }
    return $null
}

# Get process by PID
function Get-ProcessByPid {
    param([int]$Pid)
    return Get-Process -Id $Pid -ErrorAction SilentlyContinue
}

# Check if port is in use
function Test-PortInUse {
    param([int]$Port)
    $result = netstat -ano | Select-String ":$Port.*LISTENING"
    return $null -ne $result
}

# Get PID using a port
function Get-PidByPort {
    param([int]$Port)
    $result = netstat -ano | Select-String ":$Port.*LISTENING"
    if ($result) {
        $parts = $result.Line -split '\s+'
        return $parts[-1]
    }
    return $null
}

# Kill process by object
function Stop-NovaProcess {
    param(
        $Process,
        [switch]$Force
    )
    
    if (-not $Process) {
        Write-Warning2 "No process found"
        return $false
    }
    
    try {
        $pid = $Process.Id
        Write-Info "Stopping process: $($Process.Name) (PID: $pid)"
        
        if ($Force) {
            Stop-Process -Id $pid -Force -ErrorAction Stop
        } else {
            Stop-Process -Id $pid -ErrorAction Stop
        }
        
        Write-Success "Process stopped successfully"
        return $true
    }
    catch {
        Write-Error2 "Failed to stop process: $_"
        return $false
    }
}

# Wait for port to be free
function Wait-PortFree {
    param(
        [int]$Port,
        [int]$Timeout = 30
    )
    
    Write-Info "Waiting for port $Port to be free..."
    $startTime = Get-Date
    
    while (Test-PortInUse -Port $Port) {
        $elapsed = (Get-Date) - $startTime
        if ($elapsed.TotalSeconds -gt $Timeout) {
            Write-Error2 "Timeout waiting for port $Port to be free"
            return $false
        }
        Start-Sleep -Seconds 1
    }
    
    Write-Success "Port $Port is now free"
    return $true
}

# Start NovaBackup server
function Start-NovaServer {
    param(
        [string]$Args = 'server',
        [switch]$NoWindow
    )
    
    try {
        $exePath = Join-Path $WorkingDirectory 'nova-backup.exe'
        
        if (-not (Test-Path $exePath)) {
            $exePath = Join-Path $WorkingDirectory 'NovaBackup.exe'
        }
        
        if (-not (Test-Path $exePath)) {
            Write-Error2 "NovaBackup executable not found in $WorkingDirectory"
            return $false
        }
        
        Write-Info "Starting NovaBackup server..."
        
        if ($NoWindow) {
            Start-Process -FilePath $exePath -ArgumentList $Args -WindowStyle Hidden
        } else {
            Start-Process -FilePath $exePath -ArgumentList $Args -WindowStyle Normal
        }
        
        Write-Success "NovaBackup server started"
        
        # Wait for server to be ready
        if ($Wait) {
            Write-Info "Waiting for server to be ready on port $Port..."
            $maxAttempts = 30
            $attempt = 0
            while (-not (Test-PortInUse -Port $Port) -and $attempt -lt $maxAttempts) {
                Start-Sleep -Seconds 1
                $attempt++
            }
            
            if (Test-PortInUse -Port $Port) {
                Write-Success "Server is ready on http://localhost:$Port"
            } else {
                Write-Warning2 "Server may not have started correctly"
            }
        }
        
        return $true
    }
    catch {
        Write-Error2 "Failed to start server: $_"
        return $false
    }
}

# Get status
function Get-NovaStatus {
    Write-Info "=== NovaBackup Status ==="
    
    # Check processes
    $process = Get-NovaProcess
    if ($process) {
        Write-Success "Process running: $($process.Name) (PID: $($process.Id))"
        Write-Host "  CPU: $($process.CPU)% | Memory: $([math]::Round($process.WorkingSet / 1MB, 2)) MB"
    } else {
        Write-Warning2 "No NovaBackup process found"
    }
    
    # Check port
    if (Test-PortInUse -Port $Port) {
        $pid = Get-PidByPort -Port $Port
        Write-Success "Port $Port is in use (PID: $pid)"
    } else {
        Write-Warning2 "Port $Port is free"
    }
    
    # Check executable
    $exePath = Join-Path $WorkingDirectory 'nova-backup.exe'
    if (-not (Test-Path $exePath)) {
        $exePath = Join-Path $WorkingDirectory 'NovaBackup.exe'
    }
    
    if (Test-Path $exePath) {
        $version = (Get-Item $exePath).VersionInfo
        Write-Info "Executable: $exePath"
        Write-Host "  Version: $($version.FileVersion)"
    } else {
        Write-Warning2 "Executable not found"
    }
    
    Write-Host ""
}

# Kill all NovaBackup processes
function Kill-AllNova {
    param([switch]$Force)
    
    Write-Info "Killing all NovaBackup processes..."
    
    $killed = 0
    foreach ($name in $ProcessNames) {
        $processes = Get-Process -Name $name -ErrorAction SilentlyContinue
        foreach ($process in $processes) {
            Write-Info "Killing: $($process.Name) (PID: $($process.Id))"
            Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
            $killed++
        }
    }
    
    # Also kill by port if still in use
    if (Test-PortInUse -Port $Port) {
        $pid = Get-PidByPort -Port $Port
        if ($pid) {
            Write-Info "Killing process on port $Port (PID: $pid)"
            Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
            $killed++
        }
    }
    
    if ($killed -gt 0) {
        Write-Success "Killed $killed process(es)"
        
        if ($Wait) {
            Wait-PortFree -Port $Port -Timeout $TimeoutSeconds
        }
    } else {
        Write-Warning2 "No processes found to kill"
    }
    
    return $true
}

# Install as Windows Service
function Install-Service {
    param(
        [string]$ServiceName = 'NovaBackup',
        [string]$DisplayName = 'NovaBackup Service',
        [string]$Description = 'NovaBackup Enterprise Backup Service'
    )
    
    $exePath = Join-Path $WorkingDirectory 'nova-backup.exe'
    if (-not (Test-Path $exePath)) {
        $exePath = Join-Path $WorkingDirectory 'NovaBackup.exe'
    }
    
    if (-not (Test-Path $exePath)) {
        Write-Error2 "NovaBackup executable not found"
        return $false
    }
    
    try {
        Write-Info "Installing Windows service: $ServiceName"
        
        # Use sc.exe to create service
        $result = & sc.exe create $ServiceName binPath= "`"$exePath`" server" DisplayName= $DisplayName start= auto
        
        if ($LASTEXITCODE -eq 0 -or $result -match 'SUCCESS') {
            Write-Success "Service installed successfully"
            
            # Set description
            & sc.exe description $ServiceName $Description | Out-Null
            
            return $true
        } else {
            Write-Error2 "Failed to install service: $result"
            return $false
        }
    }
    catch {
        Write-Error2 "Failed to install service: $_"
        return $false
    }
}

# Uninstall Windows Service
function Uninstall-Service {
    param([string]$ServiceName = 'NovaBackup')
    
    try {
        Write-Info "Uninstalling Windows service: $ServiceName"
        
        # Stop service first
        $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
        if ($service) {
            Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        }
        
        # Remove service
        $result = & sc.exe delete $ServiceName
        
        if ($LASTEXITCODE -eq 0 -or $result -match 'SUCCESS') {
            Write-Success "Service uninstalled successfully"
            return $true
        } else {
            Write-Warning2 "Service not found or failed to uninstall"
            return $false
        }
    }
    catch {
        Write-Error2 "Failed to uninstall service: $_"
        return $false
    }
}

# Main logic
Write-Info "NovaBackup Manager - Action: $Action"
Write-Host ""

switch ($Action) {
    'kill' {
        Kill-AllNova -Force:$Force
    }
    
    'stop' {
        $process = Get-NovaProcess
        if ($process) {
            Stop-NovaProcess -Process $process -Force:$Wait
            if ($Wait) {
                Wait-PortFree -Port $Port -Timeout $TimeoutSeconds
            }
        } else {
            Write-Warning2 "No process found to stop"
        }
    }
    
    'start' {
        Start-NovaServer -Wait:$Wait
    }
    
    'restart' {
        Write-Info "Restarting NovaBackup..."
        $process = Get-NovaProcess
        if ($process) {
            Stop-NovaProcess -Process $process -Force
            Wait-PortFree -Port $Port -Timeout $TimeoutSeconds
        }
        Start-Sleep -Seconds 2
        Start-NovaServer -Wait:$Wait
    }
    
    'status' {
        Get-NovaStatus
    }
    
    'install' {
        Install-Service
    }
    
    'uninstall' {
        Uninstall-Service
    }
}

exit 0
