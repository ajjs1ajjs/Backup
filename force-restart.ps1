# Force kill and restart
# Usage: .\force-restart.ps1 [-Port 8050]

param(
    [int]$Port = 8050
)

# Get script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location -Path $scriptDir -ErrorAction SilentlyContinue

Write-Host "=== Вбиваємо всі процеси на порту $Port ===" -ForegroundColor Yellow

# Find process on port
$proc = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess -Unique
if ($proc) {
    Write-Host "Знайдено процес PID: $proc" -ForegroundColor Gray
    Stop-Process -Id $proc -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 3
}

# Also kill by name
Get-Process nova-backup -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 2

Write-Host "=== Запуск сервера ===" -ForegroundColor Green
Write-Host "Port: $Port" -ForegroundColor Gray
Write-Host "Directory: $(Get-Location)" -ForegroundColor Gray

.\nova-backup.exe server
