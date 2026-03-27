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

REM Install ALL dependencies
echo [INFO] Installing ALL dependencies...
pip install aiohttp fastapi uvicorn prometheus_client --quiet

REM Install NovaBackup
echo [INFO] Installing NovaBackup...
cd /d "%PROJECT_DIR%"
if exist "%PROJECT_DIR%\pyproject.toml" (
    pip install -e ".[api,dev]" --quiet
    if %errorlevel% equ 0 (
        echo [OK] NovaBackup installed successfully
    ) else (
        echo [WARNING] Installation completed with warnings
    )
) else (
    echo [ERROR] pyproject.toml not found in %PROJECT_DIR%
    exit /b 1
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

REM Generate secrets if script exists
if exist "%PROJECT_DIR%\generate-secrets.ps1" (
    echo [INFO] Generating secrets...
    cd /d "%PROJECT_DIR%"
    powershell -ExecutionPolicy Bypass -File generate-secrets.ps1 -All
    echo [OK] Secrets generated
)

echo.
echo ========================================
echo   Starting NovaBackup Server...
echo ========================================
echo.

REM Start the server
echo [INFO] Starting server on http://localhost:8000
echo [INFO] Opening browser in 3 seconds...
echo.

timeout /t 3 /nobreak >nul

REM Open browser
start http://localhost:8000
echo [OK] Browser opened

echo.
echo ========================================
echo   Server is Running!
echo ========================================
echo.
echo Login credentials:
echo    Username: alice
echo    Password: secret
echo.
echo Press CTRL+C to stop the server
echo.

cd /d "%PROJECT_DIR%"
python -m uvicorn novabackup.api:get_app --host 0.0.0.0 --port 8000

endlocal
