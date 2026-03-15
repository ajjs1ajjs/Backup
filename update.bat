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
set "RAW_URL=https://raw.githubusercontent.com/ajjs1ajjs/Backup/main"

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
set "DOWNLOAD_OK=0"
call :download "%GITHUB_URL%/novabackup-windows-amd64.exe" "%TEMP_DIR%\novabackup.exe"
if %errorLevel% equ 0 set "DOWNLOAD_OK=1"
if "%DOWNLOAD_OK%"=="0" (
    echo [WARNING] Release download failed. Trying raw repository...
    call :download "%RAW_URL%/novabackup.exe" "%TEMP_DIR%\novabackup.exe"
    if %errorLevel% equ 0 set "DOWNLOAD_OK=1"
)
if "%DOWNLOAD_OK%"=="0" (
    echo [ERROR] Download failed!
    echo.
    echo Please download manually from:
    echo https://github.com/ajjs1ajjs/Backup/releases
    pause
    exit /b 1
)
call :verify_exe "%TEMP_DIR%\novabackup.exe"
if %errorLevel% neq 0 (
    echo [ERROR] Downloaded file is invalid or corrupted.
    echo Ensure novabackup.exe exists in the repository main branch or use GitHub Releases.
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
exit /b 0

:download
set "DOWNLOAD_URL=%~1"
set "DOWNLOAD_OUT=%~2"
powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%DOWNLOAD_OUT%' -UseBasicParsing"
call :verify_exe "%DOWNLOAD_OUT%"
if %errorLevel% equ 0 exit /b 0
where curl.exe >nul 2>&1
if %errorLevel% equ 0 (
    curl.exe -f -L --retry 3 --retry-delay 2 -o "%DOWNLOAD_OUT%" "%DOWNLOAD_URL%"
    call :verify_exe "%DOWNLOAD_OUT%"
    if %errorLevel% equ 0 exit /b 0
)
exit /b 1

:verify_exe
set "VERIFY_FILE=%~1"
if not exist "%VERIFY_FILE%" exit /b 1
for %%A in ("%VERIFY_FILE%") do set "VERIFY_SIZE=%%~zA"
if "%VERIFY_SIZE%"=="" exit /b 1
if %VERIFY_SIZE% lss 1000000 exit /b 1
powershell -Command "$b=Get-Content -Encoding Byte -TotalCount 2 -Path '%VERIFY_FILE%'; if ($b.Length -lt 2 -or $b[0] -ne 77 -or $b[1] -ne 90) { exit 1 }"
if %errorLevel% neq 0 exit /b 1
exit /b 0
