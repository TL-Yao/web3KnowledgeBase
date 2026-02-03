#!/bin/bash

# Web3-Insight 全栈启动脚本
# 启动所有服务：PostgreSQL, Redis, Backend API, Worker, Frontend

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_ROOT/logs"

# 创建日志目录
mkdir -p "$LOG_DIR"

echo "=========================================="
echo "  Web3-Insight 全栈启动"
echo "=========================================="

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查端口是否被占用
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0  # 端口被占用
    else
        return 1  # 端口空闲
    fi
}

# 1. 启动 PostgreSQL (使用 Docker)
start_postgres() {
    log_info "检查 PostgreSQL..."
    if check_port 5432; then
        log_info "PostgreSQL 已在运行 (端口 5432)"
    else
        log_info "启动 PostgreSQL..."
        docker run -d --name web3insight-postgres \
            -e POSTGRES_USER=postgres \
            -e POSTGRES_PASSWORD=postgres \
            -e POSTGRES_DB=web3insight \
            -p 5432:5432 \
            -v web3insight-pgdata:/var/lib/postgresql/data \
            pgvector/pgvector:pg16 \
            2>/dev/null || docker start web3insight-postgres 2>/dev/null

        # 等待 PostgreSQL 就绪
        log_info "等待 PostgreSQL 就绪..."
        sleep 3
        for i in {1..30}; do
            if docker exec web3insight-postgres pg_isready -U postgres >/dev/null 2>&1; then
                log_info "PostgreSQL 就绪"
                break
            fi
            sleep 1
        done
    fi
}

# 2. 启动 Redis (使用 Docker)
start_redis() {
    log_info "检查 Redis..."
    if check_port 6379; then
        log_info "Redis 已在运行 (端口 6379)"
    else
        log_info "启动 Redis..."
        docker run -d --name web3insight-redis \
            -p 6379:6379 \
            redis:alpine \
            2>/dev/null || docker start web3insight-redis 2>/dev/null
        sleep 2
        log_info "Redis 已启动"
    fi
}

# 3. 启动 Ollama (可选)
start_ollama() {
    log_info "检查 Ollama..."
    if check_port 11434; then
        log_info "Ollama 已在运行 (端口 11434)"
    else
        if command -v ollama &> /dev/null; then
            log_info "启动 Ollama..."
            ollama serve > "$LOG_DIR/ollama.log" 2>&1 &
            echo $! > "$LOG_DIR/ollama.pid"
            sleep 2
            log_info "Ollama 已启动"
        else
            log_warn "Ollama 未安装，跳过 (LLM 功能将不可用)"
        fi
    fi
}

# 4. 启动后端 API
start_backend() {
    log_info "启动后端 API..."
    cd "$PROJECT_ROOT/backend"

    # 使用绝对路径的 Go
    GO_BIN="/usr/local/go/bin/go"

    if check_port 8080; then
        log_warn "端口 8080 已被占用，跳过后端启动"
    else
        $GO_BIN run cmd/server/main.go > "$LOG_DIR/backend.log" 2>&1 &
        echo $! > "$LOG_DIR/backend.pid"
        sleep 3

        if check_port 8080; then
            log_info "后端 API 已启动 (端口 8080)"
        else
            log_error "后端 API 启动失败，查看日志: $LOG_DIR/backend.log"
        fi
    fi
}

# 5. 启动 Worker
start_worker() {
    log_info "启动 Worker..."
    cd "$PROJECT_ROOT/backend"

    GO_BIN="/usr/local/go/bin/go"

    $GO_BIN run cmd/worker/main.go > "$LOG_DIR/worker.log" 2>&1 &
    echo $! > "$LOG_DIR/worker.pid"
    sleep 2
    log_info "Worker 已启动"
}

# 6. 启动前端
start_frontend() {
    log_info "启动前端..."
    cd "$PROJECT_ROOT/frontend"

    if check_port 3000; then
        log_warn "端口 3000 已被占用，跳过前端启动"
    else
        # 检查 node_modules
        if [ ! -d "node_modules" ]; then
            log_info "安装前端依赖..."
            pnpm install
        fi

        pnpm dev > "$LOG_DIR/frontend.log" 2>&1 &
        echo $! > "$LOG_DIR/frontend.pid"
        sleep 3

        if check_port 3000; then
            log_info "前端已启动 (端口 3000)"
        else
            log_error "前端启动失败，查看日志: $LOG_DIR/frontend.log"
        fi
    fi
}

# 执行启动
start_postgres
start_redis
start_ollama
start_backend
start_worker
start_frontend

echo ""
echo "=========================================="
echo "  启动完成！"
echo "=========================================="
echo ""
echo "服务状态:"
echo "  - PostgreSQL: http://localhost:5432"
echo "  - Redis:      http://localhost:6379"
echo "  - Backend:    http://localhost:8080"
echo "  - Frontend:   http://localhost:3000"
echo "  - Ollama:     http://localhost:11434"
echo ""
echo "日志目录: $LOG_DIR"
echo ""
echo "停止所有服务: ./scripts/stop-all.sh"
