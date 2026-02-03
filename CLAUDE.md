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
cd /path/to/backend && go build ./...

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
- `make seed` - Seed database with initial data
- `make worker` - Start async task worker

**Note**: Due to GVM shell issues, use absolute paths for running Go commands directly (see Go Build Commands section above).
