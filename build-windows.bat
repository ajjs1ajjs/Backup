@echo off
setlocal enabledelayedexpansion

echo ==========================================
echo NovaBackup Enterprise - Windows Build
echo ==========================================
echo.

:: Set variables
set VERSION=6.0.0
set BUILD_TIME=%date% %time%
set GIT_COMMIT=unknown

:: Try to get git commit
for /f "tokens=*" %%a in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%a

:: Create output directory
if not exist "dist" mkdir dist

echo [1/3] Building nova.exe (Server)...
go build -ldflags "-s -w -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT%" -o dist\nova.exe ./cmd/server
if errorlevel 1 (
    echo [ERROR] Failed to build nova.exe
    exit /b 1
)
echo [OK] dist\nova.exe built successfully

echo.
echo [2/3] Building nova-cli.exe (CLI)...
go build -ldflags "-s -w -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT%" -o dist\nova-cli.exe ./cmd/nova-cli
if errorlevel 1 (
    echo [ERROR] Failed to build nova-cli.exe
    exit /b 1
)
echo [OK] dist\nova-cli.exe built successfully

echo.
echo [3/3] Building nova-service.exe (Windows Service)...
go build -ldflags "-s -w -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT%" -o dist\nova-service.exe ./cmd/server
if errorlevel 1 (
    echo [ERROR] Failed to build nova-service.exe
    exit /b 1
)
echo [OK] dist\nova-service.exe built successfully

echo.
echo ==========================================
echo Build Complete!
echo ==========================================
echo.
echo Output files:
dir dist\*.exe /b
echo.
echo Version: %VERSION%
echo Git Commit: %GIT_COMMIT%
echo Build Time: %BUILD_TIME%
echo.
pause
