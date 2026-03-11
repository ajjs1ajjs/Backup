# NovaBackup v6.0 - Background Agent
# Works as Windows Service for scheduled backups and restores

param(
    [Parameter(Mandatory=$false)]
    [string]$Mode = "Service",
    
    [Parameter(Mandatory=$false)]
    [string]$ConfigFile = "C:\Program Files\NovaBackup\config.json",
    
    [Parameter(Mandatory=$false)]
    [switch]$Install,
    
    [Parameter(Mandatory=$false)]
    [switch]$Uninstall
)

# Global variables
$script:Config = $null
$script:Jobs = @()
$script:RunningJobs = @{}
$script:LogFile = "C:\Program Files\NovaBackup\logs\agent.log"
$script:ServiceName = "NovaBackupAgent"

# Ensure log directory exists
if (!(Test-Path "C:\Program Files\NovaBackup\logs")) {
    New-Item -ItemType Directory -Path "C:\Program Files\NovaBackup\logs" -Force
}

function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logEntry = "[$timestamp] [$Level] $Message"
    
    try {
        Add-Content -Path $script:LogFile -Value $logEntry -ErrorAction SilentlyContinue
        Write-Host $logEntry
    } catch {
        Write-Host "Failed to write to log: $_"
    }
}

function Load-Config {
    try {
        if (Test-Path $ConfigFile) {
            $script:Config = Get-Content $ConfigFile | ConvertFrom-Json
            Write-Log "Configuration loaded from $ConfigFile"
        } else {
            # Default configuration
            $script:Config = @{
                "DefaultStoragePath" = "D:\NovaBackups"
                "MaxStorageGB" = 1000
                "CompressionEnabled" = $true
                "EncryptionEnabled" = $false
                "DeduplicationEnabled" = $true
                "RetentionDays" = 30
                "EmailNotifications" = $false
                "EmailSettings" = @{
                    "SMTP" = ""
                    "From" = ""
                    "To" = ""
                }
            }
            Write-Log "Using default configuration"
        }
    } catch {
        Write-Log "Failed to load configuration: $_" "ERROR"
    }
}

function Save-Config {
    try {
        $script:Config | ConvertTo-Json -Depth 10 | Set-Content -Path $ConfigFile
        Write-Log "Configuration saved to $ConfigFile"
    } catch {
        Write-Log "Failed to save configuration: $_" "ERROR"
    }
}

function Load-Jobs {
    try {
        $jobsFile = "C:\Program Files\NovaBackup\jobs.json"
        if (Test-Path $jobsFile) {
            $script:Jobs = Get-Content $jobsFile | ConvertFrom-Json
            Write-Log "Loaded $($script:Jobs.Count) backup jobs"
        } else {
            $script:Jobs = @()
            Write-Log "No existing jobs found"
        }
    } catch {
        Write-Log "Failed to load jobs: $_" "ERROR"
    }
}

function Save-Jobs {
    try {
        $script:Jobs | ConvertTo-Json -Depth 10 | Set-Content -Path "C:\Program Files\NovaBackup\jobs.json"
        Write-Log "Jobs saved to database"
    } catch {
        Write-Log "Failed to save jobs: $_" "ERROR"
    }
}

