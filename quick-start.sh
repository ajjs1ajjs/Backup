#!/bin/bash
# NovaBackup Quick Start Script for Linux/macOS
# This script provides quick setup for different scenarios

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get script directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCENARIO="${1:-}"

# Usage function
usage() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "  NovaBackup Quick Start"
    echo -e "========================================${NC}"
    echo ""
    echo "Usage: $0 <scenario>"
    echo ""
    echo "Scenarios:"
    echo "  dev    - Development environment (SQLite, MOCK cloud)"
    echo "  prod   - Production environment (PostgreSQL, real cloud)"
    echo "  docker - Docker Compose deployment"
    echo "  test   - Testing environment with dev tools"
    echo ""
    echo "Options:"
    echo "  --skip-secrets  - Skip secret generation"
    echo "  --no-prompt     - Non-interactive mode"
    echo "  -h, --help      - Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 dev"
    echo "  $0 prod --skip-secrets"
    echo "  $0 docker --no-prompt"
    echo ""
    exit 1
}

# Parse arguments
SKIP_SECRETS=false
NO_PROMPT=false

while [[ $# -gt 0 ]]; do
    case $1 in
        dev|prod|docker|test)
            SCENARIO="$1"
            shift
            ;;
        --skip-secrets)
            SKIP_SECRETS=true
            shift
            ;;
        --no-prompt)
            NO_PROMPT=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
    esac
done

# Check if scenario is provided
if [ -z "$SCENARIO" ]; then
    usage
fi

echo -e "${CYAN}========================================${NC}"
echo -e "  ${GREEN}NovaBackup Quick Start${NC}"
echo -e "  ${YELLOW}Scenario: $SCENARIO${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

