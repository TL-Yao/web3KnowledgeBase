#!/bin/bash

# Web3-Insight 服务状态检查脚本

echo "=========================================="
echo "  Web3-Insight 服务状态"
echo "=========================================="
echo ""

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_service() {
    local name=$1
    local port=$2
    local url=$3

    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "  $name: ${GREEN}运行中${NC} (端口 $port)"
        if [ -n "$url" ]; then
            echo -e "         $url"
        fi
    else
        echo -e "  $name: ${RED}未运行${NC}"
    fi
}

check_docker() {
    local name=$1
    local container=$2

    if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${container}$"; then
        echo -e "  $name: ${GREEN}运行中${NC} (Docker: $container)"
    else
        echo -e "  $name: ${RED}未运行${NC}"
    fi
}

echo "基础服务:"
check_docker "PostgreSQL" "web3insight-postgres"
check_docker "Redis" "web3insight-redis"
echo ""

echo "应用服务:"
check_service "Backend API" 8080 "http://localhost:8080/health"
check_service "Frontend" 3000 "http://localhost:3000"
echo ""

echo "可选服务:"
check_service "Ollama" 11434 "http://localhost:11434"
echo ""

# 显示日志路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$(dirname "$SCRIPT_DIR")/logs"

if [ -d "$LOG_DIR" ]; then
    echo "日志文件:"
    for log in "$LOG_DIR"/*.log; do
        if [ -f "$log" ]; then
            echo "  - $(basename "$log")"
        fi
    done
    echo ""
    echo "查看日志: tail -f $LOG_DIR/<service>.log"
fi
