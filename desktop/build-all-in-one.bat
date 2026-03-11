@echo off
echo Building NOVA Backup All-in-One executable...
echo.

REM Change to the correct directory
cd /d "%~dp0app"

REM Check if .NET 6.0 is installed
dotnet --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: .NET 6.0 SDK is not installed
    echo Please install .NET 6.0 SDK from: https://dotnet.microsoft.com/download/dotnet/6.0
    pause
    exit /b 1
)

echo Cleaning previous builds...
dotnet clean NovaBackup.AllInOne.csproj --configuration Release --verbosity quiet

echo Restoring packages...
dotnet restore NovaBackup.AllInOne.csproj --verbosity quiet

echo Publishing All-in-One executable...
dotnet publish NovaBackup.AllInOne.csproj ^
    --configuration Release ^
    --runtime win-x64 ^
    --self-contained true ^
    --output "..\publish" ^
    --verbosity normal

if %errorlevel% neq 0 (
    echo ERROR: Build failed
    pause
    exit /b 1
)

echo.
echo Build completed successfully!
echo.
echo Output location: ..\publish\NovaBackup.Desktop.exe
echo File size:
dir "..\publish\NovaBackup.Desktop.exe" | findstr NovaBackup.Desktop.exe
echo.

REM Create installer directory
if not exist "..\installer" mkdir "..\installer"

REM Copy the single executable to installer directory
copy "..\publish\NovaBackup.Desktop.exe" "..\installer\NovaBackup.exe" >nul

echo All-in-One executable created: ..\installer\NovaBackup.exe
echo.
echo This single file contains:
echo - Complete desktop application
echo - Web console with remote access
echo - Windows Service functionality
echo - All dependencies and libraries
echo - Embedded web UI files
echo.
echo You can now run NovaBackup.exe on any Windows 10/11 machine
echo without any additional installation!
echo.

pause
