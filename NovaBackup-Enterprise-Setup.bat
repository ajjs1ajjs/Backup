@echo off
title NovaBackup v6.0 Enterprise - Self-Extracting Installer
color 0B

echo ========================================
echo    NovaBackup v6.0 Enterprise
echo    Self-Extracting Installer
echo ========================================
echo.

echo This is NovaBackup v6.0 Enterprise Installer
echo with Veeam-style interface and full functionality.
echo.

echo Creating installation directories...
if not exist "%ProgramFiles%\NovaBackup" mkdir "%ProgramFiles%\NovaBackup"
if not exist "%ProgramFiles%\NovaBackup\bin" mkdir "%ProgramFiles%\NovaBackup\bin"
if not exist "%ProgramFiles%\NovaBackup\gui" mkdir "%ProgramFiles%\NovaBackup\gui"
if not exist "%ProgramFiles%\NovaBackup\agent" mkdir "%ProgramFiles%\NovaBackup\agent"
if not exist "%ProgramFiles%\NovaBackup\config" mkdir "%ProgramFiles%\NovaBackup\config"

echo Copying NovaBackup files...
copy "nova.exe" "%ProgramFiles%\NovaBackup\bin\" >nul 2>&1
copy "NovaBackup-Enterprise.bat" "%ProgramFiles%\NovaBackup\bin\NovaBackup.exe" >nul 2>&1
copy "gui\app.py" "%ProgramFiles%\NovaBackup\gui\" >nul 2>&1
copy "gui\templates\veeam-style.html" "%ProgramFiles%\NovaBackup\gui\templates\" >nul 2>&1
copy "requirements.txt" "%ProgramFiles%\NovaBackup\gui\" >nul 2>&1
copy "NovaBackupAgent.ps1" "%ProgramFiles%\NovaBackup\agent\" >nul 2>&1
copy "nova-gui-manager.ps1" "%ProgramFiles%\NovaBackup\agent\" >nul 2>&1

echo Creating Start Menu shortcuts...
if not exist "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup" mkdir "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup"

echo [InternetShortcut] > "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\NovaBackup Enterprise.url"
echo URL=file:///%ProgramFiles%/NovaBackup/bin/NovaBackup.exe >> "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\NovaBackup Enterprise.url"
echo IconFile=%ProgramFiles%/NovaBackup/nova.exe >> "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\NovaBackup Enterprise.url"
echo IconIndex=0 >> "%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup\NovaBackup Enterprise.url"

echo Creating Desktop shortcut...
echo Set oWS = WScript.CreateObject("WScript.Shell") > "%TEMP%\CreateShortcut.vbs"
echo sLinkFile = oWS.ExpandEnvironmentStrings("%UserProfile%\Desktop\NovaBackup Enterprise.lnk") >> "%TEMP%\CreateShortcut.vbs"
echo Set oLink = oWS.CreateShortcut(sLinkFile) >> "%TEMP%\CreateShortcut.vbs"
echo oLink.TargetPath = "%ProgramFiles%\NovaBackup\bin\NovaBackup.exe" >> "%TEMP%\CreateShortcut.vbs"
echo oLink.WorkingDirectory = "%ProgramFiles%\NovaBackup\bin" >> "%TEMP%\CreateShortcut.vbs"
echo oLink.Description = "NovaBackup v6.0 Enterprise - Backup & Recovery" >> "%TEMP%\CreateShortcut.vbs"
echo oLink.Save >> "%TEMP%\CreateShortcut.vbs"
echo Set oWS = Nothing >> "%TEMP%\CreateShortcut.vbs"
echo Set oLink = Nothing >> "%TEMP%\CreateShortcut.vbs"
cscript //nologo "%TEMP%\CreateShortcut.vbs"
del "%TEMP%\CreateShortcut.vbs"

echo Installing NovaBackup service...
"%ProgramFiles%\NovaBackup\bin\nova.exe" service install >nul 2>&1
"%ProgramFiles%\NovaBackup\bin\nova.exe" service start >nul 2>&1

echo Creating configuration...
echo [General] > "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo Version=6.0.0 >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo InstallPath=%ProgramFiles%\NovaBackup >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"
echo InstallDate=%date% %time% >> "%ProgramFiles%\NovaBackup\config\novabackup.ini"

echo.
echo ========================================
echo    INSTALLATION COMPLETE!
echo ========================================
echo.
echo NovaBackup v6.0 Enterprise has been successfully installed!
echo.
echo Features installed:
echo - Veeam-style GUI interface
echo - Web-based management console  
echo - Background agent service
echo - System integration
echo - Start Menu shortcuts
echo - Desktop shortcut
echo.
echo You can now:
echo 1. Launch NovaBackup from Start Menu
echo 2. Double-click desktop shortcut
echo 3. Run: "%ProgramFiles%\NovaBackup\bin\NovaBackup.exe"
echo.
echo Thank you for choosing NovaBackup Enterprise!
echo.

timeout /t 10
