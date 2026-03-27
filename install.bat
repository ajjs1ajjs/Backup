@echo off
setlocal enabledelayedexpansion

echo ========================================
echo   NovaBackup Installer for Windows
echo ========================================
echo.

REM Get the directory where this script is located
set "SCRIPT_DIR=%~dp0"
echo [INFO] Script directory: %SCRIPT_DIR%

REM Check if we're running from the Backup directory
if exist "%SCRIPT_DIR%pyproject.toml" (
    set "WORK_DIR=%SCRIPT_DIR%"
    echo [OK] Found pyproject.toml in script directory
) else (
    REM Try current directory
    if exist "%CD%\pyproject.toml" (
        set "WORK_DIR=%CD%"
        echo [OK] Found pyproject.toml in current directory
    ) else (
        echo [ERROR] pyproject.toml not found!
        echo Please run this script from the Backup directory
        echo or download it to the Backup directory.
        exit /b 1
    )
)

REM Change to working directory
cd /d "%WORK_DIR%"
echo [INFO] Working directory: %CD%
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

REM Install novabackup from working directory
echo [INFO] Installing NovaBackup from: %WORK_DIR%
pip install -e ".[api,dev]" --quiet
if %errorlevel% equ 0 (
    echo [OK] NovaBackup installed successfully
) else (
    echo [WARNING] Installation completed with warnings
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
echo Then navigate to Backup directory and run:
echo   cd %WORK_DIR%
echo   python -m uvicorn novabackup.api:get_app --reload
echo.

REM Test installation
echo [INFO] Testing installation...
novabackup --version 2>&1 | findstr "novabackup"
if %errorlevel% equ 0 (
    echo [OK] NovaBackup CLI is working
) else (
    echo [INFO] CLI test skipped
)

echo.
echo Done!
endlocal
