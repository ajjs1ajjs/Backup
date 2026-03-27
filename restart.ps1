# Force Restart NovaBackup Server
# Usage: .\restart.ps1 [-Port 8050]
# Run this in PowerShell as Administrator

param(
    [int]$Port = 8050
)

# Get script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location -Path $scriptDir -ErrorAction SilentlyContinue

Write-Host "=== Stopping all nova-backup processes ===" -ForegroundColor Yellow
Get-Process nova-backup -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 5

Write-Host "=== Starting server ===" -ForegroundColor Green
Write-Host "Port: $Port" -ForegroundColor Gray

.\nova-backup.exe server
