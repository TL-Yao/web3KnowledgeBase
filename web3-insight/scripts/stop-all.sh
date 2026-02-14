#!/bin/bash

# Web3-Insight 全栈停止脚本
# 停止所有服务：Frontend, Worker, Backend API, Redis, PostgreSQL

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

echo "=========================================="
echo "  Web3-Insight 全栈停止"
echo "=========================================="

# 停止进程（通过 PID 文件）
stop_by_pid() {
    local name=$1
    local pid_file="$LOG_DIR/$2.pid"

    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            log_info "停止 $name (PID: $pid)..."
            kill $pid 2>/dev/null
            sleep 1
            if ps -p $pid > /dev/null 2>&1; then
                kill -9 $pid 2>/dev/null
            fi
        fi
        rm -f "$pid_file"
    fi
}

# 停止监听特定端口的进程
stop_by_port() {
    local name=$1
    local port=$2

    local pids=$(lsof -ti :$port 2>/dev/null)
    if [ -n "$pids" ]; then
        log_info "停止 $name (端口 $port)..."
        echo "$pids" | xargs kill 2>/dev/null
        sleep 1
        pids=$(lsof -ti :$port 2>/dev/null)
        if [ -n "$pids" ]; then
            echo "$pids" | xargs kill -9 2>/dev/null
        fi
    fi
}

# 1. 停止前端
log_info "停止前端..."
stop_by_pid "Frontend" "frontend"
stop_by_port "Frontend" 3000

# 2. 停止 Worker
log_info "停止 Worker..."
stop_by_pid "Worker" "worker"

# 3. 停止后端
log_info "停止后端 API..."
stop_by_pid "Backend" "backend"
stop_by_port "Backend" 8080

# 4. 停止 Ollama
log_info "停止 Ollama..."
stop_by_pid "Ollama" "ollama"
pkill -f "ollama serve" 2>/dev/null || true

# 5. 停止 Docker 服务 (PostgreSQL + Redis)
log_info "停止 Docker 服务..."
docker-compose -f "$COMPOSE_FILE" down

echo ""
echo "=========================================="
echo "  所有服务已停止"
echo "=========================================="
echo ""
echo "提示: Docker volumes 保留数据"
echo ""
echo "如需完全清理（包括数据），运行:"
echo "  docker-compose -f $COMPOSE_FILE down -v"
