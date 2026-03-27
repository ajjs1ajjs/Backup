# Force kill all nova-backup processes
# Usage: .\force_kill.ps1 [-Port 8050]

param(
    [int]$Port = 8050
)

# Get script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location -Path $scriptDir -ErrorAction SilentlyContinue

Write-Host "=== Stopping all NovaBackup processes ===" -ForegroundColor Cyan

# Kill by name patterns
$processPatterns = @("nova-backup", "novabackup", "NovaBackup")
foreach ($pattern in $processPatterns) {
    Get-Process | Where-Object {$_.Name -like "*$pattern*"} | Stop-Process -Force -ErrorAction SilentlyContinue
}

# Wait and check port
Write-Host "Waiting for port $Port to be free..." -ForegroundColor Gray
Start-Sleep -Seconds 3

$portInUse = netstat -ano | Select-String ":$Port.*LISTENING"
if ($portInUse) {
    Write-Host "Port $Port still in use, finding process..." -ForegroundColor Yellow
    foreach ($line in $portInUse) {
        $parts = $line.Line -split '\s+'
        $pid = $parts[-1]
        Write-Host "  Killing PID: $pid" -ForegroundColor Gray
        Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
    }
    Start-Sleep -Seconds 2
}

# Final check
$portInUse = netstat -ano | Select-String ":$Port.*LISTENING"
if ($portInUse) {
    Write-Host "WARNING: Port $Port still in use!" -ForegroundColor Red
} else {
    Write-Host "Port $Port is free!" -ForegroundColor Green
}

Write-Host "`nUse .\nova-manager.ps1 -Action start  to restart server" -ForegroundColor Gray
