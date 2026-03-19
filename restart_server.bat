@echo off
echo Stopping NovaBackup...
taskkill /IM nova-backup.exe /F 2>nul
timeout /t 2 /nobreak >nul
echo Starting NovaBackup...
cd /d "%~dp0"
start "" nova-backup.exe server
echo Done!
timeout /t 3 /nobreak >nul
