# Відкриття сторінок з очищенням кешу
Start-Process "http://localhost:8050/verify-changes.html"
Start-Sleep -Seconds 1
Start-Process "http://localhost:8050/restore.html"
Start-Sleep -Seconds 1
Start-Process "http://localhost:8050/jobs.html"

Write-Host "Відкрито 3 вкладки. Натисніть Ctrl+Shift+R для повного оновлення!"
