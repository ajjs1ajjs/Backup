@echo off
setlocal enabledelayedexpansion

echo ========================================
echo   NovaBackup Installer for Windows
echo ========================================
echo.

REM Check Python
where python >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Python is not found on PATH.
    echo Please install Python 3.9+ from https://python.org
    exit /b 1
)

echo [OK] Python found: 
python --version
echo.

REM Set installation directory
set "INSTALL_DIR=%USERPROFILE%\.novabackup"
set "VENV=%INSTALL_DIR%\venv"

REM Create virtual environment if not exists
if not exist "%VENV%" (
    echo [INFO] Creating virtual environment...
    mkdir "%INSTALL_DIR%" >nul 2>&1
    python -m venv "%VENV%"
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to create virtual environment
        exit /b 1
    )
    echo [OK] Virtual environment created
) else (
    echo [OK] Virtual environment exists
)

REM Activate virtual environment
echo [INFO] Activating virtual environment...
call "%VENV%\Scripts\activate.bat"

REM Upgrade pip
echo [INFO] Upgrading pip...
python -m pip install --upgrade pip --quiet

REM Install novabackup from current directory
echo [INFO] Installing NovaBackup...
if exist "pyproject.toml" (
    pip install -e ".[api,dev]" --quiet
    if %errorlevel% equ 0 (
        echo [OK] NovaBackup installed successfully
    ) else (
        echo [WARNING] Installation completed with warnings
    )
) else (
    echo [ERROR] pyproject.toml not found in current directory
    echo Please run this script from the Backup directory
    exit /b 1
)

echo.
echo ========================================
echo   Installation Complete!
echo ========================================
echo.
echo To activate NovaBackup, run:
echo   ^& "$env:USERPROFILE\.novabackup\venv\Scripts\Activate.ps1"
echo.
echo Or from Command Prompt:
echo   call "%VENV%\Scripts\activate.bat"
echo.
echo Then run:
echo   novabackup --help
echo.

REM Test installation
echo [INFO] Testing installation...
novabackup list-vms 2>&1 | findstr "id"
if %errorlevel% equ 0 (
    echo [OK] NovaBackup is working
) else (
    echo [INFO] VM list may require Hyper-V enabled
)

echo.
echo Done!
endlocal
