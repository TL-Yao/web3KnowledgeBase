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
- Only reference environment variables by name (e.g., `os.Getenv("ANTHROPIC_API_KEY")`), never by value
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
│   │   │   ├── chat_ws.go        # WebSocket 聊天
│   │   │   └── middleware.go     # CORS 中间件
│   │   ├── config/               # 配置加载 (Viper)
│   │   │   ├── config.go         # 主配置
│   │   │   ├── models.go         # 模型注册表解析
│   │   │   ├── routing.go        # 路由配置解析
│   │   │   └── prompts.go        # 主题提示词配置解析
│   │   ├── collector/            # 数据采集
│   │   │   ├── crawler.go        # 网页爬虫 (Colly)
│   │   │   ├── rss.go            # RSS 订阅
│   │   │   ├── search.go         # 搜索引擎
│   │   │   ├── serpapi.go        # SerpAPI 集成
│   │   │   └── tavily.go         # Tavily 集成
│   │   ├── database/             # 数据库
│   │   │   ├── connection.go     # 连接管理
│   │   │   └── migrate.go        # 自动迁移 (GORM AutoMigrate)
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
│   │   │   ├── research.go       # 即时研究
│   │   │   ├── semantic_search.go # 语义搜索
│   │   │   ├── keyword_pool.go   # 关键词池 (主题感知, 自动批次)
│   │   │   ├── article_generator.go # 文章生成 (Claude CLI)
│   │   │   ├── kb_update_orchestrator.go # KB更新编排器
│   │   │   ├── kb_scheduler.go   # KB更新调度器
│   │   │   ├── theme_sync.go     # 主题同步 (config→DB)
│   │   │   ├── tagger.go         # 文章自动标签 (Sonnet + balanced_v1 prompt)
│   │   │   ├── eval_tagger.go    # 标签质量评估指标
│   │   │   ├── bench_tagger.go   # 标签 Benchmark 运行引擎
│   │   │   └── article_updater.go # 文章对话更新生成 (LLM定向补充)
│   │   └── worker/               # 异步任务 (Asynq)
│   └── scripts/
│       └── clear_data.sql        # 数据清理 SQL
├── frontend/
│   ├── app/                      # Next.js App Router 页面
│   │   ├── page.tsx              # 首页
│   │   ├── knowledge/            # 知识库 (文章列表+详情)
│   │   ├── research/             # Explorer 调研
│   │   └── admin/                # 管理后台
│   │       ├── page.tsx          # 仪表板 (状态+任务)
│   │       ├── config/page.tsx   # 模型配置
│   │       ├── import/page.tsx   # 文章导入
│   │       └── kb-update/page.tsx # 知识库更新 (主题管理+调度)
│   ├── components/
│   │   ├── ui/                   # shadcn/ui 基础组件
│   │   ├── layout/               # 布局 (Sidebar, Header, MainLayout)
│   │   ├── knowledge/            # 知识库组件 (ArticleList, ArticleView, TagEditor, UpdateReviewPanel, VersionHistory)
│   │   ├── research/             # 调研组件 (ExplorerResearch, ErrorBoundary)
│   │   ├── admin/                # 管理组件 (ModelConfig, TaskMonitor, SystemStatus, SourceConfig, ArticleImport)
│   │   ├── chat/                 # 聊天组件 (FloatingChat, ChatMessage)
│   │   └── providers/            # QueryProvider
│   ├── hooks/
│   │   ├── use-chat.ts           # 聊天 Hook (多轮对话, localStorage 持久化)
│   │   └── use-feature-flag.ts   # Feature Flag Hook
│   ├── lib/
│   │   ├── api.ts                # 所有 API 类型定义和客户端方法
│   │   ├── websocket.ts          # WebSocket 客户端
│   │   └── utils.ts              # 工具函数 (cn)
│   └── __mocks__/                # MSW 测试 Mock
├── scripts/                      # 运维脚本
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
| 知识库 | 分类管理/树结构 | CategoryTree | GET /api/categories, /api/categories/tree |
| 知识库 | 全局搜索 + 语义搜索 | SearchInput | GET /api/search, /api/search/semantic |
| 聊天 | AI 实时对话 | FloatingChat | WS /ws/chat |
| 调研 | Explorer 管理 + 功能对比 | ExplorerResearch | /api/explorers (含 compare, stats, features) |
| 管理 | 系统状态 | SystemStatus | GET /health |
| 管理 | 任务监控 | TaskMonitor | GET /api/tasks, /api/tasks/stats |
| 管理 | 模型配置 (YAML + DB) | ModelConfig | /api/models (registry, tasks, selections) |
| 管理 | 数据源管理 | SourceConfig | /api/sources (CRUD + sync + validate) |
| 管理 | 文章导入导出 | ArticleImport | /api/import (batch, validate, upload, template, export) |
| 管理 | 新闻管理 | - | /api/news |
| 基础 | Feature Flags | DisabledFeature | - |
| 基础 | 错误边界 | ExplorerErrorBoundary | - |
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
| 聊天 | 多轮对话 + 上下文历史 | FloatingChat + useChat | WS /ws/chat (history support) |
| 知识库 | AI 文章更新生成 | FloatingChat → UpdateReviewPanel | POST /api/articles/:id/generate-update |
| 知识库 | 文章更新审阅 (diff 视图) | UpdateReviewPanel | POST /api/articles/:id/apply-update |
| 知识库 | 版本历史 + 回滚 | VersionHistory | GET /api/articles/:id/versions, POST /:id/versions/:versionId/rollback |

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

