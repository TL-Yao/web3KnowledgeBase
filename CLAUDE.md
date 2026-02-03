# Project Guidelines

## Go Build Commands

This environment has GVM (Go Version Manager) configured in shell profile, which causes `cd` commands to fail with:
```
cd:1: command not found: __gvm_is_function
ERROR: GVM_ROOT not set. Please source $GVM_ROOT/scripts/gvm
```

**Solution**: Use absolute path to Go binary with `-C` flag instead of `cd`:

```bash
# WRONG - do not use cd
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend && go build ./...

# CORRECT - use absolute Go path with -C flag
/usr/local/go/bin/go build -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./...
/usr/local/go/bin/go get -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend <package>
/usr/local/go/bin/go mod tidy -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend
```

The `-C` flag changes to the specified directory before executing the command, avoiding the shell `cd` issue.

## Project Structure

- `web3-insight/` - Main project directory
  - `backend/` - Go backend (Gin, GORM, Asynq)
  - `frontend/` - Next.js frontend with shadcn/ui
  - `docs/plans/` - Implementation plans

## Docker Services (Start/Stop)

The project uses Docker for PostgreSQL and Redis services:

```bash
# Start database services (PostgreSQL + Redis)
docker-compose -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/docker-compose.yml up -d postgres redis

# Stop all services
docker-compose -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/docker-compose.yml down

# Check running containers
docker ps | grep web3-insight
```

**Service Details:**
- PostgreSQL: `pgvector/pgvector:pg16` on port 5432
  - Container: `web3-insight-db`
  - User: `web3insight`, Password: `web3insight_dev`, DB: `web3insight`
- Redis: `redis:7-alpine` on port 6379
  - Container: `web3-insight-redis`

**Makefile Commands (run from web3-insight directory):**
- `make db-up` - Start database services
- `make db-down` - Stop database services
- `make dev` - Start full dev environment (db + backend + frontend)
- `make dev-backend` - Start backend only
- `make dev-frontend` - Start frontend only
- `make migrate` - Run database migrations
- `make worker` - Start async task worker

**Note**: Due to GVM shell issues, use absolute paths for running Go commands directly (see Go Build Commands section above).

## Service Startup and Shutdown

### Full Stack Startup (All Services)

**Start all services:**
```bash
# Using shell script (recommended)
./web3-insight/scripts/start-all.sh

# Creates logs in web3-insight/logs/
# - backend.log, frontend.log, worker.log, ollama.log
# - PID files: backend.pid, frontend.pid, worker.pid, ollama.pid
```

**Services started:**
1. PostgreSQL (Docker) - port 5432
2. Redis (Docker) - port 6379
3. Ollama (optional) - port 11434
4. Backend API - port 8080
5. Worker (async tasks)
6. Frontend - port 3000

**Check service status:**
```bash
./web3-insight/scripts/status.sh
```

**Stop all services:**
```bash
./web3-insight/scripts/stop-all.sh
```

### Individual Service Control

**Database Services Only:**
```bash
# Start
docker-compose -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/docker-compose.yml up -d postgres redis

# Stop (preserves data)
docker-compose -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/docker-compose.yml down

# Completely remove (deletes all data)
docker-compose -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/docker-compose.yml down -v
```

**Backend API Only:**
```bash
# Development mode (from web3-insight directory)
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend cmd/server/main.go

# Or using Makefile
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight
make dev-backend
```

**Worker Only:**
```bash
# Development mode (from web3-insight directory)
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend cmd/worker/main.go

# Or using Makefile
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight
make worker
```

**Frontend Only:**
```bash
# Development mode (from web3-insight directory)
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend
npm run dev

# Or using Makefile
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight
make dev-frontend
```

### Database Management

**Run migrations:**
```bash
# From web3-insight directory
make migrate

# Or directly
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend cmd/migrate/main.go
```

**Clear all data (preserves schema):**
```bash
# Connect to database
docker exec -it web3-insight-db psql -U web3insight -d web3insight

# Run the clear script
\i /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend/scripts/clear_data.sql

# Or copy-paste SQL from backend/scripts/clear_data.sql
```

### Build Commands

**Build backend binaries:**
```bash
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight
make build-backend

# Creates:
# - backend/bin/server
# - backend/bin/worker
```

**Build frontend:**
```bash
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight
make build-frontend

# Creates: frontend/.next/
```

**Build both:**
```bash
make build
```

### Testing

**Run backend tests:**
```bash
make test-backend
# Or: /usr/local/go/bin/go test -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./...
```

**Run frontend tests:**
```bash
make test-frontend
# Or: cd frontend && npm test
```

**Run E2E tests:**
```bash
cd frontend && npm run test:e2e
```

### Log Files

When using `scripts/start-all.sh`, logs are stored in:
- `web3-insight/logs/backend.log` - Backend API logs
- `web3-insight/logs/frontend.log` - Frontend logs
- `web3-insight/logs/worker.log` - Worker logs
- `web3-insight/logs/ollama.log` - Ollama logs (if installed)

**View live logs:**
```bash
tail -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/logs/backend.log
```

### Configuration Updates

**IMPORTANT:** When project configuration, services, or startup procedures change:
1. Update this CLAUDE.md file with the new information
2. Update scripts in `web3-insight/scripts/` if needed
3. Update Makefile if adding new commands
4. Commit changes so documentation stays current
