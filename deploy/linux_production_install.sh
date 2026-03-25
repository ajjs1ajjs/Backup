#!/usr/bin/env bash
set -euo pipefail

echo "[Production Install] Novabackup on Linux (Debian/Ubuntu)"
declare -r INSTALL_ROOT="/opt/novabackup"
declare -r VENV_DIR="$INSTALL_ROOT/venv"
declare -r REPO_URL="https://github.com/ajjs1ajjs/Backup.git"

if [ "$(id -u)" -ne 0 ]; then
  echo "This script must be run as root. Try using sudo." >&2
  exit 1
fi

apt-get update -y
apt-get install -y python3-venv python3-pip git curl ca-certificates

mkdir -p "$INSTALL_ROOT"
cd /tmp
rm -rf novabackup-prod || true
git clone --depth 1 "$REPO_URL" novabackup-prod
cd novabackup-prod

python3 -m venv "$VENV_DIR"
source "$VENV_DIR/bin/activate"
pip install -U pip
pip install -e ".[api]"  # core prod dependencies
pip install -e .         # install local package

deactivate
echo "Production installation prepared at $INSTALL_ROOT."

if [ -d /etc/systemd/system ]; then
  if [ -f /tmp/novabackup-prod/deploy/systemd/novabackup.service ]; then
    cp /tmp/novabackup-prod/deploy/systemd/novabackup.service /etc/systemd/system/novabackup.service
    systemctl daemon-reload
    systemctl enable novabackup
    systemctl start novabackup
    echo "Systemd service novabackup started."
  else
    echo "Systemd service template not found in patch; please copy deploy/systemd/novabackup.service manually and start service." 
  fi
else
  echo "Systemd not available on this platform; you can run API via uvicorn manually or install as a service using your OS tooling."
fi

echo "Optional: run migrations and verify API docs."
