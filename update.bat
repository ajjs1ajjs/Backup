@echo off
echo ========================================
echo   NovaBackup v7.0 - Update
echo   Update from GitHub
echo ========================================
echo.

REM Check admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Administrator rights required!
    echo.
    echo Please right-click -^> Run as administrator
    pause
    exit /b 1
)

echo [OK] Administrator rights confirmed
echo.

set "INSTALL_DIR=C:\Program Files\NovaBackup"
set "GITHUB_URL=https://github.com/ajjs1ajjs/Backup/releases/latest/download"

echo [*] Checking current version...
if exist "%INSTALL_DIR%\NovaBackup.exe" (
    echo      Current: %INSTALL_DIR%\NovaBackup.exe
) else (
    echo [ERROR] NovaBackup not installed!
    echo.
    Please run install.bat first
    pause
    exit /b 1
)

echo.
echo [*] Downloading latest release from GitHub...
echo      URL: %GITHUB_URL%
echo.

REM Create temporary directory
set "TEMP_DIR=%TEMP%\NovaBackup_Update"
if exist "%TEMP_DIR%" rmdir /s /q "%TEMP_DIR%"
mkdir "%TEMP_DIR%"

echo [*] Stopping NovaBackup Service...
net stop NovaBackup >nul 2>&1
taskkill /F /IM NovaBackup.exe >nul 2>&1
timeout /t 2 /nobreak >nul

echo [*] Downloading novabackup-windows-amd64.exe...
powershell -Command "Invoke-WebRequest -Uri '%GITHUB_URL%/novabackup-windows-amd64.exe' -OutFile '%TEMP_DIR%\novabackup.exe'" -UseBasicParsing
if %errorLevel% neq 0 (
    echo [ERROR] Download failed!
    echo.
    echo Please download manually from:
    echo https://github.com/ajjs1ajjs/Backup/releases
    pause
    exit /b 1
)

echo [*] Backing up current version...
if exist "%INSTALL_DIR%\NovaBackup.exe" (
    copy /Y "%INSTALL_DIR%\NovaBackup.exe" "%TEMP_DIR%\NovaBackup.old"
)

echo [*] Installing new version...
copy /Y "%TEMP_DIR%\novabackup.exe" "%INSTALL_DIR%\NovaBackup.exe"

echo [*] Starting Service...
cd /d "%INSTALL_DIR%"
NovaBackup.exe start

REM Cleanup
echo [*] Cleaning up...
timeout /t 2 /nobreak >nul
rmdir /s /q "%TEMP_DIR%"

echo.
echo ========================================
echo   Update Complete Successfully!
echo ========================================
echo.
echo Web UI: http://localhost:8050
echo Login: admin
echo Password: admin123
echo.
echo Opening Web UI...
timeout /t 2 /nobreak >nul
start "" http://localhost:8050

echo.
pause