function Start-Backup {
    param([object]$Job)
    
    $jobId = $Job.Id
    if ($script:RunningJobs.ContainsKey($jobId)) {
        Write-Log "Job '$($Job.Name)' is already running" "WARNING"
        return
    }
    
    $script:RunningJobs[$jobId] = @{
        "Name" = $Job.Name
        "StartTime" = Get-Date
        "Status" = "Running"
        "Progress" = 0
    }
    
    Write-Log "Starting backup job: $($Job.Name)"
    
    try {
        # Simulate backup process
        $sourcePath = $Job.SourcePath
        $destPath = $Job.DestinationPath
        $jobName = $Job.Name
        
        # Create destination if not exists
        if (!(Test-Path $destPath)) {
            New-Item -ItemType Directory -Path $destPath -Force
        }
        
        # Copy files with progress simulation
        $files = Get-ChildItem -Path $sourcePath -Recurse -File
        $totalFiles = $files.Count
        $processedFiles = 0
        
        foreach ($file in $files) {
            $destFile = Join-Path $destPath ($file.FullName.Replace($sourcePath, ""))
            $destDir = Split-Path $destFile -Parent
            
            if (!(Test-Path $destDir)) {
                New-Item -ItemType Directory -Path $destDir -Force
            }
            
            Copy-Item -Path $file.FullName -Destination $destFile -Force
            $processedFiles++
            
            # Update progress
            $progress = [math]::Round(($processedFiles / $totalFiles) * 100, 0)
            $script:RunningJobs[$jobId].Progress = $progress
            
            # Update job status
            $job.Status = "Running ($progress%)"
            Save-Jobs
            
            Start-Sleep -Milliseconds 10  # Simulate processing time
        }
        
        # Mark as completed
        $script:RunningJobs[$jobId].Status = "Completed"
        $script:RunningJobs[$jobId].EndTime = Get-Date
        $script:RunningJobs[$jobId].Progress = 100
        
        Write-Log "Backup job '$($Job.Name)' completed successfully"
        
        # Clean up from running jobs after delay
        Start-Job -ScriptBlock {
            param($jobIdToRemove)
            Start-Sleep -Seconds 5
            if ($script:RunningJobs.ContainsKey($jobIdToRemove)) {
                $script:RunningJobs.Remove($jobIdToRemove)
            }
        } -ArgumentList $jobId
        
    } catch {
        Write-Log "Backup job '$($Job.Name)' failed: $_" "ERROR"
        $script:RunningJobs[$jobId].Status = "Failed"
        $script:RunningJobs[$jobId].Error = $_.ToString()
    }
    
    Save-Jobs
}

function Start-Restore {
    param([object]$Job)
    
    Write-Log "Starting restore job: $($Job.Name)"
    
    try {
        $sourcePath = $Job.SourcePath
        $destPath = $Job.DestinationPath
        
        if (!(Test-Path $destPath)) {
            New-Item -ItemType Directory -Path $destPath -Force
        }
        
        # Copy files back
        $files = Get-ChildItem -Path $sourcePath -Recurse -File
        foreach ($file in $files) {
            $destFile = Join-Path $destPath ($file.FullName.Replace($sourcePath, ""))
            $destDir = Split-Path $destFile -Parent
            
            if (!(Test-Path $destDir)) {
                New-Item -ItemType Directory -Path $destDir -Force
            }
            
            Copy-Item -Path $file.FullName -Destination $destFile -Force
        }
        
        Write-Log "Restore job '$($Job.Name)' completed successfully"
        
    } catch {
        Write-Log "Restore job '$($Job.Name)' failed: $_" "ERROR"
    }
}

function Check-ScheduledJobs {
    $currentTime = Get-Date
    
    foreach ($job in $script:Jobs) {
        if ($job.Enabled -and $job.Schedule) {
            # Check if job should run now
            $shouldRun = $false
            
            switch ($job.Schedule) {
                "Daily" {
                    if ($currentTime.Hour -eq $job.ScheduleHour -and $currentTime.Minute -eq 0) {
                        $shouldRun = $true
                    }
                }
                "Weekly" {
                    if ($currentTime.DayOfWeek -eq $job.ScheduleDay -and $currentTime.Hour -eq $job.ScheduleHour -and $currentTime.Minute -eq 0) {
                        $shouldRun = $true
                    }
                }
                "Monthly" {
                    if ($currentTime.Day -eq $job.ScheduleDay -and $currentTime.Hour -eq $job.ScheduleHour -and $currentTime.Minute -eq 0) {
                        $shouldRun = $true
                    }
                }
            }
            
            if ($shouldRun) {
                Write-Log "Scheduled job '$($job.Name)' triggered"
                Start-Backup $job
            }
        }
    }
}

