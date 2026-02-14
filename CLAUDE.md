# CRITICAL SAFETY RESTRICTION

> **MANDATORY:** This rule takes precedence over all other instructions.

**Scope Boundary:** `/Users/tongleyao/claudeProjects/explorerResearch`

All file operations (create, delete, modify, read, execute) MUST be confined to this directory and its subdirectories. Any operation that would affect files **outside** this scope requires **explicit user confirmation** before proceeding.

**Before any file operation, verify:**
1. The target path starts with `/Users/tongleyao/claudeProjects/explorerResearch/`
2. The operation will not affect parent directories or sibling projects
3. Commands will not have side effects outside the allowed scope

**If an operation would affect files outside this scope:**
- STOP immediately
- Explain what operation was requested and why it's out of scope
- Ask for explicit user confirmation before proceeding

**Secrets & API Keys:**
- NEVER read, print, log, or expose any API keys, tokens, or secrets (e.g., `ANTHROPIC_API_KEY`, `.env` files)
- NEVER include API keys in code, commits, logs, or any output
- NEVER read, query, or expose API keys from the database `configs` table (keys like `api_key.anthropic`, `api_key.openai`, etc.)
- Only reference environment variables by name (e.g., `os.Getenv("ANTHROPIC_API_KEY")`), never by value
- API keys are managed via admin UI (`/admin/config` → API 密钥 tab). DB is source of truth; env vars are read-only fallbacks
- `config.yaml` is secret-free — no API key fields remain. Only non-sensitive config (enabled flags, model names, ports)
- This applies to ALL agents, teammates, and subprocesses

# Project Guidelines

## Go Build Commands

This environment has GVM (Go Version Manager) configured in shell profile, which causes `cd` commands to fail. Use absolute path to Go binary with `-C` flag:

```bash
# CORRECT - use absolute Go path with -C flag
/usr/local/go/bin/go build -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./...
/usr/local/go/bin/go test -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./...
/usr/local/go/bin/go mod tidy -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend
```

## Project Structure

