@echo off
echo Killing all nova-backup processes...
taskkill /F /IM nova-backup.exe 2>nul
timeout /t 3 /nobreak >nul
echo Checking if port 8050 is free...
netstat -ano | findstr ":8050" >nul 2>&1
if %errorlevel% equ 0 (
    echo Port 8050 still in use! Finding process...
    for /f "tokens=5" %%a in ('netstat -aon ^| findstr ":8050.*LISTENING"') do (
        echo Killing process ID %%a
        taskkill /F /PID %%a 2>nul
    )
    timeout /t 2 /nobreak >nul
) else (
    echo Port 8050 is free!
)
echo.
echo Starting new server instance...
cd /d "%~dp0"
start "" nova-backup.exe server
echo Done! Server starting...
timeout /t 5 /nobreak >nul
echo Check http://localhost:8050
