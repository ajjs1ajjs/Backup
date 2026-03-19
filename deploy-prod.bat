@echo off
chcp 65001 >nul
echo ========================================
echo NovaBackup - Deploy to Production
echo ========================================
echo.

echo [1/4] Stopping NovaBackup service...
taskkill /F /IM novabackup.exe 2>nul
taskkill /F /IM nova-service.exe 2>nul
timeout /t 2 /nobreak >nul

echo [2/4] Building new version...
cd /d "%~dp0"
go build -o novabackup.exe .\cmd\novabackup
if %ERRORLEVEL% neq 0 (
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo [3/4] Copying to installation directory...
copy /Y novabackup.exe "C:\Program Files\NovaBackup\NovaBackup.exe" 2>nul
if %ERRORLEVEL% neq 0 (
    echo WARNING: Could not copy to Program Files (may require admin rights)
    echo Please run this script as Administrator
)

echo [4/4] Starting NovaBackup service...
start "" "C:\Program Files\NovaBackup\NovaBackup.exe" server 2>nul
if %ERRORLEVEL% neq 0 (
    start "" novabackup.exe server
)

echo.
echo ========================================
echo Deployment complete!
echo ========================================
echo.
echo Web UI: http://localhost:8050
echo Login: admin / admin123
echo.
pause
