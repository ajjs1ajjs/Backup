@echo off
setlocal EnableDelayedExpansion

echo ========================================
echo   NovaBackup Enterprise v6.0 - Setup
echo ========================================
echo.

REM Check for administrator privileges
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Administrator rights required!
    echo.
    echo Please right-click this file and select:
    echo   "Run as administrator"
    echo.
    pause
    exit /b 1
)

echo [OK] Administrator privileges confirmed
echo.

REM Get the directory where the setup is located
set "SOURCE_DIR=%~dp0"
set "INSTALL_DIR=C:\Program Files\NovaBackup"
set "DATA_DIR=C:\ProgramData\NovaBackup"

echo [*] Stopping existing service (if running)...
net stop NovaBackup >nul 2>&1
sc delete NovaBackup >nul 2>&1
timeout /t 2 /nobreak >nul

echo.
echo [*] Creating directories...
mkdir "%INSTALL_DIR%" 2>nul
mkdir "%DATA_DIR%" 2>nul
mkdir "%DATA_DIR%\Logs" 2>nul
mkdir "%DATA_DIR%\Backups" 2>nul
mkdir "%DATA_DIR%\Config" 2>nul

echo.
echo [*] Copying program files...
xcopy /E /Y /I /Q "%SOURCE_DIR%*" "%INSTALL_DIR%\"

echo.
echo [*] Installing Windows Service...
"%INSTALL_DIR%\nova.exe" install
if %errorLevel% neq 0 (
    echo [ERROR] Failed to install service!
    pause
    exit /b 1
)

echo.
echo [*] Starting Service...
"%INSTALL_DIR%\nova.exe" start

echo.
echo [*] Creating Start Menu shortcuts...
set "START_MENU=%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup Enterprise"
mkdir "%START_MENU%" 2>nul

REM Create shortcut using PowerShell
powershell -Command "$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%START_MENU%\NovaBackup Enterprise.lnk'); $Shortcut.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $Shortcut.WorkingDirectory = '%INSTALL_DIR%'; $Shortcut.Description = 'NovaBackup Enterprise Console'; $Shortcut.Save()"

echo.
echo ========================================
echo   Installation Complete Successfully!
echo ========================================
echo.
echo Installation Directory: %INSTALL_DIR%
echo Data Directory: %DATA_DIR%
echo.
echo Starting NovaBackup Enterprise Console...
start "" "%INSTALL_DIR%\NovaBackup.exe"

echo.
pause
