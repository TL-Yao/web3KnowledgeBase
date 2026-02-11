# Web3-Insight

A local-first Web3 knowledge management system designed for Explorer developers to quickly build Web3 knowledge, track industry trends, and research competitor products.

## Tech Stack

- **Frontend**: Next.js (App Router), TypeScript, Tailwind CSS, shadcn/ui, TanStack Query
- **Backend**: Go (Gin, GORM, Asynq)
- **Database**: PostgreSQL 16 (pgvector), Redis 7
- **LLM**: Ollama (local-first), Claude API, OpenAI API (cloud fallback)
- **Infrastructure**: Docker Compose, WebSocket

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+
- Node.js 18+ & npm
- Ollama (optional, for local LLM)

### Start All Services

```bash
./scripts/start-all.sh
```

This starts: PostgreSQL, Redis, Backend API (port 8080), Worker, Frontend (port 3000)

### Stop All Services

```bash
./scripts/stop-all.sh
```

### Check Status

```bash
./scripts/status.sh
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  Frontend (Next.js)              │
│         http://localhost:3000                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │ Knowledge │ │ Research │ │ Admin Dashboard  │ │
│  │   Base    │ │ (Explorer│ │ (Config, Tasks,  │ │
│  │ (Articles,│ │  Compare)│ │  Import, Models) │ │
│  │  Search)  │ │          │ │                  │ │
│  └──────────┘ └──────────┘ └──────────────────┘ │
└───────────────────────┬─────────────────────────┘
                        │ REST API + WebSocket
┌───────────────────────┴─────────────────────────┐
│                 Backend (Go/Gin)                 │
│            http://localhost:8080                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │ API Layer│ │ Services │ │   LLM Clients    │ │
│  │ (Handlers│ │(Selector,│ │(Ollama, Claude,  │ │
│  │  Router) │ │ Chat,    │ │ OpenAI)          │ │
│  │          │ │ Search)  │ │                  │ │
│  └──────────┘ └──────────┘ └──────────────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │Repository│ │Collectors│ │  Worker (Asynq)  │ │
│  │  (GORM)  │ │(Crawler, │ │  (Async Tasks)   │ │
│  │          │ │ RSS, API)│ │                  │ │
│  └──────────┘ └──────────┘ └──────────────────┘ │
└───────────────────────┬─────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────┐
│              Data Layer                          │
│  ┌──────────────────┐  ┌──────────────────────┐  │
│  │ PostgreSQL 16    │  │ Redis 7              │  │
│  │ + pgvector       │  │ (Task Queue, Cache)  │  │
│  │ (port 5432)      │  │ (port 6379)          │  │
│  └──────────────────┘  └──────────────────────┘  │
└──────────────────────────────────────────────────┘
```

## Project Structure

```
web3-insight/
├── backend/
│   ├── cmd/              # Entry points
│   │   ├── server/       # API server
│   │   ├── worker/       # Async task worker
│   │   ├── bulk-tag/     # Batch article tagging CLI
│   │   ├── eval-tagger/  # Tag quality evaluation CLI
│   │   ├── bench-tagger/ # Tag benchmark CLI
│   │   └── seed-articles/# Test article generator
│   ├── config/           # YAML configs (models, routing, prompts, tags)
│   └── internal/
│       ├── api/          # HTTP handlers and router
│       ├── config/       # Config loading (Viper)
│       ├── collector/    # Data collection (crawler, RSS, search APIs)
│       ├── database/     # DB connection and auto-migration
│       ├── llm/          # LLM clients (Ollama, Claude, OpenAI)
│       ├── model/        # GORM data models
│       ├── repository/   # Data access layer
│       ├── service/      # Business logic (model selector, chat, tagger, etc.)
│       └── worker/       # Async task definitions (Asynq)
├── frontend/
│   ├── app/              # Next.js pages (knowledge, research, admin)
│   ├── components/       # React components (ui, layout, knowledge, admin, chat)
│   └── lib/              # API client, WebSocket, utilities
├── scripts/              # Ops scripts (start, stop, status)
├── docker-compose.yml    # PostgreSQL + Redis
└── Makefile              # Build and dev shortcuts
```

## API Endpoints

### Knowledge Base
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/articles | List articles (paginated, filterable by archived/tag/search) |
| GET | /api/articles/:id | Get article detail (by ID or slug) |
| POST | /api/articles | Create article |
| PUT | /api/articles/:id | Update article (supports archived field) |
| DELETE | /api/articles/:id | Delete article (cascades keyword cleanup) |
| PATCH | /api/articles/:id/archive | Toggle article archived status |
| PUT | /api/articles/:id/tags | Replace article tags |
| GET | /api/articles/:id/related | Find related articles (vector similarity) |
| GET | /api/categories | List categories |
| GET | /api/categories/tree | Get category tree |
| GET | /api/search | Keyword search |
| GET | /api/search/semantic | Semantic search |

