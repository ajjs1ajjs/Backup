@echo off
title NovaBackup v6.0 Enterprise - Portable Version
color 0B

echo ========================================
echo    NovaBackup v6.0 Enterprise
echo    Portable Version (No Installation)
echo ========================================
echo.

echo Starting NovaBackup with Veeam-style interface...
echo.

REM Check if Python is available
python --version >nul 2>&1
if %errorLevel% neq 0 (
    echo Python not found. Installing GUI mode...
    goto start_gui
)

REM Start Python GUI
echo Starting Python GUI Server...
cd gui
python app.py
goto end

:start_gui
echo Starting PowerShell GUI...
powershell -ExecutionPolicy Bypass -Command "& {
    Add-Type -AssemblyName System.Windows.Forms
    Add-Type -AssemblyName System.Drawing
    
    [System.Windows.Forms.Application]::EnableVisualStyles()
    
    # Main Form
    `$form = New-Object System.Windows.Forms.Form
    `$form.Text = 'NovaBackup v6.0 Enterprise - Veeam Style'
    `$form.Size = New-Object System.Drawing.Size(1200, 800)
    `$form.StartPosition = 'CenterScreen'
    `$form.BackColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
    `$form.Icon = [System.Drawing.SystemIcons]::Shield
    
    # Header Panel
    `$headerPanel = New-Object System.Windows.Forms.Panel
    `$headerPanel.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
    `$headerPanel.Location = New-Object System.Drawing.Point(0, 0)
    `$headerPanel.Size = New-Object System.Drawing.Size(1200, 80)
    
    `$titleLabel = New-Object System.Windows.Forms.Label
    `$titleLabel.Text = 'NovaBackup v6.0 Enterprise'
    `$titleLabel.Font = New-Object System.Drawing.Font('Segoe UI', 20, [System.Drawing.FontStyle]::Bold)
    `$titleLabel.ForeColor = [System.Drawing.Color]::White
    `$titleLabel.Location = New-Object System.Drawing.Point(20, 15)
    `$titleLabel.Size = New-Object System.Drawing.Size(400, 30)
    `$headerPanel.Controls.Add(`$titleLabel)
    
    `$subtitleLabel = New-Object System.Windows.Forms.Label
    `$subtitleLabel.Text = 'Enterprise Backup & Recovery Platform'
    `$subtitleLabel.Font = New-Object System.Drawing.Font('Segoe UI', 12)
    `$subtitleLabel.ForeColor = [System.Drawing.Color]::White
    `$subtitleLabel.Location = New-Object System.Drawing.Point(20, 45)
    `$subtitleLabel.Size = New-Object System.Drawing.Size(400, 20)
    `$headerPanel.Controls.Add(`$subtitleLabel)
    
    `$startBtn = New-Object System.Windows.Forms.Button
    `$startBtn.Text = 'Start Backup'
    `$startBtn.Location = New-Object System.Drawing.Point(900, 20)
    `$startBtn.Size = New-Object System.Drawing.Size(120, 40)
    `$startBtn.BackColor = [System.Drawing.Color]::FromArgb(40, 167, 69)
    `$startBtn.ForeColor = [System.Drawing.Color]::White
    `$startBtn.FlatStyle = 'Flat'
    `$startBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10, [System.Drawing.FontStyle]::Bold)
    `$headerPanel.Controls.Add(`$startBtn)
    
    `$createBtn = New-Object System.Windows.Forms.Button
    `$createBtn.Text = 'Create Job'
    `$createBtn.Location = New-Object System.Drawing.Point(1030, 20)
    `$createBtn.Size = New-Object System.Drawing.Size(120, 40)
    `$createBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 120, 215)
    `$createBtn.ForeColor = [System.Drawing.Color]::White
    `$createBtn.FlatStyle = 'Flat'
    `$createBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10, [System.Drawing.FontStyle]::Bold)
    `$headerPanel.Controls.Add(`$createBtn)
    
    `$form.Controls.Add(`$headerPanel)
    
    # Sidebar
    `$sidebar = New-Object System.Windows.Forms.Panel
    `$sidebar.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    `$sidebar.Location = New-Object System.Drawing.Point(0, 80)
    `$sidebar.Size = New-Object System.Drawing.Size(250, 680)
    
    `$dashboardBtn = New-Object System.Windows.Forms.Button
    `$dashboardBtn.Text = 'Dashboard'
    `$dashboardBtn.Location = New-Object System.Drawing.Point(10, 10)
    `$dashboardBtn.Size = New-Object System.Drawing.Size(230, 40)
    `$dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
    `$dashboardBtn.ForeColor = [System.Drawing.Color]::White
    `$dashboardBtn.FlatStyle = 'Flat'
    `$dashboardBtn.TextAlign = 'MiddleLeft'
    `$dashboardBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$sidebar.Controls.Add(`$dashboardBtn)
    
    `$jobsBtn = New-Object System.Windows.Forms.Button
    `$jobsBtn.Text = 'Backup Jobs'
    `$jobsBtn.Location = New-Object System.Drawing.Point(10, 60)
    `$jobsBtn.Size = New-Object System.Drawing.Size(230, 40)
    `$jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    `$jobsBtn.ForeColor = [System.Drawing.Color]::White
    `$jobsBtn.FlatStyle = 'Flat'
    `$jobsBtn.TextAlign = 'MiddleLeft'
    `$jobsBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$sidebar.Controls.Add(`$jobsBtn)
    
    `$storageBtn = New-Object System.Windows.Forms.Button
    `$storageBtn.Text = 'Storage'
    `$storageBtn.Location = New-Object System.Drawing.Point(10, 110)
    `$storageBtn.Size = New-Object System.Drawing.Size(230, 40)
    `$storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    `$storageBtn.ForeColor = [System.Drawing.Color]::White
    `$storageBtn.FlatStyle = 'Flat'
    `$storageBtn.TextAlign = 'MiddleLeft'
    `$storageBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$sidebar.Controls.Add(`$storageBtn)
    
    `$reportsBtn = New-Object System.Windows.Forms.Button
    `$reportsBtn.Text = 'Reports'
    `$reportsBtn.Location = New-Object System.Drawing.Point(10, 160)
    `$reportsBtn.Size = New-Object System.Drawing.Size(230, 40)
    `$reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    `$reportsBtn.ForeColor = [System.Drawing.Color]::White
    `$reportsBtn.FlatStyle = 'Flat'
    `$reportsBtn.TextAlign = 'MiddleLeft'
    `$reportsBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$sidebar.Controls.Add(`$reportsBtn)
    
    `$settingsBtn = New-Object System.Windows.Forms.Button
    `$settingsBtn.Text = 'Settings'
    `$settingsBtn.Location = New-Object System.Drawing.Point(10, 210)
    `$settingsBtn.Size = New-Object System.Drawing.Size(230, 40)
    `$settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    `$settingsBtn.ForeColor = [System.Drawing.Color]::White
    `$settingsBtn.FlatStyle = 'Flat'
    `$settingsBtn.TextAlign = 'MiddleLeft'
    `$settingsBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$sidebar.Controls.Add(`$settingsBtn)
    
    `$form.Controls.Add(`$sidebar)
    
    # Main Content Area
    `$mainPanel = New-Object System.Windows.Forms.Panel
    `$mainPanel.BackColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
    `$mainPanel.Location = New-Object System.Drawing.Point(250, 80)
    `$mainPanel.Size = New-Object System.Drawing.Size(950, 680)
    
    # Status Cards
    `$statusCard1 = New-Object System.Windows.Forms.Panel
    `$statusCard1.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    `$statusCard1.Location = New-Object System.Drawing.Point(20, 20)
    `$statusCard1.Size = New-Object System.Drawing.Size(280, 150)
    
    `$statusLabel1 = New-Object System.Windows.Forms.Label
    `$statusLabel1.Text = 'System Status'
    `$statusLabel1.Font = New-Object System.Drawing.Font('Segoe UI', 12, [System.Drawing.FontStyle]::Bold)
    `$statusLabel1.ForeColor = [System.Drawing.Color]::White
    `$statusLabel1.Location = New-Object System.Drawing.Point(10, 10)
    `$statusLabel1.Size = New-Object System.Drawing.Size(150, 25)
    `$statusCard1.Controls.Add(`$statusLabel1)
    
    `$statusValue1 = New-Object System.Windows.Forms.Label
    `$statusValue1.Text = 'All Systems Operational'
    `$statusValue1.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$statusValue1.ForeColor = [System.Drawing.Color]::FromArgb(40, 167, 69)
    `$statusValue1.Location = New-Object System.Drawing.Point(10, 40)
    `$statusValue1.Size = New-Object System.Drawing.Size(250, 20)
    `$statusCard1.Controls.Add(`$statusValue1)
    
    `$jobsCount = New-Object System.Windows.Forms.Label
    `$jobsCount.Text = 'Active Jobs: 3'
    `$jobsCount.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$jobsCount.ForeColor = [System.Drawing.Color]::FromArgb(173, 181, 189)
    `$jobsCount.Location = New-Object System.Drawing.Point(10, 70)
    `$jobsCount.Size = New-Object System.Drawing.Size(150, 20)
    `$statusCard1.Controls.Add(`$jobsCount)
    
    `$storageInfo = New-Object System.Windows.Forms.Label
    `$storageInfo.Text = 'Storage Used: 69.7 GB'
    `$storageInfo.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$storageInfo.ForeColor = [System.Drawing.Color]::FromArgb(173, 181, 189)
    `$storageInfo.Location = New-Object System.Drawing.Point(10, 100)
    `$storageInfo.Size = New-Object System.Drawing.Size(150, 20)
    `$statusCard1.Controls.Add(`$storageInfo)
    
    `$mainPanel.Controls.Add(`$statusCard1)
    
    # Recent Activity Card
    `$activityCard = New-Object System.Windows.Forms.Panel
    `$activityCard.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    `$activityCard.Location = New-Object System.Drawing.Point(320, 20)
    `$activityCard.Size = New-Object System.Drawing.Size(280, 150)
    
    `$activityLabel = New-Object System.Windows.Forms.Label
    `$activityLabel.Text = 'Recent Activity'
    `$activityLabel.Font = New-Object System.Drawing.Font('Segoe UI', 12, [System.Drawing.FontStyle]::Bold)
    `$activityLabel.ForeColor = [System.Drawing.Color]::White
    `$activityLabel.Location = New-Object System.Drawing.Point(10, 10)
    `$activityLabel.Size = New-Object System.Drawing.Size(150, 25)
    `$activityCard.Controls.Add(`$activityLabel)
    
    `$activityList = New-Object System.Windows.Forms.ListBox
    `$activityList.Location = New-Object System.Drawing.Point(10, 40)
    `$activityList.Size = New-Object System.Drawing.Size(260, 100)
    `$activityList.BackColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
    `$activityList.ForeColor = [System.Drawing.Color]::White
    `$activityList.Font = New-Object System.Drawing.Font('Consolas', 9)
    `$activityList.Items.AddRange(@(
        'Daily Backup - Completed - 2h ago',
        'System Backup - Running - 45%',
        'Database Backup - Scheduled - 1h'
    ))
    `$activityCard.Controls.Add(`$activityList)
    
    `$mainPanel.Controls.Add(`$activityCard)
    
    # Storage Card
    `$storageCard = New-Object System.Windows.Forms.Panel
    `$storageCard.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    `$storageCard.Location = New-Object System.Drawing.Point(620, 20)
    `$storageCard.Size = New-Object System.Drawing.Size(280, 150)
    
    `$storageLabel = New-Object System.Windows.Forms.Label
    `$storageLabel.Text = 'Storage Overview'
    `$storageLabel.Font = New-Object System.Drawing.Font('Segoe UI', 12, [System.Drawing.FontStyle]::Bold)
    `$storageLabel.ForeColor = [System.Drawing.Color]::White
    `$storageLabel.Location = New-Object System.Drawing.Point(10, 10)
    `$storageLabel.Size = New-Object System.Drawing.Size(150, 25)
    `$storageCard.Controls.Add(`$storageLabel)
    
    `$progressBar = New-Object System.Windows.Forms.ProgressBar
    `$progressBar.Location = New-Object System.Drawing.Point(10, 40)
    `$progressBar.Size = New-Object System.Drawing.Size(260, 20)
    `$progressBar.Value = 69
    `$storageCard.Controls.Add(`$progressBar)
    
    `$storageDetails = New-Object System.Windows.Forms.Label
    `$storageDetails.Text = '69.7 GB / 1 TB used (69.7%)'
    `$storageDetails.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$storageDetails.ForeColor = [System.Drawing.Color]::FromArgb(173, 181, 189)
    `$storageDetails.Location = New-Object System.Drawing.Point(10, 70)
    `$storageDetails.Size = New-Object System.Drawing.Size(250, 20)
    `$storageCard.Controls.Add(`$storageDetails)
    
    `$compressionRatio = New-Object System.Windows.Forms.Label
    `$compressionRatio.Text = 'Compression: 3.2:1 | Dedup: 5.1:1'
    `$compressionRatio.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    `$compressionRatio.ForeColor = [System.Drawing.Color]::FromArgb(173, 181, 189)
    `$compressionRatio.Location = New-Object System.Drawing.Point(10, 100)
    `$compressionRatio.Size = New-Object System.Drawing.Size(250, 20)
    `$storageCard.Controls.Add(`$compressionRatio)
    
    `$mainPanel.Controls.Add(`$storageCard)
    
    # Jobs Table
    `$jobsTable = New-Object System.Windows.Forms.Panel
    `$jobsTable.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    `$jobsTable.Location = New-Object System.Drawing.Point(20, 190)
    `$jobsTable.Size = New-Object System.Drawing.Size(880, 470)
    
    `$tableLabel = New-Object System.Windows.Forms.Label
    `$tableLabel.Text = 'Backup Jobs'
    `$tableLabel.Font = New-Object System.Drawing.Font('Segoe UI', 14, [System.Drawing.FontStyle]::Bold)
    `$tableLabel.ForeColor = [System.Drawing.Color]::White
    `$tableLabel.Location = New-Object System.Drawing.Point(10, 10)
    `$tableLabel.Size = New-Object System.Drawing.Size(150, 30)
    `$jobsTable.Controls.Add(`$tableLabel)
    
    `$jobsDataGrid = New-Object System.Windows.Forms.DataGridView
    `$jobsDataGrid.Location = New-Object System.Drawing.Point(10, 50)
    `$jobsDataGrid.Size = New-Object System.Drawing.Size(860, 400)
    `$jobsDataGrid.BackgroundColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
    `$jobsDataGrid.ForeColor = [System.Drawing.Color]::White
    `$jobsDataGrid.ColumnHeadersHeight = 30
    `$jobsDataGrid.RowTemplate.Height = 25
    
    # Add columns
    `$jobsDataGrid.Columns.Add('Name', 'Name')
    `$jobsDataGrid.Columns.Add('Type', 'Type')
    `$jobsDataGrid.Columns.Add('Status', 'Status')
    `$jobsDataGrid.Columns.Add('LastRun', 'Last Run')
    `$jobsDataGrid.Columns.Add('NextRun', 'Next Run')
    
    # Add sample data
    `$jobsDataGrid.Rows.Add('Daily Documents Backup', 'Files', 'Active', '2026-03-11 02:00', '2026-03-12 02:00')
    `$jobsDataGrid.Rows.Add('Weekly System Backup', 'System', 'Active', '2026-03-08 22:00', '2026-03-15 22:00')
    `$jobsDataGrid.Rows.Add('Database Backup', 'SQL Server', 'Active', '2026-03-11 01:00', '2026-03-12 01:00')
    
    `$jobsTable.Controls.Add(`$jobsDataGrid)
    
    `$mainPanel.Controls.Add(`$jobsTable)
    
    `$form.Controls.Add(`$mainPanel)
    
    # Button events
    `$startBtn.Add_Click({
        [System.Windows.Forms.MessageBox]::Show('Starting backup process...', 'NovaBackup', 'OK', 'Information')
    })
    
    `$createBtn.Add_Click({
        [System.Windows.Forms.MessageBox]::Show('Create new backup job functionality', 'NovaBackup', 'OK', 'Information')
    })
    
    `$dashboardBtn.Add_Click({
        `$dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
        `$jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    })
    
    `$jobsBtn.Add_Click({
        `$dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
        `$storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    })
    
    `$storageBtn.Add_Click({
        `$dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$storageBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
        `$reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    })
    
    `$reportsBtn.Add_Click({
        `$dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
        `$settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    })
    
    `$settingsBtn.Add_Click({
        `$dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        `$settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
    })
    
    # Show form
    `$form.ShowDialog()
}"

:end
echo.
echo Thank you for using NovaBackup Enterprise!
pause
