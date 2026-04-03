#!/bin/bash

set -e

VERSION="1.0.0"
INSTALL_DIR="/opt/backup-agent"
BIN_DIR="$INSTALL_DIR/bin"
CONFIG_DIR="$INSTALL_DIR/config"
LOG_DIR="$INSTALL_DIR/log"
DATA_DIR="$INSTALL_DIR/data"

SERVER_ADDR=""
AGENT_TOKEN=""
AGENT_TYPE="hyperv"
AUTO_START=false
FORCE=false
INSTALL_MODE="agent"
SKIP_SSL=false
SOURCE_URL=""
LOCAL_SOURCE=""
REPO_URL="https://github.com/ajjs1ajjs/Backup.git"
BUILD_DIR="/tmp/backup-build"

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"; }
error() { echo "[ERROR] $1" >&2; exit 1; }

show_help() {
    cat << EOF
Backup System Installer v$VERSION

Usage: $0 [OPTIONS]

Options:
    --server ADDR         Management server address (host:port)
    --token TOKEN         Agent registration token
    --agent-type TYPE     Agent type: hyperv, vmware, kvm, mssql, postgres, oracle
    --install-dir DIR     Installation directory (default: $INSTALL_DIR)
    --mode MODE           Installation mode: agent, server, all (default: agent)
    --auto-start         Start service after installation
    --force              Force reinstallation
    --uninstall          Uninstall
    --skip-ssl           Skip SSL certificate verification
    --local-source PATH  Use local source code
    -h, --help           Show this help

Examples:
    # Install agent only
    $0 --server localhost:8000 --token ABCD-1234 --agent-type hyperv --auto-start

    # Install server only
    $0 --mode server --auto-start

    # Install both server and agent
    $0 --mode all --server localhost:8000 --token ABCD-1234 --agent-type hyperv --auto-start

EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --server) SERVER_ADDR="$2"; shift 2 ;;
            --token) AGENT_TOKEN="$2"; shift 2 ;;
            --agent-type) AGENT_TYPE="$2"; shift 2 ;;
            --install-dir) INSTALL_DIR="$2"; shift 2 ;;
            --mode) INSTALL_MODE="$2"; shift 2 ;;
            --auto-start) AUTO_START=true; shift ;;
            --force) FORCE=true; shift ;;
            --uninstall) UNINSTALL=true; shift ;;
            --skip-ssl) SKIP_SSL=true; shift ;;
            --local-source) LOCAL_SOURCE="$2"; shift 2 ;;
            -h|--help) show_help; exit 0 ;;
            *) error "Unknown option: $1" ;;
        esac
    done
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root (use sudo)"
    fi
}

install_dotnet() {
    if command -v dotnet &> /dev/null; then
        log ".NET SDK already installed: $(dotnet --version)"
        return 0
    fi

    log "Installing .NET SDK 8.0..."
    wget -q https://dot.net/v1/dotnet-install.sh -O /tmp/dotnet-install.sh
    chmod +x /tmp/dotnet-install.sh
    /tmp/dotnet-install.sh --channel 8.0 --install-dir /opt/dotnet
    export DOTNET_ROOT=/opt/dotnet
    export PATH="$PATH:/opt/dotnet"
    ln -sf /opt/dotnet/dotnet /usr/local/bin/dotnet
    log ".NET SDK installed successfully"
}

clone_repo() {
    if [[ -n "$LOCAL_SOURCE" && -d "$LOCAL_SOURCE" ]]; then
        log "Using local source: $LOCAL_SOURCE"
        rm -rf "$BUILD_DIR"
        cp -r "$LOCAL_SOURCE" "$BUILD_DIR"
        return 0
    fi

    if [[ -d "$BUILD_DIR/.git" && "$FORCE" == "false" ]]; then
        log "Updating existing repository..."
        cd "$BUILD_DIR"
        git pull || true
        return 0
    fi

    log "Cloning repository..."
    rm -rf "$BUILD_DIR"
    git clone "$REPO_URL" "$BUILD_DIR"
}

install_server() {
    log "Installing Backup Server..."

    install_dotnet

    local server_src="$BUILD_DIR/src/server/Backup.Server"
    local server_install="/opt/backup-server"

    if [[ ! -f "$server_src/Backup.Server.csproj" ]]; then
        error "Server source not found at $server_src"
    fi

    log "Building server..."
    cd "$server_src"
    dotnet restore
    dotnet publish -c Release -o "$server_install/publish"

    # Create service file
    cat > /etc/systemd/system/backup-server.service << EOF
[Unit]
Description=Backup Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$server_install/publish
ExecStart=/opt/dotnet/dotnet $server_install/publish/Backup.Server.dll --urls=http://localhost:8000
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment=ASPNETCORE_ENVIRONMENT=Production
Environment=Jwt__Key=CHANGE_ME_TO_A_STRONG_SECRET_KEY_12345
Environment=Server__PublicUrl=http://localhost:8000

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload

    if [[ "$AUTO_START" == "true" ]]; then
        log "Starting server..."
        systemctl enable backup-server
        systemctl start backup-server
        sleep 5

        if systemctl is-active --quiet backup-server; then
            log "Server started successfully on http://localhost:8000"
        else
            log "Warning: Server failed to start. Check logs: journalctl -u backup-server"
        fi
    fi

    log "Server installed at: $server_install"
}

