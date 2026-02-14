#!/bin/bash

# Web3-Insight 全栈启动脚本
# 启动所有服务：PostgreSQL, Redis, Backend API, Worker, Frontend

set -e

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/lib.sh"

mkdir -p "$LOG_DIR"

echo "=========================================="
echo "  Web3-Insight 全栈启动"
echo "=========================================="

# 1. 启动 PostgreSQL + Redis (Docker Compose)
start_docker_services() {
    log_info "启动 Docker 服务 (PostgreSQL + Redis)..."
    if check_port 5432 && check_port 6379; then
        log_info "PostgreSQL 和 Redis 已在运行"
    else
        docker-compose -f "$COMPOSE_FILE" up -d postgres redis
        log_info "等待 Docker 服务就绪..."
        for i in {1..30}; do
            if docker-compose -f "$COMPOSE_FILE" exec -T postgres pg_isready -U web3insight >/dev/null 2>&1; then
                log_info "PostgreSQL 就绪"
                break
            fi
            sleep 1
        done
        log_info "Redis 已启动"
    fi
}

# 2. 启动 Ollama (可选)
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

# 3. 启动后端 API
start_backend() {
    log_info "启动后端 API..."
    GO_BIN="/usr/local/go/bin/go"

    if check_port 8080; then
        log_warn "端口 8080 已被占用，跳过后端启动"
    else
        $GO_BIN run -C "$PROJECT_ROOT/backend" cmd/server/main.go > "$LOG_DIR/backend.log" 2>&1 &
        echo $! > "$LOG_DIR/backend.pid"
        sleep 3

        if check_port 8080; then
            log_info "后端 API 已启动 (端口 8080)"
        else
            log_error "后端 API 启动失败，查看日志: $LOG_DIR/backend.log"
        fi
    fi
}

# 4. 启动 Worker
start_worker() {
    log_info "启动 Worker..."
    GO_BIN="/usr/local/go/bin/go"

    $GO_BIN run -C "$PROJECT_ROOT/backend" cmd/worker/main.go > "$LOG_DIR/worker.log" 2>&1 &
    echo $! > "$LOG_DIR/worker.pid"
    sleep 2
    log_info "Worker 已启动"
}

# 5. 启动前端
start_frontend() {
    log_info "启动前端..."

    if check_port 3000; then
        log_warn "端口 3000 已被占用，跳过前端启动"
    else
        if [ ! -d "$PROJECT_ROOT/frontend/node_modules" ]; then
            log_info "安装前端依赖..."
            npm --prefix "$PROJECT_ROOT/frontend" install
        fi

        npm --prefix "$PROJECT_ROOT/frontend" run dev > "$LOG_DIR/frontend.log" 2>&1 &
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
start_docker_services
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
