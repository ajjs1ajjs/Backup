#!/bin/bash

set -e

VERSION="1.0.0"
INSTALL_DIR="/opt/backup"
BUILD_DIR="/tmp/backup-build"
JWT_KEY=""
POSTGRES_PASSWORD="postgres"
AUTO_START=true

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"; }
error() { echo "[ERROR] $1" >&2; exit 1; }

show_help() {
    cat << EOF
Backup System v$VERSION - Universal Installer

Usage: $0 [OPTIONS]

Options:
    --auto-start         Start services after installation (default: true)
    --jwt-key KEY        JWT secret key (auto-generated if not provided)
    --postgres-password PASSWORD  PostgreSQL password (default: postgres)
    -h, --help           Show this help

Example:
    $0 --auto-start

EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --auto-start) AUTO_START=true; shift ;;
            --jwt-key) JWT_KEY="$2"; shift 2 ;;
            --postgres-password) POSTGRES_PASSWORD="$2"; shift 2 ;;
            -h|--help) show_help; exit 0 ;;
            *) shift ;;
        esac
    done
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root (use sudo)"
    fi
}

install_postgres() {
    log "Installing PostgreSQL..."
    
    if command -v pg_isready &> /dev/null; then
        log "PostgreSQL already installed"
        systemctl enable postgresql 2>/dev/null || true
        systemctl start postgresql 2>/dev/null || true
        return 0
    fi
    
    if command -v apt-get &> /dev/null; then
        apt-get update -qq
        apt-get install -y -qq postgresql postgresql-contrib
    elif command -v yum &> /dev/null; then
        yum install -y postgresql-server postgresql
    elif command -v dnf &> /dev/null; then
        dnf install -y postgresql-server postgresql
    fi
    
    systemctl enable postgresql 2>/dev/null || true
    systemctl start postgresql 2>/dev/null || true
    
    sleep 3
    
    if command -v psql &> /dev/null; then
        log "Configuring PostgreSQL..."
        sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD '$POSTGRES_PASSWORD';" 2>/dev/null || true
        sudo -u postgres psql -c "CREATE DATABASE backup;" 2>/dev/null || true
    fi
    
    log "PostgreSQL ready"
}

install_dependencies() {
    log "Installing dependencies..."
    
    local missing=()
    for cmd in dotnet node npm git cmake make g++; do
        if ! command -v $cmd &> /dev/null; then
            missing+=($cmd)
        fi
    done
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        if command -v apt-get &> /dev/null; then
            apt-get update -qq
            apt-get install -y -qq wget curl git cmake build-essential
        fi
    fi
    
    if ! command -v dotnet &> /dev/null; then
        log "Installing .NET SDK 8.0..."
        wget -q https://dot.net/v1/dotnet-install.sh -O /tmp/dotnet-install.sh
        chmod +x /tmp/dotnet-install.sh
        /tmp/dotnet-install.sh --channel 8.0 --install-dir /opt/dotnet
        ln -sf /opt/dotnet/dotnet /usr/local/bin/dotnet
    fi
    
    if ! command -v node &> /dev/null; then
        log "Installing Node.js 18..."
        if command -v apt-get &> /dev/null; then
            curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
            apt-get install -y -qq nodejs
        fi
    fi
    
    log "Dependencies ready"
}

clone_repo() {
    log "Cloning repository..."
    rm -rf "$BUILD_DIR"
    git clone https://github.com/ajjs1ajjs/Backup.git "$BUILD_DIR"
}

generate_jwt_key() {
    if [[ -z "$JWT_KEY" ]]; then
        JWT_KEY=$(openssl rand -base64 32 2>/dev/null || head -c 32 /dev/urandom | base64)
    fi
}

install_server() {
    log "Building Backup Server..."
    
    local server_src="$BUILD_DIR/src/server/Backup.Server"
    local server_install="$INSTALL_DIR/server"
    mkdir -p "$server_install"
    
    cd "$server_src"
    dotnet restore
    dotnet publish -c Release -o "$server_install/publish"

    local public_url="http://$(hostname -I | awk '{print $1}'):8000"
    
    cat > "$server_install/publish/appsettings.json" << EOF
{
  "ConnectionStrings": {
    "DefaultConnection": "Host=localhost;Database=backup;Username=postgres;Password=$POSTGRES_PASSWORD"
  },
  "Jwt": {
    "Key": "$JWT_KEY",
    "Issuer": "BackupServer",
    "Audience": "BackupClients"
  },
  "Server": {
    "PublicUrl": "$public_url"
  },
  "BootstrapAdmin": {
    "Username": "admin",
    "Email": "admin@backupsystem.com",
    "Password": "admin123"
  }
}
EOF

    cat > /etc/systemd/system/backup-server.service << EOF
[Unit]
Description=Backup Server
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=$server_install/publish
ExecStart=/opt/dotnet/dotnet $server_install/publish/Backup.Server.dll --urls=http://0.0.0.0:8000
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload

    if [[ "$AUTO_START" == "true" ]]; then
        log "Starting server..."
        systemctl enable backup-server
        systemctl start backup-server
        sleep 5
    fi

    log "Server installed"
}

install_ui() {
    log "Building Backup UI..."
    
    local ui_src="$BUILD_DIR/src/ui"
    mkdir -p "$INSTALL_DIR/ui"
    
    cd "$ui_src"
    npm install --production 2>/dev/null || npm install
    npm run build
    
    if [[ -d "build" ]]; then
        cp -r build/* "$INSTALL_DIR/ui/"
    fi
    
    log "UI installed"
}

configure_nginx() {
    if command -v nginx &> /dev/null; then
        log "Configuring Nginx..."
        
        cat > /etc/nginx/sites-available/backup << EOF
server {
    listen 80;
    server_name _;
    
    root $INSTALL_DIR/ui;
    index index.html;
    
    location / {
        try_files \$uri \$uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }
    
    location /grpc {
        grpc_pass localhost:8000;
    }
}
EOF

        ln -sf /etc/nginx/sites-available/backup /etc/nginx/sites-enabled/backup
        nginx -t && systemctl reload nginx
    else
        log "Nginx not found, UI available at $INSTALL_DIR/ui"
    fi
}

main() {
    parse_args "$@"
    check_root
    
    generate_jwt_key
    
    log ""
    log "========================================="
    log "Installing Backup System v$VERSION..."
    log "========================================="
    
    install_postgres
    install_dependencies
    clone_repo
    install_server
    install_ui
    configure_nginx
    
    log ""
    log "========================================="
    log "Installation Complete!"
    log "========================================="
    log ""
    log "Access the application:"
    log "  UI: http://localhost:80"
    log "  API: http://localhost:8000"
    log "  Swagger: http://localhost:8000/swagger"
    log ""
    log "Login credentials:"
    log "  Username: admin"
    log "  Password: admin123"
    log ""
    log "IMPORTANT: Change password on first login!"
    log ""
    log "Check status: systemctl status backup-server"
    log ""
}

main "$@"
