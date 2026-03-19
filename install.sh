#!/bin/bash

# ========================================
#   NovaBackup v7.0 - Auto Installation
#   One-command fully automated install
# ========================================

echo "========================================"
echo "  NovaBackup v7.0 - Auto Installation"
echo "========================================"
echo ""

# Check root
if [ "$EUID" -ne 0 ]; then
    echo "[OK] Restarting with sudo..."
    exec sudo bash "$0" "$@"
fi

echo "[OK] Root privileges confirmed"
echo ""

INSTALL_DIR="/opt/novabackup"
DATA_DIR="/var/lib/novabackup"
SYSTEMD_DIR="/etc/systemd/system"
RAW_URL="https://raw.githubusercontent.com/ajjs1ajjs/Backup/main"

echo "[*] Downloading from GitHub..."
cd /tmp

# Download latest release
curl -sL -o novabackup "$RAW_URL/novabackup-linux-amd64"
if [ ! -f novabackup ] || [ ! -s novabackup ]; then
    echo "[ERROR] Download failed!"
    exit 1
fi

echo "[*] Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$DATA_DIR/logs"
mkdir -p "$DATA_DIR/backups"
mkdir -p "$DATA_DIR/config"

echo "[*] Installing..."
chmod +x novabackup
cp novabackup "$INSTALL_DIR/NovaBackup"

echo "[*] Creating systemd service..."
cat > "$SYSTEMD_DIR/novabackup.service" << EOF
[Unit]
Description=NovaBackup Enterprise v7.0
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/NovaBackup server
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

echo "[*] Reloading systemd..."
systemctl daemon-reload

echo "[*] Enabling service..."
systemctl enable novabackup

echo "[*] Starting service..."
systemctl start novabackup

# Wait for service to start
sleep 3

# Verify and fallback
if ! systemctl is-active --quiet novabackup; then
    echo "[WARNING] Service failed, starting manually..."
    nohup "$INSTALL_DIR/NovaBackup" server > /dev/null 2>&1 &
    sleep 2
fi

# Cleanup
rm -f /tmp/novabackup

echo ""
echo "========================================"
echo "  Installation Complete!"
echo "========================================"
echo ""
echo "Installation: $INSTALL_DIR"
echo "Data: $DATA_DIR"
echo ""

if systemctl is-active --quiet novabackup; then
    echo "[OK] Service: RUNNING"
else
    echo "[!] Service: Background mode"
fi

echo ""
echo "Web UI: http://localhost:8050"
echo "Login: admin"
echo "Password: admin123"
echo ""

# Check if web server is responding
echo "[*] Checking Web UI..."
for i in 1 2 3 4 5; do
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8050 2>/dev/null | grep -qE "200|302"; then
        echo "[OK] Web UI is responding"
        break
    fi
    sleep 1
done

echo ""
echo "Done!"
sleep 2

# Open browser if available
if command -v xdg-open &> /dev/null; then
    xdg-open http://localhost:8050 2>/dev/null &
fi

exit 0
