# Backup System - Installation

## Quick Install (One Command)

```bash
curl -fsSL https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh | sudo bash -s -- --auto-start
```

Or save script first:
```bash
curl -fsSL -o install.sh https://raw.githubusercontent.com/ajjs1ajjs/Backup/main/install.sh
sudo chmod +x install.sh && sudo ./install.sh --auto-start
```

## What Gets Installed

- PostgreSQL database
- .NET 8 SDK
- Node.js 18
- Backup Server (port 8000)
- Backup UI (port 80 via Nginx)
- Swagger API documentation

## Access

After installation:
- **UI**: http://localhost
- **API**: http://localhost:8000
- **Swagger**: http://localhost:8000/swagger

## Login

- Username: `admin`
- Password: `admin123`

**Important**: Change password on first login!

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `--auto-start` | Start services after install | true |
| `--jwt-key KEY` | JWT secret key | auto-generated |
| `--postgres-password` | PostgreSQL password | postgres |

## Status & Logs

```bash
# Check server status
systemctl status backup-server

# View logs
journalctl -u backup-server -f
```

## Uninstall

```bash
systemctl stop backup-server
systemctl disable backup-server
rm -rf /opt/backup
rm -f /etc/systemd/system/backup-server.service
systemctl daemon-reload
```
