@echo off
setlocal EnableDelayedExpansion

echo ========================================
echo   NovaBackup - Build Installer
echo ========================================
echo.

set "PROJECT_DIR=%~dp0"
set "BUILD_DIR=%PROJECT_DIR%build\installer"
set "CMD_DIR=%PROJECT_DIR%cmd\nova-service"
set "GUI_DIR=%PROJECT_DIR%cmd\nova-wpf"

REM Create build directory
echo [*] Creating build directory...
mkdir "%BUILD_DIR%" 2>nul

REM Build Go service
echo.
echo [*] Building Go service (nova.exe)...
cd /d "%PROJECT_DIR%"
go build -o "%BUILD_DIR%\nova.exe" ./cmd/nova-service/
if %errorLevel% neq 0 (
    echo [ERROR] Failed to build Go service!
    pause
    exit /b 1
)

REM Build WPF GUI
echo.
echo [*] Building WPF GUI (NovaBackup.exe)...
cd /d "%GUI_DIR%"
dotnet build -c Release /p:Platform=x64
if %errorLevel% neq 0 (
    echo [ERROR] Failed to build WPF GUI!
    pause
    exit /b 1
)

REM Copy GUI files
echo.
echo [*] Copying GUI files...
xcopy /E /Y /I /Q "%GUI_DIR%\bin\Release\net8.0-windows\*" "%BUILD_DIR%\"

REM Copy Setup.bat
echo.
echo [*] Copying Setup.bat...
copy /Y "%PROJECT_DIR%Setup.bat" "%BUILD_DIR%\"

echo.
echo ========================================
echo   Build Complete!
echo ========================================
echo.
echo Installer location: %BUILD_DIR%
echo.
echo To install: Run Setup.bat as Administrator
echo.

cd /d "%PROJECT_DIR%"
pause
