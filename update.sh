#!/bin/bash

echo "========================================"
echo "  NovaBackup v7.0 - Update"
echo "  Update from GitHub"
echo "========================================"
echo ""

# Check root
if [ "$EUID" -ne 0 ]; then
    echo "[ERROR] Please run as root (sudo ./update.sh)"
    echo ""
    exit 1
fi

echo "[OK] Root privileges confirmed"
echo ""

INSTALL_DIR="/opt/novabackup"
GITHUB_URL="https://github.com/ajjs1ajjs/Backup/releases/latest/download"
ENV_FILE="/etc/novabackup.env"
SYSTEMD_FILE="/etc/systemd/system/novabackup.service"

# Check if installed
if [ ! -f "$INSTALL_DIR/NovaBackup" ]; then
    echo "[ERROR] NovaBackup not installed!"
    echo ""
    echo "Please run install.sh first"
    exit 1
fi

echo "[*] Current version: $INSTALL_DIR/NovaBackup"
echo ""
echo "[*] Downloading latest release from GitHub..."
echo "      URL: $GITHUB_URL"
echo ""

# Stop service
echo "[*] Stopping NovaBackup service..."
systemctl stop novabackup

if [ -n "$NOVABACKUP_MASTER_KEY" ]; then
    echo "[*] Updating NOVABACKUP_MASTER_KEY..."
    echo "NOVABACKUP_MASTER_KEY=$NOVABACKUP_MASTER_KEY" > "$ENV_FILE"
    chmod 600 "$ENV_FILE"
fi

if [ -f "$SYSTEMD_FILE" ]; then
    if ! grep -q "EnvironmentFile=-$ENV_FILE" "$SYSTEMD_FILE"; then
        sed -i "/\\[Service\\]/a EnvironmentFile=-$ENV_FILE" "$SYSTEMD_FILE"
        systemctl daemon-reload
    fi
fi

# Backup current version
echo "[*] Backing up current version..."
cp "$INSTALL_DIR/NovaBackup" "/tmp/NovaBackup.old"

# Download latest
echo "[*] Downloading novabackup-linux-amd64..."
cd /tmp
curl -L -o novabackup "$GITHUB_URL/novabackup-linux-amd64"
if [ $? -ne 0 ]; then
    echo "[ERROR] Download failed!"
    echo ""
    echo "Restoring previous version..."
    cp "/tmp/NovaBackup.old" "$INSTALL_DIR/NovaBackup"
    systemctl start novabackup
    exit 1
fi

# Install new version
echo "[*] Installing new version..."
chmod +x novabackup
cp novabackup "$INSTALL_DIR/NovaBackup"

# Start service
echo "[*] Starting service..."
systemctl start novabackup

# Cleanup
echo "[*] Cleaning up..."
sleep 2
rm -f /tmp/novabackup /tmp/NovaBackup.old

echo ""
echo "========================================"
echo "  Update Complete Successfully!"
echo "========================================"
echo ""
echo "Service Status:"
systemctl status novabackup --no-pager -l
echo ""
echo "Web UI: http://localhost:8050"
echo ""

# Open browser if available
if command -v xdg-open &> /dev/null; then
    echo "Opening Web UI..."
    sleep 2
    xdg-open http://localhost:8050
fi

echo ""
