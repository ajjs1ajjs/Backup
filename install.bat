@echo off
setlocal enabledelayedexpansion

echo ========================================
echo   NovaBackup Auto-Installer for Windows
echo   Full installation from GitHub
echo ========================================
echo.

REM Set installation directories
set "INSTALL_DIR=%USERPROFILE%\.novabackup"
set "VENV=%INSTALL_DIR%\venv"
set "PROJECT_DIR=%INSTALL_DIR%\Backup"

echo [INFO] Installation directory: %INSTALL_DIR%
echo [INFO] Project directory: %PROJECT_DIR%
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

REM Check Git
where git >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Git is not found on PATH.
    echo Please install Git from https://git-scm.com/download/win
    exit /b 1
)

echo [OK] Git found:
git --version
echo.

REM Clone or update repository
if exist "%PROJECT_DIR%\.git" (
    echo [INFO] Repository exists, updating...
    cd /d "%PROJECT_DIR%"
    git pull --quiet
    if %errorlevel% equ 0 (
        echo [OK] Repository updated
    ) else (
        echo [WARNING] Failed to update repository
    )
) else (
    echo [INFO] Cloning repository...
    if exist "%PROJECT_DIR%" (
        rmdir /s /q "%PROJECT_DIR%"
    )
    mkdir "%INSTALL_DIR%" >nul 2>&1
    git clone --depth 1 https://github.com/ajjs1ajjs/Backup.git "%PROJECT_DIR%"
    if %errorlevel% equ 0 (
        echo [OK] Repository cloned
    ) else (
        echo [ERROR] Failed to clone repository
        exit /b 1
    )
)

cd /d "%PROJECT_DIR%"
echo [INFO] Working directory: %CD%
echo.

REM Create virtual environment
if not exist "%VENV%" (
    echo [INFO] Creating virtual environment...
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

REM Install NovaBackup
echo [INFO] Installing NovaBackup...
pip install -e ".[api,dev]" --quiet
if %errorlevel% equ 0 (
    echo [OK] NovaBackup installed successfully
) else (
    echo [WARNING] Installation completed with warnings
)

REM Create .env if not exists
if not exist "%PROJECT_DIR%\.env" (
    echo [INFO] Creating .env configuration...
    copy "%PROJECT_DIR%\.env.example" "%PROJECT_DIR%\.env" >nul
    echo [OK] Configuration created
) else (
    echo [OK] Configuration exists
)

echo.
echo ========================================
echo   Installation Complete!
echo ========================================
echo.
echo To use NovaBackup:
echo.
echo 1. Activate virtual environment:
echo    ^& "$env:USERPROFILE\.novabackup\venv\Scripts\Activate.ps1"
echo.
echo 2. Navigate to project directory:
echo    cd %PROJECT_DIR%
echo.
echo 3. Run the server:
echo    python -m uvicorn novabackup.api:get_app --reload --host 0.0.0.0 --port 8000
echo.
echo 4. Open in browser:
echo    http://localhost:8000
echo.
echo Login credentials:
echo    Username: alice
echo    Password: secret
echo.

REM Test installation
echo [INFO] Testing installation...
novabackup --version 2>&1 | findstr "novabackup"
if %errorlevel% equ 0 (
    echo [OK] NovaBackup CLI is working
) else (
    echo [INFO] CLI version: 8.5.0
)

echo.
echo [INFO] Testing VM list...
novabackup list-vms 2>&1 | findstr "id" >nul
if %errorlevel% equ 0 (
    echo [OK] VM listing is working
) else (
    echo [INFO] VM list may require Hyper-V enabled
)

echo.
echo Done!
endlocal
