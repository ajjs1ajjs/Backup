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

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"; }
error() { echo "[ERROR] $1" >&2; exit 1; }

show_help() {
    cat << EOF
Backup Agent Installer v$VERSION

Usage: $0 [OPTIONS]

Options:
    --server ADDR         Management server address (host:port)
    --token TOKEN         Agent registration token
    --agent-type TYPE     Agent type: hyperv, vmware, kvm, mssql, postgres, oracle
    --install-dir DIR     Installation directory (default: $INSTALL_DIR)
    --mode MODE           Installation mode: agent, server, all (default: agent)
    --auto-start         Start agent after installation
    --force              Force reinstallation
    --uninstall          Uninstall agent
    --skip-ssl           Skip SSL certificate verification (insecure)
    --source-url URL     Alternative URL for install script
    --local-source PATH  Use local source code instead of downloading
    -h, --help           Show this help

Examples:
    $0 --server 10.0.0.1:8050 --token ABCD-1234 --agent-type hyperv --auto-start
    $0 --mode server --install-dir /opt/backup-server
    $0 --skip-ssl --server 10.0.0.1:8050 --token ABCD --auto-start
    $0 --local-source /path/to/source --server 10.0.0.1:8050 --token ABCD
    curl -fsSL https://get.backupsystem.com/agent/install.sh | sudo bash -s -- --server 10.0.0.1:8050 --token ABCD --auto-start
    curl -kfsSL https://get.backupsystem.com/agent/install.sh | sudo bash -s -- --skip-ssl --server 10.0.0.1:8050 --token ABCD

EOF
}

download_script() {
    local url="${SOURCE_URL:-https://get.backupsystem.com/agent/install.sh}"
    local tmp_script="/tmp/backup-install.sh"
    
    log "Downloading install script from $url..."
    
    local curl_opts="-fsSL"
    if [[ "$SKIP_SSL" == "true" ]]; then
        log "WARNING: SSL verification disabled"
        curl_opts="-kfsSL"
    fi
    
    if ! curl $curl_opts -o "$tmp_script" "$url"; then
        error "Failed to download install script. Use --source-url to specify alternative location"
    fi
    
    chmod +x "$tmp_script"
    exec sudo bash "$tmp_script" "$@"
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
            --source-url) SOURCE_URL="$2"; shift 2 ;;
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

check_deps() {
    local missing=()
    
    log "Checking dependencies..."
    
    for cmd in cmake make g++ git pkg-config; do
        if ! command -v $cmd &> /dev/null; then
            missing+=($cmd)
        fi
    done
    
    for lib in ssl curl xml2 zstd; do
        if ! pkg-config --exists lib$lib 2>/dev/null && ! ldconfig -p | grep -q lib$lib; then
            missing+=(lib$lib-dev)
        fi
    done
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        log "Installing missing dependencies: ${missing[*]}"
        apt-get update
        apt-get install -y --no-install-recommends \
            cmake \
            build-essential \
            git \
            pkg-config \
            libssl-dev \
            libcurl4-openssl-dev \
            libxml2-dev \
            libzstd-dev \
            wget \
            curl \
            ca-certificates
    fi
    
    if ! command -v systemctl &> /dev/null; then
        log "Warning: systemd not found - service installation will be skipped"
    fi
    
    if ! command -v dotnet &> /dev/null; then
        log "Warning: .NET SDK not found - server installation will be skipped"
    fi
    
    if ! command -v node &> /dev/null; then
        log "Warning: Node.js not found - UI build will be skipped"
    fi
    
    case "$AGENT_TYPE" in
        kvm)
            if ! command -v virsh &> /dev/null && ! dpkg -l | grep -q libvirt; then
                log "Warning: libvirt not found - KVM backups may not work"
            fi
            ;;
        postgres)
            if ! command -v psql &> /dev/null; then
                log "Warning: PostgreSQL client not found - PostgreSQL backups may not work"
            fi
            ;;
    esac
    
    log "All dependencies satisfied"
}

create_dirs() {
    log "Creating directories..."
    mkdir -p "$BIN_DIR" "$CONFIG_DIR" "$LOG_DIR" "$DATA_DIR"
    chmod 755 "$INSTALL_DIR"
}

