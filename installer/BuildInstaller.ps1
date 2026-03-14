param(
    [string]$Configuration = "Release",
    [string]$Version = "4.0.0.0",
    [string]$OutputPath = ".\NovaBackup-4.0.0.msi"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  NovaBackup v4.0 - MSI Installer Build" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Set error handling
$ErrorActionPreference = "Stop"

# Configuration
$SolutionPath = "..\NovaBackup.sln"
$PublishPath = "..\publish\installer"
$InstallerProject = "Installer.wixproj"

Write-Host "Step 1: Publishing NovaBackup..." -ForegroundColor Yellow
dotnet publish $SolutionPath -c $Configuration -o $PublishPath --self-contained false -r win-x64

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Publishing failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 2: Building MSI Installer..." -ForegroundColor Yellow

# Build WiX project
dotnet build $InstallerProject -c $Configuration /p:Version=$Version /p:PublishDir=$PublishPath

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: MSI build failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 3: Copying MSI to output..." -ForegroundColor Yellow

# Copy MSI to output path
$BuiltMSI = ".\bin\$Configuration\NovaBackup.msi"
if (Test-Path $BuiltMSI) {
    Copy-Item $BuiltMSI $OutputPath -Force
    Write-Host "SUCCESS: MSI created at $OutputPath" -ForegroundColor Green

    # Show file info
    $FileInfo = Get-Item $OutputPath
    Write-Host ""
    Write-Host "File Information:" -ForegroundColor Cyan
    Write-Host "  Name: $($FileInfo.Name)"
    Write-Host "  Size: $([math]::Round($FileInfo.Length / 1MB, 2)) MB"
    Write-Host "  Created: $($FileInfo.CreationTime)"
    Write-Host ""
} else {
    Write-Host "ERROR: MSI file not found!" -ForegroundColor Red
    exit 1
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Build Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "To install, run:" -ForegroundColor White
Write-Host "  msiexec /i `"$OutputPath`"" -ForegroundColor Green
Write-Host ""
Write-Host "For silent install:" -ForegroundColor White
Write-Host "  msiexec /i `"$OutputPath`" /quiet /norestart" -ForegroundColor Green
Write-Host ""
