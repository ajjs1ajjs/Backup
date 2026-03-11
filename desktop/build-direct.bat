@echo off
echo ================================================
echo    NOVA BACKUP - DIRECT BUILD
echo ================================================
echo.

cd /d "%~dp0final"

echo Building NOVA Backup directly...

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

echo Creating simple project file...
echo ^<Project Sdk="Microsoft.NET.Sdk"^> > NovaBackup.Direct.csproj
echo   ^<PropertyGroup^> >> NovaBackup.Direct.csproj
echo     ^<OutputType^>WinExe^</OutputType^> >> NovaBackup.Direct.csproj
echo     ^<TargetFramework^>net6.0-windows^</TargetFramework^> >> NovaBackup.Direct.csproj
echo     ^<UseWindowsForms^>true^</UseWindowsForms^> >> NovaBackup.Direct.csproj
echo     ^<PublishSingleFile^>true^</PublishSingleFile^> >> NovaBackup.Direct.csproj
echo     ^<SelfContained^>true^</SelfContained^> >> NovaBackup.Direct.csproj
echo     ^<RuntimeIdentifier^>win-x64^</RuntimeIdentifier^> >> NovaBackup.Direct.csproj
echo     ^<PublishReadyToRun^>true^</PublishReadyToRun^> >> NovaBackup.Direct.csproj
echo   ^</PropertyGroup^> >> NovaBackup.Direct.csproj
echo ^</Project^> >> NovaBackup.Direct.csproj

echo Building...
dotnet build NovaBackup.Direct.csproj --configuration Release --verbosity minimal

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
if exist "bin\Release\net6.0-windows\win-x64\NovaBackup.Direct.exe" (
    copy "bin\Release\net6.0-windows\win-x64\NovaBackup.Direct.exe" "..\installer\NovaBackup.exe" >nul
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
    echo Expected location: bin\Release\net6.0-windows\win-x64\NovaBackup.Direct.exe
    pause
    exit /b 1
)

echo.
pause
