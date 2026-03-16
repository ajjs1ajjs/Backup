@echo off
echo Stopping NovaBackup...
taskkill /F /IM nova-backup.exe 2>nul
timeout /t 3 /nobreak >nul
echo Starting NovaBackup with new version...
cd /d "%~dp0"
start "" nova-backup.exe server
echo Done! Server is restarting...
timeout /t 5 /nobreak >nul
echo You can close this window.
