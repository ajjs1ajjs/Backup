# Backup System Installer - Enterprise Edition (Full)
$ErrorActionPreference = "Stop"

Write-Host "[*] Checking dependencies..."

Function Install-Dep {
    param($Id, $Cmd)
    if (!(Get-Command $Cmd -ErrorAction SilentlyContinue)) {
        Write-Host "[!] $Id not found. Installing..."
        winget install --id $Id --silent --accept-package-agreements
    }
}

Install-Dep "Microsoft.DotNet.SDK.8" "dotnet"
Install-Dep "OpenJS.NodeJS" "node"
Install-Dep "PostgreSQL.PostgreSQL.16" "psql"
# Встановлення Build Tools для C++ та VSS
Install-Dep "Microsoft.VisualStudio.2022.BuildTools" "msbuild"

$path = "C:\BackupServer"
if (!(Test-Path $path)) { New-Item -ItemType Directory -Path $path }

Write-Host "[*] Cloning repository..."
$temp = Join-Path $env:TEMP ("Backup-" + [Guid]::NewGuid().ToString())
git clone https://github.com/ajjs1ajjs/Backup.git $temp

Write-Host "[*] Building server (cleaning obj)..."
Remove-Item -Path "$temp\src\server\Backup.Server\obj" -Recurse -Force -ErrorAction SilentlyContinue
Set-Location "$temp\src\server\Backup.Server"
dotnet publish -c Release -o "$path\publish"

Write-Host "[*] Building C++ Agent with VSS support..."
Set-Location "$temp\src\agent\Backup.Agent"
mkdir build
Set-Location build
cmake ..
cmake --build . --config Release
Copy-Item "Release\Backup.Agent.exe" "$path\publish\Backup.Agent.exe"

# Config
if (!(Test-Path "$path\appsettings.json")) {
    $json = '{ "ConnectionStrings": { "DefaultConnection": "Host=localhost;Database=backup_db;Username=postgres;Password=postgres" }, "Server": { "PublicUrl": "http://localhost:8000" } }'
    $json | Out-File -FilePath "$path\appsettings.json" -Encoding ascii
}

# Service setup
$exePath = Join-Path $path "publish\Backup.Server.exe"
$serviceName = "BackupServer"

Write-Host "[*] Configuring Windows Service..."
sc.exe stop $serviceName 2>$null
sc.exe delete $serviceName 2>$null
sc.exe create $serviceName binPath= "$exePath" start= auto DisplayName= "Fortress Backup Server"
sc.exe start $serviceName

Write-Host "[======================================]"
Write-Host "[+] Installation complete."
Write-Host "[+] Server and VSS-Agent installed!"
Write-Host "[======================================]"
