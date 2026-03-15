@echo off
echo ========================================
echo   NovaBackup v7.0 - Build All Platforms
echo ========================================
echo.

cd /d "%~dp0"

REM Create build directory
if not exist "build" mkdir "build"

REM Windows AMD64
echo [1/4] Building for Windows AMD64...
set GOOS=windows
set GOARCH=amd64
go build -o build\novabackup-windows-amd64.exe ./cmd/novabackup/
if %errorlevel% neq 0 (
    echo ERROR: Windows build failed!
) else (
    echo OK: Windows AMD64 built successfully
)

REM Linux AMD64
echo.
echo [2/4] Building for Linux AMD64...
set GOOS=linux
set GOARCH=amd64
go build -o build\novabackup-linux-amd64 ./cmd/novabackup/
if %errorlevel% neq 0 (
    echo ERROR: Linux build failed!
) else (
    echo OK: Linux AMD64 built successfully
)

REM macOS AMD64
echo.
echo [3/4] Building for macOS AMD64...
set GOOS=darwin
set GOARCH=amd64
go build -o build\novabackup-macos-amd64 ./cmd/novabackup/
if %errorlevel% neq 0 (
    echo ERROR: macOS AMD64 build failed!
) else (
    echo OK: macOS AMD64 built successfully
)

REM macOS ARM64
echo.
echo [4/4] Building for macOS ARM64 (M1/M2)...
set GOOS=darwin
set GOARCH=arm64
go build -o build\novabackup-macos-arm64 ./cmd/novabackup/
if %errorlevel% neq 0 (
    echo ERROR: macOS ARM64 build failed!
) else (
    echo OK: macOS ARM64 built successfully
)

REM Reset GOOS/GOARCH
set GOOS=
set GOARCH=

echo.
echo ========================================
echo   Build Summary
echo ========================================
dir build /b
echo.
echo Done!
pause
