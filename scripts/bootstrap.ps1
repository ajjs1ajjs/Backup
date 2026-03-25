Param()
@"
Note: This script bootstraps a Python venv and installs the package (Windows).
"@
$venv = ".\venv"
if(!(Test-Path $venv)) { python -m venv $venv }
& $venv\Scripts\Activate.ps1
pip install -e .
Write-Host "Bootstrap complete. Run: & '.\venv\Scripts\activate.ps1'; python -m novabackup list-vms" -ForegroundColor Green
