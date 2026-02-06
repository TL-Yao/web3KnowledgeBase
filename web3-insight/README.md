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
│   ├── cmd/              # Entry points (server, worker, cleardata)
│   ├── config/           # YAML configs (models.yaml, routing.yaml)
│   └── internal/
│       ├── api/          # HTTP handlers and router
│       ├── config/       # Config loading (Viper)
│       ├── collector/    # Data collection (crawler, RSS, search APIs)
│       ├── database/     # DB connection and auto-migration
│       ├── llm/          # LLM clients (Ollama, Claude, OpenAI)
│       ├── model/        # GORM data models
│       ├── repository/   # Data access layer
│       ├── service/      # Business logic (model selector, chat, search)
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
| GET | /api/articles | List articles (paginated) |
| GET | /api/articles/:id | Get article detail |
| POST | /api/articles | Create article |
| PUT | /api/articles/:id | Update article |
| DELETE | /api/articles/:id | Delete article |
| GET | /api/categories | List categories |
| GET | /api/categories/tree | Get category tree |
| GET | /api/search | Keyword search |
| GET | /api/search/semantic | Semantic search |

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
| POST | /api/kb-update/trigger | Trigger knowledge base update |
| GET | /api/kb-update/status/:jobId | Get update job status |
| GET | /api/kb-update/history | List update history (paginated) |
| POST | /api/kb/keywords/init | Initialize keyword pool |
| GET | /api/kb/keywords/stats | Get keyword pool stats |
| GET | /api/kb/scheduler/status | Get scheduler status |
| POST | /api/kb/scheduler/start | Start auto-update scheduler |
| POST | /api/kb/scheduler/stop | Stop auto-update scheduler |

### Chat
| Method | Endpoint | Description |
|--------|----------|-------------|
| WS | /ws/chat | AI chat via WebSocket |

## KB Auto-Update

The system includes an automated knowledge base update feature that generates Web3 technical articles using Claude Code CLI.

**How it works:**
1. A keyword pool (target: 200) is maintained and auto-replenished when below 30
2. Every 4 hours (or manually triggered), 3 keywords are picked for article generation
3. Each keyword is sent to Claude Code CLI (`--print` mode) which researches and writes a Chinese technical article
4. Articles are saved to the knowledge base automatically

**Key components:**
- `ClaudeExecutor` - Claude Code CLI wrapper with per-call session isolation
- `KeywordPoolService` - Keyword pool management with auto-replenishment
- `ArticleGeneratorService` - Article generation with 60-min timeout and delimiter-based output parsing
- `KBUpdateOrchestrator` - Job orchestration with locking and orphaned job cleanup
- `KBScheduler` - Cron-based scheduler (`0 */4 * * *`)

**Reliability:** Job locking (409 on duplicate), orphaned job cleanup (>30 min), process group termination, per-article failure isolation.

**Frontend:** Admin page at `/admin/kb-update` with manual trigger, keyword pool stats, live progress polling, and update history.

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