install_agent() {
    log "Installing Backup Agent..."

    # Check dependencies
    local missing=()
    for cmd in cmake make g++ git; do
        if ! command -v $cmd &> /dev/null; then
            missing+=($cmd)
        fi
    done

    if [[ ${#missing[@]} -gt 0 ]]; then
        log "Installing build dependencies: ${missing[*]}"
        apt-get update -qq
        apt-get install -y -qq cmake build-essential git pkg-config libssl-dev libcurl4-openssl-dev
    fi

    # Create directories
    mkdir -p "$BIN_DIR" "$CONFIG_DIR" "$LOG_DIR" "$DATA_DIR"

    # Build agent from source
    local agent_src="$BUILD_DIR/src/agent/Backup.Agent"

    if [[ -f "$agent_src/CMakeLists.txt" ]]; then
        log "Building agent from source..."
        cd "$agent_src"
        mkdir -p build && cd build
        cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR"
        make -j$(nproc)
        cp backup-agent "$BIN_DIR/" 2>/dev/null || find . -name "backup-agent*" -type f -executable -exec cp {} "$BIN_DIR/" \;
        chmod +x "$BIN_DIR/backup-agent"
    else
        log "Warning: Agent source not found, skipping agent build"
    fi

    # Generate config
    if [[ -n "$SERVER_ADDR" && -n "$AGENT_TOKEN" ]]; then
        cat > "$CONFIG_DIR/agent.conf" << EOF
# Backup Agent Configuration
server=$SERVER_ADDR
token=$AGENT_TOKEN
agent_type=$AGENT_TYPE
log_dir=$LOG_DIR
data_dir=$DATA_DIR
log_level=info
EOF
        chmod 600 "$CONFIG_DIR/agent.conf"
        log "Configuration generated at $CONFIG_DIR/agent.conf"
    fi

    # Create service
    cat > /etc/systemd/system/backup-agent.service << EOF
[Unit]
Description=Backup Agent Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$BIN_DIR/backup-agent --config $CONFIG_DIR/agent.conf
Restart=on-failure
RestartSec=10
StandardOutput=append:$LOG_DIR/agent.log
StandardError=append:$LOG_DIR/agent.log

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload

    if [[ "$AUTO_START" == "true" ]]; then
        log "Starting agent..."
        systemctl enable backup-agent
        systemctl start backup-agent
        sleep 2

        if systemctl is-active --quiet backup-agent; then
            log "Agent started successfully"
        else
            log "Warning: Agent failed to start. Check logs: journalctl -u backup-agent"
        fi
    fi

    log "Agent installed at: $INSTALL_DIR"
}

uninstall() {
    log "Uninstalling..."

    # Stop and remove server
    if systemctl is-active --quiet backup-server 2>/dev/null; then
        systemctl stop backup-server
    fi
    systemctl disable backup-server 2>/dev/null || true
    rm -f /etc/systemd/system/backup-server.service

    # Stop and remove agent
    if systemctl is-active --quiet backup-agent 2>/dev/null; then
        systemctl stop backup-agent
    fi
    systemctl disable backup-agent 2>/dev/null || true
    rm -f /etc/systemd/system/backup-agent.service

    systemctl daemon-reload

    # Remove files
    rm -rf /opt/backup-server
    rm -rf "$INSTALL_DIR"
    rm -rf "$BUILD_DIR"

    log "Uninstall complete"
}

main() {
    parse_args "$@"

    if [[ "$UNINSTALL" == "true" ]]; then
        check_root
        uninstall
        exit 0
    fi

    check_root

    # Clone repository first
    clone_repo

    case "$INSTALL_MODE" in
        server)
            install_server
            ;;
        all)
            install_server
            install_agent
            ;;
        agent|*)
            if [[ -z "$SERVER_ADDR" || -z "$AGENT_TOKEN" ]]; then
                show_help
                error "--server and --token are required for agent installation"
            fi
            install_agent
            ;;
    esac

    log ""
    log "========================================="
    log "Installation completed successfully!"
    log "========================================="

    if [[ "$INSTALL_MODE" == "server" || "$INSTALL_MODE" == "all" ]]; then
        log "Server: http://localhost:8000"
        log "Swagger: http://localhost:8000/swagger"
        log "Check status: systemctl status backup-server"
    fi

    if [[ "$INSTALL_MODE" == "agent" || "$INSTALL_MODE" == "all" ]]; then
        log "Agent: $INSTALL_DIR"
        log "Config: $CONFIG_DIR/agent.conf"
        log "Check status: systemctl status backup-agent"
    fi
}

main "$@"
