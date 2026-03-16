# Force kill all nova-backup processes
Get-Process | Where-Object {$_.Name -like "*nova*"} | Stop-Process -Force -ErrorAction SilentlyContinue

# Kill process by PID if still running
$processId = 12428
$process = Get-Process -Id $processId -ErrorAction SilentlyContinue
if ($process) {
    Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
    Write-Host "Killed process $processId"
} else {
    Write-Host "Process $processId not found"
}

# Wait and check port
Start-Sleep -Seconds 3
$port = netstat -ano | Select-String ":8050.*LISTENING"
if ($port) {
    Write-Host "Port 8050 still in use!"
    Write-Host $port
} else {
    Write-Host "Port 8050 is free!"
}

# Start new server
Write-Host "Starting new server..."
cd D:\WORK_CODE\Backup
Start-Process ".\nova-backup.exe" -ArgumentList "server" -WindowStyle Normal
