@echo off
title NovaBackup v6.0 - Desktop GUI
color 0B

echo ========================================
echo    NovaBackup v6.0 - Desktop GUI
echo ========================================
echo.

REM Check if nova.exe exists
if not exist "nova.exe" (
    echo ERROR: nova.exe not found!
    echo Please run from the NovaBackup installation directory.
    pause
    exit /b 1
)

echo Starting NovaBackup Desktop Interface...
echo.

REM Create a simple desktop GUI using PowerShell
powershell -ExecutionPolicy Bypass -File "nova-desktop.ps1"

pause
