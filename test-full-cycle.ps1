# PowerShell Script to test full backup/restore cycle
$ErrorActionPreference = "Stop"

Write-Host "=== Тестування повного циклу Backup/Restore ===" -ForegroundColor Cyan

# Step 1: Login
Write-Host "`n1. Логін..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8050/api/auth/login" `
        -Method Post `
        -Body $loginBody `
        -ContentType "application/json"

    $token = $loginResponse.token
    Write-Host "   ✅ Успішний вхід! Token отримано" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Помилка логіну: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = $token
    "Content-Type" = "application/json"
}

# Step 2: Check existing jobs
Write-Host "`n2. Перевірка існуючих завдань..." -ForegroundColor Yellow
try {
    $jobs = Invoke-RestMethod -Uri "http://localhost:8050/api/jobs" -Headers $headers
    Write-Host "   Знайдено $($jobs.jobs.Count) завдань" -ForegroundColor Gray

    if ($jobs.jobs.Count -gt 0) {
        $jobs.jobs | ForEach-Object {
            Write-Host "   - $($_.name) (ID: $($_.id))" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "   ❌ Помилка: $_" -ForegroundColor Red
}

# Step 3: Check backup sessions
Write-Host "`n3. Перевірка сесій бекапу..." -ForegroundColor Yellow
try {
    $sessions = Invoke-RestMethod -Uri "http://localhost:8050/api/backup/sessions" -Headers $headers
    Write-Host "   Знайдено $($sessions.sessions.Count) сесій" -ForegroundColor Gray

    if ($sessions.sessions.Count -gt 0) {
        $sessions.sessions | ForEach-Object {
            $sizeGB = [math]::Round($_.total_size / 1GB, 2)
            Write-Host "   - $($_.job_name): $sizeGB GB ($($_.status))" -ForegroundColor Gray
        }
    } else {
        Write-Host "   ⚠️ Сесій немає. Створіть бекап через UI" -ForegroundColor Yellow
    }
} catch {
    Write-Host "   ❌ Помилка: $_" -ForegroundColor Red
}

# Step 4: Open browser pages
Write-Host "`n4. Відкриття сторінок для перевірки..." -ForegroundColor Yellow

Write-Host "   📄 Restore page..." -ForegroundColor Gray
Start-Process "http://localhost:8050/restore.html"

Start-Sleep -Seconds 2

Write-Host "   📄 Verify changes page..." -ForegroundColor Gray
Start-Process "http://localhost:8050/verify-changes.html"

Write-Host "`n✅ Тестування завершено!" -ForegroundColor Green
Write-Host "`nВідкрийте:" -ForegroundColor Cyan
Write-Host "  - http://localhost:8050/verify-changes.html (перевірка змін)" -ForegroundColor White
Write-Host "  - http://localhost:8050/restore.html (відновлення)" -ForegroundColor White
Write-Host "  - http://localhost:8050/jobs.html (створення бекапів)" -ForegroundColor White