### Tags
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/tags | List tags (filter by status/theme) |
| GET | /api/tags/search | Tag autocomplete search |
| GET | /api/tags/in-use | Tags in use with article counts |
| GET | /api/tags/stats | Tag statistics |
| PUT | /api/tags/:id/status | Update tag status |
| POST | /api/tags/:id/approve | Approve pending tag |
| POST | /api/tags/bulk-tag | Bulk-tag all articles |

### Explorer Research
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/explorers | List explorers |
| POST | /api/explorers | Create explorer |
| GET | /api/explorers/stats | Get statistics |
| GET | /api/explorers/compare | Compare explorers |
| GET | /api/explorers/features | List features |

### Admin
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | System health check |
| GET | /api/tasks | List tasks |
| GET | /api/models/registry | Get model registry |
| PUT | /api/models/selections | Update model selections |
| GET | /api/sources | List data sources |
| POST | /api/import | Batch import articles |

### KB Auto-Update
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/kb/update/trigger | Trigger knowledge base update |
| GET | /api/kb/update/jobs | List update history |
| GET | /api/kb/update/jobs/:job_id | Get update job status |
| POST | /api/kb/keywords/init | Initialize keyword pool |
| GET | /api/kb/keywords/stats | Get keyword pool stats |
| GET | /api/kb/themes | List themes |
| GET | /api/kb/themes/active | Get active theme |
| PUT | /api/kb/themes/:id/activate | Set active theme |
| GET | /api/kb/config | Get KB config |
| PUT | /api/kb/config/batch-size | Update batch size |
| GET | /api/kb/scheduler/status | Get scheduler status |
| POST | /api/kb/scheduler/start | Start auto-update scheduler |
| POST | /api/kb/scheduler/stop | Stop auto-update scheduler |

### Chat
| Method | Endpoint | Description |
|--------|----------|-------------|
| WS | /ws/chat | AI chat via WebSocket |

## KB Auto-Update

The system includes an automated knowledge base update feature that generates Web3 technical articles using Claude Code CLI, organized by configurable themes.

**How it works:**
1. Themes are defined in `config/prompts.yaml` (9 themes, synced to DB on startup)
2. Each theme has a keyword pool (auto-replenished when below threshold)
3. Every 4 hours (or manually triggered), keywords are picked from the active theme for article generation
4. Each keyword is sent to Claude Code CLI (`--print` mode) which researches and writes a Chinese technical article
5. Articles are saved, auto-tagged, and added to the knowledge base

**Key components:**
- `ClaudeExecutor` - Claude Code CLI wrapper with per-call session isolation
- `KeywordPoolService` - Theme-aware keyword pool management with auto-replenishment
- `ArticleGeneratorService` - Article generation with 60-min timeout and delimiter-based output parsing
- `KBUpdateOrchestrator` - Job orchestration with locking and orphaned job cleanup
- `KBScheduler` - Cron-based scheduler (`0 */4 * * *`)

**Reliability:** Job locking (409 on duplicate), orphaned job cleanup (>30 min), process group termination, per-article failure isolation.

**Frontend:** Admin page at `/admin/kb-update` with theme management, manual trigger, keyword pool stats, live progress polling, and update history.

## Article Management

Articles support lifecycle management through the knowledge base UI:

- **Archive/Unarchive:** Toggle archived status from article detail page. Archived articles are hidden from the default list view but accessible via filter.
- **Delete:** Permanently delete articles with confirmation dialog. Cleans up keyword references.
- **Tag editing:** Inline tag editor on article detail page with autocomplete from the tag registry. Supports add, remove, and free-text entry.
- **Archived filter:** Checkbox on knowledge list page to show/hide archived articles (hidden by default).

## Tagging System

Articles are automatically tagged using LLM (Claude Sonnet) against a curated tag registry (92 tags across 9 themes, defined in `config/tags.yaml`).

**Auto-tagging:** When new articles are created via KB update, the tagger assigns 3-7 tags from the registry. Tags are validated for registry compliance, with keyword-based fallback for minimum tag count.

**Manual editing:** Article detail pages include an inline tag editor with autocomplete search against the tag registry.

**CLI tools:**
- `cmd/bulk-tag` — Batch-tag all untagged (or all with `--force`) articles
- `cmd/eval-tagger` — Evaluate tagging quality metrics (F1, registry rate, distribution)
- `cmd/bench-tagger` — Benchmark different model+prompt combinations

## Development

### Individual Services

```bash
# Database only
docker-compose up -d postgres redis

# Backend only
make dev-backend

# Frontend only
make dev-frontend

# Worker only
make worker
```

### Build

```bash
make build           # Build all
make build-backend   # Build backend binaries
make build-frontend  # Build frontend
```

### Test

```bash
make test-backend    # Go tests
make test-frontend   # Jest tests
```

## Model Configuration

The system uses a two-layer configuration:
1. **YAML files** (`config/models.yaml`, `config/routing.yaml`) - developer-maintained model registry
2. **Database** - user-selected model preferences per task type

Each AI task has a primary and fallback model. If the primary is unavailable, the system automatically uses the fallback.
