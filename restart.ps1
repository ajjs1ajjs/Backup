# Force Restart NovaBackup Server
# Run this in PowerShell as Administrator

Write-Host "=== Stopping all nova-backup processes..." -ForegroundColor Yellow
Get-Process nova-backup -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 5

Write-Host "=== Starting server..." -ForegroundColor Green
cd D:\WORK_CODE\Backup
.\nova-backup.exe server
