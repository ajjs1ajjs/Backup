#!/usr/bin/env bash
set -euo pipefail
echo "Novabackup installer (Linux/macOS)"

if ! command -v python3 >/dev/null 2>&1; then
  echo "Python3 is required. Please install Python 3.9+." >&2
  exit 1
fi
if ! command -v pip3 >/dev/null 2>&1; then
  echo "pip3 is required. Please install pip." >&2
  exit 1
fi

INSTALL_ROOT="$HOME/.novabackup"
VENV_DIR="$INSTALL_ROOT/venv"

mkdir -p "$INSTALL_ROOT"

if [ ! -d "$VENV_DIR" ]; then
  python3 -m venv "$VENV_DIR"
fi

source "$VENV_DIR/bin/activate"

pip install -e ".[dev,api]"

echo "Novabackup installed in virtual environment: $VENV_DIR"
echo "To use: source $VENV_DIR/bin/activate"
echo "Then run: novabackup list-vms"

# Try to fetch repo if not present (git required)
if command -v git >/dev/null 2>&1; then
  REPO_URL_DEFAULT=${NOVABACKUP_REPO_URL:-https://github.com/ajjs1ajjs/Backup}
  TMPDIR=$(mktemp -d)
  git clone --depth 1 "$REPO_URL_DEFAULT" "$TMPDIR/novabackup" || true
  if [ -d "$TMPDIR/novabackup" ]; then
    echo "Found repo at $REPO_URL_DEFAULT, attempting to install from source."
    cd "$TMPDIR/novabackup"
    if [ -f pyproject.toml ] || [ -f setup.py ]; then
      pip install -e .
    fi
  fi
else
  echo "Git is not installed. To install from source, clone the repository manually and run 'pip install -e .[dev,api]'."
fi