```
web3-insight/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go        # API 服务入口
│   │   ├── worker/main.go        # 异步任务 Worker
│   │   ├── cleardata/main.go     # 数据清理工具
│   │   ├── bulk-tag/main.go      # 批量标签工具 (标记未打标文章)
│   │   ├── eval-tagger/main.go   # 标签质量评估 CLI
│   │   ├── bench-tagger/main.go  # 标签 Benchmark 评估 CLI
│   │   └── seed-articles/main.go # 测试文章生成脚本
│   ├── config/
│   │   ├── models.yaml           # 模型注册表 (本地/云端模型定义)
│   │   ├── routing.yaml          # 任务路由配置 (任务→模型映射)
│   │   ├── prompts.yaml          # 主题提示词模板 (9个主题 + 生成配置)
│   │   ├── research.yaml         # 即时研究域模板 (8个领域 + 生成设置)
│   │   └── tags.yaml             # 标签注册表 (92个标签, 按主题分组)
│   ├── internal/
│   │   ├── api/                  # HTTP 路由和处理器
│   │   │   ├── router.go         # 路由注册
│   │   │   ├── article.go        # 文章 CRUD
│   │   │   ├── category.go       # 分类管理
│   │   │   ├── config.go         # 系统配置
│   │   │   ├── model_config.go   # 模型配置 API
│   │   │   ├── explorer.go       # Explorer 调研
│   │   │   ├── datasource.go     # 数据源管理
│   │   │   ├── news.go           # 新闻管理
│   │   │   ├── importer.go       # 导入导出
│   │   │   ├── search.go         # 搜索 (关键词+语义)
│   │   │   ├── task.go           # 任务监控
│   │   │   ├── kb_update.go      # 知识库更新 API (主题、关键词、调度)
│   │   │   ├── tag.go            # 标签管理 API
│   │   │   ├── api_key.go        # API Key 管理 (列表/保存/测试)
│   │   │   ├── chat_ws.go        # WebSocket 聊天
│   │   │   ├── research.go       # 即时研究 REST API (会话/计划/固定/整合)
│   │   │   ├── research_chat_ws.go # 即时研究 WebSocket 聊天
│   │   │   └── middleware.go     # CORS 中间件
│   │   ├── config/               # 配置加载 (Viper)
│   │   │   ├── config.go         # 主配置
│   │   │   ├── models.go         # 模型注册表解析
│   │   │   ├── routing.go        # 路由配置解析
│   │   │   ├── prompts.go        # 主题提示词配置解析
│   │   │   ├── research.go       # 即时研究配置解析 (领域模板+生成设置)
│   │   │   └── tags.go           # 标签注册表解析
│   │   ├── collector/            # 数据采集
│   │   │   ├── crawler.go        # 网页爬虫 (Colly)
│   │   │   ├── rss.go            # RSS 订阅
│   │   │   ├── search.go         # 搜索引擎
│   │   │   ├── serpapi.go        # SerpAPI 集成
│   │   │   └── tavily.go         # Tavily 集成
│   │   ├── database/             # 数据库 (connection.go, migrate.go)
│   │   ├── llm/                  # LLM 客户端
│   │   │   ├── ollama.go         # Ollama (本地)
│   │   │   ├── claude.go         # Claude API
│   │   │   ├── openai.go         # OpenAI API
│   │   │   ├── router.go         # LLM 路由
│   │   │   └── embedding.go      # 向量嵌入
│   │   ├── model/                # 数据模型 (GORM structs)
│   │   ├── repository/           # 数据访问层
│   │   ├── service/              # 业务逻辑层
│   │   │   ├── model_selector.go # 模型选择 (主选/备选 fallback)
│   │   │   ├── chat.go           # 聊天服务
│   │   │   ├── generator.go      # 内容生成
│   │   │   ├── summarizer.go     # 摘要生成
│   │   │   ├── classifier.go     # 自动分类
│   │   │   ├── research.go       # 即时研究编排 (计划→研究→撰写, Claude CLI)
│   │   │   ├── research_chat.go  # 即时研究聊天 (会话上下文+领域提示)
│   │   │   ├── semantic_search.go # 语义搜索
│   │   │   ├── keyword_pool.go   # 关键词池 (主题感知, 自动批次)
│   │   │   ├── article_generator.go # 文章生成 (Claude CLI)
│   │   │   ├── kb_update_orchestrator.go # KB更新编排器
│   │   │   ├── kb_scheduler.go   # KB更新调度器
│   │   │   ├── theme_sync.go     # 主题同步 (config→DB)
│   │   │   ├── tagger.go         # 文章自动标签 (Sonnet + balanced_v1 prompt)
│   │   │   ├── eval_tagger.go    # 标签质量评估指标
│   │   │   ├── bench_tagger.go   # 标签 Benchmark 运行引擎
│   │   │   ├── article_updater.go # 文章对话更新生成 (LLM定向补充)
│   │   │   ├── cli_article_updater.go # CLI文章更新 (订阅认证, 零成本)
│   │   │   ├── claude_executor.go # Claude CLI 执行器 (env过滤, 进程组管理)
│   │   │   └── key_provider.go   # API Key 提供者 (DB-backed, 30s缓存)
│   │   └── worker/               # 异步任务 (Asynq)
│   └── scripts/
│       └── clear_data.sql        # 数据清理 SQL
├── frontend/
│   ├── app/                      # Next.js App Router 页面
│   │   ├── page.tsx              # 首页
│   │   ├── knowledge/            # 知识库 (文章列表+详情)
│   │   ├── research/             # 即时研究
│   │   │   ├── page.tsx          # 研究首页 (搜索+领域+历史)
│   │   │   └── [id]/page.tsx     # 研究会话详情 (报告+聊天双栏)
│   │   └── admin/                # 管理后台
│   │       ├── page.tsx          # 仪表板 (状态+任务)
│   │       ├── config/page.tsx   # 模型配置 + API 密钥
│   │       ├── import/page.tsx   # 文章导入
│   │       ├── tags/page.tsx     # 标签管理
│   │       └── kb-update/page.tsx # 知识库更新 (主题管理+调度)
│   ├── components/
│   │   ├── ui/                   # shadcn/ui 基础组件
│   │   ├── layout/               # 布局 (Sidebar, Header, MainLayout)
│   │   ├── knowledge/            # 知识库组件 (ArticleList, ArticleView, ArticleEditor, EditorToolbar, TagEditor, UpdateReviewPanel, VersionHistory)
│   │   ├── admin/                # 管理组件 (ModelConfig, ApiKeyConfig, TaskMonitor, SystemStatus, SourceConfig, ArticleImport)
│   │   ├── chat/                 # 聊天组件 (ChatSidebar, ChatMessage, SidebarToggle)
│   │   ├── research/             # 即时研究组件 (DomainSelector, PlanReview, ReportViewer, ResearchChat, SessionList, IntegrateFindings, PinButton)
│   │   └── providers/            # QueryProvider
│   ├── hooks/
│   │   ├── use-chat.ts           # 聊天 Hook (多轮对话, localStorage 持久化)
│   │   ├── use-research.ts       # 研究会话 Hook (状态轮询, 固定/整合, 分页历史)
│   │   ├── use-research-chat.ts  # 研究聊天 Hook (WebSocket, 模型选择)
│   │   ├── use-resize.ts         # 拖拽调整大小 Hook (rAF 节流)
│   │   └── use-feature-flag.ts   # Feature Flag Hook
│   ├── lib/
│   │   ├── api.ts                # 所有 API 类型定义和客户端方法
│   │   ├── websocket.ts          # WebSocket 客户端
│   │   └── utils.ts              # 工具函数 (cn)
│   └── __mocks__/                # MSW 测试 Mock
├── scripts/                      # 运维脚本
│   ├── lib.sh                    # 共享函数 (Docker Compose 检测)
│   ├── start-all.sh              # 一键启动所有服务
│   ├── stop-all.sh               # 停止所有服务
│   └── status.sh                 # 检查服务状态
├── docker-compose.yml            # Docker 服务 (PostgreSQL + Redis)
└── Makefile                      # 常用命令快捷方式
```

