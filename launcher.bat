@echo off
title NovaBackup v6.0 Launcher
color 0B

echo ========================================
echo    NovaBackup v6.0 - Enterprise Backup
echo ========================================
echo.
echo Starting NovaBackup Services...
echo.

REM Start CLI Backend
echo [1/3] Starting NovaBackup CLI Backend...
start "NovaBackup CLI" /min cmd /c "nova.exe api start --port 8081"
timeout /t 3 /nobreak >nul

REM Start Web GUI
echo [2/3] Starting Web GUI Interface...
start "NovaBackup Web GUI" cmd /c "cd web-ui && python3 -m http.server 8080"
timeout /t 2 /nobreak >nul

REM Open Browser
echo [3/3] Opening Web Interface...
start http://localhost:8080

echo.
echo ========================================
echo NovaBackup v6.0 is now running!
echo.
echo Web Interface: http://localhost:8080
echo API Server:   http://localhost:8081
echo Swagger Docs: http://localhost:8081/swagger
echo.
echo Press any key to stop all services...
pause >nul

REM Stop services
echo.
echo Stopping NovaBackup services...
taskkill /f /im python.exe >nul 2>&1
taskkill /f /im nova.exe >nul 2>&1
echo Services stopped.
timeout /t 2 /nobreak >nul
exit
