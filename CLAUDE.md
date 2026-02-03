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

## Browser Testing

Claude Code has Chrome plugin capability enabled for visual verification of frontend changes.

**When to use:**
- After implementing UI changes to verify actual browser appearance
- When debugging layout or styling issues
- To confirm empty states display correctly
- To validate user interactions and component behavior

**How to use:**
```bash
# Open frontend in Chrome (home page)
open -a "Google Chrome" http://localhost:3000

# Open specific pages
open -a "Google Chrome" http://localhost:3000/admin
open -a "Google Chrome" http://localhost:3000/knowledge
open -a "Google Chrome" http://localhost:3000/research
```

**Best practices:**
- Use this capability when implementing or reviewing UI changes
- Verify both desktop and responsive views if needed
- Check console for JavaScript errors after opening
- Confirm loading states, empty states, and error states display properly

**Note:** This supplements (not replaces) automated testing. Use for visual verification and UX validation.

### Chrome Automation Capabilities

**IMPORTANT:** Claude Code has full Chrome browser automation capabilities through the `claude-in-chrome` extension. This is a powerful tool that should be used proactively during development.

**Available capabilities:**

1. **Navigation and Page Control**
   - Navigate to URLs
   - Go forward/back in browser history
   - Refresh pages
   - Create new tabs
   - Resize browser windows

2. **Page Interaction**
   - Click elements (buttons, links, etc.)
   - Fill form inputs
   - Type text
   - Scroll pages
   - Drag and drop
   - Hover over elements

3. **Content Reading and Analysis**
   - Read page accessibility tree
   - Extract page text content
   - Find elements using natural language queries
   - Take screenshots (full page or specific regions)
   - Read console messages (logs, errors, warnings)
   - Read network requests (XHR, Fetch, API calls)

4. **Advanced Automation**
   - Execute JavaScript code in page context
   - Upload images to forms or drag-drop targets
   - Record GIF animations of browser interactions
   - Execute shortcuts and workflows

**When to use Chrome automation:**
- Verify frontend changes visually after implementation
- Debug UI issues by inspecting actual rendered elements
- Test user interaction flows (form submissions, navigation, etc.)
- Capture console errors or network issues
- Validate API responses in browser DevTools
- Create visual documentation (screenshots, GIFs) of features
- Test responsive design by resizing windows
- Verify empty states, loading states, and error states
- Check that JavaScript executes correctly
- Monitor performance and network activity

**Best practices:**
- Always start by calling `tabs_context_mcp` to get current browser state
- Create new tabs for each testing session to avoid conflicts
- Use natural language queries with `find` tool to locate elements
- Take screenshots before and after interactions to document changes
- Read console messages with pattern filters to avoid noise
- Record GIFs when demonstrating multi-step interactions
- Clean up tabs after testing to maintain organized workspace

**Example workflows:**

```bash
# 1. Start services
./web3-insight/scripts/start-all.sh

# 2. Use Chrome automation to:
#    - Open http://localhost:3000 in new tab
#    - Navigate through pages
#    - Fill forms and submit
#    - Take screenshots of results
#    - Read console for errors
#    - Inspect network requests
#    - Verify API responses
```

**Important notes:**
- Chrome automation is available in ALL Claude Code sessions
- This is NOT just opening Chrome - it's full programmatic control
- Use this proactively without waiting for user requests
- Prefer automated verification over manual "please check" requests
- Screenshots and console logs provide definitive evidence
- GIF recordings are excellent for documenting complex interactions

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

**Migrations:**
Migrations run automatically when the backend server starts. There is no separate migration command.

The backend automatically:
1. Connects to the database
2. Runs all pending migrations via GORM AutoMigrate
3. Creates/updates tables as needed
4. Then starts the API server

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

## Plan Completion Protocol

**IMPORTANT:** After completing any implementation plan, follow this protocol to maintain clean documentation and capture learnings.

### Step 1: Clean Up Plan Files

After a plan in `docs/plans/` is fully implemented and verified:

1. **Review the plan file** - Read through to identify what was completed
2. **Archive or remove** - Decide whether to:
   - Delete the plan file if it's no longer needed
   - Move it to `docs/plans/archive/` if it has historical value
   - Keep it if it documents ongoing work

### Step 2: Document Completion Summary

Add a brief summary to this CLAUDE.md file under the appropriate section:

