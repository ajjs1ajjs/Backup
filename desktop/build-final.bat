@echo off
echo ================================================
echo    NOVA BACKUP - FINAL BUILD
echo ================================================
echo.

cd /d "%~dp0final"

echo Building NOVA Backup Final executable...

REM Check if .NET 6.0 is installed
dotnet --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: .NET 6.0 SDK is not installed
    echo Please install .NET 6.0 SDK from: https://dotnet.microsoft.com/download/dotnet/6.0
    pause
    exit /b 1
)

echo Cleaning previous builds...
if exist "bin" rmdir /s /q bin

echo Building...
dotnet build NovaBackup.Simple.csproj --configuration Release --verbosity minimal

if %errorlevel% neq 0 (
    echo ERROR: Build failed
    pause
    exit /b 1
)

echo.
echo Build completed successfully!
echo.

REM Copy to installer directory
if not exist "..\installer" mkdir "..\installer"

REM Copy executable
if exist "bin\Release\net6.0-windows\win-x64\NovaBackup.Simple.exe" (
    copy "bin\Release\net6.0-windows\win-x64\NovaBackup.Simple.exe" "..\installer\NovaBackup.exe" >nul
    echo SUCCESS: NovaBackup.exe created!
    echo Location: ..\installer\NovaBackup.exe
    echo.
    echo This single file contains:
    echo - Complete desktop application with full GUI
    echo - Web console with remote access
    echo - Windows Service functionality
    echo - All dependencies and libraries
    echo - Embedded web UI files
    echo.
    echo USAGE:
    echo   1. Run NovaBackup.exe directly for GUI
    echo   2. Run start-nova-backup.bat for launcher
    echo.
    echo Web Console: http://localhost:8080
    echo Remote Access: http://[IP]:8080
    echo Default Credentials: admin / admin
    echo.
    echo You can now run NovaBackup.exe on any Windows 10/11 machine
    echo without any additional installation!
) else (
    echo ERROR: Executable not found
    echo Expected location: bin\Release\net6.0-windows\win-x64\NovaBackup.Simple.exe
    pause
    exit /b 1
)

echo.
pause