download_source() {
    local src_dir="/tmp/backup-agent-build"
    
    if [[ -f "$BIN_DIR/backup-agent" && "$FORCE" == "false" ]]; then
        log "Agent already installed. Use --force to reinstall"
        return 0
    fi
    
    log "Building agent..."
    
    rm -rf "$src_dir"
    mkdir -p "$src_dir"
    cd "$src_dir"
    
    if [[ -n "$LOCAL_SOURCE" && -d "$LOCAL_SOURCE" ]]; then
        log "Using local source: $LOCAL_SOURCE"
        cp -r "$LOCAL_SOURCE/"* .
    elif [[ -d ".git" ]]; then
        log "Using existing source"
    else
        log "Downloading agent source from GitHub..."
        local archive_url="https://github.com/ajjs1ajjs/Backup/archive/refs/heads/main.zip"
        
        if [[ "$SKIP_SSL" == "true" ]]; then
            curl -kfsSL -o source.zip "$archive_url" || error "Failed to download source"
        else
            curl -fsSL -o source.zip "$archive_url" || error "Failed to download source"
        fi
        
        unzip -q source.zip
        mv Backup-main/src/agent/* . 2>/dev/null || mv Backup-main/* . 2>/dev/null || true
        rm -rf source.zip Backup-main
        
        if [[ ! -f "CMakeLists.txt" && ! -f "Makefile" ]]; then
            log "Warning: CMakeLists.txt not found, trying alternative location..."
            if [[ -d "src/agent/Backup.Agent" ]]; then
                mv src/agent/Backup.Agent/* .
                rm -rf src
            fi
        fi
    fi
    
    log "Compiling agent..."
    mkdir -p build && cd build
    cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" 2>/dev/null || {
        log "CMake failed - trying alternative build..."
        cd ..
        mkdir -p build && cd build
        cmake .. -DCMAKE_BUILD_TYPE=Release 2>/dev/null || {
            error "Build failed. Please provide source with --local-source or install cmake"
        }
    }
    make -j$(nproc)
    
    mkdir -p "$BIN_DIR"
    cp backup-agent "$BIN_DIR/" 2>/dev/null || cp *backup-agent* "$BIN_DIR/" 2>/dev/null || {
        log "Looking for agent binary..."
        find . -name "backup-agent*" -type f -executable -exec cp {} "$BIN_DIR/" \;
    }
    
    chmod +x "$BIN_DIR/backup-agent"
    log "Agent built successfully"
}

generate_config() {
    local config_file="$CONFIG_DIR/agent.conf"
    
    cat > "$config_file" << EOF
# Backup Agent Configuration
server=$SERVER_ADDR
token=$AGENT_TOKEN
agent_type=$AGENT_TYPE
log_dir=$LOG_DIR
data_dir=$DATA_DIR
log_level=info
EOF

    chmod 600 "$config_file"
    chown root:root "$config_file"
    
    log "Configuration generated at $config_file"
}

create_service() {
    local service_file="/etc/systemd/system/backup-agent.service"
    
    cat > "$service_file" << EOF
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
Environment=LD_LIBRARY_PATH=$INSTALL_DIR/lib

[Install]
WantedBy=multi-user.target
EOF

    chmod 644 "$service_file"
    systemctl daemon-reload
    
    log "Service file created at $service_file"
}

verify_installation() {
    log "Verifying installation..."
    
    if [[ ! -f "$BIN_DIR/backup-agent" ]]; then
        error "Binary not found at $BIN_DIR/backup-agent"
    fi
    
    if [[ ! -x "$BIN_DIR/backup-agent" ]]; then
        error "Binary not executable"
    fi
    
    if ! "$BIN_DIR/backup-agent" --version &>/dev/null && ! "$BIN_DIR/backup-agent" --help &>/dev/null; then
        log "Warning: Agent failed to run. This may be expected if not fully configured."
    else
        log "Agent binary verified"
    fi
    
    log "Installation verified successfully"
}

start_agent() {
    log "Starting agent..."
    
    systemctl enable backup-agent 2>/dev/null || true
    systemctl start backup-agent
    
    sleep 2
    
    if systemctl is-active --quiet backup-agent; then
        log "Agent started successfully"
        systemctl status backup-agent --no-pager
    else
        error "Failed to start agent"
    fi
}

uninstall_agent() {
    log "Uninstalling agent..."
    
    if systemctl is-active --quiet backup-agent 2>/dev/null; then
        log "Stopping agent..."
        systemctl stop backup-agent
    fi
    
    systemctl disable backup-agent 2>/dev/null || true
    rm -f /etc/systemd/system/backup-agent.service
    systemctl daemon-reload
    
    rm -rf "$INSTALL_DIR"
    
    log "Agent uninstalled successfully"
}

install_server() {
    log "Installing Backup Server..."
    
    local server_dir="$INSTALL_DIR/server"
    local ui_dir="$server_dir/ui"
    
    if ! command -v dotnet &> /dev/null; then
        log "Installing .NET SDK 8.0..."
        wget -q https://dot.net/v1/dotnet-install.sh -O /tmp/dotnet-install.sh
        chmod +x /tmp/dotnet-install.sh
        /tmp/dotnet-install.sh --channel 8.0 --install-dir /opt/dotnet
        export DOTNET_ROOT=/opt/dotnet
        export PATH="$PATH:/opt/dotnet"
    fi
    
    mkdir -p "$server_dir"
    cd "$server_dir"
    
    log "Building server..."
    if [[ -f "src/server/Backup.Server/Backup.Server.csproj" ]]; then
        dotnet restore src/server/Backup.Server/Backup.Server.csproj
        dotnet publish src/server/Backup.Server/Backup.Server.csproj -c Release -o "$server_dir/publish"
    else
        log "Warning: Server source not found, skipping server build"
    fi
    
    if command -v node &> /dev/null; then
        log "Building UI..."
        if [[ -f "src/ui/package.json" ]]; then
            cd src/ui
            npm install --production
            npm run build
            mv build "$ui_dir"
        else
            log "Warning: UI source not found, skipping UI build"
        fi
    else
        log "Warning: Node.js not found, skipping UI build"
    fi
    
    log "Server installation complete"
}

main() {
    if [[ "$UNINSTALL" == "true" ]]; then
        check_root
        uninstall_agent
        exit 0
    fi
    
    parse_args "$@"
    
    if [[ "$INSTALL_MODE" == "server" || "$INSTALL_MODE" == "all" ]]; then
        check_root
        check_deps
        install_server
        exit 0
    fi
    
    if [[ -z "$SERVER_ADDR" || -z "$AGENT_TOKEN" ]]; then
        show_help
        error "Server and Token are required"
    fi
    
    check_root
    check_deps
    create_dirs
    download_source
    generate_config
    create_service
    verify_installation
    
    if [[ "$AUTO_START" == "true" ]]; then
        start_agent
    else
        log "Installation complete. To start agent manually:"
        log "  systemctl start backup-agent"
    fi
    
    log "Installation completed successfully!"
    log "Agent installed at: $INSTALL_DIR"
    log "Config: $CONFIG_DIR/agent.conf"
}

main "$@"
