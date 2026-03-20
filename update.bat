@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM ========================================
REM   NovaBackup - Update Script
REM   Full update with service restart
REM ========================================

echo ========================================
echo   NovaBackup - Update
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

REM Stop service
echo [*] Stopping NovaBackup service...
sc stop NovaBackup >nul 2>&1
timeout /t 3 /nobreak >nul

REM Kill any running process
taskkill /F /IM NovaBackup.exe >nul 2>&1
timeout /t 2 /nobreak >nul

REM Backup current version
echo [*] Creating backup...
set "BACKUP_DIR=C:\Program Files\NovaBackup\backup_%date:~-4,4%%date:~-7,2%%date:~-10,2%"
if exist "%BACKUP_DIR%" rmdir /s /q "%BACKUP_DIR%"
mkdir "%BACKUP_DIR%" 2>nul
copy /Y "C:\Program Files\NovaBackup\NovaBackup.exe" "%BACKUP_DIR%\" 2>nul

REM Download latest version from GitHub Releases
echo [*] Downloading latest version from GitHub...
set "RELEASE_URL=https://github.com/ajjs1ajjs/Backup/releases/latest/download/novabackup.exe"

powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%RELEASE_URL%' -OutFile 'C:\Program Files\NovaBackup\NovaBackup.exe' -UseBasicParsing"
if %errorLevel% neq 0 (
    echo [ERROR] Download failed!
    echo Restoring backup...
    copy /Y "%BACKUP_DIR%\NovaBackup.exe" "C:\Program Files\NovaBackup\NovaBackup.exe"
    pause
    exit /b 1
)

REM Copy web files
echo [*] Updating web interface...
set "WEB_DIR=C:\Program Files\NovaBackup\web"
if not exist "%WEB_DIR%" mkdir "%WEB_DIR%"

REM Download web files from raw GitHub
set "RAW_URL=https://raw.githubusercontent.com/ajjs1ajjs/Backup/main"
powershell -Command "Invoke-WebRequest -Uri '%RAW_URL%/web/index.html' -OutFile '%WEB_DIR%\index.html' -UseBasicParsing"
powershell -Command "Invoke-WebRequest -Uri '%RAW_URL%/web/quick-backup.html' -OutFile '%WEB_DIR%\quick-backup.html' -UseBasicParsing"
powershell -Command "Invoke-WebRequest -Uri '%RAW_URL%/web/database-backup.html' -OutFile '%WEB_DIR%\database-backup.html' -UseBasicParsing"
powershell -Command "Invoke-WebRequest -Uri '%RAW_URL%/web/vm-backup.html' -OutFile '%WEB_DIR%\vm-backup.html' -UseBasicParsing"
powershell -Command "Invoke-WebRequest -Uri '%RAW_URL%/web/restore.html' -OutFile '%WEB_DIR%\restore.html' -UseBasicParsing"

REM Start service
echo [*] Starting NovaBackup service...
sc start NovaBackup >nul
if %errorLevel% neq 0 (
    echo [WARNING] Service start failed, starting in background...
    start "" "C:\Program Files\NovaBackup\NovaBackup.exe" server
    timeout /t 2 /nobreak >nul
)

REM Wait for service
timeout /t 3 /nobreak >nul

REM Verify
sc query NovaBackup | find "RUNNING" >nul
if %errorLevel% equ 0 (
    echo [OK] Service: RUNNING
) else (
    echo [!] Service: Background mode
)

REM Cleanup old backups (keep last 3)
echo [*] Cleaning up old backups...
for /f "delims=" %%i in ('dir /b /o-d "%BACKUP_DIR:~0,-18%" ^| findstr /n "^" ^| findstr /v "^1:" ^| findstr /v "^2:" ^| findstr /v "^3:"') do (
    rmdir /s /q "%BACKUP_DIR:~0,-18%\%%i" 2>nul
)

echo.
echo ========================================
echo   Update Complete!
echo ========================================
echo.
echo Web UI: http://localhost:8050
echo.
echo New Features:
echo   - VM Backup in Quick Backup page
echo   - Database Backup (SQL Server, PostgreSQL, Oracle)
echo   - VM Backup (Hyper-V, KVM)
echo.
echo Opening Web UI...
timeout /t 2 /nobreak >nul
start "" http://localhost:8050

echo.
timeout /t 2 /nobreak >nul
exit /b 0
