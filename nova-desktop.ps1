Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

# Main Form
$form = New-Object System.Windows.Forms.Form
$form.Text = "NovaBackup v6.0 - Enterprise Backup & Recovery"
$form.Size = New-Object System.Drawing.Size(1200,800)
$form.StartPosition = "CenterScreen"
$form.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 48)

# Title Label
$titleLabel = New-Object System.Windows.Forms.Label
$titleLabel.Text = "NovaBackup v6.0"
$titleLabel.Font = New-Object System.Drawing.Font("Arial", 16, [System.Drawing.FontStyle]::Bold)
$titleLabel.ForeColor = [System.Drawing.Color]::White
$titleLabel.Location = New-Object System.Drawing.Point(20, 20)
$titleLabel.Size = New-Object System.Drawing.Size(300, 30)
$form.Controls.Add($titleLabel)

# Subtitle Label
$subtitleLabel = New-Object System.Windows.Forms.Label
$subtitleLabel.Text = "Enterprise Backup & Recovery Platform"
$subtitleLabel.Font = New-Object System.Drawing.Font("Arial", 10)
$subtitleLabel.ForeColor = [System.Drawing.Color]::Gray
$subtitleLabel.Location = New-Object System.Drawing.Point(20, 55)
$subtitleLabel.Size = New-Object System.Drawing.Size(300, 20)
$form.Controls.Add($subtitleLabel)

# Status Panel
$statusPanel = New-Object System.Windows.Forms.Panel
$statusPanel.Location = New-Object System.Drawing.Point(20, 90)
$statusPanel.Size = New-Object System.Drawing.Size(1160, 60)
$statusPanel.BackColor = [System.Drawing.Color]::FromArgb(62, 62, 66)
$statusPanel.BorderStyle = [System.Windows.Forms.BorderStyle]::FixedSingle
$form.Controls.Add($statusPanel)

# Status Label
$statusLabel = New-Object System.Windows.Forms.Label
$statusLabel.Text = "Status: Ready"
$statusLabel.Font = New-Object System.Drawing.Font("Arial", 12)
$statusLabel.ForeColor = [System.Drawing.Color]::LightGreen
$statusLabel.Location = New-Object System.Drawing.Point(10, 15)
$statusLabel.Size = New-Object System.Drawing.Size(200, 30)
$statusPanel.Controls.Add($statusLabel)

# Buttons
$createJobBtn = New-Object System.Windows.Forms.Button
$createJobBtn.Text = "Create Job"
$createJobBtn.Location = New-Object System.Drawing.Point(750, 15)
$createJobBtn.Size = New-Object System.Drawing.Size(120, 30)
$createJobBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 120, 215)
$createJobBtn.ForeColor = [System.Drawing.Color]::White
$createJobBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$createJobBtn.FlatAppearance.BorderSize = 0
$statusPanel.Controls.Add($createJobBtn)

$runBackupBtn = New-Object System.Windows.Forms.Button
$runBackupBtn.Text = "Run Now"
$runBackupBtn.Location = New-Object System.Drawing.Point(880, 15)
$runBackupBtn.Size = New-Object System.Drawing.Size(120, 30)
$runBackupBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 153, 0)
$runBackupBtn.ForeColor = [System.Drawing.Color]::White
$runBackupBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$runBackupBtn.FlatAppearance.BorderSize = 0
$statusPanel.Controls.Add($runBackupBtn)

$settingsBtn = New-Object System.Windows.Forms.Button
$settingsBtn.Text = "Settings"
$settingsBtn.Location = New-Object System.Drawing.Point(1010, 15)
$settingsBtn.Size = New-Object System.Drawing.Size(120, 30)
$settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(100, 100, 100)
$settingsBtn.ForeColor = [System.Drawing.Color]::White
$settingsBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$settingsBtn.FlatAppearance.BorderSize = 0
$statusPanel.Controls.Add($settingsBtn)

# Jobs Panel
$jobsPanel = New-Object System.Windows.Forms.Panel
$jobsPanel.Location = New-Object System.Drawing.Point(20, 170)
$jobsPanel.Size = New-Object System.Drawing.Size(700, 400)
$jobsPanel.BackColor = [System.Drawing.Color]::FromArgb(62, 62, 66)
$jobsPanel.BorderStyle = [System.Windows.Forms.BorderStyle]::FixedSingle
$form.Controls.Add($jobsPanel)

