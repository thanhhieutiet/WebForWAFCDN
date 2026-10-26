#!/usr/bin/env bash
# ==============================================================================
# CMC Telecom LabShop - Automated Load & Security Test Suite Controller
# Supports: Performance Benchmark (4 Cases), Multi-ISP Routing Test, WAF Defense Test
# Usage:
#   ./run_test_suite.sh [mode] [target_host] [workers]
# Example:
#   ./run_test_suite.sh perf https://app.thanhhieutiet.shop 4
#   ./run_test_suite.sh security https://app.thanhhieutiet.shop 4
#   ./run_test_suite.sh network https://app.thanhhieutiet.shop 2
#   ./run_test_suite.sh down
# ==============================================================================

set -e

# Default settings
MODE="${1:-interactive}"
HOST="${2:-https://app.thanhhieutiet.shop}"
WORKERS="${3:-4}"

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

print_header() {
    echo -e "${CYAN}====================================================================${NC}"
    echo -e "${BOLD}${GREEN}  CMC TELECOM LABSHOP - DISTRIBUTED LOAD TESTING CONTROLLER  ${NC}"
    echo -e "${CYAN}====================================================================${NC}"
}

print_menu() {
    print_header
    echo -e "${BOLD}Select Test Mode to Execute:${NC}"
    echo -e "  ${GREEN}1)${NC} ${BOLD}Performance Benchmark (4 Cases: Direct, WAF, CDN, Full)${NC}"
    echo -e "     - TEST_MODE=performance | USE_RANGE_HEADER=true (Anti-Bottleneck)"
    echo -e "     - Recommended Load: 300 - 500 CCU | Ramp-up: 20 | Duration: 5 mins"
    echo -e "     - WAF probe disabled (weight=0) to prevent IP rate-limit blast radius."
    echo ""
    echo -e "  ${BLUE}2)${NC} ${BOLD}Multi-ISP / Multi-Network Routing Test (Office vs Home vs 4G)${NC}"
    echo -e "     - TEST_MODE=performance | USE_RANGE_HEADER=true"
    echo -e "     - Recommended Load: 30 CCU | Ramp-up: 10 | Duration: 2 mins"
    echo -e "     - Measures network latency, TTL & ISP jitter without saturating uplink."
    echo ""
    echo -e "  ${YELLOW}3)${NC} ${BOLD}WAF Defense & Anti-Brute-Force Security Test${NC}"
    echo -e "     - TEST_MODE=security | WafRateLimitProbeUser only"
    echo -e "     - Recommended Load: 100 CCU | Ramp-up: 20 | Duration: 2 mins"
    echo -e "     - Tests SQLi block (403) and login rate-limiting (429)."
    echo ""
    echo -e "  ${CYAN}4)${NC} ${BOLD}Scale Locust Workers${NC} (Current default: $WORKERS workers)"
    echo -e "  ${RED}5)${NC} ${BOLD}Stop All Load Testing Services (docker compose down)${NC}"
    echo -e "  ${BOLD}0)${NC} Exit"
    echo -e "${CYAN}--------------------------------------------------------------------${NC}"
    read -rp "Enter choice [1-5]: " CHOICE
}

start_stack() {
    local test_mode="$1"
    local use_range="$2"
    local target_host="$3"
    local worker_count="$4"

    echo -e "${YELLOW}>> Configuring Environment:${NC}"
    echo -e "   - TEST_MODE=${BOLD}${test_mode}${NC}"
    echo -e "   - USE_RANGE_HEADER=${BOLD}${use_range}${NC}"
    echo -e "   - TARGET_HOST=${BOLD}${target_host}${NC}"
    echo -e "   - WORKERS=${BOLD}${worker_count} replicas${NC}"

    export TEST_MODE="$test_mode"
    export USE_RANGE_HEADER="$use_range"
    export LOCUST_HOST="$target_host"

    echo -e "${BLUE}>> Starting Locust Master-Worker cluster and Telemetry pipeline...${NC}"
    docker compose up -d --scale locust-worker="$worker_count"

    echo ""
    echo -e "${GREEN}====================================================================${NC}"
    echo -e "${BOLD}  LOAD TEST CLUSTER IS READY & RUNNING!  ${NC}"
    echo -e "${GREEN}====================================================================${NC}"
    echo -e "  * ${BOLD}Locust Web UI:${NC}         ${CYAN}http://localhost:8089${NC}"
    echo -e "  * ${BOLD}Grafana Dashboard:${NC}     ${CYAN}http://localhost:3000${NC} (admin / admin)"
    echo -e "  * ${BOLD}Prometheus Targets:${NC}    ${CYAN}http://localhost:9090/targets${NC}"
    echo -e "  * ${BOLD}Locust Exporter:${NC}       ${CYAN}http://localhost:9646/metrics${NC}"
    echo -e "${GREEN}--------------------------------------------------------------------${NC}"
    echo -e ">> Open ${CYAN}http://localhost:8089${NC} in your browser, enter CCU & Ramp-up, and click ${BOLD}'Start swarming'${NC}!"
    echo -e ">> Watch live real-time metrics at ${CYAN}http://localhost:3000${NC}"
}

# If interactive mode, show menu
if [ "$MODE" = "interactive" ]; then
    print_menu
    case "$CHOICE" in
        1)
            read -rp "Enter target host [$HOST]: " INPUT_HOST
            TARGET="${INPUT_HOST:-$HOST}"
            start_stack "performance" "true" "$TARGET" "$WORKERS"
            ;;
        2)
            read -rp "Enter target host [$HOST]: " INPUT_HOST
            TARGET="${INPUT_HOST:-$HOST}"
            start_stack "performance" "true" "$TARGET" "2"
            ;;
        3)
            read -rp "Enter target host [$HOST]: " INPUT_HOST
            TARGET="${INPUT_HOST:-$HOST}"
            start_stack "security" "false" "$TARGET" "$WORKERS"
            ;;
        4)
            read -rp "Enter desired number of Locust workers: " NEW_WORKERS
            docker compose up -d --scale locust-worker="$NEW_WORKERS"
            echo -e "${GREEN}Scaled locust-worker to $NEW_WORKERS instances.${NC}"
            ;;
        5)
            echo -e "${YELLOW}Stopping all services...${NC}"
            docker compose down
            echo -e "${GREEN}All services stopped successfully.${NC}"
            ;;
        *)
            echo "Exiting."
            exit 0
            ;;
    esac
elif [ "$MODE" = "perf" ] || [ "$MODE" = "performance" ]; then
    start_stack "performance" "true" "$HOST" "$WORKERS"
elif [ "$MODE" = "network" ] || [ "$MODE" = "multi-isp" ]; then
    start_stack "performance" "true" "$HOST" "2"
elif [ "$MODE" = "security" ] || [ "$MODE" = "waf" ]; then
    start_stack "security" "false" "$HOST" "$WORKERS"
elif [ "$MODE" = "down" ] || [ "$MODE" = "stop" ]; then
    echo -e "${YELLOW}Stopping all services...${NC}"
    docker compose down
    echo -e "${GREEN}All services stopped.${NC}"
elif [ "$MODE" = "status" ] || [ "$MODE" = "ps" ]; then
    docker compose ps
else
    echo -e "${RED}Unknown mode: $MODE${NC}"
    echo "Usage: ./run_test_suite.sh [perf|network|security|down|status] [target_host] [workers]"
    exit 1
fi
