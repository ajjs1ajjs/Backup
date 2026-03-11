@echo off  
  
echo ========================================  
echo   NovaBackup v6.0 - Build Installer  
echo ========================================  
echo.  
echo Looking for Inno Setup...  
  
if exist \"C:\Program Files (x86)\Inno Setup 6\ISCC.exe\" (  
    \"C:\Program Files (x86)\Inno Setup 6\ISCC.exe\" novabackup_full.iss  
) else if exist \"C:\Program Files\Inno Setup 6\ISCC.exe\" (  
    \"C:\Program Files\Inno Setup 6\ISCC.exe\" novabackup_full.iss  
) else (  
    echo Inno Setup not found!  
    echo Please download from: https://jrsoftware.org/isdl.php  
    pause  
    exit /b 1  
)  
  
echo ========================================  
echo   Build Complete!  
echo ========================================  
echo.  
echo Installer created: output\  
  
@echo off
title NovaBackup v6.0 - Building Enterprise Installer
color 0B

echo ========================================
echo    NovaBackup v6.0 Enterprise
echo    Building Professional Installer
echo ========================================
echo.

REM Check if Inno Setup is installed
where iscc >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: Inno Setup is not installed!
    echo Please download and install Inno Setup from: https://jrsoftware.org/isinfo.php
    pause
    exit /b 1
)

echo Checking required files...
if not exist "..\nova.exe" (
    echo ERROR: nova.exe not found!
    echo Please build the main application first.
    pause
    exit /b 1
)

if not exist "..\NovaBackup-Enterprise.bat" (
    echo ERROR: NovaBackup-Enterprise.bat not found!
    pause
    exit /b 1
)

if not exist "..\gui\app.py" (
    echo ERROR: GUI files not found!
    pause
    exit /b 1
)

echo All required files found.
echo.

echo Creating installer directories...
if not exist "output" mkdir "output"
if not exist "icons" mkdir "icons"
if not exist "images" mkdir "images"

echo Creating icons...
if not exist "icons\novabackup.ico" (
    echo Creating NovaBackup icon...
    powershell -Command "Add-Type -AssemblyName System.Drawing; $bitmap = New-Object System.Drawing.Bitmap 64, 64; $graphics = [System.Drawing.Graphics]::FromImage($bitmap); $graphics.Clear([System.Drawing.Color]::Blue); $graphics.DrawString('NB', [System.Drawing.Font]::new('Arial', 12, [System.Drawing.FontStyle]::Bold), [System.Drawing.Brushes]::White, 10, 20); $bitmap.Save('icons\novabackup.ico', [System.Drawing.Imaging.ImageFormat]::Icon); $graphics.Dispose(); $bitmap.Dispose()"
)

echo Creating wizard images...
if not exist "images\wizard.bmp" (
    echo Creating wizard image...
    powershell -Command "Add-Type -AssemblyName System.Drawing; $bitmap = New-Object System.Drawing.Bitmap 164, 314; $graphics = [System.Drawing.Graphics]::FromImage($bitmap); $graphics.Clear([System.Drawing.Color]::FromArgb(0, 168, 230)); $graphics.DrawString('NovaBackup', [System.Drawing.Font]::new('Arial', 16, [System.Drawing.FontStyle]::Bold), [System.Drawing.Brushes]::White, 20, 50); $graphics.DrawString('Enterprise', [System.Drawing.Font]::new('Arial', 12), [System.Drawing.Brushes]::White, 20, 80); $graphics.DrawString('Backup & Recovery', [System.Drawing.Font]::new('Arial', 10), [System.Drawing.Brushes]::White, 20, 100); $bitmap.Save('images\wizard.bmp', [System.Drawing.Imaging.ImageFormat]::Bmp); $graphics.Dispose(); $bitmap.Dispose()"
)

if not exist "images\wizard-small.bmp" (
    echo Creating small wizard image...
    powershell -Command "Add-Type -AssemblyName System.Drawing; $bitmap = New-Object System.Drawing.Bitmap 55, 58; $graphics = [System.Drawing.Graphics]::FromImage($bitmap); $graphics.Clear([System.Drawing.Color]::FromArgb(0, 168, 230)); $graphics.DrawString('NB', [System.Drawing.Font]::new('Arial', 8, [System.Drawing.FontStyle]::Bold), [System.Drawing.Brushes]::White, 15, 20); $bitmap.Save('images\wizard-small.bmp', [System.Drawing.Imaging.ImageFormat]::Bmp); $graphics.Dispose(); $bitmap.Dispose()"
)

echo Creating license file...
echo NovaBackup v6.0 Enterprise License Agreement > license.txt
echo. >> license.txt
echo This software is provided "as-is" without warranty of any kind. >> license.txt
echo. >> license.txt
echo Copyright (c) 2026 NovaBackup Technologies. All rights reserved. >> license.txt

echo Creating readme file...
echo NovaBackup v6.0 Enterprise > readme.txt
echo ============================ >> readme.txt
echo. >> readme.txt
echo NovaBackup is an enterprise-grade backup and recovery solution >> readme.txt
echo designed for Windows environments. >> readme.txt
echo. >> readme.txt
echo Features: >> readme.txt
echo - Veeam-style interface >> readme.txt
echo - 15-component architecture >> readme.txt
echo - Advanced deduplication >> readme.txt
echo - Compression and encryption >> readme.txt
echo - Web-based management >> readme.txt
echo - Background agent service >> readme.txt
echo. >> readme.txt
echo System Requirements: >> readme.txt
echo - Windows 7 or later >> readme.txt
echo - 4GB RAM minimum >> readme.txt
echo - 10GB disk space >> readme.txt
echo - Administrator privileges >> readme.txt

echo.
echo Building installer...
echo.

REM Build the installer
iscc NovaBackup-Setup.iss

if %errorLevel% equ 0 (
    echo.
    echo ========================================
    echo    INSTALLER BUILT SUCCESSFULLY!
    echo ========================================
    echo.
    echo Installer location: output\NovaBackup-Enterprise-Setup.exe
    echo.
    echo The installer includes:
    echo - Veeam-style GUI interface
    echo - Web-based management console
    echo - Background agent service
    echo - Complete documentation
    echo - System integration
    echo.
    echo Ready for distribution!
) else (
    echo.
    echo ========================================
    echo    BUILD FAILED!
    echo ========================================
    echo.
    echo Please check the error messages above.
)

pause