## Completed Features

| 模块 | 功能 | 前端组件 | 后端端点 |
|------|------|----------|----------|
| 知识库 | 文章列表/详情/CRUD | ArticleList, ArticleView | GET/POST/PUT/DELETE /api/articles |
| 知识库 | 分类管理/树结构 | - | GET /api/categories, /api/categories/tree |
| 知识库 | 全局搜索 + 语义搜索 | SearchInput | GET /api/search, /api/search/semantic |
| 聊天 | AI 实时对话 (可调宽侧边栏) | ChatSidebar, SidebarToggle, ResizeHandle | WS /ws/chat |
| 调研 | Explorer 管理 + 功能对比 | - (backend only) | /api/explorers (含 compare, stats, features) |
| 管理 | 系统状态 | SystemStatus | GET /health |
| 管理 | 任务监控 | TaskMonitor | GET /api/tasks, /api/tasks/stats |
| 管理 | 模型配置 (YAML + DB) | ModelConfig | /api/models (registry, tasks, selections) |
| 管理 | 数据源管理 | SourceConfig | /api/sources (CRUD + sync + validate) |
| 管理 | 文章导入导出 | ArticleImport | /api/import (batch, validate, upload, template, export) |
| 管理 | 新闻管理 | - | /api/news |
| 基础 | Feature Flags | DisabledFeature | - |
| 基础 | 错误边界 | ErrorBoundary | - |
| 管理 | 知识库自动更新 (主题驱动) | KBUpdatePage | /api/kb (trigger, jobs, keywords, scheduler, themes, config) |
| 管理 | 标签管理 (CRUD+注册表+生命周期) | TagsPage | /api/tags (list, create, update, delete, stats, status, approve) |
| 标签 | 文章自动标签 (Sonnet + balanced_v1) | - | tagger.go (TagArticle) |
| 标签 | 自动标签开关 (config toggle) | Switch on TagsPage | PUT /api/config/auto_tagging_enabled |
| 标签 | 标签评估管道 | - | eval-tagger CLI |
| 标签 | 标签 Benchmark 评估工具 | - | bench-tagger CLI + bench_tagger.go |
| 基础 | CORS 中间件 | - | middleware.go |
| 知识库 | 文章归档/取消归档 | ArticleView (dropdown) | PATCH /api/articles/:id/archive |
| 知识库 | 文章删除 (确认对话框) | ArticleView + AlertDialog | DELETE /api/articles/:id |
| 知识库 | 手动标签编辑 (自动补全) | TagEditor | PUT /api/articles/:id/tags |
| 知识库 | 标签搜索 (自动补全) | TagEditor | GET /api/tags/search |
| 知识库 | 已归档文章过滤 | Checkbox on KnowledgePage | GET /api/articles?archived= |
| 聊天 | 多轮对话 + 上下文历史 | ChatSidebar + useChat | WS /ws/chat (history support) |
| 聊天 | Markdown 渲染 (助手消息) | ChatMessage (ReactMarkdown + remarkGfm) | - |
| 聊天 | 侧边栏拖拽调整宽度 (320-600px) | ResizeHandle + useResize | - |
| 知识库 | AI 文章更新生成 | ChatSidebar → UpdateReviewPanel | POST /api/articles/:id/generate-update |
| 知识库 | 文章更新审阅 (inline diff 视图) | UpdateReviewPanel (inline/overlay variants) | POST /api/articles/:id/apply-update |
| 知识库 | 版本历史 + 回滚 | VersionHistory | GET /api/articles/:id/versions, POST /:id/versions/:versionId/rollback |
| 知识库 | CLI 文章更新 (订阅认证, API fallback) | ChatSidebar → UpdateReviewPanel | POST /api/articles/:id/generate-update (cli_article_updater.go → claude_executor.go) |
| 聊天 | 模型选择器 (Haiku/Sonnet/Opus, 默认 Sonnet) | Select in ChatSidebar header + useChat | WS /ws/chat (model field → resolveModelName → adapter) |
| 管理 | API Key DB存储 + 管理UI (动态keyFunc, env fallback) | ApiKeyConfig | GET/PUT /api/models/keys, POST /api/models/keys/test |
| 知识库 | WYSIWYG 文章编辑器 (Tiptap, markdown round-trip) | ArticleEditor, EditorToolbar | POST /api/articles/:id/save-edit |
| 研究 | 即时研究首页 (搜索+领域选择+历史) | research/page.tsx, DomainSelector, SessionList | GET /api/research/domains, /sessions |
| 研究 | 研究会话详情 (报告+聊天双栏) | research/[id]/page.tsx, ReportViewer, ResearchChat | GET /api/research/sessions/:id, /status |
| 研究 | 研究计划审核 (编辑+批准) | PlanReview | POST /api/research/sessions/:id/approve-plan |
| 研究 | 研究报告生成 (Claude CLI, 订阅认证) | ReportViewer (阶段指示器) | POST /api/research/sessions (202 异步) |
| 研究 | 研究聊天 (会话上下文+模型选择) | ResearchChat | WS /ws/research-chat |
| 研究 | 固定发现 + 整合到报告 | PinButton, IntegrateFindings | POST /sessions/:id/pin, /integrate |
| 研究 | 会话取消 (CLI 进程终止) | - | POST /api/research/sessions/:id/cancel |
| 研究 | 会话删除 (含关联文章) | SessionList (确认对话框) | DELETE /api/research/sessions/:id |

