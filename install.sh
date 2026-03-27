#!/usr/bin/env bash
set -euo pipefail

echo "========================================"
echo "  NovaBackup Installer for Linux/macOS"
echo "========================================"
echo ""

# Check Python
if ! command -v python3 >/dev/null 2>&1; then
    echo "[ERROR] Python3 is not found."
    echo "Please install Python 3.9+ from https://python.org"
    exit 1
fi

echo "[OK] Python found:"
python3 --version
echo ""

# Set installation directory
INSTALL_ROOT="$HOME/.novabackup"
VENV_DIR="$INSTALL_ROOT/venv"

# Create virtual environment
if [ ! -d "$VENV_DIR" ]; then
    echo "[INFO] Creating virtual environment..."
    mkdir -p "$INSTALL_ROOT"
    python3 -m venv "$VENV_DIR"
    if [ $? -ne 0 ]; then
        echo "[ERROR] Failed to create virtual environment"
        exit 1
    fi
    echo "[OK] Virtual environment created"
else
    echo "[OK] Virtual environment exists"
fi

# Activate virtual environment
echo "[INFO] Activating virtual environment..."
source "$VENV_DIR/bin/activate"

# Upgrade pip
echo "[INFO] Upgrading pip..."
pip install --upgrade pip --quiet

# Install novabackup from current directory
echo "[INFO] Installing NovaBackup..."
if [ -f "pyproject.toml" ]; then
    pip install -e ".[api,dev]" --quiet
    if [ $? -eq 0 ]; then
        echo "[OK] NovaBackup installed successfully"
    else
        echo "[WARNING] Installation completed with warnings"
    fi
else
    echo "[ERROR] pyproject.toml not found in current directory"
    echo "Please run this script from the Backup directory"
    exit 1
fi

echo ""
echo "========================================"
echo "  Installation Complete!"
echo "========================================"
echo ""
echo "To activate NovaBackup, run:"
echo "  source $VENV_DIR/bin/activate"
echo ""
echo "Then run:"
echo "  novabackup --help"
echo ""

# Test installation
echo "[INFO] Testing installation..."
if command -v novabackup >/dev/null 2>&1; then
    echo "[OK] NovaBackup is working"
    novabackup list-vms 2>&1 | head -5 || echo "[INFO] VM list may require libvirt/Hyper-V"
else
    echo "[WARNING] novabackup command not found"
fi

echo ""
echo "Done!"
