@echo off
title NovaBackup v6.0 Enterprise - Simple GUI
color 0B

echo ========================================
echo    NovaBackup v6.0 Enterprise
echo    Simple Veeam-Style Interface
echo ========================================
echo.

echo Starting NovaBackup GUI...
echo.

powershell -ExecutionPolicy Bypass -File "NovaBackup-GUI.ps1"

echo.
echo NovaBackup GUI closed.
pause
