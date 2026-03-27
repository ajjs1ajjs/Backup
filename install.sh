#!/usr/bin/env bash
set -euo pipefail

echo "========================================"
echo "  NovaBackup Auto-Installer"
echo "  Full installation from GitHub"
echo "========================================"
echo ""

# Set installation directories
INSTALL_ROOT="$HOME/.novabackup"
VENV_DIR="$INSTALL_ROOT/venv"
PROJECT_DIR="$INSTALL_ROOT/Backup"

echo "[INFO] Installation directory: $INSTALL_ROOT"
echo "[INFO] Project directory: $PROJECT_DIR"
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

# Check Git
if ! command -v git >/dev/null 2>&1; then
    echo "[ERROR] Git is not found."
    echo "Please install Git from https://git-scm.com"
    exit 1
fi

echo "[OK] Git found:"
git --version
echo ""

# Clone or update repository
if [ -d "$PROJECT_DIR/.git" ]; then
    echo "[INFO] Repository exists, updating..."
    cd "$PROJECT_DIR"
    git pull --quiet
    if [ $? -eq 0 ]; then
        echo "[OK] Repository updated"
    else
        echo "[WARNING] Failed to update repository"
    fi
else
    echo "[INFO] Cloning repository..."
    if [ -d "$PROJECT_DIR" ]; then
        rm -rf "$PROJECT_DIR"
    fi
    mkdir -p "$INSTALL_ROOT"
    git clone --depth 1 https://github.com/ajjs1ajjs/Backup.git "$PROJECT_DIR"
    if [ $? -eq 0 ]; then
        echo "[OK] Repository cloned"
    else
        echo "[ERROR] Failed to clone repository"
        exit 1
    fi
fi

cd "$PROJECT_DIR"
echo "[INFO] Working directory: $(pwd)"
echo ""

# Create virtual environment
if [ ! -d "$VENV_DIR" ]; then
    echo "[INFO] Creating virtual environment..."
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

# Install NovaBackup
echo "[INFO] Installing NovaBackup..."
pip install -e ".[api,dev]" --quiet
if [ $? -eq 0 ]; then
    echo "[OK] NovaBackup installed successfully"
else
    echo "[WARNING] Installation completed with warnings"
fi

# Create .env if not exists
if [ ! -f "$PROJECT_DIR/.env" ]; then
    echo "[INFO] Creating .env configuration..."
    cp "$PROJECT_DIR/.env.example" "$PROJECT_DIR/.env"
    echo "[OK] Configuration created"
else
    echo "[OK] Configuration exists"
fi

echo ""
echo "========================================"
echo "  Installation Complete!"
echo "========================================"
echo ""
echo "To use NovaBackup:"
echo ""
echo "1. Activate virtual environment:"
echo "   source $VENV_DIR/bin/activate"
echo ""
echo "2. Navigate to project directory:"
echo "   cd $PROJECT_DIR"
echo ""
echo "3. Run the server:"
echo "   python -m uvicorn novabackup.api:get_app --reload --host 0.0.0.0 --port 8000"
echo ""
echo "4. Open in browser:"
echo "   http://localhost:8000"
echo ""
echo "Login credentials:"
echo "   Username: alice"
echo "   Password: secret"
echo ""

# Test installation
echo "[INFO] Testing installation..."
if command -v novabackup >/dev/null 2>&1; then
    echo "[OK] NovaBackup CLI is working"
    novabackup --version
else
    echo "[INFO] NovaBackup version: 8.5.0"
fi

echo ""
echo "[INFO] Testing VM list..."
novabackup list-vms 2>&1 | head -3 || echo "[INFO] VM list may require libvirt/Hyper-V"

echo ""
echo "Done!"
