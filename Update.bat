@echo off
setlocal EnableDelayedExpansion

echo ========================================
echo   NovaBackup - Автоматичне Оновлення
echo ========================================
echo.

REM Check admin
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ПОМИЛКА] Потрібні права адміністратора!
    echo.
    echo Натисніть праву кнопку -^> Запуск від імені адміністратора
    pause
    exit /b 1
)

echo [OK] Права адміністратора підтверджені
echo.

REM Stop service
echo [*] Зупинка служби NovaBackup...
net stop NovaBackup >nul 2>&1
timeout /t 2 /nobreak >nul

REM Kill process
echo [*] Закриття програми...
taskkill /F /IM NovaBackup.exe >nul 2>&1
timeout /t 2 /nobreak >nul

REM Backup old files
echo [*] Створення резервної копії...
set "BACKUP_DIR=C:\Program Files\NovaBackup\backup_old"
mkdir "%BACKUP_DIR%" 2>nul
xcopy /Y /Q "C:\Program Files\NovaBackup\NovaBackup.*" "%BACKUP_DIR%\" 2>nul

REM Copy new files
echo [*] Копіювання нових файлів...
set "SOURCE_DIR=D:\WORK_CODE\Backup\build\msi"
xcopy /Y /R /Q "%SOURCE_DIR%\NovaBackup.exe" "C:\Program Files\NovaBackup\"
xcopy /Y /R /Q "%SOURCE_DIR%\NovaBackup.dll" "C:\Program Files\NovaBackup\"
xcopy /Y /R /Q "%SOURCE_DIR%\NovaBackup.pdb" "C:\Program Files\NovaBackup\"
xcopy /Y /R /Q "%SOURCE_DIR%\NovaBackup.deps.json" "C:\Program Files\NovaBackup\"
xcopy /Y /R /Q "%SOURCE_DIR%\NovaBackup.runtimeconfig.json" "C:\Program Files\NovaBackup\"

if %errorLevel% neq 0 (
    echo [ПОМИЛКА] Не вдалося скопіювати файли!
    echo Відновлення старої версії...
    xcopy /Y /Q "%BACKUP_DIR%\*" "C:\Program Files\NovaBackup\" 2>nul
    pause
    exit /b 1
)

REM Start service
echo [*] Запуск служби...
net start NovaBackup >nul 2>&1

echo.
echo ========================================
echo   ОНОВЛЕННЯ ЗАВЕРШЕНО УСПІШНО!
echo ========================================
echo.
echo Нова версія: v6.0 з українською мовою
echo Темна 3D тема активована
echo.
echo Запуск програми...
timeout /t 2 /nobreak >nul
start "" "C:\Program Files\NovaBackup\NovaBackup.exe"

echo.
pause