# Jobs Title
$jobsTitleLabel = New-Object System.Windows.Forms.Label
$jobsTitleLabel.Text = "Backup Jobs"
$jobsTitleLabel.Font = New-Object System.Drawing.Font("Arial", 12, [System.Drawing.FontStyle]::Bold)
$jobsTitleLabel.ForeColor = [System.Drawing.Color]::White
$jobsTitleLabel.Location = New-Object System.Drawing.Point(10, 10)
$jobsTitleLabel.Size = New-Object System.Drawing.Size(300, 25)
$jobsPanel.Controls.Add($jobsTitleLabel)

# Jobs DataGridView
$jobsGrid = New-Object System.Windows.Forms.DataGridView
$jobsGrid.Location = New-Object System.Drawing.Point(10, 40)
$jobsGrid.Size = New-Object System.Drawing.Size(680, 300)
$jobsGrid.BackgroundColor = [System.Drawing.Color]::FromArgb(45, 45, 48)
$jobsGrid.DefaultCellStyle.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 48)
$jobsGrid.DefaultCellStyle.ForeColor = [System.Drawing.Color]::White
$jobsGrid.DefaultCellStyle.SelectionBackColor = [System.Drawing.Color]::FromArgb(0, 120, 215)
$jobsGrid.DefaultCellStyle.SelectionForeColor = [System.Drawing.Color]::White
$jobsGrid.ColumnHeadersDefaultCellStyle.BackColor = [System.Drawing.Color]::FromArgb(80, 80, 80)
$jobsGrid.ColumnHeadersDefaultCellStyle.ForeColor = [System.Drawing.Color]::White
$jobsGrid.BorderStyle = [System.Windows.Forms.BorderStyle]::None
$jobsGrid.AllowUserToAddRows = $false
$jobsGrid.AllowUserToDeleteRows = $false
$jobsGrid.ReadOnly = $true
$jobsPanel.Controls.Add($jobsGrid)

# Add columns to DataGridView
$jobsGrid.Columns.Add("Name", "Job Name") | Out-Null
$jobsGrid.Columns.Add("Type", "Type") | Out-Null
$jobsGrid.Columns.Add("Status", "Status") | Out-Null
$jobsGrid.Columns.Add("LastRun", "Last Run") | Out-Null
$jobsGrid.Columns.Add("NextRun", "Next Run") | Out-Null
$jobsGrid.Columns.Add("Schedule", "Schedule") | Out-Null

# Set column widths
$jobsGrid.Columns["Name"].Width = 180
$jobsGrid.Columns["Type"].Width = 80
$jobsGrid.Columns["Status"].Width = 100
$jobsGrid.Columns["LastRun"].Width = 140
$jobsGrid.Columns["NextRun"].Width = 140
$jobsGrid.Columns["Schedule"].Width = 100

# Add sample data
$jobsGrid.Rows.Add("Daily Documents Backup", "Files", "Active", "2026-03-11 02:00", "2026-03-12 02:00", "Daily 2AM") | Out-Null
$jobsGrid.Rows.Add("Weekly System Backup", "System", "Active", "2026-03-08 22:00", "2026-03-15 22:00", "Weekly Sun 10PM") | Out-Null

# Job Buttons
$editJobBtn = New-Object System.Windows.Forms.Button
$editJobBtn.Text = "Edit"
$editJobBtn.Location = New-Object System.Drawing.Point(10, 350)
$editJobBtn.Size = New-Object System.Drawing.Size(100, 30)
$editJobBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 120, 215)
$editJobBtn.ForeColor = [System.Drawing.Color]::White
$editJobBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$editJobBtn.FlatAppearance.BorderSize = 0
$jobsPanel.Controls.Add($editJobBtn)

$deleteJobBtn = New-Object System.Windows.Forms.Button
$deleteJobBtn.Text = "Delete"
$deleteJobBtn.Location = New-Object System.Drawing.Point(120, 350)
$deleteJobBtn.Size = New-Object System.Drawing.Size(100, 30)
$deleteJobBtn.BackColor = [System.Drawing.Color]::FromArgb(215, 0, 0)
$deleteJobBtn.ForeColor = [System.Drawing.Color]::White
$deleteJobBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$deleteJobBtn.FlatAppearance.BorderSize = 0
$jobsPanel.Controls.Add($deleteJobBtn)

