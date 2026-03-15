@echo off
echo ========================================
echo   NovaBackup - Оновлення Служби
echo ========================================
echo.

echo [*] Зупинка служби...
net stop NovaBackup 2>nul
timeout /t 2 /nobreak >nul

echo [*] Видалення старої служби...
sc delete NovaBackup 2>nul
timeout /t 2 /nobreak >nul

echo [*] Копіювання нових файлів...
xcopy /Y /Q "D:\WORK_CODE\Backup\build\msi\nova.exe" "C:\Program Files\NovaBackup\"
xcopy /Y /Q "D:\WORK_CODE\Backup\build\msi\NovaBackup.*" "C:\Program Files\NovaBackup\"

echo [*] Встановлення служби...
"C:\Program Files\NovaBackup\nova.exe" install

echo [*] Запуск служби...
net start NovaBackup

echo.
echo ========================================
echo   ГОТОВО!
echo ========================================
echo.
echo Перевірка статусу...
sc query NovaBackup | findstr "STATE"

echo.
echo Перевірка API...
curl -s http://localhost:8080/api/health

echo.
pause
