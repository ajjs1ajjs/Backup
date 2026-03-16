# Відкрити сторінки для перевірки змін
$pages = @(
    "http://localhost:8050/verify-changes.html",
    "http://localhost:8050/restore.html",
    "http://localhost:8050/jobs.html"
)

foreach ($page in $pages) {
    Start-Process $page
    Start-Sleep -Seconds 1
}

Write-Host "Відкрито 3 вкладки. Перевірте зміни!" -ForegroundColor Green
Write-Host "Натисніть Ctrl+Shift+R для повного оновлення кешу" -ForegroundColor Yellow
