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
│   │   │   ├── tagger.go         # 文章自动标签 (LLM + 注册表验证)
│   │   │   └── eval_tagger.go    # 标签质量评估指标
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
│   │   ├── knowledge/            # 知识库组件 (ArticleList, CategoryTree, ArticleView)
│   │   ├── research/             # 调研组件 (ExplorerResearch, ErrorBoundary)
│   │   ├── admin/                # 管理组件 (ModelConfig, TaskMonitor, SystemStatus, SourceConfig, ArticleImport)
│   │   ├── chat/                 # 聊天组件 (FloatingChat, ChatMessage)
│   │   └── providers/            # QueryProvider
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
| 管理 | 标签管理 (注册表+生命周期) | TagsPage | /api/tags (list, stats, status, approve) |
| 标签 | 文章自动标签 (LLM Haiku) | - | tagger.go (TagArticle) |
| 标签 | 标签评估管道 | - | eval-tagger CLI |
| 基础 | CORS 中间件 | - | middleware.go |

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

Evaluated on 27 articles (2026-02-10). Tagger uses Claude Haiku with tag registry validation.

| 指标 | 目标 | 实际 | 状态 | 说明 |
|------|------|------|------|------|
| 每篇标签数 (3-7) | >95% in range | 63% (avg 3.0) | FAIL | 10 articles with <3 tags |
| 注册表合规率 | >95% | 91% (75/82) | FAIL | 4 off-registry: Lido, Rocket Pool, 分布式账本, 算法稳定币 |
| 标签复用率 (>=3篇) | >80% | 23% | FAIL | 10/44 tags reused (expected low at 27 articles) |
| 孤儿标签率 (仅1篇) | <15% | 64% | FAIL | 28/44 tags orphaned (expected high at 27 articles) |
| 最大覆盖率 | <=40% | 44% | FAIL | "DeFi" covers 12/27 articles |
| Gini 系数 | <=0.60 | 0.36 | PASS | Reasonable distribution |
| 空标签率 | 0% | 0% | PASS | All 27 articles tagged |

**Known issues**: (1) "DeFi" tag overused - appears in 44% of articles; prompt needs to discourage overly generic tags. (2) LLM occasionally invents tags (4 off-registry out of 82). (3) Reuse/orphan metrics inherently poor with only 27 articles.

**评估命令:**
```bash
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./cmd/eval-tagger --limit 50
/usr/local/go/bin/go run -C /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/backend ./cmd/eval-tagger --export review.md
```

### Tagging Lessons (2026-02-10)

- **LLM tag compliance**: Even with explicit "only choose from this list" prompts, Claude Haiku generates off-registry tags ~34% of the time. The `resolveTag()` function in `tagger.go` handles case-insensitive matching and parenthetical stripping, but fundamentally the LLM creates new tags. Consider retry-on-failure or post-filter with fuzzy matching.
- **Viper env expansion**: `config.yaml` uses `${ANTHROPIC_API_KEY}` syntax but Viper doesn't auto-expand. CLI tools must call `os.ExpandEnv()` on API key config values before creating LLM router.
- **Reuse/orphan metrics require scale**: With only 27 articles, tag reuse metrics are inherently low. These targets (>80% reuse, <15% orphan) are meaningful only at 100+ articles.
- **Bulk tagging CLI**: `cmd/bulk-tag` with `--force` flag re-tags all articles. Uses 500ms delay between articles to avoid rate limits. `--limit N` for partial runs.

## Plan Completion Protocol

After completing any implementation plan:
1. **整理文档**: 检查 `docs/` 和 `web3-insight/backend/docs/` 目录，将有价值的经验教训提取到 CLAUDE.md，将新功能描述更新到 README.md，然后删除这些文档文件
2. Delete the plan file from `docs/plans/`
3. Update relevant CLAUDE.md sections if project structure/features changed
4. **提交代码**: 将所有改动（代码 + 文档更新 + 文件清理）commit 到 git
