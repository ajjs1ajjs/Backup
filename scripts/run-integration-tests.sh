#!/bin/bash
# Integration Test Runner Script for Backup System
# Usage: ./run-integration-tests.sh [-t <All|Unit|Integration|Stress>] [-v]

set -e

# Colors
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

# Default values
TEST_TYPE="All"
VERBOSE=""
SKIP_DOCKER=false
POSTGRES_HOST="localhost"
POSTGRES_PORT="5432"

# Parse arguments
while getopts "t:vhs" opt; do
    case $opt in
        t) TEST_TYPE="$OPTARG" ;;
        v) VERBOSE="--logger \"console;verbosity=detailed\"" ;;
        h) 
            echo "Usage: $0 [-t <All|Unit|Integration|Stress>] [-v] [-s]"
            echo "  -t  Test type: All, Unit, Integration, Stress (default: All)"
            echo "  -v  Verbose output"
            echo "  -s  Skip Docker"
            exit 0
            ;;
        s) SKIP_DOCKER=true ;;
        \?) echo "Invalid option -$OPTARG"; exit 1 ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_PATH="$(dirname "$SCRIPT_DIR")"
SERVER_PATH="$ROOT_PATH/src/server/Backup.Server"
TEST_PROJECT_PATH="$ROOT_PATH/src/server/Backup.Server.IntegrationTests"
UNIT_TEST_PATH="$ROOT_PATH/src/server/Backup.Server.Tests"

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  Backup System Integration Test Runner${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

write_step() {
    echo -e "\n${YELLOW}► $1${NC}"
}

write_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

write_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check prerequisites
write_step "Checking prerequisites..."

if ! command -v dotnet &> /dev/null; then
    write_error ".NET SDK not found. Please install .NET 8.0 SDK."
    exit 1
fi
DOTNET_VERSION=$(dotnet --version)
write_success ".NET SDK $DOTNET_VERSION found"

if [ "$SKIP_DOCKER" = false ]; then
    if ! command -v docker &> /dev/null; then
        echo -e "${YELLOW}⚠ Docker not found. Some tests may be skipped.${NC}"
    else
        write_success "Docker found"
    fi
fi

# Start PostgreSQL if using Docker
if [ "$SKIP_DOCKER" = false ]; then
    write_step "Starting PostgreSQL test container..."
    
    EXISTING_CONTAINER=$(docker ps -a --filter "name=backup-integration-test-db" --format "{{.Names}}")
    
    if [ -n "$EXISTING_CONTAINER" ]; then
        echo -e "${GRAY}Container already exists. Starting...${NC}"
        docker start backup-integration-test-db > /dev/null 2>&1
    else
        docker run -d \
            --name backup-integration-test-db \
            -p ${POSTGRES_PORT}:5432 \
            -e POSTGRES_DB=backup_test \
            -e POSTGRES_USER=postgres \
            -e POSTGRES_PASSWORD=postgres \
            postgres:14
        
        write_success "PostgreSQL container started"
    fi
    
    echo -e "${GRAY}Waiting for PostgreSQL to be ready...${NC}"
    sleep 5
    
    # Check if PostgreSQL is ready
    MAX_RETRIES=10
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if docker exec backup-integration-test-db pg_isready -U postgres 2>/dev/null | grep -q "accepting connections"; then
            write_success "PostgreSQL is ready"
            break
        else
            echo -e "${GRAY}Waiting for PostgreSQL... ($((RETRY_COUNT + 1))/$MAX_RETRIES)${NC}"
            sleep 2
            RETRY_COUNT=$((RETRY_COUNT + 1))
        fi
    done
    
    if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
        write_error "PostgreSQL failed to start. Tests may fail."
    fi
fi

# Run Unit Tests
if [ "$TEST_TYPE" = "All" ] || [ "$TEST_TYPE" = "Unit" ]; then
    write_step "Running Unit Tests..."
    
    cd "$UNIT_TEST_PATH"
    
    if [ -n "$VERBOSE" ]; then
        dotnet test --logger "console;verbosity=detailed" --results-directory "$ROOT_PATH/test-results/unit"
    else
        dotnet test --logger "console;verbosity=normal" --results-directory "$ROOT_PATH/test-results/unit"
    fi
    
    if [ $? -eq 0 ]; then
        write_success "Unit Tests passed"
    else
        write_error "Unit Tests failed"
    fi
fi

# Run Integration Tests
if [ "$TEST_TYPE" = "All" ] || [ "$TEST_TYPE" = "Integration" ]; then
    write_step "Running Integration Tests..."
    
    cd "$TEST_PROJECT_PATH"
    
    if [ -n "$VERBOSE" ]; then
        dotnet test --logger "console;verbosity=detailed" --results-directory "$ROOT_PATH/test-results/integration"
    else
        dotnet test --logger "console;verbosity=normal" --results-directory "$ROOT_PATH/test-results/integration"
    fi
    
    if [ $? -eq 0 ]; then
        write_success "Integration Tests passed"
    else
        write_error "Integration Tests failed"
    fi
fi

# Run Stress Tests (via API)
if [ "$TEST_TYPE" = "All" ] || [ "$TEST_TYPE" = "Stress" ]; then
    write_step "Running Stress Tests..."
    
    # Check if server is running
    if curl -s --connect-timeout 5 http://localhost:8080/health > /dev/null 2>&1; then
        write_success "Backup server is running"
        
        echo -e "${GRAY}Running stress test for 100 VMs...${NC}"
        
        RESULT=$(curl -s -X POST http://localhost:8080/api/stresstest/run \
            -H "Content-Type: application/json" \
            -d '{"vmCount":100,"concurrentCount":50}')
        
        if [ $? -eq 0 ]; then
            write_success "Stress Test completed"
            
            # Parse JSON result (requires jq)
            if command -v jq &> /dev/null; then
                echo -e "  ${CYAN}Total Backups:$(echo $RESULT | jq -r '.totalBackups')${NC}"
                echo -e "  ${GREEN}Successful: $(echo $RESULT | jq -r '.successfulBackups')${NC}"
                FAILED=$(echo $RESULT | jq -r '.failedBackups')
                if [ "$FAILED" -gt 0 ]; then
                    echo -e "  ${RED}Failed: $FAILED${NC}"
                else
                    echo -e "  ${GRAY}Failed: $FAILED${NC}"
                fi
                echo -e "  ${CYAN}Average Duration: $(echo $RESULT | jq -r '.averageDurationMs') ms${NC}"
                echo -e "  ${CYAN}P95 Duration: $(echo $RESULT | jq -r '.percentile95DurationMs') ms${NC}"
            else
                echo -e "${GRAY}Install 'jq' for formatted JSON output${NC}"
                echo "  Raw result: $RESULT"
            fi
        else
            write_error "Stress Test failed"
        fi
    else
        echo -e "${YELLOW}⚠ Backup server not running.${NC}"
        echo -e "${GRAY}   Please run: cd $SERVER_PATH && dotnet run${NC}"
        echo -e "${GRAY}   Skipping stress tests...${NC}"
    fi
fi

# Summary
echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  Test Run Summary${NC}"
echo -e "${CYAN}========================================${NC}"

if [ -d "$ROOT_PATH/test-results" ]; then
    echo ""
    echo -e "Test results saved to:"
    echo -e "  ${GRAY}$ROOT_PATH/test-results${NC}"
fi

echo ""
echo -e "To view detailed results:"
echo -e "  ${GRAY}dotnet test --logger \"trx\" --results-directory test-results${NC}"

echo ""
echo -e "To clean up Docker container:"
echo -e "  ${GRAY}docker stop backup-integration-test-db${NC}"
echo -e "  ${GRAY}docker rm backup-integration-test-db${NC}"

echo ""
write_success "Test run completed!"
echo ""