Configuration: `backend/config/models.yaml` (registry) + `backend/config/routing.yaml` (routing) + DB `configs` table (user selections)

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

**Known Limitation**: Chrome MCP tools only work from the main Claude Code session. Team subagents (spawned via Task tool) cannot access Chrome MCP — all calls return "Browser extension is not connected." Browser E2E testing must be performed by the team leader (main session) directly, not delegated to qa-engineer or other teammates.

## Key Patterns & Lessons

- **GVM workaround**: Always use `/usr/local/go/bin/go -C /path` instead of `cd && go`
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

- **Claude CLI Session ID**: Every call MUST use a fresh `uuid.UUID`. Reuse causes "Session ID already in use" error. Retry loops also need new IDs per attempt. Stale session files at `~/.claude/projects/*/session-*` can also cause this — clean up on startup.
- **Process Group Killing**: When killing a timed-out Claude CLI process, use `syscall.SysProcAttr{Setpgid: true}` + `syscall.Kill(-pid, SIGKILL)` to kill the entire process group. Simple `cmd.Process.Kill()` leaves zombie child processes.
- **Orphaned Job Cleanup**: Backend restarts leave jobs stuck in "running". On startup, mark any job running > 30 min as "failed". Without this, job locking prevents new updates.
- **LLM Output Parsing**: JSON output from LLMs is ~75% reliable even with strong prompts. Use two-layer defense: prompt constraints + `extractJSON()` code extraction. For complex content (markdown, Chinese), use delimiter format (`===TITLE_START===`/`===TITLE_END===`) for 100% success.
- **Article Generation Timeout**: Simple articles: 2-4 min. Research-heavy articles (WebFetch): 60-90 min. Default timeout: 60 min. Initial 5-10 min timeouts failed all articles.
- **goroutines Must Use context.Background()**: HTTP handler goroutines must NOT reuse `c.Request.Context()` — it's canceled when the response is sent. Always create fresh `context.Background()` for long-running background tasks.
- **Async HTTP Pattern for Long Tasks**: Return HTTP 202 immediately, run work in goroutine with `context.Background()`. Use `sync.Mutex.TryLock()` or DB checks for 409 Conflict on duplicate requests.
- **Nil struct vs nil field**: Passing a struct with nil internal fields causes deep nil pointer panics. Pass `nil` for the whole struct and nil-check before use: `if classifier != nil { ... }`.
- **GORM Updates**: Use `Updates(map[string]interface{}{...})` for partial updates, not full struct (which zeros out unset fields).
- **pq.StringArray**: PostgreSQL `TEXT[]` fields need explicit `pq.StringArray(slice)` conversion. Array append: `gorm.Expr("array_append(field, ?)", value)`.
- **Next.js useSearchParams**: Requires `<Suspense>` boundary wrapper in App Router, otherwise build fails.
- **Repository Testing**: PostgreSQL-specific features (UUID, pq.StringArray, array_append) make SQLite-based tests impossible. Use integration tests against real PostgreSQL.
- **LLM Prompt Specificity**: Concrete keywords ("Gas optimization tips") produce reliable structured output. Abstract keywords ("Byzantine fault tolerance") cause LLMs to output free-form text instead.

