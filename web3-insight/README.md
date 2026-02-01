# Web3-Insight

A local Web3 knowledge management system for Explorer developers.

## Tech Stack

- **Frontend:** Next.js 14, TypeScript, Tailwind CSS, shadcn/ui
- **Backend:** Go 1.21+, Gin, GORM, Asynq
- **Database:** PostgreSQL 15+ with pgvector, Redis 7+
- **LLM:** Ollama (local), Claude API, OpenAI API

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Node.js 18+
- Go 1.21+
- pnpm

### Development

```bash
# Start databases
docker-compose up -d

# Start backend
cd backend && go run .

# Start frontend
cd frontend && pnpm dev
```

## Project Structure

```
web3-insight/
├── frontend/       # Next.js frontend
├── backend/        # Go backend
├── docs/           # Documentation
└── docker-compose.yml
```
