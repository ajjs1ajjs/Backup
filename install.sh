#!/bin/bash

echo "========================================"
echo "  NovaBackup v7.0 - Installation"
echo "  Installation from GitHub"
echo "========================================"
echo ""

# Check root
if [ "$EUID" -ne 0 ]; then
    echo "[ERROR] Please run as root (sudo ./install.sh)"
    echo ""
    exit 1
fi

echo "[OK] Root privileges confirmed"
echo ""

INSTALL_DIR="/opt/novabackup"
DATA_DIR="/var/lib/novabackup"
SYSTEMD_DIR="/etc/systemd/system"
GITHUB_URL="https://github.com/ajjs1ajjs/Backup/releases/latest/download"

echo "[*] Downloading latest release from GitHub..."
echo "      URL: $GITHUB_URL"
echo ""

# Create directories
echo "[*] Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$DATA_DIR/logs"
mkdir -p "$DATA_DIR/backups"
mkdir -p "$DATA_DIR/config"

# Download latest release
echo "[*] Downloading novabackup-linux-amd64..."
cd /tmp
curl -L -o novabackup "$GITHUB_URL/novabackup-linux-amd64"
if [ $? -ne 0 ]; then
    echo "[ERROR] Download failed!"
    echo ""
    echo "Please download manually from:"
    echo "https://github.com/ajjs1ajjs/Backup/releases"
    exit 1
fi

# Install
echo "[*] Installing..."
chmod +x novabackup
cp novabackup "$INSTALL_DIR/NovaBackup"

# Create systemd service
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
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
echo "[*] Reloading systemd..."
systemctl daemon-reload

# Enable and start service
echo "[*] Enabling and starting service..."
systemctl enable novabackup
systemctl start novabackup

# Wait for service to start
echo "[*] Waiting for service to start..."
sleep 3

# Verify service is running
if systemctl is-active --quiet novabackup; then
    echo "[OK] NovaBackup service is RUNNING"
else
    echo "[WARNING] Service failed to start, trying manual start..."
    "$INSTALL_DIR/NovaBackup" server &
    sleep 2
fi

# Cleanup
echo "[*] Cleaning up..."
rm -f /tmp/novabackup

echo ""
echo "========================================"
echo "  Installation Complete Successfully!"
echo "========================================"
echo ""
echo "Installation Directory: $INSTALL_DIR"
echo "Data Directory: $DATA_DIR"
echo ""
echo "Service Status:"
systemctl status novabackup --no-pager -l
echo ""

# Check if web server is responding
echo "[*] Checking Web UI..."
for i in 1 2 3 4 5; do
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8050 | grep -q "200\|302"; then
        echo "[OK] Web UI is responding on http://localhost:8050"
        break
    fi
    sleep 1
done

echo ""
echo "Web UI: http://localhost:8050"
echo "Login: admin"
echo "Password: admin123"
echo ""

# Open browser if available
if command -v xdg-open &> /dev/null; then
    echo "Opening Web UI..."
    sleep 2
    xdg-open http://localhost:8050
fi

echo ""