$runJobBtn = New-Object System.Windows.Forms.Button
$runJobBtn.Text = "Run"
$runJobBtn.Location = New-Object System.Drawing.Point(230, 350)
$runJobBtn.Size = New-Object System.Drawing.Size(100, 30)
$runJobBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 153, 0)
$runJobBtn.ForeColor = [System.Drawing.Color]::White
$runJobBtn.FlatStyle = [System.Windows.Forms.FlatStyle]::Flat
$runJobBtn.FlatAppearance.BorderSize = 0
$jobsPanel.Controls.Add($runJobBtn)

# Status Panel (Right)
$systemPanel = New-Object System.Windows.Forms.Panel
$systemPanel.Location = New-Object System.Drawing.Point(740, 170)
$systemPanel.Size = New-Object System.Drawing.Size(440, 400)
$systemPanel.BackColor = [System.Drawing.Color]::FromArgb(62, 62, 66)
$systemPanel.BorderStyle = [System.Windows.Forms.BorderStyle]::FixedSingle
$form.Controls.Add($systemPanel)

# System Title
$systemTitleLabel = New-Object System.Windows.Forms.Label
$systemTitleLabel.Text = "System Status"
$systemTitleLabel.Font = New-Object System.Drawing.Font("Arial", 12, [System.Drawing.FontStyle]::Bold)
$systemTitleLabel.ForeColor = [System.Drawing.Color]::White
$systemTitleLabel.Location = New-Object System.Drawing.Point(10, 10)
$systemTitleLabel.Size = New-Object System.Drawing.Size(300, 25)
$systemPanel.Controls.Add($systemTitleLabel)

# Progress Label
$progressLabel = New-Object System.Windows.Forms.Label
$progressLabel.Text = "Backup Progress:"
$progressLabel.Font = New-Object System.Drawing.Font("Arial", 10)
$progressLabel.ForeColor = [System.Drawing.Color]::White
$progressLabel.Location = New-Object System.Drawing.Point(10, 45)
$progressLabel.Size = New-Object System.Drawing.Size(200, 20)
$systemPanel.Controls.Add($progressLabel)

# Progress Bar
$progressBar = New-Object System.Windows.Forms.ProgressBar
$progressBar.Location = New-Object System.Drawing.Point(10, 70)
$progressBar.Size = New-Object System.Drawing.Size(410, 25)
$progressBar.Style = [System.Windows.Forms.ProgressBarStyle]::Continuous
$systemPanel.Controls.Add($progressBar)

# Log Label
$logLabel = New-Object System.Windows.Forms.Label
$logLabel.Text = "Activity Log:"
$logLabel.Font = New-Object System.Drawing.Font("Arial", 10)
$logLabel.ForeColor = [System.Drawing.Color]::White
$logLabel.Location = New-Object System.Drawing.Point(10, 105)
$logLabel.Size = New-Object System.Drawing.Size(200, 20)
$systemPanel.Controls.Add($logLabel)

# Log TextBox
$logTextBox = New-Object System.Windows.Forms.TextBox
$logTextBox.Location = New-Object System.Drawing.Point(10, 130)
$logTextBox.Size = New-Object System.Drawing.Size(410, 200)
$logTextBox.Multiline = $true
$logTextBox.ScrollBars = [System.Windows.Forms.ScrollBars]::Vertical
$logTextBox.ReadOnly = $true
$logTextBox.BackColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
$logTextBox.ForeColor = [System.Drawing.Color]::White
$logTextBox.Font = New-Object System.Drawing.Font("Consolas", 9)
$systemPanel.Controls.Add($logTextBox)

# Statistics Panel
$statsPanel = New-Object System.Windows.Forms.Panel
$statsPanel.Location = New-Object System.Drawing.Point(20, 590)
$statsPanel.Size = New-Object System.Drawing.Size(1160, 60)
$statsPanel.BackColor = [System.Drawing.Color]::FromArgb(62, 62, 66)
$statsPanel.BorderStyle = [System.Windows.Forms.BorderStyle]::FixedSingle
$form.Controls.Add($statsPanel)

# Stats Labels
$statsTitleLabel = New-Object System.Windows.Forms.Label
$statsTitleLabel.Text = "Storage Statistics:"
$statsTitleLabel.Font = New-Object System.Drawing.Font("Arial", 12, [System.Drawing.FontStyle]::Bold)
$statsTitleLabel.ForeColor = [System.Drawing.Color]::White
$statsTitleLabel.Location = New-Object System.Drawing.Point(10, 15)
$statsTitleLabel.Size = New-Object System.Drawing.Size(200, 30)
$statsPanel.Controls.Add($statsTitleLabel)