## Tagging Quality Metrics (Success Matrix)

**Current production config**: Claude Sonnet 4 + balanced_v1 prompt (upgraded from Haiku + default prompt on 2026-02-11).
Fallback model: Claude Haiku 4.5 (previously qwen2.5:32b).

Benchmark result (Sonnet + balanced_v1): Macro-F1 ~79.6%, Registry Rate ~92%, Avg Tags ~4.3 (vs baseline Haiku: F1 72.6%, Reg 74.6%, Avg 3.6).

Evaluated on 27 articles (2026-02-10). Previous eval uses Haiku with tag registry validation + keyword fallback.

| 指标 | 目标 | 实际 | 状态 | 说明 |
|------|------|------|------|------|
| 每篇标签数 (3-7) | >95% in range | 96% (avg 4.0) | PASS | 26/27 in range, 1 article at 2 tags |
| 注册表合规率 | >95% | 100% (107/107) | PASS | All tags from registry |
| 标签复用率 (>=3篇) | >80% | 20% | FAIL* | 11/54 reused (dataset size limit) |
| 孤儿标签率 (仅1篇) | <15% | 57% | FAIL* | 31/54 orphaned (dataset size limit) |
| 最大覆盖率 | <=40% | 30% | PASS | "DeFi协议" covers 8/27 articles |
| Gini 系数 | <=0.60 | 0.37 | PASS | Good distribution |
| 空标签率 | 0% | 0% | PASS | All 27 articles tagged |

*Reuse/orphan metrics are mathematically limited at 27 articles. With 54 unique tags × 3 articles = 162 min assignments needed for 80% reuse, but only 107 total assignments exist. These targets become meaningful at 50+ articles.

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
- **Config env expansion**: `config.yaml` uses `${ANTHROPIC_API_KEY}` syntax. Config loader now calls `os.ExpandEnv()` automatically. CLI tools no longer need manual expansion.
- **Reuse/orphan metrics require scale**: With only 27 articles, tag reuse metrics are inherently low. These targets (>80% reuse, <15% orphan) are meaningful only at 50+ articles.
- **Bulk tagging**: Use `POST /api/tags/bulk-tag?force=true` endpoint or `cmd/bulk-tag --force` CLI. API endpoint runs in background, CLI is synchronous.
- **Tag auto-creation disabled**: The `handleNewTagSuggestion()` was causing DB pollution (196 auto-created tags). Now disabled - LLM suggestions are logged but not auto-created.

## Plan Completion Protocol

After completing any implementation plan:
1. **整理文档**: 检查 `docs/` 和 `web3-insight/backend/docs/` 目录，将有价值的经验教训提取到 CLAUDE.md，将新功能描述更新到 README.md，然后删除这些文档文件
2. Delete the plan file from `docs/plans/`
3. Update relevant CLAUDE.md sections if project structure/features changed
4. **提交代码**: 将所有改动（代码 + 文档更新 + 文件清理）commit 到 git
