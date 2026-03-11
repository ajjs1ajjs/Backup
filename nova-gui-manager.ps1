# NovaBackup v6.0 - GUI Manager
# Communicates with the background agent

Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

# Global variables
$script:AgentProcess = $null
$script:ConfigFile = "C:\Program Files\NovaBackup\config.json"
$script:JobsFile = "C:\Program Files\NovaBackup\jobs.json"

function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message"
}

function Start-Agent {
    Write-Log "Starting NovaBackup Agent..."
    
    try {
        # Check if agent is already running
        $agentPath = "C:\Program Files\NovaBackup\NovaBackupAgent.ps1"
        if (!(Test-Path $agentPath)) {
            Write-Log "Agent not found. Installing..."
            Copy-Item -Path ".\NovaBackupAgent.ps1" -Destination $agentPath -Force
        }
        
        # Start agent in background
        $script:AgentProcess = Start-Process -FilePath "powershell.exe" -ArgumentList "-ExecutionPolicy Bypass -File `"$agentPath`" -Mode Service" -WindowStyle Hidden -PassThru
        
        Write-Log "Agent started with PID: $($script:AgentProcess.Id)"
        return $true
        
    } catch {
        Write-Log "Failed to start agent: $_" "ERROR"
        return $false
    }
}

function Stop-Agent {
    Write-Log "Stopping NovaBackup Agent..."
    
    try {
        if ($script:AgentProcess) {
            Stop-Process -Id $script:AgentProcess.Id -Force
            $script:AgentProcess = $null
            Write-Log "Agent stopped"
        } else {
            # Try to find and stop existing agent process
            $processes = Get-Process | Where-Object { $_.ProcessName -eq "powershell" -and $_.MainWindowTitle -eq "" }
            foreach ($proc in $processes) {
                try {
                    Stop-Process -Id $proc.Id -Force
                    Write-Log "Stopped agent process: $($proc.Id)"
                } catch {
                    Write-Log "Failed to stop process $($proc.Id): $_"
                }
            }
        }
    } catch {
        Write-Log "Failed to stop agent: $_" "ERROR"
    }
}

function Get-Agent-Status {
    try {
        if ($script:AgentProcess -and !$script:AgentProcess.HasExited) {
            return @{
                "Status" = "Running"
                "PID" = $script:AgentProcess.Id
                "StartTime" = $script:AgentProcess.StartTime
            }
        } else {
            return @{
                "Status" = "Stopped"
                "PID" = $null
                "StartTime" = $null
            }
        }
    } catch {
        return @{
            "Status" = "Unknown"
            "Error" = $_.ToString()
        }
    }
}

function Load-Jobs {
    try {
        if (Test-Path $script:JobsFile) {
            return Get-Content $script:JobsFile | ConvertFrom-Json
        } else {
            return @()
        }
    } catch {
        Write-Log "Failed to load jobs: $_" "ERROR"
        return @()
    }
}

function Save-Jobs {
    param([array]$Jobs)
    
    try {
        # Create directory if not exists
        $jobsDir = Split-Path $script:JobsFile -Parent
        if (!(Test-Path $jobsDir)) {
            New-Item -ItemType Directory -Path $jobsDir -Force
        }
        
        $Jobs | ConvertTo-Json -Depth 10 | Set-Content -Path $script:JobsFile
        Write-Log "Jobs saved successfully"
        return $true
    } catch {
        Write-Log "Failed to save jobs: $_" "ERROR"
        return $false
    }
}

function Send-Command-To-Agent {
    param([string]$Command, [hashtable]$Data = @{})
    
    try {
        # Create temporary command file
        $commandFile = "C:\Program Files\NovaBackup\commands\command_$((Get-Date).ToString('yyyyMMddHHmmss')).json"
        $commandDir = Split-Path $commandFile -Parent
        
        if (!(Test-Path $commandDir)) {
            New-Item -ItemType Directory -Path $commandDir -Force
        }
        
        $commandData = @{
            "Command" = $Command
            "Data" = $Data
            "Timestamp" = Get-Date
            "Id" = [Guid]::NewGuid().ToString()
        }
        
        $commandData | ConvertTo-Json -Depth 10 | Set-Content -Path $commandFile
        
        Write-Log "Command sent to agent: $Command"
        return $true
        
    } catch {
        Write-Log "Failed to send command to agent: $_" "ERROR"
        return $false
    }
}

# Main GUI Form
$form = New-Object System.Windows.Forms.Form
$form.Text = "NovaBackup v6.0 - Manager"
$form.Size = New-Object System.Drawing.Size(1000, 700)
$form.StartPosition = "CenterScreen"
$form.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 48)
$form.FormClosing = {
    Stop-Agent
}

# Title
$titleLabel = New-Object System.Windows.Forms.Label
$titleLabel.Text = "NovaBackup v6.0 Manager"
$titleLabel.Font = New-Object System.Drawing.Font("Arial", 16, [System.Drawing.FontStyle]::Bold)
$titleLabel.ForeColor = [System.Drawing.Color]::White
$titleLabel.Location = New-Object System.Drawing.Point(20, 20)
$titleLabel.Size = New-Object System.Drawing.Size(400, 30)
$form.Controls.Add($titleLabel)

# Agent Status Panel
$statusPanel = New-Object System.Windows.Forms.Panel
$statusPanel.Location = New-Object System.Drawing.Point(20, 70)
$statusPanel.Size = New-Object System.Drawing.Size(960, 80)
$statusPanel.BackColor = [System.Drawing.Color]::FromArgb(62, 62, 66)
$statusPanel.BorderStyle = [System.Windows.Forms.BorderStyle]::FixedSingle
$form.Controls.Add($statusPanel)

$agentStatusLabel = New-Object System.Windows.Forms.Label
$agentStatusLabel.Text = "Agent Status: Unknown"
$agentStatusLabel.Font = New-Object System.Drawing.Font("Arial", 12)
$agentStatusLabel.ForeColor = [System.Drawing.Color]::White
$agentStatusLabel.Location = New-Object System.Drawing.Point(10, 10)
$agentStatusLabel.Size = New-Object System.Drawing.Size(300, 25)
$statusPanel.Controls.Add($agentStatusLabel)

$agentDetailsLabel = New-Object System.Windows.Forms.Label
$agentDetailsLabel.Text = "PID: N/A | Started: N/A"
$agentDetailsLabel.Font = New-Object System.Drawing.Font("Arial", 10)
$agentDetailsLabel.ForeColor = [System.Drawing.Color]::Gray
$agentDetailsLabel.Location = New-Object System.Drawing.Point(10, 40)
$agentDetailsLabel.Size = New-Object System.Drawing.Size(500, 25)
$statusPanel.Controls.Add($agentDetailsLabel)

# Control Buttons
$startBtn = New-Object System.Windows.Forms.Button
$startBtn.Text = "Start Agent"
$startBtn.Location = New-Object System.Drawing.Point(600, 15)
$startBtn.Size = New-Object System.Drawing.Size(120, 30)
$startBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 153, 0)
$startBtn.ForeColor = [System.Drawing.Color]::White
$startBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$statusPanel.Controls.Add($startBtn)

$stopBtn = New-Object System.Windows.Forms.Button
$stopBtn.Text = "Stop Agent"
$stopBtn.Location = New-Object System.Drawing.Point(730, 15)
$stopBtn.Size = New-Object System.Drawing.Size(120, 30)
$stopBtn.BackColor = [System.Drawing.Color]::FromArgb(215, 0, 0)
$stopBtn.ForeColor = [System.Drawing.Color]::White
$stopBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$statusPanel.Controls.Add($stopBtn)

$refreshBtn = New-Object System.Windows.Forms.Button
$refreshBtn.Text = "Refresh Status"
$refreshBtn.Location = New-Object System.Drawing.Point(860, 15)
$refreshBtn.Size = New-Object System.Drawing.Size(120, 30)
$refreshBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 120, 215)
$refreshBtn.ForeColor = [System.Drawing.Color]::White
$refreshBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$statusPanel.Controls.Add($refreshBtn)

# Jobs Panel
$jobsPanel = New-Object System.Windows.Forms.Panel
$jobsPanel.Location = New-Object System.Drawing.Point(20, 170)
$jobsPanel.Size = New-Object System.Drawing.Size(960, 300)
$jobsPanel.BackColor = [System.Drawing.Color]::FromArgb(62, 62, 66)
$jobsPanel.BorderStyle = [System.Windows.Forms.BorderStyle]::FixedSingle
$form.Controls.Add($jobsPanel)

$jobsTitleLabel = New-Object System.Windows.Forms.Label
$jobsTitleLabel.Text = "Backup Jobs"
$jobsTitleLabel.Font = New-Object System.Drawing.Font("Arial", 12, [System.Drawing.FontStyle]::Bold)
$jobsTitleLabel.ForeColor = [System.Drawing.Color]::White
$jobsTitleLabel.Location = New-Object System.Drawing.Point(10, 10)
$jobsTitleLabel.Size = New-Object System.Drawing.Size(200, 25)
$jobsPanel.Controls.Add($jobsTitleLabel)

# Jobs DataGridView
$jobsGrid = New-Object System.Windows.Forms.DataGridView
$jobsGrid.Location = New-Object System.Drawing.Point(10, 40)
$jobsGrid.Size = New-Object System.Drawing.Size(940, 200)
$jobsGrid.BackgroundColor = [System.Drawing.Color]::FromArgb(45, 45, 48)
$jobsGrid.DefaultCellStyle.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 48)
$jobsGrid.DefaultCellStyle.ForeColor = [System.Drawing.Color]::White
$jobsGrid.ColumnHeadersDefaultCellStyle.BackColor = [System.Drawing.Color]::FromArgb(80, 80, 80)
$jobsGrid.ColumnHeadersDefaultCellStyle.ForeColor = [System.Drawing.Color]::White
$jobsGrid.BorderStyle = [System.Windows.Forms.BorderStyle]::None
$jobsPanel.Controls.Add($jobsGrid)

# Add columns
$jobsGrid.Columns.Add("Name", "Name") | Out-Null
$jobsGrid.Columns.Add("Type", "Type") | Out-Null
$jobsGrid.Columns.Add("Status", "Status") | Out-Null
$jobsGrid.Columns.Add("Schedule", "Schedule") | Out-Null
$jobsGrid.Columns.Add("LastRun", "Last Run") | Out-Null

# Set column widths
$jobsGrid.Columns["Name"].Width = 200
$jobsGrid.Columns["Type"].Width = 80
$jobsGrid.Columns["Status"].Width = 100
$jobsGrid.Columns["Schedule"].Width = 120
$jobsGrid.Columns["LastRun"].Width = 150
$jobsGrid.Columns["NextRun"].Width = 150

# Job Control Buttons
$jobButtonsPanel = New-Object System.Windows.Forms.Panel
$jobButtonsPanel.Location = New-Object System.Drawing.Point(10, 250)
$jobButtonsPanel.Size = New-Object System.Drawing.Size(940, 40)
$jobsPanel.Controls.Add($jobButtonsPanel)

$addJobBtn = New-Object System.Windows.Forms.Button
$addJobBtn.Text = "Add Job"
$addJobBtn.Location = New-Object System.Drawing.Point(0, 5)
$addJobBtn.Size = New-Object System.Drawing.Size(100, 30)
$addJobBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 120, 215)
$addJobBtn.ForeColor = [System.Drawing.Color]::White
$addJobBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$jobButtonsPanel.Controls.Add($addJobBtn)

$editJobBtn = New-Object System.Windows.Forms.Button
$editJobBtn.Text = "Edit Job"
$editJobBtn.Location = New-Object System.Drawing.Point(110, 5)
$editJobBtn.Size = New-Object System.Drawing.Size(100, 30)
$editJobBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 120, 215)
$editJobBtn.ForeColor = [System.Drawing.Color]::White
$editJobBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$jobButtonsPanel.Controls.Add($editJobBtn)

$deleteJobBtn = New-Object System.Windows.Forms.Button
$deleteJobBtn.Text = "Delete Job"
$deleteJobBtn.Location = New-Object System.Drawing.Point(220, 5)
$deleteJobBtn.Size = New-Object System.Drawing.Size(100, 30)
$deleteJobBtn.BackColor = [System.Drawing.Color]::FromArgb(215, 0, 0)
$deleteJobBtn.ForeColor = [System.Drawing.Color]::White
$deleteJobBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$jobButtonsPanel.Controls.Add($deleteJobBtn)

$runJobBtn = New-Object System.Windows.Forms.Button
$runJobBtn.Text = "Run Now"
$runJobBtn.Location = New-Object System.Drawing.Point(330, 5)
$runJobBtn.Size = New-Object System.Drawing.Size(100, 30)
$runJobBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 153, 0)
$runJobBtn.ForeColor = [System.Drawing.Color]::White
$runJobBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$jobButtonsPanel.Controls.Add($runJobBtn)

# Log Panel
$logPanel = New-Object System.Windows.Forms.Panel
$logPanel.Location = New-Object System.Drawing.Point(20, 490)
$logPanel.Size = New-Object System.Drawing.Size(960, 180)
$logPanel.BackColor = [System.Drawing.Color]::FromArgb(62, 62, 66)
$logPanel.BorderStyle = [System.Windows.Forms.BorderStyle]::FixedSingle
$form.Controls.Add($logPanel)

$logTitleLabel = New-Object System.Windows.Forms.Label
$logTitleLabel.Text = "Activity Log"
$logTitleLabel.Font = New-Object System.Drawing.Font("Arial", 12, [System.Drawing.FontStyle]::Bold)
$logTitleLabel.ForeColor = [System.Drawing.Color]::White
$logTitleLabel.Location = New-Object System.Drawing.Point(10, 10)
$logTitleLabel.Size = New-Object System.Drawing.Size(200, 25)
$logPanel.Controls.Add($logTitleLabel)

$logTextBox = New-Object System.Windows.Forms.TextBox
$logTextBox.Location = New-Object System.Drawing.Point(10, 40)
$logTextBox.Size = New-Object System.Drawing.Size(940, 130)
$logTextBox.Multiline = $true
$logTextBox.ScrollBars = [System.Windows.Forms.ScrollBars]::Vertical
$logTextBox.ReadOnly = $true
$logTextBox.BackColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
$logTextBox.ForeColor = [System.Drawing.Color]::White
$logTextBox.Font = New-Object System.Drawing.Font("Consolas", 9)
$logPanel.Controls.Add($logTextBox)

# Button Events
$startBtn.Add_Click({
    if (Start-Agent) {
        $agentStatusLabel.Text = "Agent Status: Running"
        $agentStatusLabel.ForeColor = [System.Drawing.Color]::LightGreen
        $agentDetailsLabel.Text = "PID: $($script:AgentProcess.Id) | Started: $($script:AgentProcess.StartTime)"
    }
})

$stopBtn.Add_Click({
    Stop-Agent
    $agentStatusLabel.Text = "Agent Status: Stopped"
    $agentStatusLabel.ForeColor = [System.Drawing.Color]::Red
    $agentDetailsLabel.Text = "PID: N/A | Started: N/A"
})

$refreshBtn.Add_Click({
    $status = Get-Agent-Status
    $agentStatusLabel.Text = "Agent Status: $($status.Status)"
    $agentDetailsLabel.Text = "PID: $($status.PID) | Started: $($status.StartTime)"
    
    if ($status.Status -eq "Running") {
        $agentStatusLabel.ForeColor = [System.Drawing.Color]::LightGreen
    } else {
        $agentStatusLabel.ForeColor = [System.Drawing.Color]::Red
    }
    
    Refresh-Jobs
})

$addJobBtn.Add_Click({
    [System.Windows.Forms.MessageBox]::Show("Add Job functionality", "NovaBackup", [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information)
})

$editJobBtn.Add_Click({
    [System.Windows.Forms.MessageBox]::Show("Edit Job functionality", "NovaBackup", [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information)
})

$deleteJobBtn.Add_Click({
    [System.Windows.Forms.MessageBox]::Show("Delete Job functionality", "NovaBackup", [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information)
})

$runJobBtn.Add_Click({
    [System.Windows.Forms.MessageBox]::Show("Run Job functionality", "NovaBackup", [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information)
})

function Refresh-Jobs {
    $jobs = Load-Jobs
    $jobsGrid.Rows.Clear()
    
    foreach ($job in $jobs) {
        $jobsGrid.Rows.Add(@(
            $job.Name,
            $job.Type,
            $job.Status,
            $job.Schedule,
            $job.LastRun,
            $job.NextRun
        )) | Out-Null
    }
}

# Status update timer
$timer = New-Object System.Windows.Forms.Timer
$timer.Interval = 5000  # 5 seconds
$timer.Add_Tick({
    $status = Get-Agent-Status
    $agentStatusLabel.Text = "Agent Status: $($status.Status)"
    
    if ($status.Status -eq "Running") {
        $agentStatusLabel.ForeColor = [System.Drawing.Color]::LightGreen
        if ($status.PID) {
            $agentDetailsLabel.Text = "PID: $($status.PID) | Started: $($status.StartTime)"
        }
    } else {
        $agentStatusLabel.ForeColor = [System.Drawing.Color]::Red
        $agentDetailsLabel.Text = "PID: N/A | Started: N/A"
    }
})

$timer.Start()

# Initial setup
Write-Log "NovaBackup Manager starting..."
Refresh-Jobs

# Show form
$form.ShowDialog() | Out-Null