# Function to check prerequisites
check_prerequisites() {
    echo -e "${GREEN}[1/6] Checking prerequisites...${NC}"
    
    # Check Python
    if command -v python3 &> /dev/null; then
        PYTHON_CMD="python3"
    elif command -v python &> /dev/null; then
        PYTHON_CMD="python"
    else
        echo -e "${RED}  ✗ Python not found${NC}"
        echo -e "${YELLOW}  Please install Python 3.9+${NC}"
        exit 1
    fi
    
    PYTHON_VERSION=$($PYTHON_CMD --version 2>&1)
    echo -e "${GREEN}  ✓ Python found: $PYTHON_VERSION${NC}"
    
    # Check version >= 3.9
    VERSION=$($PYTHON_CMD -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
    MAJOR=$(echo $VERSION | cut -d. -f1)
    MINOR=$(echo $VERSION | cut -d. -f2)
    
    if [ "$MAJOR" -lt 3 ] || ([ "$MAJOR" -eq 3 ] && [ "$MINOR" -lt 9 ]); then
        echo -e "${RED}  ✗ Python 3.9+ required, found $VERSION${NC}"
        exit 1
    fi
    
    # Check Git
    if command -v git &> /dev/null; then
        GIT_VERSION=$(git --version 2>&1)
        echo -e "${GREEN}  ✓ Git found: $GIT_VERSION${NC}"
    else
        echo -e "${YELLOW}  ⚠ Git not found (optional for existing projects)${NC}"
    fi
    
    # Check Docker (for docker scenario)
    if [ "$SCENARIO" = "docker" ]; then
        if command -v docker &> /dev/null; then
            DOCKER_VERSION=$(docker --version 2>&1)
            echo -e "${GREEN}  ✓ Docker found: $DOCKER_VERSION${NC}"
            
            COMPOSE_VERSION=$(docker compose version 2>&1)
            echo -e "${GREEN}  ✓ Docker Compose found: $COMPOSE_VERSION${NC}"
        else
            echo -e "${RED}  ✗ Docker not found${NC}"
            echo -e "${YELLOW}  Please install Docker Desktop${NC}"
            exit 1
        fi
    fi
    
    echo ""
}

# Function to create virtual environment
create_venv() {
    echo -e "${GREEN}[2/6] Setting up virtual environment...${NC}"
    
    if [ -d "$PROJECT_ROOT/venv" ]; then
        echo -e "${YELLOW}  ℹ Virtual environment already exists${NC}"
    else
        $PYTHON_CMD -m venv venv
        echo -e "${GREEN}  ✓ Virtual environment created${NC}"
    fi
    
    # Activate
    source "$PROJECT_ROOT/venv/bin/activate"
    echo -e "${GREEN}  ✓ Virtual environment activated${NC}"
    echo ""
}

# Function to install dependencies
install_dependencies() {
    echo -e "${GREEN}[3/6] Installing dependencies...${NC}"
    
    # Upgrade pip
    pip install --upgrade pip --quiet
    
    if [ "$SCENARIO" = "test" ]; then
        pip install -e ".[api,db,dev]" --quiet
        echo -e "${GREEN}  ✓ Dependencies installed (including dev tools)${NC}"
    elif [ "$SCENARIO" = "docker" ]; then
        echo -e "${YELLOW}  ℹ Skipping pip install (using Docker)${NC}"
    else
        pip install -e ".[api,db]" --quiet
        echo -e "${GREEN}  ✓ Dependencies installed${NC}"
    fi
    
    echo ""
}

# Function to configure environment
configure_environment() {
    echo -e "${GREEN}[4/6] Configuring environment...${NC}"
    
    ENV_FILE="$PROJECT_ROOT/.env"
    
    if [ -f "$ENV_FILE" ]; then
        echo -e "${YELLOW}  ℹ .env file already exists${NC}"
        if [ "$NO_PROMPT" = false ]; then
            read -p "Overwrite existing .env? (y/n) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                echo -e "${YELLOW}  ℹ Keeping existing .env${NC}"
                echo ""
                return
            fi
        fi
    fi
    
    # Copy template
    cp "$PROJECT_ROOT/.env.example" "$ENV_FILE"
    echo -e "${GREEN}  ✓ Created .env from template${NC}"
    
    # Generate secrets
    if [ "$SKIP_SECRETS" = true ]; then
        echo -e "${YELLOW}  ⚠ Skipping secret generation (use ./generate-secrets.sh -All manually)${NC}"
    else
        if [ -f "$PROJECT_ROOT/generate-secrets.sh" ]; then
            bash "$PROJECT_ROOT/generate-secrets.sh" --all
            echo -e "${GREEN}  ✓ Secrets generated${NC}"
        elif [ -f "$PROJECT_ROOT/generate-secrets.ps1" ]; then
            echo -e "${YELLOW}  ⚠ PowerShell script found, please run generate-secrets.ps1 manually${NC}"
        else
            # Generate secrets inline
            echo -e "${YELLOW}  Generating secrets...${NC}"
            
            MASTER_KEY=$(openssl rand -hex 32)
            JWT_SECRET=$(openssl rand -hex 32)
            API_KEY=$(openssl rand -hex 16)
            
            sed -i.bak "s|NOVABACKUP_MASTER_KEY=.*|NOVABACKUP_MASTER_KEY=$MASTER_KEY|" "$ENV_FILE"
            sed -i.bak "s|NOVABACKUP_JWT_SECRET=.*|NOVABACKUP_JWT_SECRET=$JWT_SECRET|" "$ENV_FILE"
            sed -i.bak "s|NOVABACKUP_API_KEY=.*|NOVABACKUP_API_KEY=$API_KEY|" "$ENV_FILE"
            rm "$ENV_FILE.bak" 2>/dev/null || true
            
            echo -e "${GREEN}  ✓ Secrets generated${NC}"
        fi
    fi
    
    # Configure based on scenario
    echo -e "${GREEN}  Configuring for $SCENARIO scenario...${NC}"
    
    case $SCENARIO in
        dev)
            sed -i.bak "s|NOVABACKUP_DATABASE_URL=.*|NOVABACKUP_DATABASE_URL=sqlite:///./novabackup.db|" "$ENV_FILE"
            sed -i.bak "s|NOVABACKUP_CLOUD_PROVIDERS=.*|NOVABACKUP_CLOUD_PROVIDERS=MOCK|" "$ENV_FILE"
            sed -i.bak "s|NOVABACKUP_DEBUG=.*|NOVABACKUP_DEBUG=true|" "$ENV_FILE"
            sed -i.bak "s|NOVABACKUP_PORT=.*|NOVABACKUP_PORT=8000|" "$ENV_FILE"
            rm "$ENV_FILE.bak" 2>/dev/null || true
            ;;
        prod)
            sed -i.bak "s|NOVABACKUP_DEBUG=.*|NOVABACKUP_DEBUG=false|" "$ENV_FILE"
            sed -i.bak "s|NOVABACKUP_PORT=.*|NOVABACKUP_PORT=8050|" "$ENV_FILE"
            echo -e "${RED}  ⚠ Production config created - UPDATE SECRETS AND DATABASE URL!${NC}"
            ;;
        test)
            sed -i.bak "s|NOVABACKUP_DATABASE_URL=.*|NOVABACKUP_DATABASE_URL=sqlite:///:memory:|" "$ENV_FILE"
            sed -i.bak "s|NOVABACKUP_CLOUD_PROVIDERS=.*|NOVABACKUP_CLOUD_PROVIDERS=MOCK|" "$ENV_FILE"
            sed -i.bak "s|NOVABACKUP_DEBUG=.*|NOVABACKUP_DEBUG=true|" "$ENV_FILE"
            sed -i.bak "s|NOVABACKUP_TESTING_MODE=.*|NOVABACKUP_TESTING_MODE=true|" "$ENV_FILE"
            rm "$ENV_FILE.bak" 2>/dev/null || true
            ;;
    esac
    
    echo -e "${GREEN}  ✓ Environment configured for $SCENARIO${NC}"
    echo ""
}

