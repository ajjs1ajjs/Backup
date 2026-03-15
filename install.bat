@echo off
echo ========================================
echo   NovaBackup v7.0 - Installation
echo   Installation from GitHub
echo ========================================
echo.

REM Check admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Administrator rights required!
    echo.
    echo Please right-click -^> Run as administrator
    pause
    exit /b 1
)

echo [OK] Administrator rights confirmed
echo.

REM Set installation directory
set "INSTALL_DIR=C:\Program Files\NovaBackup"
set "DATA_DIR=C:\ProgramData\NovaBackup"
set "GITHUB_URL=https://github.com/ajjs1ajjs/Backup/releases/latest/download"
set "RAW_URL=https://raw.githubusercontent.com/ajjs1ajjs/Backup/main"

echo [*] Downloading latest release from GitHub...
echo      URL: %GITHUB_URL%
echo.

REM Create temporary directory
set "TEMP_DIR=%TEMP%\NovaBackup_Install"
if exist "%TEMP_DIR%" rmdir /s /q "%TEMP_DIR%"
mkdir "%TEMP_DIR%"

REM Download latest release
echo [*] Downloading novabackup-windows-amd64.exe...
call :download "%GITHUB_URL%/novabackup-windows-amd64.exe" "%TEMP_DIR%\novabackup.exe"
if %errorLevel% neq 0 (
    echo [WARNING] Release download failed. Trying raw repository...
    call :download "%RAW_URL%/novabackup.exe" "%TEMP_DIR%\novabackup.exe"
)
if %errorLevel% neq 0 (
    echo [ERROR] Download failed!
    echo.
    echo Please download manually from:
    echo https://github.com/ajjs1ajjs/Backup/releases
    pause
    exit /b 1
)

echo [*] Creating installation directory...
mkdir "%INSTALL_DIR%" 2>nul
mkdir "%DATA_DIR%" 2>nul
mkdir "%DATA_DIR%\Logs" 2>nul
mkdir "%DATA_DIR%\Backups" 2>nul
mkdir "%DATA_DIR%\Config" 2>nul

echo [*] Copying files...
copy /Y "%TEMP_DIR%\novabackup.exe" "%INSTALL_DIR%\NovaBackup.exe"

echo [*] Installing Windows Service...
cd /d "%INSTALL_DIR%"
NovaBackup.exe install
if %errorLevel% neq 0 (
    echo [ERROR] Service installation failed!
    pause
    exit /b 1
)

echo [*] Starting Service...
NovaBackup.exe start

echo [*] Creating shortcuts...
powershell -Command "$WshShell = New-Object -ComObject WScript.Shell; $S = $WshShell.CreateShortcut('%USERPROFILE%\Desktop\NovaBackup.lnk'); $S.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $S.WorkingDirectory = '%INSTALL_DIR%'; $S.Description = 'NovaBackup Enterprise v7.0'; $S.Save()"

powershell -Command "$WshShell = New-Object -ComObject WScript.Shell; $S = $WshShell.CreateShortcut('%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup.lnk'); $S.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $S.WorkingDirectory = '%INSTALL_DIR%'; $S.Description = 'NovaBackup Enterprise v7.0'; $S.Save()"

REM Cleanup
echo [*] Cleaning up...
rmdir /s /q "%TEMP_DIR%"

echo.
echo ========================================
echo   Installation Complete Successfully!
echo ========================================
echo.
echo Installation Directory: %INSTALL_DIR%
echo Data Directory: %DATA_DIR%
echo.
echo Web UI: http://localhost:8050
echo Login: admin
echo Password: admin123
echo.
echo Starting NovaBackup...
timeout /t 2 /nobreak >nul
start "" http://localhost:8050

echo.
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
