@echo off
echo ========================================
echo NovaBackup - Clean and Restart
echo ========================================

:: Kill running server
echo Stopping NovaBackup server...
taskkill /F /IM nova-backup.exe 2>nul
timeout /t 2 /nobreak >nul

:: Start new server
echo Starting new server...
cd /d "%~dp0"
start "" nova-backup.exe server

echo.
echo Server started!
echo.
echo IMPORTANT: Wait 5 seconds for migration to complete, then:
echo 1. Open http://localhost:8050/quick-backup.html
echo 2. Create NEW backup job (DO NOT reuse old jobs)
echo 3. When entering paths, use forward slashes: D:/Documents/MySoft
echo.
pause
