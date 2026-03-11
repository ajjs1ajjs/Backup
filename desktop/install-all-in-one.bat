@echo off
echo NOVA Backup All-in-One Installer
echo ================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: This installer requires administrator privileges
    echo Please right-click and "Run as administrator"
    pause
    exit /b 1
)

REM Create installation directory
set INSTALL_DIR=C:\Program Files\NovaBackup
echo Installing to: %INSTALL_DIR%

if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
)

REM Copy the executable
echo Copying NOVA Backup executable...
copy "NovaBackup.exe" "%INSTALL_DIR%\NovaBackup.exe" >nul

if %errorlevel% neq 0 (
    echo ERROR: Failed to copy executable
    pause
    exit /b 1
)

REM Create shortcuts
echo Creating shortcuts...

REM Desktop shortcut
powershell "$WshShell = New-Object -comObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%PUBLIC%\Desktop\NovaBackup.lnk'); $Shortcut.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $Shortcut.IconLocation = '%INSTALL_DIR%\NovaBackup.exe, 0'; $Shortcut.Save()"

REM Start Menu shortcut
if not exist "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup" (
    mkdir "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup"
)

powershell "$WshShell = New-Object -comObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\NOVA Backup.lnk'); $Shortcut.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $Shortcut.IconLocation = '%INSTALL_DIR%\NovaBackup.exe, 0'; $Shortcut.Save()"

REM Web Console shortcut
powershell "$WshShell = New-Object -comObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\Web Console.lnk'); $Shortcut.TargetPath = 'http://localhost:8080'; $Shortcut.Save()"

REM Uninstall shortcut
powershell "$WshShell = New-Object -comObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\Uninstall.lnk'); $Shortcut.TargetPath = '%INSTALL_DIR%\uninstall.bat'; $Shortcut.Save()"

REM Create uninstall script
echo Creating uninstall script...
echo @echo off > "%INSTALL_DIR%\uninstall.bat"
echo echo Uninstalling NOVA Backup... >> "%INSTALL_DIR%\uninstall.bat"
echo. >> "%INSTALL_DIR%\uninstall.bat"
echo REM Stop running processes >> "%INSTALL_DIR%\uninstall.bat"
echo taskkill /f /im NovaBackup.exe 2^>nul >> "%INSTALL_DIR%\uninstall.bat"
echo. >> "%INSTALL_DIR%\uninstall.bat"
echo REM Remove shortcuts >> "%INSTALL_DIR%\uninstall.bat"
echo del /f /q "%PUBLIC%\Desktop\NovaBackup.lnk" 2^>nul >> "%INSTALL_DIR%\uninstall.bat"
echo rmdir /s /q "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup" 2^>nul >> "%INSTALL_DIR%\uninstall.bat"
echo. >> "%INSTALL_DIR%\uninstall.bat"
echo REM Remove program files >> "%INSTALL_DIR%\uninstall.bat"
echo cd /d ^"%%ProgramFiles^%" >> "%INSTALL_DIR%\uninstall.bat"
echo rmdir /s /q NovaBackup >> "%INSTALL_DIR%\uninstall.bat"
echo. >> "%INSTALL_DIR%\uninstall.bat"
echo echo NOVA Backup uninstalled successfully! >> "%INSTALL_DIR%\uninstall.bat"
echo pause >> "%INSTALL_DIR%\uninstall.bat"

REM Add to registry for Windows uninstall
echo Adding to registry...
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\NovaBackup" /v "DisplayName" /t REG_SZ /d "NOVA Backup" /f >nul
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\NovaBackup" /v "InstallLocation" /t REG_SZ /d "%INSTALL_DIR%" /f >nul
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\NovaBackup" /v "UninstallString" /t REG_SZ /d "\"%INSTALL_DIR%\uninstall.bat\"" /f >nul
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\NovaBackup" /v "DisplayIcon" /t REG_SZ /d "\"%INSTALL_DIR%\NovaBackup.exe\"" /f >nul
reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\NovaBackup" /v "Publisher" /t REG_SZ /d "NOVA Backup" /f >nul

REM Configure firewall for web console
echo Configuring firewall...
netsh advfirewall firewall delete rule name="NOVA Backup Web Console" >nul 2>&1
netsh advfirewall firewall add rule name="NOVA Backup Web Console" dir=in action=allow protocol=TCP localport=8080 >nul

REM Create data directories
echo Creating data directories...
if not exist "%PROGRAMDATA%\NovaBackup" mkdir "%PROGRAMDATA%\NovaBackup"
if not exist "%PROGRAMDATA%\NovaBackup\logs" mkdir "%PROGRAMDATA%\NovaBackup\logs"
if not exist "%PROGRAMDATA%\NovaBackup\config" mkdir "%PROGRAMDATA%\NovaBackup\config"
if not exist "%PROGRAMDATA%\NovaBackup\backups" mkdir "%PROGRAMDATA%\NovaBackup\backups"

REM Create default configuration
echo Creating default configuration...
echo { > "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo   "Settings": { >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo     "DefaultBackupPath": "%PROGRAMDATA%\NovaBackup\backups", >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo     "MaxConcurrentBackups": 3, >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo     "CompressionLevel": "normal", >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo     "EnableEncryption": true, >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo     "EnableNotifications": true, >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo     "WebConsoleEnabled": true, >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo     "WebConsolePort": 8080 >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo   }, >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo   "BackupJobs": [], >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo   "Schedules": [], >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo   "UpdatedAt": "%DATE% %TIME%" >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"
echo } >> "%PROGRAMDATA%\NovaBackup\config\backup-config.json"

echo.
echo Installation completed successfully!
echo.
echo NOVA Backup has been installed to: %INSTALL_DIR%
echo.
echo Shortcuts created:
echo - Desktop: NOVA Backup
echo - Start Menu: NOVA Backup, Web Console, Uninstall
echo.
echo Web Console: http://localhost:8080
echo Default credentials for remote access: admin / admin
echo.
echo Would you like to launch NOVA Backup now? (Y/N)
set /p launch=
if /i "%launch%"=="Y" (
    start "" "%INSTALL_DIR%\NovaBackup.exe"
)

echo.
echo Installation complete! Enjoy using NOVA Backup!
pause
