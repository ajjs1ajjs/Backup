# NovaBackup MSI Installer Build Script
param(
    [string]$Version = "6.0.0.0"
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$InstallerDir = Join-Path $ProjectRoot "installer"
$OutputDir = Join-Path $ProjectRoot "build\msi"
$BuildInstaller = Join-Path $ProjectRoot "build\installer"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  NovaBackup MSI Builder v6.0" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Create output directory
if (!(Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Step 1: Build Go service
Write-Host "[1/4] Building Go service..." -ForegroundColor Yellow
Set-Location $ProjectRoot
go build -o "$BuildInstaller\nova.exe" ./cmd/nova-service/
Write-Host "      Done" -ForegroundColor Green

# Step 2: Build WPF GUI
Write-Host "[2/4] Building WPF GUI..." -ForegroundColor Yellow
Set-Location (Join-Path $ProjectRoot "cmd\nova-wpf")
dotnet build -c Release /p:Platform=x64 | Out-Null
Write-Host "      Done" -ForegroundColor Green

# Step 3: Copy files to build\installer
Write-Host "[3/4] Copying files..." -ForegroundColor Yellow
$WpfBin = Join-Path $ProjectRoot "cmd\nova-wpf\bin\x64\Release\net8.0-windows"
$Files = @("NovaBackup.exe", "NovaBackup.dll", "NovaBackup.pdb", "NovaBackup.deps.json", "NovaBackup.runtimeconfig.json",
           "MaterialDesignColors.dll", "MaterialDesignThemes.Wpf.dll", "Microsoft.Xaml.Behaviors.dll",
           "System.ServiceProcess.ServiceController.dll")
foreach ($f in $Files) {
    Copy-Item (Join-Path $WpfBin $f) -Destination $BuildInstaller -Force
}
Copy-Item (Join-Path $ProjectRoot "nova.exe") -Destination $BuildInstaller -Force
Write-Host "      Done" -ForegroundColor Green

# Step 4: Build MSI with WiX
Write-Host "[4/4] Building MSI installer..." -ForegroundColor Yellow
Set-Location $InstallerDir

# Find WiX
$wixPath = ""
if (Test-Path "C:\Program Files (x86)\WiX Toolset v3.14\bin") {
    $wixPath = "C:\Program Files (x86)\WiX Toolset v3.14\bin"
} elseif (Test-Path "C:\Program Files (x86)\WiX Toolset v3.11\bin") {
    $wixPath = "C:\Program Files (x86)\WiX Toolset v3.11\bin"
}

if ($wixPath -eq "" -or !(Test-Path "$wixPath\candle.exe")) {
    Write-Host "      WiX not found! Creating standalone EXE installer..." -ForegroundColor Yellow

    # Create self-extracting installer
    $SetupContent = @"
@echo off
setlocal EnableDelayedExpansion
echo ========================================
echo   NovaBackup Enterprise v6.0 - Setup
echo ========================================
echo.
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Administrator rights required!
    echo Right-click this file -^> Run as administrator
    pause
    exit /b 1
)
set "SOURCE_DIR=%~dp0"
set "INSTALL_DIR=C:\Program Files\NovaBackup"
set "DATA_DIR=C:\ProgramData\NovaBackup"
echo [*] Stopping existing service...
net stop NovaBackup >nul 2>&1
sc delete NovaBackup >nul 2>&1
timeout /t 2 /nobreak >nul
echo.
echo [*] Creating directories...
mkdir "%INSTALL_DIR%" 2>nul
mkdir "%DATA_DIR%" 2>nul
mkdir "%DATA_DIR%\Logs" 2>nul
mkdir "%DATA_DIR%\Backups" 2>nul
mkdir "%DATA_DIR%\Config" 2>nul
echo.
echo [*] Copying program files...
xcopy /E /Y /I /Q "%SOURCE_DIR%*" "%INSTALL_DIR%\"
echo.
echo [*] Installing Windows Service...
cd /d "%INSTALL_DIR%"
nova.exe install
if %errorLevel% neq 0 (
    echo [ERROR] Failed to install service!
    pause
    exit /b 1
)
echo.
echo [*] Starting Service...
nova.exe start
echo.
echo [*] Creating shortcuts...
powershell -Command "$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%APPDATA%\Microsoft\Windows\Start Menu\Programs\NovaBackup Enterprise.lnk'); $Shortcut.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $Shortcut.WorkingDirectory = '%INSTALL_DIR%'; $Shortcut.Save()"
powershell -Command "$WshShell = New-Object -ComObject WScript.Shell; $Shortcut = $WshShell.CreateShortcut('%USERPROFILE%\Desktop\NovaBackup Enterprise.lnk'); $Shortcut.TargetPath = '%INSTALL_DIR%\NovaBackup.exe'; $Shortcut.WorkingDirectory = '%INSTALL_DIR%'; $Shortcut.Save()"
echo.
echo ========================================
echo   Installation Complete!
echo ========================================
echo.
echo Installation: %INSTALL_DIR%
echo Data: %DATA_DIR%
echo.
start "" "%INSTALL_DIR%\NovaBackup.exe"
echo.
pause
"@
    Set-Content -Path (Join-Path $OutputDir "NovaBackup-Setup.bat") -Value $SetupContent -Force

    # Create ZIP archive
    Compress-Archive -Path "$BuildInstaller\*" -DestinationPath (Join-Path $OutputDir "NovaBackup-$Version.zip") -Force

    Write-Host "      Created: NovaBackup-Setup.bat and NovaBackup-$Version.zip" -ForegroundColor Green
} else {
    # Build with WiX
    & "$wixPath\candle.exe" -arch x64 -ext WixUIExtension -out "$OutputDir\" Product.wxs
    & "$wixPath\light.exe" -ext WixUIExtension -out "$OutputDir\NovaBackup-$Version-x64.msi" "$OutputDir\Product.wixobj"
    Write-Host "      Created: NovaBackup-$Version-x64.msi" -ForegroundColor Green
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Build Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Output: $OutputDir" -ForegroundColor Cyan
Write-Host ""
