@echo off
setlocal EnableDelayedExpansion

:: ============================================================================
:: NovaBackup Enterprise v7.0 - Production Installation Script
:: For Windows Server / Enterprise Deployment
:: ============================================================================

echo.
echo ============================================================================
echo   NovaBackup Enterprise v7.0 - Production Installation
echo   Copyright (c) 2024 NovaBackup Team. All rights reserved.
echo ============================================================================
echo.

:: Check for administrator privileges
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Administrator privileges required!
    echo.
    echo Please right-click this file and select:
    echo   "Run as administrator"
    echo.
    pause
    exit /b 1
)

echo [OK] Administrator privileges confirmed
echo.

:: Configuration
set "INSTALL_DIR=C:\Program Files\NovaBackup"
set "DATA_DIR=C:\ProgramData\NovaBackup"
set "LOGS_DIR=%DATA_DIR%\Logs"
set "CONFIG_DIR=%DATA_DIR\Config"
set "BACKUPS_DIR=%DATA_DIR\Backups"
set "GITHUB_URL=https://github.com/ajjs1ajjs/Backup/releases/latest/download"
set "RAW_URL=https://raw.githubusercontent.com/ajjs1ajjs/Backup/main"

:: Get version from command line or use latest
set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=latest"

echo [*] NovaBackup Enterprise v7.0 Installation
echo      Version: %VERSION%
echo      Install Directory: %INSTALL_DIR%
echo      Data Directory: %DATA_DIR%
echo.

:: Step 1: Stop existing service
echo [1/7] Stopping existing NovaBackup service...
net stop NovaBackup >nul 2>&1
taskkill /F /IM NovaBackup.exe >nul 2>&1
timeout /t 2 /nobreak >nul
echo      Done
echo.

:: Step 2: Create directories
echo [2/7] Creating directories...
mkdir "%INSTALL_DIR%" 2>nul
mkdir "%DATA_DIR%" 2>nul
mkdir "%LOGS_DIR%" 2>nul
mkdir "%CONFIG_DIR%" 2>nul
mkdir "%BACKUPS_DIR%" 2>nul
echo      Done
echo.

:: Step 3: Download latest release
echo [3/7] Downloading NovaBackup from GitHub...
set "DOWNLOAD_FILE=%TEMP%\novabackup.zip"

call :download "%GITHUB_URL%/novabackup-windows-amd64.zip" "%DOWNLOAD_FILE%"
if %errorLevel% neq 0 (
    echo [WARNING] Release download failed. Trying raw repository...
    call :download "%RAW_URL%/novabackup.exe" "%TEMP%\novabackup.exe"
    if %errorLevel% equ 0 (
        copy /Y "%TEMP%\novabackup.exe" "%INSTALL_DIR%\\NovaBackup.exe"
    ) else (
        echo [WARNING] Raw download failed! Using local build...
        if exist "novabackup.exe" (
            copy /Y "novabackup.exe" "%INSTALL_DIR%\"
        ) else (
            echo [ERROR] No local build found!
            pause
            exit /b 1
        )
    )
) else (
    :: Extract zip
    powershell -Command "Expand-Archive -Path '%DOWNLOAD_FILE%' -DestinationPath '%INSTALL_DIR%' -Force"
    del /Q "%DOWNLOAD_FILE%"
)
echo      Done
echo.

:: Step 4: Copy configuration files
echo [4/7] Installing configuration files...

:: Create default config
(
echo {
echo   "server": {
echo     "ip": "0.0.0.0",
echo     "port": 8050,
echo     "https": false,
echo     "https_port": 8443
echo   },
echo   "backup": {
echo     "default_path": "%BACKUPS_DIR%",
echo     "retention_days": 30
echo   },
echo   "logging": {
echo     "level": "info",
echo     "file": "%LOGS_DIR%\novabackup.log"
echo   },
echo   "database": {
echo     "path": "%DATA_DIR%\novabackup.db"
echo   },
echo   "version": "7.0.0"
echo }
) > "%CONFIG_DIR%\config.json"
echo      Done
echo.

:: Step 5: Install Windows Service
echo [5/7] Installing Windows Service...
cd /d "%INSTALL_DIR%"
NovaBackup.exe install
if %errorLevel% neq 0 (
    echo [ERROR] Service installation failed!
    pause
    exit /b 1
)
echo      Done
echo.

:: Step 6: Configure Firewall
echo [6/7] Configuring Windows Firewall...
netsh advfirewall firewall add rule name="NovaBackup Web UI" dir=in action=allow protocol=TCP localport=8050 >nul 2>&1
netsh advfirewall firewall add rule name="NovaBackup HTTPS" dir=in action=allow protocol=TCP localport=8443 >nul 2>&1
echo      Done
echo.

:: Step 7: Start Service
echo [7/7] Starting NovaBackup Service...
net start NovaBackup
if %errorLevel% neq 0 (
    echo [WARNING] Service failed to start. Trying manual start...
    timeout /t 3 /nobreak >nul
    start "" "%INSTALL_DIR%\NovaBackup.exe" start
)
echo      Done
echo.

:: Create desktop shortcut
echo [*] Creating desktop shortcut...
powershell -Command "$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%USERPROFILE%\Desktop\NovaBackup Enterprise.lnk'); $Shortcut.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $Shortcut.WorkingDirectory = '%INSTALL_DIR%'; $Shortcut.Description = 'NovaBackup Enterprise v7.0'; $Shortcut.Save()"

:: Installation summary
echo.
echo ============================================================================
echo   Installation Complete Successfully!
echo ============================================================================
echo.
echo Installation Details:
echo   Product: NovaBackup Enterprise v7.0
echo   Install Directory: %INSTALL_DIR%
echo   Data Directory: %DATA_DIR%
echo   Configuration: %CONFIG_DIR%\config.json
echo   Logs: %LOGS_DIR%
echo.
echo Access Information:
echo   Web UI: http://localhost:8050
echo   Default Login: admin
echo   Default Password: admin123
echo.
echo IMPORTANT: Please change the default password after first login!
echo.
echo Services:
echo   NovaBackup Service: Running
echo   Auto-start: Enabled
echo.
echo Next Steps:
echo   1. Open http://localhost:8050 in your browser
echo   2. Login with admin/admin123
echo   3. Change the default password
echo   4. Create your first backup job
echo.
echo Support:
echo   Documentation: https://github.com/ajjs1ajjs/Backup/wiki
echo   Issues: https://github.com/ajjs1ajjs/Backup/issues
echo   Email: support@novabackup.local
echo.
echo ============================================================================
echo.

:: Open web UI
echo [*] Opening Web UI...
timeout /t 3 /nobreak >nul
start http://localhost:8050

pause
exit /b 0

:download
set "DOWNLOAD_URL=%~1"
set "DOWNLOAD_OUT=%~2"
powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%DOWNLOAD_OUT%' -UseBasicParsing"
if %errorLevel% equ 0 exit /b 0
where curl.exe >nul 2>&1
if %errorLevel% equ 0 (
    curl.exe -L --retry 3 --retry-delay 2 -o "%DOWNLOAD_OUT%" "%DOWNLOAD_URL%"
    if %errorLevel% equ 0 exit /b 0
)
exit /b 1
