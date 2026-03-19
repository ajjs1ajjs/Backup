@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM ========================================
REM   NovaBackup v7.0 - Auto Installation
REM   One-command fully automated install
REM ========================================

echo ========================================
echo   NovaBackup v7.0 - Auto Installation
echo ========================================
echo.

REM Check admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [OK] Restarting with administrator privileges...
    powershell -Command "Start-Process cmd -ArgumentList '/c', '%~f0' -Verb RunAs"
    exit /b 0
)

echo [OK] Administrator rights confirmed
echo.

REM Set installation directory
set "INSTALL_DIR=C:\Program Files\NovaBackup"
set "DATA_DIR=C:\ProgramData\NovaBackup"
set "RAW_URL=https://raw.githubusercontent.com/ajjs1ajjs/Backup/main"

echo [*] Downloading from GitHub...
echo.

REM Create temporary directory
set "TEMP_DIR=%TEMP%\NovaBackup_Install"
if exist "%TEMP_DIR%" rmdir /s /q "%TEMP_DIR%"
mkdir "%TEMP_DIR%"

REM Download latest release
powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%RAW_URL%/novabackup.exe' -OutFile '%TEMP_DIR%\novabackup.exe' -UseBasicParsing"
if %errorLevel% neq 0 (
    echo [ERROR] Download failed!
    rmdir /s /q "%TEMP_DIR%"
    pause
    exit /b 1
)

echo [*] Creating directories...
mkdir "%INSTALL_DIR%" 2>nul
mkdir "%DATA_DIR%" 2>nul
mkdir "%DATA_DIR%\Logs" 2>nul
mkdir "%DATA_DIR%\Backups" 2>nul
mkdir "%DATA_DIR%\Config" 2>nul

echo [*] Copying files...
copy /Y "%TEMP_DIR%\novabackup.exe" "%INSTALL_DIR%\NovaBackup.exe"

echo [*] Installing Windows Service...
cd /d "%INSTALL_DIR%"

REM Stop and remove existing service
sc stop NovaBackup >nul 2>&1
timeout /t 1 /nobreak >nul
sc delete NovaBackup >nul 2>&1
timeout /t 1 /nobreak >nul

REM Create new service with auto-start
sc create NovaBackup binPath= "\"%INSTALL_DIR%\NovaBackup.exe\" server" start= auto DisplayName= "NovaBackup" >nul
if %errorLevel% neq 0 (
    echo [ERROR] Service creation failed!
    pause
    exit /b 1
)

echo [*] Starting Service...
sc start NovaBackup >nul
if %errorLevel% neq 0 (
    timeout /t 2 /nobreak >nul
    sc start NovaBackup >nul
)

REM Wait for service to start
timeout /t 3 /nobreak >nul

REM Verify and fallback
sc query NovaBackup | find "RUNNING" >nul
if %errorLevel% neq 0 (
    start "" "%INSTALL_DIR%\NovaBackup.exe" server
    timeout /t 2 /nobreak >nul
)

REM Cleanup
rmdir /s /q "%TEMP_DIR%"

REM Create shortcuts
powershell -Command "$S = (New-Object -ComObject WScript.Shell).CreateShortcut('%USERPROFILE%\Desktop\NovaBackup.lnk'); $S.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $S.WorkingDirectory = '%INSTALL_DIR%'; $S.Save()"
powershell -Command "$S = (New-Object -ComObject WScript.Shell).CreateShortcut('%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup.lnk'); $S.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $S.WorkingDirectory = '%INSTALL_DIR%'; $S.Save()"

echo.
echo ========================================
echo   Installation Complete!
echo ========================================
echo.
echo Installation: %INSTALL_DIR%
echo Data: %DATA_DIR%
echo.
sc query NovaBackup | find "RUNNING" >nul && echo [OK] Service: RUNNING || echo [!] Service: Background mode
echo.
echo Web UI: http://localhost:8050
echo Login: admin
echo Password: admin123
echo.
echo Opening Web UI...
timeout /t 2 /nobreak >nul
start "" http://localhost:8050
echo.
echo Done!
timeout /t 2 /nobreak >nul
exit /b 0
