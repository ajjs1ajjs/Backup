@echo off
setlocal EnableDelayedExpansion

echo ========================================
echo   NovaBackup v7.0 - Windows Setup
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

set "RAW_BASE=https://raw.githubusercontent.com/ajjs1ajjs/Backup/main"
set "TEMP_DIR=%TEMP%\NovaBackup_Setup"
set "CACHE_BUST=%RANDOM%%RANDOM%"

if exist "%TEMP_DIR%" rmdir /s /q "%TEMP_DIR%"
mkdir "%TEMP_DIR%"

:menu
echo.
echo Select action:
echo   [1] Install / Repair
echo   [2] Update
echo   [3] Remove Service
echo   [4] Exit
echo.
set /p "ACTION=Enter choice (1-4): "

if "%ACTION%"=="1" (
    call :run_install
    goto menu
)
if "%ACTION%"=="2" (
    call :run_update
    goto menu
)
if "%ACTION%"=="3" (
    call :run_remove
    goto menu
)
if "%ACTION%"=="4" (
    goto cleanup
)

echo Invalid choice. Please try again.
goto menu

:run_install
echo.
echo [*] Downloading install script...
call :download "%RAW_BASE%/install.bat?%CACHE_BUST%" "%TEMP_DIR%\install.bat"
if %errorLevel% neq 0 (
    echo [ERROR] Failed to download install.bat
    exit /b 1
)
call "%TEMP_DIR%\install.bat"
exit /b 0

:run_update
echo.
echo [*] Downloading update script...
call :download "%RAW_BASE%/update.bat?%CACHE_BUST%" "%TEMP_DIR%\update.bat"
if %errorLevel% neq 0 (
    echo [ERROR] Failed to download update.bat
    exit /b 1
)
call "%TEMP_DIR%\update.bat"
exit /b 0

:run_remove
echo.
echo [*] Removing NovaBackup service...
set "INSTALL_DIR=C:\Program Files\NovaBackup"
if exist "%INSTALL_DIR%\NovaBackup.exe" (
    "%INSTALL_DIR%\NovaBackup.exe" remove
) else (
    sc stop NovaBackup >nul 2>&1
    sc delete NovaBackup >nul 2>&1
)
echo [OK] Remove command sent
exit /b 0

:download
set "DOWNLOAD_URL=%~1"
set "DOWNLOAD_OUT=%~2"
for %%D in ("%DOWNLOAD_OUT%") do if not exist "%%~dpD" mkdir "%%~dpD"
powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%DOWNLOAD_OUT%' -UseBasicParsing"
call :check_file "%DOWNLOAD_OUT%" 100
if %errorLevel% equ 0 exit /b 0
where curl.exe >nul 2>&1
if %errorLevel% equ 0 (
    curl.exe -f -L --retry 3 --retry-delay 2 -o "%DOWNLOAD_OUT%" "%DOWNLOAD_URL%"
    call :check_file "%DOWNLOAD_OUT%" 100
    if %errorLevel% equ 0 exit /b 0
)
exit /b 1

:check_file
set "CHECK_FILE=%~1"
set "MIN_SIZE=%~2"
if not exist "%CHECK_FILE%" exit /b 1
for %%A in ("%CHECK_FILE%") do set "CHECK_SIZE=%%~zA"
if "%CHECK_SIZE%"=="" exit /b 1
if %CHECK_SIZE% lss %MIN_SIZE% exit /b 1
exit /b 0

:cleanup
if exist "%TEMP_DIR%" rmdir /s /q "%TEMP_DIR%"
exit /b 0