# Function to run database migrations
run_migrations() {
    echo -e "${GREEN}[5/6] Running database migrations...${NC}"
    
    if [ "$SCENARIO" = "docker" ]; then
        echo -e "${YELLOW}  ℹ Migrations will run automatically in Docker${NC}"
    else
        if $PYTHON_CMD -m novabackup.migrate 2>/dev/null; then
            echo -e "${GREEN}  ✓ Database migrations completed${NC}"
        else
            echo -e "${YELLOW}  ⚠ Migration script not found or failed${NC}"
            echo -e "${YELLOW}  ℹ Database will initialize on first run${NC}"
        fi
    fi
    
    echo ""
}

# Function to start the application
start_application() {
    echo -e "${GREEN}[6/6] Starting NovaBackup...${NC}"
    
    if [ "$SCENARIO" = "docker" ]; then
        echo -e "${GREEN}  Starting Docker Compose...${NC}"
        docker-compose -f docker-compose-prod.yml up -d
        echo -e "${GREEN}  ✓ Docker containers started${NC}"
        echo ""
        echo -e "${CYAN}  Access points:${NC}"
        echo -e "${WHITE}  - API:       http://localhost:8000${NC}"
        echo -e "${WHITE}  - Dashboard: http://localhost:8080${NC}"
        echo -e "${WHITE}  - Docs:      http://localhost:8000/docs${NC}"
        echo ""
        echo -e "${YELLOW}  View logs: docker-compose logs -f${NC}"
        echo -e "${YELLOW}  Stop:      docker-compose down${NC}"
    elif [ "$SCENARIO" = "test" ]; then
        echo -e "${GREEN}  Running tests...${NC}"
        pytest tests/ -v
        echo -e "${GREEN}  ✓ Tests completed${NC}"
    else
        echo -e "${GREEN}  Starting development server...${NC}"
        echo ""
        echo -e "${CYAN}  Access points:${NC}"
        echo -e "  - API:       http://localhost:8000"
        echo -e "  - Dashboard: http://localhost:8000/static/index.html"
        echo -e "  - Docs:      http://localhost:8000/docs"
        echo ""
        echo -e "${YELLOW}  Default credentials:${NC}"
        echo -e "  - Username: alice"
        echo -e "  - Password: secret"
        echo ""
        echo -e "${YELLOW}  Press Ctrl+C to stop the server${NC}"
        echo ""
        
        python -m uvicorn novabackup.api:get_app --reload --host 0.0.0.0 --port 8000
    fi
}

# Main execution
check_prerequisites

if [ "$SCENARIO" != "docker" ]; then
    create_venv
    install_dependencies
    configure_environment
    run_migrations
fi

start_application

echo -e "${CYAN}========================================${NC}"
echo -e "  ${GREEN}Setup Complete!${NC}"
echo -e "${CYAN}========================================${NC}"
