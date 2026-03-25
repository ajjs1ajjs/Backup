@echo off
setlocal
echo Novabackup installer (Windows)

where python >nul 2>&1
if %errorlevel% neq 0 (
  echo Python is not found on PATH. Please install Python 3.9+ and ensure it's on PATH.
  exit /b 1
)

set "INSTALL_DIR=%USERPROFILE%\.novabackup"
set "VENV=%INSTALL_DIR%\venv"
if not exist "%VENV%" (
  mkdir "%INSTALL_DIR%" >nul 2>&1
)
python -m venv "%VENV%"
call "%VENV%\Scripts\activate.bat"

pip install --upgrade pip
pip install -e ".[api,dev]"

echo Novabackup installed. Use: call "%VENV%\Scripts\activate.bat" and run 'novabackup'"
novabackup list-vms || echo "Note: VM list may require Windows Hyper-V to be enabled."

REM Optional: Try to fetch repo for source installation if git is available
where git >nul 2>&1
if %errorlevel% equ 0 (
  echo Fetching repository for source installation...
  set "REPO_URL_DEFAULT=https://github.com/ajjs1ajjs/Backup"
  set "TMPDIR=%TEMP%\novabackup_install"
  if exist "%TMPDIR%" (rmdir /s /q "%TMPDIR%")
  mkdir "%TMPDIR%"
  git clone --depth 1 "%REPO_URL_DEFAULT%" "%TMPDIR%\novabackup" >nul 2>&1 || (
    echo Failed to clone repository. You can install manually after cloning.
  )
  if exist "%TMPDIR%\novabackup\pyproject.toml" (
    call "%VENV%\Scripts\activate.bat"
    cd /d "%TMPDIR%\novabackup"
    pip install -e .
  )
)
