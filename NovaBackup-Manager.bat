@echo off
title NovaBackup v6.0 Manager
color 0B

echo ========================================
echo    NovaBackup v6.0 - Manager
echo ========================================
echo.

REM Check if running as Administrator
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: Administrator privileges required!
    echo Please run as Administrator.
    pause
    exit /b 1
)

echo Starting NovaBackup Manager...
echo.

REM Create program files directory
if not exist "C:\Program Files\NovaBackup" (
    echo Creating NovaBackup program directory...
    mkdir "C:\Program Files\NovaBackup"
    mkdir "C:\Program Files\NovaBackup\logs"
    mkdir "C:\Program Files\NovaBackup\commands"
    mkdir "C:\Program Files\NovaBackup\jobs"
)

REM Copy files to program directory
echo Installing NovaBackup components...
copy "NovaBackupAgent.ps1" "C:\Program Files\NovaBackup\" >nul 2>&1
copy "nova-gui-manager.ps1" "C:\Program Files\NovaBackup\" >nul 2>&1
copy "nova.exe" "C:\Program Files\NovaBackup\" >nul 2>&1

REM Create desktop shortcut
echo Creating desktop shortcut...
powershell -Command "$WshShell = New-Object -comObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('C:\Users\%USERNAME%\Desktop\NovaBackup Manager.lnk'); $Shortcut.TargetPath = 'C:\Program Files\NovaBackup\nova-gui-manager.ps1'; $Shortcut.Save()"

REM Create Start Menu shortcut
echo Creating Start Menu shortcut...
powershell -Command "$WshShell = New-Object -comObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('C:\ProgramData\Microsoft\Windows\Start Menu\Programs\NovaBackup Manager.lnk'); $Shortcut.TargetPath = 'C:\Program Files\NovaBackup\nova-gui-manager.ps1'; $Shortcut.Save()"

echo.
echo ========================================
echo NovaBackup v6.0 Installation Complete!
echo ========================================
echo.
echo Components installed:
echo - Background Agent: NovaBackupAgent.ps1
echo - GUI Manager: nova-gui-manager.ps1
echo - CLI Engine: nova.exe
echo.
echo Shortcuts created:
echo - Desktop: NovaBackup Manager
echo - Start Menu: NovaBackup Manager
echo.
echo The NovaBackup Agent will run in the background
echo and perform scheduled backups even when the GUI is closed.
echo.
echo To start the GUI Manager, use the desktop shortcut.
echo.
pause