function Install-Service {
    Write-Log "Installing NovaBackup Agent as Windows Service..."
    
    try {
        # Check if running as Administrator
        $currentPrincipal = [Security.Principal.WindowsIdentity]::GetCurrent()
        $currentPrincipal = [Security.Principal.WindowsPrincipal]::new($currentPrincipal)
        if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
            Write-Log "Administrator privileges required for service installation" "ERROR"
            return
        }
        
        # Create service directory
        $serviceDir = "C:\Program Files\NovaBackup"
        if (!(Test-Path $serviceDir)) {
            New-Item -ItemType Directory -Path $serviceDir -Force
        }
        
        # Copy agent files
        Copy-Item -Path $PSCommandPath -Destination "$serviceDir\NovaBackupAgent.ps1" -Force
        
        # Create service using PowerShell
        $serviceCommand = "powershell.exe -ExecutionPolicy Bypass -File `"$serviceDir\NovaBackupAgent.ps1`" -Mode Service"
        
        New-Service -Name $script:ServiceName -DisplayName "NovaBackup Agent" -BinaryPathName $serviceCommand -StartupType Automatic -Description "NovaBackup v6.0 Background Agent" -ErrorAction SilentlyContinue
        
        Start-Service -Name $script:ServiceName -ErrorAction SilentlyContinue
        
        Write-Log "Service installed and started successfully"
        
    } catch {
        Write-Log "Failed to install service: $_" "ERROR"
    }
}

function Uninstall-Service {
    Write-Log "Uninstalling NovaBackup Agent Service..."
    
    try {
        Stop-Service -Name $script:ServiceName -Force -ErrorAction SilentlyContinue
        Remove-Service -Name $script:ServiceName -ErrorAction SilentlyContinue
        
        Write-Log "Service uninstalled successfully"
        
    } catch {
        Write-Log "Failed to uninstall service: $_" "ERROR"
    }
}

function Service-Main {
    Write-Log "NovaBackup Agent Service starting..."
    
    # Load configuration and jobs
    Load-Config
    Load-Jobs
    
    # Main service loop
    while ($true) {
        try {
            # Check scheduled jobs every minute
            Check-ScheduledJobs
            
            # Clean up completed jobs
            $completedJobs = @()
            foreach ($jobId in $script:RunningJobs.Keys) {
                if ($script:RunningJobs[$jobId].Status -eq "Completed" -or $script:RunningJobs[$jobId].Status -eq "Failed") {
                    $completedJobs += $jobId
                }
            }
            
            foreach ($jobId in $completedJobs) {
                if ($script:RunningJobs.ContainsKey($jobId)) {
                    $script:RunningJobs.Remove($jobId)
                }
            }
            
            # Sleep for 60 seconds
            Start-Sleep -Seconds 60
            
        } catch {
            Write-Log "Service loop error: $_" "ERROR"
            Start-Sleep -Seconds 30
        }
    }
}

function Interactive-Mode {
    Write-Log "NovaBackup Agent running in interactive mode"
    Write-Host "NovaBackup v6.0 Agent - Interactive Mode"
    Write-Host "Commands: status, jobs, backup <jobid>, restore <jobid>, config, exit"
    
    Load-Config
    Load-Jobs
    
    while ($true) {
        Write-Host "NovaBackup> " -NoNewline
        $input = Read-Host
        
        switch ($input.ToLower()) {
            "status" {
                Write-Host "Running Jobs:"
                foreach ($jobId in $script:RunningJobs.Keys) {
                    $job = $script:RunningJobs[$jobId]
                    Write-Host "  $($job.Name): $($job.Status) ($($job.Progress)%)"
                }
            }
            "jobs" {
                Write-Host "Configured Jobs:"
                foreach ($job in $script:Jobs) {
                    Write-Host "  $($job.Name): $($job.Type) - $($job.Schedule)"
                }
            }
            "backup" {
                if ($input -match "backup (\d+)") {
                    $jobId = $matches[1]
                    $job = $script:Jobs | Where-Object { $_.Id -eq $jobId }
                    if ($job) {
                        Start-Backup $job
                    } else {
                        Write-Host "Job not found: $jobId"
                    }
                }
            }
            "config" {
                Write-Host "Current Configuration:"
                Write-Host ($script:Config | ConvertTo-Json)
            }
            "exit" {
                Write-Log "Agent exiting interactive mode"
                break
            }
            default {
                Write-Host "Unknown command. Available: status, jobs, backup <id>, restore <id>, config, exit"
            }
        }
    }
}

# Main execution
switch ($Mode) {
    "Install" {
        if ($Install) {
            Install-Service
        }
    }
    "Uninstall" {
        if ($Uninstall) {
            Uninstall-Service
        }
    }
    "Service" {
        Service-Main
    }
    default {
        Interactive-Mode
    }
}
