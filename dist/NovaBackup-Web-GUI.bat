@echo off
title NovaBackup v6.0 Web GUI
color 0B

echo ========================================
echo    NovaBackup v6.0 - Web Interface
echo ========================================
echo.

REM Check if nova.exe exists
if not exist "..\nova.exe" (
    echo ERROR: nova.exe not found!
    echo Please run from the NovaBackup installation directory.
    pause
    exit /b 1
)

REM Start API Server
echo Starting NovaBackup API Server...
start "NovaBackup API" /min cmd /c "..\nova.exe api start --port 8080"
timeout /t 3 /nobreak >nul

REM Start Web Server
echo Starting Web Interface...
cd web-ui
python3 -m http.server 8081
pause
