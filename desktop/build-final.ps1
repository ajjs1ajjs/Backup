# NOVA BACKUP - FINAL BUILD SCRIPT

Write-Host "Building NOVA Backup Final All-in-One executable..." -ForegroundColor Green
Write-Host ""

# Check if .NET 6.0 is installed
try {
    $dotnetVersion = & dotnet --version 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: .NET 6.0 SDK is not installed" -ForegroundColor Red
        Write-Host "Please install .NET 6.0 SDK from: https://dotnet.microsoft.com/download/dotnet/6.0" -ForegroundColor Red
        exit 1
    }
    Write-Host "✓ .NET 6.0 SDK found: $dotnetVersion" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Failed to check .NET SDK" -ForegroundColor Red
    exit 1
}

# Clean previous builds
Write-Host "Cleaning previous builds..." -ForegroundColor Yellow
& dotnet clean NovaBackup.Final.csproj --configuration Release --verbosity quiet

# Restore packages
Write-Host "Restoring packages..." -ForegroundColor Yellow
& dotnet restore NovaBackup.Final.csproj --verbosity quiet

# Build the project
Write-Host "Building Final All-in-One executable..." -ForegroundColor Yellow
try {
    & dotnet build NovaBackup.Final.csproj --configuration Release --verbosity minimal
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Build failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✓ Build completed successfully!" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Build failed with exception" -ForegroundColor Red
    exit 1
}

# Publish the project
Write-Host "Publishing Final All-in-One executable..." -ForegroundColor Yellow
try {
    & dotnet publish NovaBackup.Final.csproj `
        --configuration Release `
        --runtime win-x64 `
        --self-contained true `
        --output "../publish" `
        --verbosity normal
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Publish failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✓ Publish completed successfully!" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Publish failed with exception" -ForegroundColor Red
    exit 1
}

# Check if executable was created
$exePath = "../publish/NovaBackup.Desktop.exe"
if (Test-Path $exePath) {
    Write-Host ""
    Write-Host "🎉 SUCCESS: All-in-One executable created!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📁 Location: $exePath" -ForegroundColor Cyan
    $fileInfo = Get-Item $exePath
    $fileSize = [math]::Round($fileInfo.Length / 1MB, 2)
    Write-Host "📊 Size: $fileSize MB" -ForegroundColor Cyan
    Write-Host ""
    
    # Create installer directory
    if (!(Test-Path "../installer")) {
        New-Item -ItemType Directory -Path "../installer" | Out-Null
    }
    
    # Copy executable to installer directory
    Copy-Item $exePath "../installer/NovaBackup.exe" -Force
    Write-Host "📦 Copied to: ../installer/NovaBackup.exe" -ForegroundColor Green
    
    # Copy launcher script
    if (Test-Path "../start-nova-backup.bat") {
        Copy-Item "../start-nova-backup.bat" "../installer/start-nova-backup.bat" -Force
        Write-Host "📦 Copied launcher: ../installer/start-nova-backup.bat" -ForegroundColor Green
    }
    
    Write-Host ""
    Write-Host "🌟 FEATURES INCLUDED:" -ForegroundColor Yellow
    Write-Host "✅ Complete desktop application with full GUI" -ForegroundColor Green
    Write-Host "✅ Web console with remote access" -ForegroundColor Green
    Write-Host "✅ Windows Service functionality" -ForegroundColor Green
    Write-Host "✅ All dependencies and libraries" -ForegroundColor Green
    Write-Host "✅ Embedded web UI files" -ForegroundColor Green
    Write-Host ""
    Write-Host "🌐 ACCESS METHODS:" -ForegroundColor Yellow
    Write-Host "• Local: http://localhost:8080 (no auth)" -ForegroundColor Cyan
    Write-Host "• Remote: http://[IP]:8080 (admin/admin)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "🚀 READY TO USE!" -ForegroundColor Green
    Write-Host "Run NovaBackup.exe directly or use start-nova-backup.bat" -ForegroundColor White
    
} else {
    Write-Host "ERROR: Executable not found at $exePath" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Press any key to exit..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
