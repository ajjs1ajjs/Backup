@echo off
title NovaBackup - Quick Test
color 0A

echo.
echo ╔══════════════════════════════════════════════════════════════════╗
echo ║                    NovaBackup Quick Test                         ║
echo ╚══════════════════════════════════════════════════════════════════╝
echo.
echo [TEST 1] Перевірка nova.exe...
nova.exe --help >nul 2>&1
if errorlevel 1 (
    echo [FAIL] nova.exe не працює
) else (
    echo [OK] nova.exe працює
)

echo.
echo [TEST 2] Перевірка nova-cli.exe...
nova-cli.exe --help >nul 2>&1
if errorlevel 1 (
    echo [FAIL] nova-cli.exe не працює
) else (
    echo [OK] nova-cli.exe працює
)

echo.
echo [TEST 3] Перевірка команд vmware...
nova-cli.exe vmware --help >nul 2>&1
if errorlevel 1 (
    echo [FAIL] vmware command не працює
) else (
    echo [OK] vmware command працює
)

echo.
echo ═══════════════════════════════════════════════════════════════════
echo Якщо всі тести [OK] - можна встановлювати!
echo Запустіть install.bat від імені адміністратора
echo.
pause
