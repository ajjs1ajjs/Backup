#!/usr/bin/env bash
set -euo pipefail

# Linux/macOS bootstrap: create venv and install package in editable mode
if [ -d "venv" ]; then
  echo "venv already exists. Activating..."
else
  python3 -m venv venv
fi
source venv/bin/activate
pip install -e .
echo "Bootstrap complete. Use: source venv/bin/activate and run 'python -m novabackup'"
