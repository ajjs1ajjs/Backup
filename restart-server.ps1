# Restart NovaBackup Server
Write-Host "=== Зупинка сервера..." -ForegroundColor Yellow
Stop-Process -Name "nova-backup" -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 5

Write-Host "=== Запуск сервера..." -ForegroundColor Yellow
cd D:\WORK_CODE\Backup
.\nova-backup.exe server