## Model Fallback Pattern

All AI-dependent features MUST use `ModelSelector` service:

```go
result, err := modelSelector.SelectModelForTask("simple_generation")
if err != nil {
    return nil, fmt.Errorf("cannot execute task: %w", err) // Both unavailable
}
if result.IsFallback {
    log.Printf("Using fallback model: %s", result.Reason)
}
```

Configuration: `backend/config/models.yaml` (registry) + `backend/config/routing.yaml` (routing) + DB `configs` table (user selections + API keys)

**API Key Resolution:** All LLM adapters use `keyFunc func() string` pattern. Keys are resolved at request time from `KeyProvider` (DB-backed, 30s cache). Fallback order: DB → environment variable → empty string. Env vars are read-only fallbacks (never written to DB). DB keys: `api_key.anthropic`, `api_key.openai`, `api_key.tavily`, `api_key.serpapi`. Config structs (`ClaudeConfig`, `OpenAIConfig`, `TavilyConfig`, `SerpAPIConfig`) no longer contain `APIKey` fields — all key management is through DB + admin UI.

## Service Management

**Start all:** `./web3-insight/scripts/start-all.sh`
**Stop all:** `./web3-insight/scripts/stop-all.sh`
**Status:** `./web3-insight/scripts/status.sh`

**Ports:** PostgreSQL 5432, Redis 6379, Backend 8080, Frontend 3000, Ollama 11434

