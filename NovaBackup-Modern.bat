@echo off
title NovaBackup v6.0 - Modern GUI
color 0B

echo ========================================
echo    NovaBackup v6.0 - Modern GUI
echo    Powered by Python + React
echo ========================================
echo.

REM Check if Python is installed
python --version >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: Python is not installed!
    echo Please install Python 3.7+ from https://python.org
    pause
    exit /b 1
)

REM Check if pip is available
pip --version >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: pip is not available!
    echo Please install pip first
    pause
    exit /b 1
)

echo Installing required Python packages...
pip install -r requirements.txt

if %errorLevel% neq 0 (
    echo ERROR: Failed to install Python packages!
    pause
    exit /b 1
)

echo.
echo Starting NovaBackup Modern GUI...
echo.

REM Change to GUI directory
cd gui

echo Launching web interface...
echo The GUI will open in your default browser...
echo.
python app.py

pause
