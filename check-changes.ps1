# PowerShell script to check changes in restore.html
Write-Host "=== Перевірка змін у restore.html ===" -ForegroundColor Cyan

# Use script directory as base path (no hardcoded paths)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = if ($scriptDir) { $scriptDir } else { "." }
$filePath = Join-Path $projectRoot "web\restore.html"

if (Test-Path $filePath) {
    $content = Get-Content $filePath -Raw

    Write-Host "`n1. Кнопка оновлення:" -ForegroundColor Yellow
    if ($content -match "🔄 Оновити") {
        Write-Host "   ✅ Знайдено кнопку оновлення" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Кнопку не знайдено" -ForegroundColor Red
    }

    Write-Host "`n2. Console logging:" -ForegroundColor Yellow
    if ($content -match "console.log.*Loading backup points") {
        Write-Host "   ✅ Logging додано" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Logging не знайдено" -ForegroundColor Red
    }

    Write-Host "`n3. API integration:" -ForegroundColor Yellow
    if ($content -match "async function loadBackupSessions") {
        Write-Host "   ✅ Функція завантаження сесій існує" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Функція не знайдена" -ForegroundColor Red
    }

    Write-Host "`n4. Start restore function:" -ForegroundColor Yellow
    if ($content -match "async function startRestore" -and $content -match "/api/restore/files") {
        Write-Host "   ✅ Відновлення через API" -ForegroundColor Green
    } else {
        Write-Host "   ❌ API відновлення не знайдено" -ForegroundColor Red
    }
} else {
    Write-Host "❌ Файл не знайдено!" -ForegroundColor Red
}

Write-Host "`n=== Перевірка jobs.html ===" -ForegroundColor Cyan
$jobsFile = Join-Path $projectRoot "web\jobs.html"

if (Test-Path $jobsFile) {
    $content = Get-Content $jobsFile -Raw

    Write-Host "`n1. Інкрементальне бекапування:" -ForegroundColor Yellow
    if ($content -match "job-incremental") {
        Write-Host "   ✅ Опція додана" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Опція не знайдена" -ForegroundColor Red
    }

    Write-Host "`n2. Дедуплікація:" -ForegroundColor Yellow
    if ($content -match "job-deduplication") {
        Write-Host "   ✅ Опція додана" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Опція не знайдена" -ForegroundColor Red
    }

    Write-Host "`n3. Політика зберігання:" -ForegroundColor Yellow
    if ($content -match "job-retention-days" -and $content -match "job-retention-copies") {
        Write-Host "   ✅ Налаштування додані" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Налаштування не знайдені" -ForegroundColor Red
    }
}

Write-Host "`n=== Перевірка сервера ===" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8050" -TimeoutSec 5 -UseBasicParsing
    Write-Host "✅ Сервер працює на http://localhost:8050" -ForegroundColor Green
    Write-Host "   Status: $($response.StatusCode)" -ForegroundColor Gray
} catch {
    Write-Host "❌ Сервер не відповідає" -ForegroundColor Red
}

Write-Host "`n"