**Location:** Add summaries to a new "Implementation History" section below

**Format:**
```markdown
### [Date] - [Feature/Task Name]

**What was completed:**
- Brief bullet points of main changes
- Key files modified
- New functionality added

**Important takeaways:**
- Lessons learned
- Patterns established
- Things to remember for future development
- Gotchas or issues encountered and solved

**Related commits:** [commit range or key commits]
```

### Step 3: Update Project Structure (if needed)

If the implementation changed:
- Project structure or organization
- Service configurations
- Startup/shutdown procedures
- Environment requirements

→ Update the relevant sections of this CLAUDE.md file

---

## Implementation History

### 2026-02-03 - Mock Data Removal and Startup Documentation

**What was completed:**
- Added comprehensive startup/shutdown procedures to CLAUDE.md
- Removed all backend seed data (categories, articles)
- Removed all frontend mock data from MSW handlers
- Verified system works with empty state
- Full integration test passed

**Important takeaways:**
- **Migrations are automatic**: No separate migration command exists. The backend runs `database.Migrate(db)` automatically on startup via `cmd/server/main.go`. This was incorrectly documented initially and had to be corrected.
- **GVM issues**: The environment has GVM configured which causes `cd` commands to fail. Always use `/usr/local/go/bin/go run -C /path/to/directory` instead of `cd /path && go run`.
- **Package manager**: The frontend uses `npm` (not `pnpm`). Check `package.json` scripts when documenting commands.
- **Test fixtures**: After removing default mock data, tests that relied on it failed (expected). Tests should provide their own fixtures rather than depending on global mocks.
- **Docker service detection**: The `status.sh` script has a bug where it incorrectly reports PostgreSQL and Redis as "not running" even when they are. Services are confirmed via `docker ps` and port checks.

**Related commits:** `1d7d98f` through `1b8a8b1` (9 commits total)

**Files modified:**
- CLAUDE.md (documentation)
- backend/cmd/server/main.go (removed seed call)
- backend/internal/database/seed.go (emptied function)
- frontend/__mocks__/data.ts (emptied mock data)
- frontend/__mocks__/handlers.ts (return empty data)
- Makefile (removed seed and migrate targets)

**Key patterns established:**
- All documentation goes in CLAUDE.md for easy reference
- Use absolute paths in documentation for copy-paste reliability
- Verify functionality before documenting it
- Keep test infrastructure but remove default test data

### 2026-02-03 - Remove Hardcoded UI Data (Follow-up)

**What was completed:**
- Removed hardcoded admin page stats (12 articles, 847 API calls, $0.32 cost)
- Removed hardcoded task monitor mock data (3 sample tasks)
- Replaced with proper API queries returning empty defaults
- Added empty state UI to task monitor component
- Added TypeScript interfaces, error handling, and loading states

**Important takeaways:**
- **Frontend mock data locations**: Mock data can hide in multiple places:
  1. `__mocks__/` directory (MSW handlers) - covered in previous cleanup
  2. Component-level hardcoded values (admin page stats)
  3. Query function return values (task-monitor.tsx)
- Always check both page files and component files for hardcoded data
- Empty states should be explicit and user-friendly, not just blank screens
- Use `?? 0` pattern for safe default values in stats display
- Follow codebase patterns: add TypeScript types, loading/error states, TODO comments
- Currency should use `.toFixed(2)` for consistent formatting

**Related commits:** 5d5cfb3, c065a70, 44834fb, cbb0cf7

**Files modified:**
- app/admin/page.tsx (converted to 'use client', added useQuery for stats, added error/loading states)
- components/admin/task-monitor.tsx (removed mock tasks, added empty state, added loading state)

### 2026-02-04 - Fix Frontend-Backend API Response Mismatch

**What was completed:**
- Fixed ArticleListResponse interface to match backend response format
- Changed 'data' field to 'articles' to match backend's actual response
- Changed 'limit' field to 'pageSize' to match backend's actual response
- Fixed request parameter mapping: frontend 'limit' now maps to backend 'page_size'
- Fixed system status component to check backend health endpoint directly
- Verified all other API endpoints have correct interfaces