**Docker services:**
```bash
docker-compose -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/docker-compose.yml up -d postgres redis
docker-compose -f /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/docker-compose.yml down
```

**Database:** PostgreSQL (`pgvector/pgvector:pg16`), User: `web3insight`, Password: `web3insight_dev`, DB: `web3insight`

**Migrations:** Automatic on backend startup via GORM AutoMigrate. No separate command.

**Research Feature:** Uses Claude CLI (`claude --print`) for report generation via subscription auth (StripAPIKey: true, $0 API cost). Each research session spawns a CLI subprocess with fresh UUID. Orphaned sessions cleaned up on startup (45-min cutoff). Max 3 concurrent sessions.

**Logs:** `web3-insight/logs/{backend,frontend,worker,ollama}.log`

## Build & Test

```bash
# Backend
/usr/local/go/bin/go build -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./...
/usr/local/go/bin/go test -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./...

# Frontend
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm run build
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm test
```

## Browser Testing

Chrome browser automation is available via `claude-in-chrome` extension. Use proactively for visual verification after UI changes. Always call `tabs_context_mcp` first.

### Chrome MCP Connection Issue Protocol

- When QA encounters Chrome MCP browser extension disconnection during testing, they should STOP immediately and report to the team leader
- The team leader will notify the user to fix the Chrome connection (run `/chrome` command)
- QA should NOT skip browser testing or fall back to API-only testing — browser E2E testing is required
- After the user fixes the connection, QA resumes browser testing

## Key Patterns & Lessons

- **Package manager**: Frontend uses `npm` (not pnpm)
- **API contracts**: Frontend TypeScript interfaces must exactly match backend JSON responses. Use `curl` to verify actual responses.
- **React Query**: Use `useQuery`/`useMutation` for all data fetching. Never `useEffect` + `fetch`.
- **Empty states**: Always provide user-friendly empty states with Chinese text
- **Stats null safety**: Use `?? 0` pattern for safe defaults
- **Next.js cache**: Use `rm -rf .next` to force clean rebuild when encountering persistent errors
- **CSS compatibility**: Avoid `field-sizing: content` (limited browser support)

### Theme System Lessons (2026-02-10)

- **JSON field naming consistency**: Backend Go struct JSON tags MUST match frontend request body keys. Use camelCase throughout (`json:"batchSize"`, not `json:"batch_size"`). The `binding:"required"` validator will silently reject mismatched field names, making the bug hard to trace.
- **Composite unique indexes for multi-tenant data**: When data is partitioned by a foreign key (e.g., keywords by theme), unique constraints must be composite: `uniqueIndex:idx_keyword_theme` on BOTH `Keyword` and `ThemeID` fields. Single-field unique index prevents valid cross-theme duplicates.
- **Config-to-DB sync pattern**: Define config as source of truth in YAML, sync to DB on startup. Create new records as "paused", update metadata for existing (preserve runtime status). Auto-activate first item if none active.
- **Go template rendering for prompts**: Use `text/template` with struct variables (`{{.Keyword}}`, `{{.Count}}`). Keep prompts in YAML config, never hardcode in Go. Mark prompt fields `json:"-"` to prevent API exposure.
- **GORM AutoMigrate ordering**: Tables with foreign keys must appear AFTER the referenced table in the `AutoMigrate()` call. Theme must come before Keyword.

### KB Auto-Update Lessons (2026-02-07)

- **Orphaned Job Cleanup**: On startup, mark any job running > 30 min as "failed" to prevent stale locks.
- **LLM Output Parsing**: JSON output from LLMs is ~75% reliable even with strong prompts. Use two-layer defense: prompt constraints + `extractJSON()` code extraction. For complex content (markdown, Chinese), use delimiter format (`===TITLE_START===`/`===TITLE_END===`) for 100% success.
- **Article Generation Timeout**: Simple articles: 2-4 min. Research-heavy articles (WebFetch): 60-90 min. Default timeout: 60 min. Initial 5-10 min timeouts failed all articles.
- **goroutines Must Use context.Background()**: HTTP handler goroutines must NOT reuse `c.Request.Context()` — it's canceled when the response is sent. Always create fresh `context.Background()` for long-running background tasks.
- **Async HTTP Pattern for Long Tasks**: Return HTTP 202 immediately, run work in goroutine with `context.Background()`. Use `sync.Mutex.TryLock()` or DB checks for 409 Conflict on duplicate requests.
- **Repository Testing**: PostgreSQL-specific features (UUID, pq.StringArray, array_append) make SQLite-based tests impossible. Use integration tests against real PostgreSQL.
- **LLM Prompt Specificity**: Concrete keywords ("Gas optimization tips") produce reliable structured output. Abstract keywords ("Byzantine fault tolerance") cause LLMs to output free-form text instead.

