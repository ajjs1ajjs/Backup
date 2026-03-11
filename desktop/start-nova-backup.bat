@echo off
echo ================================================
echo    NOVA BACKUP - ALL-IN-ONE LAUNCHER
echo ================================================
echo.

REM Check if NovaBackup.exe exists
if exist "NovaBackup.exe" (
    echo Starting NOVA Backup...
    start "" NovaBackup.exe
    echo.
    echo NOVA Backup started successfully!
    echo Web Console: http://localhost:8080
    echo Remote Access: http://[YOUR_IP]:8080
    echo Default Credentials: admin / admin
    echo.
    echo Press any key to exit...
    pause >nul
    exit /b 0
)

echo ERROR: NovaBackup.exe not found!
echo.
echo Please make sure NovaBackup.exe is in the same directory
echo as this launcher script.
echo.
echo You can build NovaBackup.exe using:
echo   build-simple.bat
echo.
pause
