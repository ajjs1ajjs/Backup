@echo off 
chcp 65001 >nul
title NovaBackup Enterprise - Installation
color 0A

echo ========================================  
echo   NovaBackup v6.0 Installation        
echo ========================================  
echo.  
echo IMPORTANT: Run as Administrator!  
echo Right-click this file -^> Run as Administrator  
echo.  
pause  
echo.  
echo [1/4] Checking files...  
if not exist "nova.exe" (
    echo ERROR: nova.exe not found!
    echo Please run this from the dist folder.
    pause
    exit /b 1
)
if not exist "nova-cli.exe" (
    echo ERROR: nova-cli.exe not found!
    pause
    exit /b 1
)
echo [OK] Files found
echo.  
echo [2/4] Installing to Program Files...  
mkdir "C:\Program Files\NovaBackup" 2>nul
copy /Y nova.exe "C:\Program Files\NovaBackup\"  
copy /Y nova-cli.exe "C:\Program Files\NovaBackup\"
echo [OK] Files copied
echo.  
echo [3/4] Installing Windows Service...  
sc query NovaBackupService >nul 2>&1
if %errorlevel% == 0 (
    echo Stopping existing service...
    net stop NovaBackupService >nul 2>&1
    sc delete NovaBackupService >nul 2>&1
    timeout /t 2 >nul
)
sc create NovaBackupService binPath= ""C:\Program Files\NovaBackup\nova.exe" --service" displayName= "NovaBackup Service" start= auto
echo [OK] Service installed
echo.  
echo [4/4] Starting Service...  
net start NovaBackupService
echo.  
echo ========================================  
echo   Installation Complete!  
echo ========================================  
echo.  
echo Web Console: http://localhost:8080  
echo CLI: "C:\Program Files\NovaBackup\nova-cli.exe" --help  
echo.  
pause