## Tagging Quality Metrics

**Current production config**: Claude Sonnet 4 + balanced_v1 prompt. Fallback: Claude Haiku 4.5.

Benchmark (Sonnet + balanced_v1): Macro-F1 ~79.6%, Registry Rate ~92%, Avg Tags ~4.3. Run `eval-tagger` or `bench-tagger` for latest metrics.

**评估命令:**
```bash
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./cmd/eval-tagger --limit 50
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./cmd/eval-tagger --export review.md
```

**Benchmark 命令:**
```bash
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./cmd/bench-tagger
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./cmd/bench-tagger --method sonnet-balanced-v1 --verbose
```

**批量标签 API:**
```bash
curl -X POST "http://localhost:8080/api/tags/bulk-tag?force=true"
```

**自动标签开关:**
```bash
# 查询状态
curl http://localhost:8080/api/config/auto_tagging_enabled
# 关闭
curl -X PUT http://localhost:8080/api/config/auto_tagging_enabled -H "Content-Type: application/json" -d '{"value":"false"}'
# 开启
curl -X PUT http://localhost:8080/api/config/auto_tagging_enabled -H "Content-Type: application/json" -d '{"value":"true"}'
```

### Tagging Lessons (2026-02-10, updated 2026-02-11)

- **Model capability > prompt engineering**: Sonnet vs Haiku on same prompt shows 5-8pp F1 difference. Registry compliance jumps from ~75% to ~92%. Model's instruction-following ability is the biggest lever.
- **Benchmark-driven optimization**: 12 method combinations (4 models × 6 prompts) tested via bench-tagger CLI. balanced_v1 + Sonnet won on F1 (79.6%) while keeping cost under $0.01/article.
- **Auto-tagging toggle**: Uses `configs` table key `auto_tagging_enabled`. Frontend sends string "true"/"false", backend stores as JSON `"true"`/`"false"` (with quotes from json.Marshal). Check uses `string(cfg.Value) == \`"false"\``.
- **LLM tag compliance**: Even with explicit "only choose from this list" prompts, Claude Haiku generates off-registry tags ~15% of the time. The `ResolveTag()` function handles case-insensitive matching and parenthetical stripping. Code-level validation filters these out to achieve 100% compliance.
- **Keyword fallback for minimum tags**: After LLM tag validation, if fewer than 3 tags remain, auto-supplement from universal tags by keyword matching against article title/summary. This raised in-range from 89% to 96%.
- **Reuse/orphan metrics require scale**: With only 27 articles, tag reuse metrics are inherently low. These targets (>80% reuse, <15% orphan) are meaningful only at 50+ articles.
- **Bulk tagging**: Use `POST /api/tags/bulk-tag?force=true` endpoint or `cmd/bulk-tag --force` CLI. API endpoint runs in background, CLI is synchronous.
- **Tag auto-creation disabled**: The `handleNewTagSuggestion()` was causing DB pollution (196 auto-created tags). Now disabled - LLM suggestions are logged but not auto-created.

### CLI Article Updater & Chat Model Selector Lessons (2026-02-12)

- **Claude CLI subprocess patterns**: Every call MUST use a fresh `uuid.UUID` session ID (reuse causes errors). Strip `ANTHROPIC_API_KEY` from env via `ClaudeExecutorOptions{StripAPIKey: true}` to force subscription auth ($0 cost). Use `syscall.SysProcAttr{Setpgid: true}` + `syscall.Kill(-pid, SIGKILL)` to kill entire process group on timeout.
- **Model name resolution pattern**: Frontend sends short names ("haiku"/"sonnet"/"opus"), backend `resolveModelName()` maps to adapter names ("claude-haiku-4-5"/"claude-sonnet-4"/"claude-opus-4"). Frontend `formatModelName()` reverse-maps using `includes()` for display. Keep both maps in sync.
- **Explicit model selection = no fallback**: When user explicitly selects a model via `GenerateStreamWithModel()`, failures return errors directly (no automatic fallback to another model). This is correct — user should know if their chosen model is unavailable.
- **CLI → API fallback pattern**: `ArticleHandler.GenerateUpdate()` tries `cliUpdater` first, falls back to `updater` (API-based) on any error. Both are injected separately in `NewArticleHandler()`. Log the fallback reason for debugging.
- **Ref pattern for stale closures**: Frontend `useChat` hook uses `modelRef.current` (not `model` state) inside `sendMessage` callback to avoid stale closure over React state. Same pattern for `messagesRef`.

