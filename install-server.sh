#!/bin/bash
set -e

echo "[$(date)] Starting server installation..."

# Install Docker if not present
if ! command -v docker &> /dev/null; then
    echo "[$(date)] Installing Docker..."
    sudo apt update
    sudo apt install -y docker.io
    sudo systemctl start docker
    sudo systemctl enable docker
fi

# Clear cache
echo "[$(date)] Clearing Docker cache..."
sudo docker system prune -af --volumes 2>/dev/null || true

# Build and run server
echo "[$(date)] Building and running server..."

cd /tmp
rm -rf Backup-main server.zip 2>/dev/null

curl -fsSL -o server.zip https://github.com/ajjs1ajjs/Backup/archive/refs/heads/main.zip
unzip -q server.zip

cd Backup-main/src/server

# Stop and remove old container
sudo docker stop backup-server 2>/dev/null || true
sudo docker rm backup-server 2>/dev/null || true

# Build
sudo docker build -t backup-server .

# Run
sudo docker run -d -p 8050:8050 --name backup-server backup-server

echo "[$(date)] Server started on http://localhost:8050"
echo "[$(date)] Login: admin / admin123"
