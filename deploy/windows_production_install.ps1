Param()
@"
Windows Production Installer for Novabackup
"@
$installDir = "C:\\ProgramData\\Novabackup"
if(!(Test-Path $installDir)) { New-Item -ItemType Directory -Force -Path $installDir | Out-Null }
if(!(Get-Command python -ErrorAction SilentlyContinue)) {
  Write-Error "Python is required to install production."
  exit 1
}
python -m venv "$installDir\\venv" | Out-Null
& "$installDir\\venv\\Scripts\\activate.ps1"; pip install --upgrade pip
pip install -e ".[api]"; pip install -e .
Write-Host "Production install prepared at $installDir" -ForegroundColor Green
