# Restart NovaBackup Server
# Usage: .\restart-server.ps1 [-Port 8050]

param(
    [int]$Port = 8050
)

# Get script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location -Path $scriptDir -ErrorAction SilentlyContinue

Write-Host "=== Зупинка сервера ===" -ForegroundColor Yellow
Stop-Process -Name "nova-backup" -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 5

Write-Host "=== Запуск сервера ===" -ForegroundColor Yellow
Write-Host "Port: $Port" -ForegroundColor Gray

.\nova-backup.exe server
