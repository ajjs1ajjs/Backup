@echo off
title NovaBackup v6.0 Enterprise - Veeam Style GUI
color 0B

echo ========================================
echo    NovaBackup v6.0 Enterprise
echo    Veeam Style GUI Interface
echo ========================================
echo.

echo Starting NovaBackup GUI...
echo.

powershell -ExecutionPolicy Bypass -Command "& {
    Add-Type -AssemblyName System.Windows.Forms
    Add-Type -AssemblyName System.Drawing
    
    # Main Form
    $form = New-Object System.Windows.Forms.Form
    $form.Text = 'NovaBackup v6.0 Enterprise - Veeam Style'
    $form.Size = New-Object System.Drawing.Size(1200, 800)
    $form.StartPosition = 'CenterScreen'
    $form.BackColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
    
    # Header
    $header = New-Object System.Windows.Forms.Panel
    $header.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
    $header.Location = New-Object System.Drawing.Point(0, 0)
    $header.Size = New-Object System.Drawing.Size(1200, 80)
    
    $title = New-Object System.Windows.Forms.Label
    $title.Text = 'NovaBackup v6.0 Enterprise'
    $title.Font = New-Object System.Drawing.Font('Segoe UI', 20, [System.Drawing.FontStyle]::Bold)
    $title.ForeColor = [System.Drawing.Color]::White
    $title.Location = New-Object System.Drawing.Point(20, 15)
    $title.Size = New-Object System.Drawing.Size(400, 30)
    $header.Controls.Add($title)
    
    $subtitle = New-Object System.Windows.Forms.Label
    $subtitle.Text = 'Enterprise Backup & Recovery Platform'
    $subtitle.Font = New-Object System.Drawing.Font('Segoe UI', 12)
    $subtitle.ForeColor = [System.Drawing.Color]::White
    $subtitle.Location = New-Object System.Drawing.Point(20, 45)
    $subtitle.Size = New-Object System.Drawing.Size(400, 20)
    $header.Controls.Add($subtitle)
    
    $startBtn = New-Object System.Windows.Forms.Button
    $startBtn.Text = 'Start Backup'
    $startBtn.Location = New-Object System.Drawing.Point(900, 20)
    $startBtn.Size = New-Object System.Drawing.Size(120, 40)
    $startBtn.BackColor = [System.Drawing.Color]::FromArgb(40, 167, 69)
    $startBtn.ForeColor = [System.Drawing.Color]::White
    $startBtn.FlatStyle = 'Flat'
    $startBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10, [System.Drawing.FontStyle]::Bold)
    $header.Controls.Add($startBtn)
    
    $createBtn = New-Object System.Windows.Forms.Button
    $createBtn.Text = 'Create Job'
    $createBtn.Location = New-Object System.Drawing.Point(1030, 20)
    $createBtn.Size = New-Object System.Drawing.Size(120, 40)
    $createBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 120, 215)
    $createBtn.ForeColor = [System.Drawing.Color]::White
    $createBtn.FlatStyle = 'Flat'
    $createBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10, [System.Drawing.FontStyle]::Bold)
    $header.Controls.Add($createBtn)
    
    $form.Controls.Add($header)
    
    # Sidebar
    $sidebar = New-Object System.Windows.Forms.Panel
    $sidebar.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    $sidebar.Location = New-Object System.Drawing.Point(0, 80)
    $sidebar.Size = New-Object System.Drawing.Size(250, 680)
    
    $dashboardBtn = New-Object System.Windows.Forms.Button
    $dashboardBtn.Text = 'Dashboard'
    $dashboardBtn.Location = New-Object System.Drawing.Point(10, 10)
    $dashboardBtn.Size = New-Object System.Drawing.Size(230, 40)
    $dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
    $dashboardBtn.ForeColor = [System.Drawing.Color]::White
    $dashboardBtn.FlatStyle = 'Flat'
    $dashboardBtn.TextAlign = 'MiddleLeft'
    $dashboardBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $sidebar.Controls.Add($dashboardBtn)
    
    $jobsBtn = New-Object System.Windows.Forms.Button
    $jobsBtn.Text = 'Backup Jobs'
    $jobsBtn.Location = New-Object System.Drawing.Point(10, 60)
    $jobsBtn.Size = New-Object System.Drawing.Size(230, 40)
    $jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    $jobsBtn.ForeColor = [System.Drawing.Color]::White
    $jobsBtn.FlatStyle = 'Flat'
    $jobsBtn.TextAlign = 'MiddleLeft'
    $jobsBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $sidebar.Controls.Add($jobsBtn)
    
    $storageBtn = New-Object System.Windows.Forms.Button
    $storageBtn.Text = 'Storage'
    $storageBtn.Location = New-Object System.Drawing.Point(10, 110)
    $storageBtn.Size = New-Object System.Drawing.Size(230, 40)
    $storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    $storageBtn.ForeColor = [System.Drawing.Color]::White
    $storageBtn.FlatStyle = 'Flat'
    $storageBtn.TextAlign = 'MiddleLeft'
    $storageBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $sidebar.Controls.Add($storageBtn)
    
    $reportsBtn = New-Object System.Windows.Forms.Button
    $reportsBtn.Text = 'Reports'
    $reportsBtn.Location = New-Object System.Drawing.Point(10, 160)
    $reportsBtn.Size = New-Object System.Drawing.Size(230, 40)
    $reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    $reportsBtn.ForeColor = [System.Drawing.Color]::White
    $reportsBtn.FlatStyle = 'Flat'
    $reportsBtn.TextAlign = 'MiddleLeft'
    $reportsBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $sidebar.Controls.Add($reportsBtn)
    
    $settingsBtn = New-Object System.Windows.Forms.Button
    $settingsBtn.Text = 'Settings'
    $settingsBtn.Location = New-Object System.Drawing.Point(10, 210)
    $settingsBtn.Size = New-Object System.Drawing.Size(230, 40)
    $settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    $settingsBtn.ForeColor = [System.Drawing.Color]::White
    $settingsBtn.FlatStyle = 'Flat'
    $settingsBtn.TextAlign = 'MiddleLeft'
    $settingsBtn.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $sidebar.Controls.Add($settingsBtn)
    
    $form.Controls.Add($sidebar)
    
    # Main Content
    $mainPanel = New-Object System.Windows.Forms.Panel
    $mainPanel.BackColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
    $mainPanel.Location = New-Object System.Drawing.Point(250, 80)
    $mainPanel.Size = New-Object System.Drawing.Size(950, 680)
    
    # Status Cards
    $card1 = New-Object System.Windows.Forms.Panel
    $card1.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    $card1.Location = New-Object System.Drawing.Point(20, 20)
    $card1.Size = New-Object System.Drawing.Size(280, 150)
    
    $statusTitle = New-Object System.Windows.Forms.Label
    $statusTitle.Text = 'System Status'
    $statusTitle.Font = New-Object System.Drawing.Font('Segoe UI', 12, [System.Drawing.FontStyle]::Bold)
    $statusTitle.ForeColor = [System.Drawing.Color]::White
    $statusTitle.Location = New-Object System.Drawing.Point(10, 10)
    $statusTitle.Size = New-Object System.Drawing.Size(150, 25)
    $card1.Controls.Add($statusTitle)
    
    $statusValue = New-Object System.Windows.Forms.Label
    $statusValue.Text = 'All Systems Operational'
    $statusValue.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $statusValue.ForeColor = [System.Drawing.Color]::FromArgb(40, 167, 69)
    $statusValue.Location = New-Object System.Drawing.Point(10, 40)
    $statusValue.Size = New-Object System.Drawing.Size(250, 20)
    $card1.Controls.Add($statusValue)
    
    $jobsCount = New-Object System.Windows.Forms.Label
    $jobsCount.Text = 'Active Jobs: 3'
    $jobsCount.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $jobsCount.ForeColor = [System.Drawing.Color]::FromArgb(173, 181, 189)
    $jobsCount.Location = New-Object System.Drawing.Point(10, 70)
    $jobsCount.Size = New-Object System.Drawing.Size(150, 20)
    $card1.Controls.Add($jobsCount)
    
    $storageInfo = New-Object System.Windows.Forms.Label
    $storageInfo.Text = 'Storage Used: 69.7 GB'
    $storageInfo.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $storageInfo.ForeColor = [System.Drawing.Color]::FromArgb(173, 181, 189)
    $storageInfo.Location = New-Object System.Drawing.Point(10, 100)
    $storageInfo.Size = New-Object System.Drawing.Size(150, 20)
    $card1.Controls.Add($storageInfo)
    
    $mainPanel.Controls.Add($card1)
    
    # Activity Card
    $card2 = New-Object System.Windows.Forms.Panel
    $card2.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    $card2.Location = New-Object System.Drawing.Point(320, 20)
    $card2.Size = New-Object System.Drawing.Size(280, 150)
    
    $activityTitle = New-Object System.Windows.Forms.Label
    $activityTitle.Text = 'Recent Activity'
    $activityTitle.Font = New-Object System.Drawing.Font('Segoe UI', 12, [System.Drawing.FontStyle]::Bold)
    $activityTitle.ForeColor = [System.Drawing.Color]::White
    $activityTitle.Location = New-Object System.Drawing.Point(10, 10)
    $activityTitle.Size = New-Object System.Drawing.Size(150, 25)
    $card2.Controls.Add($activityTitle)
    
    $activityList = New-Object System.Windows.Forms.ListBox
    $activityList.Location = New-Object System.Drawing.Point(10, 40)
    $activityList.Size = New-Object System.Drawing.Size(260, 100)
    $activityList.BackColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
    $activityList.ForeColor = [System.Drawing.Color]::White
    $activityList.Font = New-Object System.Drawing.Font('Consolas', 9)
    $activityList.Items.Add('Daily Backup - Completed - 2h ago')
    $activityList.Items.Add('System Backup - Running - 45%')
    $activityList.Items.Add('Database Backup - Scheduled - 1h')
    $card2.Controls.Add($activityList)
    
    $mainPanel.Controls.Add($card2)
    
    # Storage Card
    $card3 = New-Object System.Windows.Forms.Panel
    $card3.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    $card3.Location = New-Object System.Drawing.Point(620, 20)
    $card3.Size = New-Object System.Drawing.Size(280, 150)
    
    $storageTitle = New-Object System.Windows.Forms.Label
    $storageTitle.Text = 'Storage Overview'
    $storageTitle.Font = New-Object System.Drawing.Font('Segoe UI', 12, [System.Drawing.FontStyle]::Bold)
    $storageTitle.ForeColor = [System.Drawing.Color]::White
    $storageTitle.Location = New-Object System.Drawing.Point(10, 10)
    $storageTitle.Size = New-Object System.Drawing.Size(150, 25)
    $card3.Controls.Add($storageTitle)
    
    $progressBar = New-Object System.Windows.Forms.ProgressBar
    $progressBar.Location = New-Object System.Drawing.Point(10, 40)
    $progressBar.Size = New-Object System.Drawing.Size(260, 20)
    $progressBar.Value = 69
    $card3.Controls.Add($progressBar)
    
    $storageDetails = New-Object System.Windows.Forms.Label
    $storageDetails.Text = '69.7 GB / 1 TB used (69.7%)'
    $storageDetails.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $storageDetails.ForeColor = [System.Drawing.Color]::FromArgb(173, 181, 189)
    $storageDetails.Location = New-Object System.Drawing.Point(10, 70)
    $storageDetails.Size = New-Object System.Drawing.Size(250, 20)
    $card3.Controls.Add($storageDetails)
    
    $compressionInfo = New-Object System.Windows.Forms.Label
    $compressionInfo.Text = 'Compression: 3.2:1 | Dedup: 5.1:1'
    $compressionInfo.Font = New-Object System.Drawing.Font('Segoe UI', 10)
    $compressionInfo.ForeColor = [System.Drawing.Color]::FromArgb(173, 181, 189)
    $compressionInfo.Location = New-Object System.Drawing.Point(10, 100)
    $compressionInfo.Size = New-Object System.Drawing.Size(250, 20)
    $card3.Controls.Add($compressionInfo)
    
    $mainPanel.Controls.Add($card3)
    
    # Jobs Table
    $tablePanel = New-Object System.Windows.Forms.Panel
    $tablePanel.BackColor = [System.Drawing.Color]::FromArgb(45, 45, 45)
    $tablePanel.Location = New-Object System.Drawing.Point(20, 190)
    $tablePanel.Size = New-Object System.Drawing.Size(880, 470)
    
    $tableTitle = New-Object System.Windows.Forms.Label
    $tableTitle.Text = 'Backup Jobs'
    $tableTitle.Font = New-Object System.Drawing.Font('Segoe UI', 14, [System.Drawing.FontStyle]::Bold)
    $tableTitle.ForeColor = [System.Drawing.Color]::White
    $tableTitle.Location = New-Object System.Drawing.Point(10, 10)
    $tableTitle.Size = New-Object System.Drawing.Size(150, 30)
    $tablePanel.Controls.Add($tableTitle)
    
    $grid = New-Object System.Windows.Forms.DataGridView
    $grid.Location = New-Object System.Drawing.Point(10, 50)
    $grid.Size = New-Object System.Drawing.Size(860, 400)
    $grid.BackgroundColor = [System.Drawing.Color]::FromArgb(30, 30, 30)
    $grid.ForeColor = [System.Drawing.Color]::White
    $grid.ColumnHeadersHeight = 30
    $grid.RowTemplate.Height = 25
    
    $grid.Columns.Add('Name', 'Name')
    $grid.Columns.Add('Type', 'Type')
    $grid.Columns.Add('Status', 'Status')
    $grid.Columns.Add('LastRun', 'Last Run')
    $grid.Columns.Add('NextRun', 'Next Run')
    
    $grid.Rows.Add('Daily Documents Backup', 'Files', 'Active', '2026-03-11 02:00', '2026-03-12 02:00')
    $grid.Rows.Add('Weekly System Backup', 'System', 'Active', '2026-03-08 22:00', '2026-03-15 22:00')
    $grid.Rows.Add('Database Backup', 'SQL Server', 'Active', '2026-03-11 01:00', '2026-03-12 01:00')
    
    $tablePanel.Controls.Add($grid)
    
    $mainPanel.Controls.Add($tablePanel)
    
    $form.Controls.Add($mainPanel)
    
    # Button Events
    $startBtn.Add_Click({
        [System.Windows.Forms.MessageBox]::Show('Starting backup process...', 'NovaBackup', 'OK', 'Information')
    })
    
    $createBtn.Add_Click({
        [System.Windows.Forms.MessageBox]::Show('Create new backup job functionality', 'NovaBackup', 'OK', 'Information')
    })
    
    $dashboardBtn.Add_Click({
        $dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
        $jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    })
    
    $jobsBtn.Add_Click({
        $dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
        $storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    })
    
    $storageBtn.Add_Click({
        $dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $storageBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
        $reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    })
    
    $reportsBtn.Add_Click({
        $dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
        $settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
    })
    
    $settingsBtn.Add_Click({
        $dashboardBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $jobsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $storageBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $reportsBtn.BackColor = [System.Drawing.Color]::FromArgb(60, 60, 60)
        $settingsBtn.BackColor = [System.Drawing.Color]::FromArgb(0, 168, 230)
    })
    
    $form.ShowDialog()
}"

echo.
echo NovaBackup GUI closed.
pause
