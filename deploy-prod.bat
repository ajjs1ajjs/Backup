@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM ========================================
REM   NovaBackup - Deploy to Production
REM   Deploy local build to production server
REM ========================================

echo ========================================
echo   NovaBackup - Production Deploy
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

REM Set paths
set "SOURCE=D:\PROJECT\Backup\novabackup.exe"
set "DEST=C:\Program Files\NovaBackup\NovaBackup.exe"
set "BACKUP_DIR=C:\Program Files\NovaBackup\backup_%date:~-4,4%%date:~-7,2%%date:~-10,2%"

REM Check if source exists
if not exist "%SOURCE%" (
    echo [ERROR] Source file not found: %SOURCE%
    echo Please build first: go build -o novabackup.exe ./cmd/novabackup
    pause
    exit /b 1
)

echo [*] Stopping NovaBackup service...
sc stop NovaBackup >nul 2>&1
timeout /t 3 /nobreak >nul

REM Kill any running process
taskkill /F /IM NovaBackup.exe >nul 2>&1
timeout /t 2 /nobreak >nul

REM Backup current version
echo [*] Creating backup of current version...
mkdir "%BACKUP_DIR%" 2>nul
copy /Y "%DEST%" "%BACKUP_DIR%\" 2>nul
echo     Backup saved to: %BACKUP_DIR%

REM Deploy new version
echo [*] Deploying new version...
copy /Y "%SOURCE%" "%DEST%"
if %errorLevel% neq 0 (
    echo [ERROR] Copy failed!
    echo Restoring backup...
    copy /Y "%BACKUP_DIR%\NovaBackup.exe" "%DEST%"
    pause
    exit /b 1
)

REM Start service
echo [*] Starting NovaBackup service...
sc start NovaBackup >nul
if %errorLevel% neq 0 (
    echo [WARNING] Service start failed, starting in background...
    start "" "%DEST%" server
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

echo.
echo ========================================
echo   Deploy Complete!
echo ========================================
echo.
echo Source: %SOURCE%
echo Destination: %DEST%
echo Backup: %BACKUP_DIR%
echo.
echo Web UI: http://localhost:8050
echo.
timeout /t 2 /nobreak >nul
exit /b 0