$totalBackupsLabel = New-Object System.Windows.Forms.Label
$totalBackupsLabel.Text = "Total Backups: 0"
$totalBackupsLabel.Font = New-Object System.Drawing.Font("Arial", 10)
$totalBackupsLabel.ForeColor = [System.Drawing.Color]::White
$totalBackupsLabel.Location = New-Object System.Drawing.Point(250, 20)
$totalBackupsLabel.Size = New-Object System.Drawing.Size(150, 20)
$statsPanel.Controls.Add($totalBackupsLabel)

$storageUsedLabel = New-Object System.Windows.Forms.Label
$storageUsedLabel.Text = "Storage Used: 0 GB"
$storageUsedLabel.Font = New-Object System.Drawing.Font("Arial", 10)
$storageUsedLabel.ForeColor = [System.Drawing.Color]::White
$storageUsedLabel.Location = New-Object System.Drawing.Point(420, 20)
$storageUsedLabel.Size = New-Object System.Drawing.Size(150, 20)
$statsPanel.Controls.Add($storageUsedLabel)

$dedupeLabel = New-Object System.Windows.Forms.Label
$dedupeLabel.Text = "Deduplication: 0:1"
$dedupeLabel.Font = New-Object System.Drawing.Font("Arial", 10)
$dedupeLabel.ForeColor = [System.Drawing.Color]::White
$dedupeLabel.Location = New-Object System.Drawing.Point(590, 20)
$dedupeLabel.Size = New-Object System.Drawing.Size(150, 20)
$statsPanel.Controls.Add($dedupeLabel)

$compressionLabel = New-Object System.Windows.Forms.Label
$compressionLabel.Text = "Compression: 0:1"
$compressionLabel.Font = New-Object System.Drawing.Font("Arial", 10)
$compressionLabel.ForeColor = [System.Drawing.Color]::White
$compressionLabel.Location = New-Object System.Drawing.Point(760, 20)
$compressionLabel.Size = New-Object System.Drawing.Size(150, 20)
$statsPanel.Controls.Add($compressionLabel)

# Button Events
$createJobBtn.Add_Click({
    [System.Windows.Forms.MessageBox]::Show("Create Backup Job", "NovaBackup", [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information)
})

$runBackupBtn.Add_Click({
    $statusLabel.Text = "Status: Running backup..."
    $statusLabel.ForeColor = [System.Drawing.Color]::Orange
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logTextBox.AppendText("[$timestamp] Starting backup process...`r`n")
    
    # Simulate progress
    for ($i = 0; $i -le 100; $i += 5) {
        $progressBar.Value = $i
        Start-Sleep -Milliseconds 100
        $form.Refresh()
    }
    
    $statusLabel.Text = "Status: Backup completed"
    $statusLabel.ForeColor = [System.Drawing.Color]::LightGreen
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logTextBox.AppendText("[$timestamp] Backup completed successfully`r`n")
})

$settingsBtn.Add_Click({
    [System.Windows.Forms.MessageBox]::Show("NovaBackup Settings", "NovaBackup", [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information)
})

$editJobBtn.Add_Click({
    [System.Windows.Forms.MessageBox]::Show("Edit Backup Job", "NovaBackup", [System.Windows.Forms.MessageBoxButtons]::OK, [System.Windows.Forms.MessageBoxIcon]::Information)
})

$deleteJobBtn.Add_Click({
    $result = [System.Windows.Forms.MessageBox]::Show("Delete this backup job?", "Confirm", [System.Windows.Forms.MessageBoxButtons]::YesNo, [System.Windows.Forms.MessageBoxIcon]::Question)
    if ($result -eq [System.Windows.Forms.DialogResult]::Yes) {
        $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        $logTextBox.AppendText("[$timestamp] Job deleted`r`n")
    }
})

$runJobBtn.Add_Click({
    $runBackupBtn.PerformClick()
})

# Background monitoring
$timer = New-Object System.Windows.Forms.Timer
$timer.Interval = 5000  # 5 seconds
$timer.Add_Tick({
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logTextBox.AppendText("[$timestamp] System monitoring active`r`n")
    
    # Keep log size manageable
    $lines = $logTextBox.Lines
    if ($lines.Count -gt 50) {
        $logTextBox.Lines = $lines | Select-Object -Last 50
    }
    
    $logTextBox.ScrollToCaret()
})
$timer.Start()

# Initial log entry
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$logTextBox.AppendText("[$timestamp] NovaBackup v6.0 started`r`n")
$logTextBox.AppendText("[$timestamp] System ready for operation`r`n")

# Show form
$form.ShowDialog() | Out-Null
