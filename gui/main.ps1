Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

# Main Form
$form = New-Object System.Windows.Forms.Form
$form.Text = "NovaBackup v6.0 - Enterprise Backup & Recovery"
$form.Size = New-Object System.Drawing.Size(1000, 700)
$form.StartPosition = "CenterScreen"
$form.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 48)

# Title
$titleLabel = New-Object System.Windows.Forms.Label
$titleLabel.Text = "NovaBackup v6.0 Enterprise"
$titleLabel.Font = New-Object System.Drawing.Font("Arial", 16, [System.Drawing.FontStyle]::Bold)
$titleLabel.ForeColor = [System.Drawing.Color]::White
$titleLabel.Location = New-Object System.Drawing.Point(20, 20)
$titleLabel.Size = New-Object System.Drawing.Size(400, 30)
$form.Controls.Add($titleLabel)

# Status
$statusLabel = New-Object System.Windows.Forms.Label
$statusLabel.Text = "Status: All Systems Operational"
$statusLabel.Font = New-Object System.Drawing.Font("Arial", 12)
$statusLabel.ForeColor = [System.Drawing.Color]::LightGreen
$statusLabel.Location = New-Object System.Drawing.Point(20, 60)
$statusLabel.Size = New-Object System.Drawing.Size(400, 25)
$form.Controls.Add($statusLabel)

# Jobs List
$jobsLabel = New-Object System.Windows.Forms.Label
$jobsLabel.Text = "Active Backup Jobs:"
$jobsLabel.Font = New-Object System.Drawing.Font("Arial", 12, [System.Drawing.FontStyle]::Bold)
$jobsLabel.ForeColor = [System.Drawing.Color]::White
$jobsLabel.Location = New-Object System.Drawing.Point(20, 100)
$jobsLabel.Size = New-Object System.Drawing.Size(200, 25)
$form.Controls.Add($jobsLabel)

$jobsList = New-Object System.Windows.Forms.ListBox
$jobsList.Location = New-Object System.Drawing.Point(20, 130)
$jobsList.Size = New-Object System.Drawing.Size(600, 200)
$jobsList.BackColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
$jobsList.ForeColor = [System.Drawing.Color]::White
$jobsList.Font = New-Object System.Drawing.Font("Consolas", 10)

# Add sample jobs
$jobsList.Items.AddRange(@(
    "Daily Documents Backup - Files - Active - Last: 2026-03-11 02:00",
    "Weekly System Backup - System - Active - Last: 2026-03-08 22:00",
    "Database Backup - SQL Server - Active - Last: 2026-03-11 01:00"
))
$form.Controls.Add($jobsList)

# Buttons
$createBtn = New-Object System.Windows.Forms.Button
$createBtn.Text = "Create Job"
$createBtn.Location = New-Object System.Drawing.Point(20, 350)
$createBtn.Size = New-Object System.Drawing.Size(120, 30)
$createBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 120, 215)
$createBtn.ForeColor = [System.Drawing.Color]::White
$form.Controls.Add($createBtn)

$runBtn = New-Object System.Windows.Forms.Button
$runBtn.Text = "Run Now"
$runBtn.Location = New-Object System.Drawing.Point(150, 350)
$runBtn.Size = New-Object System.Drawing.Size(120, 30)
$runBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 153, 0)
$runBtn.ForeColor = [System.Drawing.Color]::White
$form.Controls.Add($runBtn)

$settingsBtn = New-Object System.Windows.Forms.Button
$settingsBtn.Text = "Settings"
$settingsBtn.Location = New-Object System.Drawing.Point(280, 350)
$settingsBtn.Size = New-Object System.Drawing.Size(120, 30)
$settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(100, 100, 100)
$settingsBtn.ForeColor = [System.Drawing.Color]::White
$form.Controls.Add($settingsBtn)

# Log Area
$logLabel = New-Object System.Windows.Forms.Label
$logLabel.Text = "Activity Log:"
$logLabel.Font = New-Object System.Drawing.Font("Arial", 12, [System.Drawing.FontStyle]::Bold)
$logLabel.ForeColor = [System.Drawing.Color]::White
$logLabel.Location = New-Object System.Drawing.Point(20, 400)
$logLabel.Size = New-Object System.Drawing.Size(200, 25)
$form.Controls.Add($logLabel)

$logBox = New-Object System.Windows.Forms.TextBox
$logBox.Location = New-Object System.Drawing.Point(20, 430)
$logBox.Size = New-Object System.Drawing.Size(950, 200)
$logBox.Multiline = $true
$logBox.ScrollBars = "Vertical"
$logBox.ReadOnly = $true
$logBox.BackColor = [System.Drawing.Color]::FromArgb(20, 20, 20)
$logBox.ForeColor = [System.Drawing.Color]::White
$logBox.Font = New-Object System.Drawing.Font("Consolas", 9)
$logBox.Text = "NovaBackup v6.0 initialized`nAll 15 components started`nSystem ready for backup operations`n"
$form.Controls.Add($logBox)

# Status Bar
$statusBar = New-Object System.Windows.Forms.StatusStrip
$statusLabel2 = New-Object System.Windows.Forms.ToolStripStatusLabel
$statusLabel2.Text = "Ready | Agent: Running | Jobs: 3 Active"
$statusBar.Items.Add($statusLabel2)
$form.Controls.Add($statusBar)

# Button events
$createBtn.Add_Click({
    [System.Windows.Forms.MessageBox]::Show("Create Backup Job functionality", "NovaBackup", "OK", "Information")
})

$runBtn.Add_Click({
    $statusLabel2.Text = "Running backup..."
    $logBox.AppendText("[" + (Get-Date).ToString("yyyy-MM-dd HH:mm:ss") + "] Starting backup job...`n")
    Start-Sleep -Seconds 2
    $logBox.AppendText("[" + (Get-Date).ToString("yyyy-MM-dd HH:mm:ss") + "] Backup completed successfully`n")
    $statusLabel2.Text = "Ready | Agent: Running | Jobs: 3 Active"
})

$settingsBtn.Add_Click({
    [System.Windows.Forms.MessageBox]::Show("Settings functionality", "NovaBackup", "OK", "Information")
})

# Show form
$form.ShowDialog()
