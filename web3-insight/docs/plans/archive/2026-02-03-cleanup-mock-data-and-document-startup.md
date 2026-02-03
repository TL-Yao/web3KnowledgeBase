# Cleanup Mock Data and Document Startup Procedures

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Document all service startup/shutdown procedures in CLAUDE.md and remove all mock/fake data from backend and frontend

**Architecture:** Two-phase approach: First update documentation, then remove all seeding and mock data initialization code

**Tech Stack:** Go (backend), Next.js (frontend), Docker (PostgreSQL, Redis), Shell scripts

---

## Task 1: Update CLAUDE.md with Complete Startup Documentation

**Files:**
- Modify: `/Users/tongleyao/claudeProjects/explorerResearch/CLAUDE.md`

**Step 1: Read current CLAUDE.md content**

Already read - current content includes Docker services and Makefile commands but missing detailed script information.

**Step 2: Add comprehensive startup/shutdown documentation**

Add the following sections to CLAUDE.md after the existing content:

```markdown
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
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend
/usr/local/go/bin/go run cmd/server/main.go

# Or using Makefile
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight
make dev-backend
```

**Worker Only:**
```bash
# Development mode (from web3-insight directory)
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend
/usr/local/go/bin/go run cmd/worker/main.go

# Or using Makefile
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight
make worker
```

**Frontend Only:**
```bash
# Development mode (from web3-insight directory)
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend
pnpm dev

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
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend
/usr/local/go/bin/go run cmd/migrate/main.go
```

**Clear all data (preserves schema):**
```bash
# Connect to database
docker exec -it web3-insight-db psql -U web3insight -d web3insight

# Run the clear script
\i /path/to/backend/scripts/clear_data.sql

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
# Or: cd backend && /usr/local/go/bin/go test ./...
```

**Run frontend tests:**
```bash
make test-frontend
# Or: cd frontend && pnpm test
```

**Run E2E tests:**
```bash
cd frontend && pnpm test:e2e
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
```

**Step 3: Commit the documentation update**

```bash
git add CLAUDE.md
git commit -m "docs: add comprehensive startup/shutdown procedures to CLAUDE.md"
```

---

## Task 2: Remove Backend Seed Data

**Files:**
- Modify: `/Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend/cmd/server/main.go:30-35`
- Modify: `/Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend/internal/database/seed.go`

**Step 1: Remove seed call from main.go**

In `backend/cmd/server/main.go`, remove lines that call `database.Seed(db)`:

```go
// OLD - Remove these lines (30-35)
	// Seed initial data
	if err := database.Seed(db); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}
	log.Println("Seed data loaded")
```

**Step 2: Empty the Seed function**

In `backend/internal/database/seed.go`, replace the entire function body with a no-op:

```go
package database

import (
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	// Seed function intentionally empty - no mock data
	// Use admin UI or API to create categories and content
	return nil
}
```

**Step 3: Verify backend starts without seeding**

Run: `/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend cmd/server/main.go`

Expected: Server starts successfully without "Seed data loaded" message

**Step 4: Commit backend changes**

```bash
git add backend/cmd/server/main.go backend/internal/database/seed.go
git commit -m "feat: remove automatic seed data from backend startup"
```

---

## Task 3: Remove Frontend Mock Data for Tests

**Files:**
- Modify: `/Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend/__mocks__/data.ts`
- Modify: `/Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend/__mocks__/handlers.ts`

**Step 1: Replace mock data with empty arrays**

In `frontend/__mocks__/data.ts`:

```typescript
// __mocks__/data.ts
import type { Article, Category, DataSource } from '@/lib/api'

// Empty mock data - tests should provide their own fixtures
export const mockCategories: Category[] = []

export const mockArticles: Article[] = []

export const mockDataSources: DataSource[] = []
```

**Step 2: Update handlers to return empty data**

In `frontend/__mocks__/handlers.ts`, keep the same structure but return empty arrays:

```typescript
// __mocks__/handlers.ts
import { http, HttpResponse } from 'msw'
import { mockArticles, mockCategories, mockDataSources } from './data'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export const handlers = [
  // Categories - return empty arrays by default
  http.get(`${API_BASE}/api/categories`, () => {
    return HttpResponse.json([])
  }),

  http.get(`${API_BASE}/api/categories/tree`, () => {
    return HttpResponse.json([])
  }),

  // Articles - return empty results
  http.get(`${API_BASE}/api/articles`, ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    const pageSize = parseInt(url.searchParams.get('page_size') || '10')

    return HttpResponse.json({
      articles: [],
      total: 0,
      page,
      pageSize,
    })
  }),

  http.get(`${API_BASE}/api/articles/:slug`, () => {
    return new HttpResponse(null, { status: 404 })
  }),

  // Data Sources - return empty array
  http.get(`${API_BASE}/api/sources`, () => {
    return HttpResponse.json([])
  }),

  http.post(`${API_BASE}/api/sources/:id/sync`, () => {
    return HttpResponse.json({ message: 'Sync started', itemsFound: 0, itemsNew: 0 })
  }),

  // Search - return empty results
  http.get(`${API_BASE}/api/search`, () => {
    return HttpResponse.json({
      articles: [],
      total: 0,
    })
  }),
]
```

**Step 3: Run frontend tests to verify they still work**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && pnpm test:run`

Expected: Tests pass (some may need updating to provide their own fixtures)

**Step 4: Commit frontend mock data removal**

```bash
git add frontend/__mocks__/data.ts frontend/__mocks__/handlers.ts
git commit -m "feat: remove default mock data from MSW handlers"
```

---

## Task 4: Verify Empty State

**Files:** N/A (verification only)

**Step 1: Start fresh database**

```bash
# Stop all services
./web3-insight/scripts/stop-all.sh

# Remove database volume to start fresh
docker-compose -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/docker-compose.yml down -v

# Start database services
docker-compose -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/docker-compose.yml up -d postgres redis
```

**Step 2: Run migrations only (no seed)**

```bash
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend
/usr/local/go/bin/go run cmd/migrate/main.go
```

Expected: Migrations complete, tables created, but no data inserted

**Step 3: Start backend and verify empty**

```bash
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend cmd/server/main.go
```

Expected: Server starts, no "Seed data loaded" message

**Step 4: Check API returns empty data**

```bash
# Check categories
curl http://localhost:8080/api/categories

# Check articles
curl http://localhost:8080/api/articles
```

Expected: Both return empty arrays or empty paginated results

**Step 5: Start frontend and verify empty state**

```bash
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend
pnpm dev
```

Visit http://localhost:3000/knowledge - should show empty state (no articles, no categories)

**Step 6: Document verification results**

No files to modify - just confirm everything is empty and working.

---

## Task 5: Final Integration Test

**Files:** N/A (testing only)

**Step 1: Stop all services**

```bash
./web3-insight/scripts/stop-all.sh
```

**Step 2: Start all services with start-all.sh**

```bash
./web3-insight/scripts/start-all.sh
```

Expected: All services start successfully

**Step 3: Check service status**

```bash
./web3-insight/scripts/status.sh
```

Expected: All services show as running

**Step 4: Verify each service**

- PostgreSQL: `docker exec -it web3-insight-db psql -U web3insight -d web3insight -c "SELECT COUNT(*) FROM categories;"`
  - Expected: 0 rows
- Backend: `curl http://localhost:8080/health`
  - Expected: 200 OK
- Frontend: Visit http://localhost:3000
  - Expected: Page loads with empty state

**Step 5: Create final commit**

```bash
git status
# Verify all changes are committed
```

---

## Summary

This plan:
1. Documents all startup/shutdown procedures in CLAUDE.md
2. Removes seed data from backend startup
3. Empties mock data in frontend test mocks
4. Verifies system works with empty state
5. Tests full integration with scripts

After completion, the system will:
- Have comprehensive startup documentation
- Start with completely empty database (only schema)
- Show empty state in UI until data is added via admin or API
- Keep test infrastructure (MSW) but without default mock data