### Config Consolidation Lessons (2026-02-12)

- **Shared config path resolution**: All CLI tools (eval-tagger, bench-tagger, seed-articles, bulk-tag) should use `config.FindConfigFile()` and `config.LoadTagsFromConfigDir()` instead of duplicating path-search logic. This eliminated ~50 lines of duplicated code across 3 files.
- **Sensitive vs non-sensitive config separation**: `config.yaml` has only the DB password (dev default `web3insight_dev`), no API keys. The other 4 YAML files (models, routing, prompts, tags) are non-sensitive. `.claudeignore` and `settings.local.json` deny rules block only `config.yaml`.
- **Hardcoded DSN is a code smell**: `seed-articles` had a hardcoded DB connection string. Always use `config.Load()` + `database.Connect()` — it respects config file values and env var expansion.

### WYSIWYG Editor Lessons (2026-02-15)

- **ContentHTML stale data bug**: When articles have a pre-existing `contentHtml` field (from import/crawl), any content update (manual edit, AI update, rollback) MUST clear `ContentHTML = ""` so the markdown renderer takes over. This is easy to miss on new content-modifying endpoints — check all paths that update `article.Content`.
- **Tiptap markdown round-trip**: Use `tiptap-markdown` extension for load (content as string) and save (`editor.storage.markdown.getMarkdown()`). Minor formatting differences (extra newlines, list style) are expected and acceptable.
- **Mutual exclusion for edit modes**: When multiple editing modes exist (manual edit vs AI update), lift state to the common parent and use boolean flags (`isEditing`, `isGenerating`, `isReviewOpen`) to disable conflicting actions. Simpler than a state machine for two modes.

### Instant Research Lessons (2026-02-15)

- **Cancellable goroutines with sync.Map**: Long-running background goroutines (CLI subprocesses) need cancellation support. Store `context.CancelFunc` per session ID in `sync.Map`, call it on cancel, defer cleanup on completion. Using bare `context.Background()` without cancel is a bug — the process runs forever even after user cancellation.
- **GORM column/JSON name collision**: A field with `gorm:"column:category" json:"category"` collides with a `Category *Category` relationship that also serializes to `"category"`. Always use distinct names — `gorm:"column:article_type" json:"articleType"`.
- **Column rename migration**: GORM AutoMigrate adds columns but won't rename them. Use a pre-migration check: query `information_schema.columns` for old name, then `ALTER TABLE RENAME COLUMN` before AutoMigrate runs.
- **Frontend-backend API contract verification**: Field name mismatches (`content` vs `messageContent`, `.id` vs `.sessionId`) cause silent failures. Always verify contracts from both sides during code review. The `binding:"required"` validator silently returns 400 for mismatched field names.
- **Delimiter parsing for multi-section LLM output**: For research reports with title + content + summary + citations, use multiple delimiter pairs (`===REPORT_TITLE_START===`/`===REPORT_TITLE_END===`). Add fallbacks for each section — title falls back to question, content falls back to full response.
- **Domain config ID consistency**: Frontend domain selector IDs must exactly match backend YAML config IDs. Use the full IDs (`tech-engineering`, not `tech`) to avoid silent routing failures.

## Plan Completion Protocol

After completing any implementation plan:
1. **整理文档**: 检查 `docs/` 和 `web3-insight/backend/docs/` 目录，将有价值的经验教训提取到 CLAUDE.md，将新功能描述更新到 README.md，然后删除这些文档文件
2. Delete the plan file from `docs/plans/`
3. Update relevant CLAUDE.md sections if project structure/features changed
4. **提交代码**: 将所有改动（代码 + 文档更新 + 文件清理）commit 到 git
