# Force kill and restart
Write-Host "=== Вбиваємо всі процеси на порту 8050..." -ForegroundColor Yellow

# Find process on port 8050
$proc = Get-NetTCPConnection -LocalPort 8050 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique
if ($proc) {
    Write-Host "Знайдено процес PID: $proc" -ForegroundColor Gray
    Stop-Process -Id $proc -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 3
}

# Also kill by name
Get-Process nova-backup -ErrorAction SilentlyContinue | Stop-Process -Force

Start-Sleep -Seconds 2

Write-Host "=== Запуск сервера..." -ForegroundColor Green
cd D:\WORK_CODE\Backup
.\nova-backup.exe server