**Important takeaways:**
- **API contract alignment**: Frontend TypeScript interfaces must exactly match backend JSON responses
- **Response field naming**: Backend returns `{"articles": [...]}` not `{"data": [...]}`
- **Pagination field naming**: Backend uses `pageSize` not `limit` in response
- **Request parameter mapping**: Backend expects `page_size` query parameter, not `limit`
- **Health check endpoint**: System status must fetch from backend URL directly (localhost:8080/health)
- **Error manifestation**: Mismatched field names cause "无法连接后端服务" error even when backend is running fine
- **Testing approach**: Use curl to verify actual backend responses before writing frontend interfaces

**Related commits:** cac15b6, 6aa95af, 667c450

**Files modified:**
- web3-insight/frontend/lib/api.ts (ArticleListResponse interface, parameter mapping)
- web3-insight/frontend/components/knowledge/article-list.tsx (response field access)
- web3-insight/frontend/components/admin/system-status.tsx (health check endpoint)
- CLAUDE.md (this documentation)

### 2026-02-04 - Fix Explorer Research Tab Error

**What was completed:**
- Fixed CSS compatibility issue in Textarea component (removed unsupported field-sizing-content property)
- Added ErrorBoundary component for graceful error handling
- Verified null reference safety in explorer stats display
- Tested Add Explorer dialog functionality (successfully created 2 test explorers)
- Frontend restarted with clean cache to clear Next.js compilation errors
- Verified both "知识库" and "浏览器调研" tabs load correctly with no errors

**Important takeaways:**
- **CSS compatibility**: The `field-sizing: content` CSS property is not widely supported (only in Safari 17.4+). Caused console warnings and potential layout issues. Must remove or use feature detection.
- **Error boundaries**: Critical for production apps. React error boundaries catch rendering errors and prevent white screens. Should be added to all major feature sections.
- **Next.js cache issues**: Turbopack cache can cause persistent errors even after fixes. Use `rm -rf .next` to force clean rebuild.
- **Stats null safety**: Always check for null/undefined when displaying counts. Use `?? 0` pattern or optional chaining.
- **API response structure**: Explorer API returns `{"count": N, "data": [...]}` format. Stats API returns `{"total": N, "byChain": [...], "byStatus": {...}}`.
- **Frontend logs location**: When services are started with scripts, check `/Users/tongleyao/claudeProjects/explorerResearch/web3-insight/logs/frontend.log` for errors.

**Verification results:**
- ✅ "浏览器调研" tab loads with 200 OK response
- ✅ No CSS warnings in console (field-sizing-content removed)
- ✅ Explorer stats display correctly (total: 2, by chain, by status)
- ✅ Add Explorer dialog works (created 2 test explorers)
- ✅ Error boundary catches and displays errors gracefully
- ✅ No null reference errors

**Related commits:** f429a17, 9ef3d3d

**Files modified:**
- web3-insight/frontend/components/ui/textarea.tsx (removed field-sizing-content)
- web3-insight/frontend/components/research/error-boundary.tsx (new component)
- web3-insight/frontend/app/research/page.tsx (wrapped with ErrorBoundary)

### 2026-02-04 - Fetch Categories from Backend

**What was completed:**
- Removed hardcoded mockCategories array from frontend
- Added Category interface matching backend model to lib/api.ts
- Added categoryAPI.getTree() method to fetch hierarchical categories
- Updated CategoryTree component to use useQuery for data fetching
- Added loading, error, and empty states to CategoryTree
- Added test coverage for loading state
- Verified end-to-end integration with backend API

**Important takeaways:**
- **React Query pattern**: Frontend uses TanStack Query (React Query) for data fetching throughout the app. Always use useQuery/useMutation instead of useEffect + fetch.
- **Empty state UX**: Always show user-friendly empty states ("暂无分类") rather than blank screens when no data exists.
- **Backend category API**: The backend provides two endpoints:
  - `GET /api/categories` - flat list of all categories
  - `GET /api/categories/tree` - hierarchical tree structure
  - Tree endpoint recursively loads children for building nested category UI
- **TypeScript interfaces**: Frontend Category interface must match backend model.Category struct fields
- **Testing approach**: Write tests first to verify loading states, then implement the component

**Related commits:** [commit hash to be added]

**Files modified:**
- web3-insight/frontend/lib/api.ts (Category interface, categoryAPI.getTree)
- web3-insight/frontend/components/knowledge/category-tree.tsx (replaced mock with useQuery)
- web3-insight/frontend/components/knowledge/__tests__/category-tree.test.tsx (new test)
