# Web3-Insight 完整实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建一个本地运行的 Web3 知识管理系统，帮助 Explorer 程序员快速建立 Web3 知识体系，跟进行业动态，了解竞品功能。

**Architecture:**
- 前端使用 Next.js 14 (App Router) + TypeScript + Tailwind + shadcn/ui
- 后端使用 Go (Gin + GORM + Asynq)
- 数据库使用 PostgreSQL + pgvector + Redis
- LLM 使用本地 Ollama 优先，云端 API 作为 Fallback
- 24 小时后台运行，持续收集和生成内容

**Tech Stack:**
- Frontend: Next.js 14, TypeScript, Tailwind CSS, shadcn/ui, React Query, Zustand, Recharts
- Backend: Go 1.21+, Gin, GORM, Asynq, Colly, WebSocket
- Database: PostgreSQL 15+ with pgvector, Redis 7+
- LLM: Ollama (本地), Claude API, OpenAI API, Gemini API (可选)
- Tools: Docker Compose (开发环境), pnpm (前端包管理)

---

## 项目概述

### 目标用户
- Crypto.com 新入职的 Explorer 程序员
- 需要快速建立 Web3 知识体系
- 需要持续跟进行业动态
- 需要了解各链 Explorer 产品功能

### 核心模块

1. **知识库 (Knowledge Base)** - Web3 技术演进史、概念深度解析、技术关系图谱
2. **新闻聚合 (News Aggregator)** - 多源抓取、智能分类、自动摘要
3. **Explorer 调研 (Competitor Intelligence)** - 主流 explorer 功能对比、用户反馈、链上数据分析
4. **即时研究 (Instant Research)** - 临时查询任意技术名词，智能保存到合适分类
5. **后台管理 (Admin Dashboard)** - 配置管理、任务监控、成本分析、日志查看

### 设计原则

- **简洁现代专业** - 参考 Linear/Notion/Stripe Docs 设计语言
- **配色** - 黑白灰 + 蓝色强调色 (#0066FF)
- **多语言输入，中文输出** - 收集英文信息，用中文呈现，保留专业术语
- **本地优先** - 本地模型处理简单任务，云端处理复杂任务
- **阅读时编辑** - 内容自动生成发布，阅读时按需修改/补充/重写

---

## 数据模型设计

### 核心表结构

```sql
-- 分类表 (树形结构 + 标签)
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    slug VARCHAR(100) UNIQUE NOT NULL,
    parent_id UUID REFERENCES categories(id),
    description TEXT,
    icon VARCHAR(50),
    sort_order INT DEFAULT 0,
    auto_created BOOLEAN DEFAULT FALSE,
    article_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 文章表
CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    content_html TEXT,
    summary TEXT,
    category_id UUID REFERENCES categories(id),
    tags TEXT[], -- PostgreSQL 数组
    status VARCHAR(20) DEFAULT 'published', -- draft, published, archived
    source_urls TEXT[], -- 原始来源
    source_language VARCHAR(10), -- 原文语言
    model_used VARCHAR(50), -- 生成时使用的模型
    generation_prompt TEXT, -- 生成时使用的 prompt
    view_count INT DEFAULT 0,
    embedding vector(1536), -- pgvector 向量
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 文章版本历史
CREATE TABLE article_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    edited_by VARCHAR(20) DEFAULT 'ai', -- ai, human
    change_summary TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 问答记录
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID REFERENCES articles(id),
    session_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL, -- user, assistant
    content TEXT NOT NULL,
    model_used VARCHAR(50),
    saved_to_article BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 新闻条目
CREATE TABLE news_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    original_title VARCHAR(500),
    content TEXT,
    summary TEXT,
    source_url VARCHAR(1000) UNIQUE NOT NULL,
    source_name VARCHAR(100),
    source_language VARCHAR(10),
    category VARCHAR(50), -- tech, finance, product, company, regulation
    tags TEXT[],
    published_at TIMESTAMP,
    fetched_at TIMESTAMP DEFAULT NOW(),
    processed BOOLEAN DEFAULT FALSE,
    embedding vector(1536)
);

-- Explorer 调研
CREATE TABLE explorer_research (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_name VARCHAR(100) NOT NULL,
    explorer_name VARCHAR(100) NOT NULL,
    explorer_url VARCHAR(500) NOT NULL,
    features JSONB, -- 功能列表
    screenshots TEXT[], -- 截图路径
    analysis TEXT, -- AI 分析
    popularity_score FLOAT,
    last_updated TIMESTAMP DEFAULT NOW()
);

-- 任务记录
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL, -- rss_sync, web_crawl, content_generate, classify
    status VARCHAR(20) DEFAULT 'pending', -- pending, running, completed, failed
    payload JSONB,
    result JSONB,
    error TEXT,
    model_used VARCHAR(50),
    tokens_used INT,
    cost_usd DECIMAL(10, 6),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 配置表
CREATE TABLE configs (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 数据源表
CREATE TABLE data_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- rss, api, crawl
    url VARCHAR(1000),
    config JSONB,
    enabled BOOLEAN DEFAULT TRUE,
    fetch_interval INT DEFAULT 3600, -- 秒
    last_fetched_at TIMESTAMP,
    last_error TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 初始分类结构

```yaml
categories:
  - name: "基础技术"
    name_en: "Fundamentals"
    children:
      - name: "区块链原理"
        name_en: "Blockchain Basics"
        children:
          - name: "共识机制"
            name_en: "Consensus Mechanisms"
          - name: "密码学基础"
            name_en: "Cryptography"
          - name: "数据结构"
            name_en: "Data Structures"
      - name: "智能合约"
        name_en: "Smart Contracts"
        children:
          - name: "EVM"
            name_en: "EVM"
          - name: "Solidity"
            name_en: "Solidity"
  - name: "扩容方案"
    name_en: "Scaling Solutions"
    children:
      - name: "Layer 2"
        name_en: "Layer 2"
        children:
          - name: "Optimistic Rollup"
            name_en: "Optimistic Rollup"
          - name: "ZK Rollup"
            name_en: "ZK Rollup"
      - name: "Layer 0"
        name_en: "Layer 0"
      - name: "侧链"
        name_en: "Sidechains"
  - name: "跨链技术"
    name_en: "Cross-chain"
    children:
      - name: "桥接协议"
        name_en: "Bridges"
      - name: "IBC"
        name_en: "IBC"
  - name: "生态系统"
    name_en: "Ecosystems"
    children:
      - name: "以太坊"
        name_en: "Ethereum"
      - name: "Cosmos"
        name_en: "Cosmos"
      - name: "Solana"
        name_en: "Solana"
  - name: "Explorer 技术"
    name_en: "Explorer Tech"
    children:
      - name: "索引器"
        name_en: "Indexer"
      - name: "RPC 节点"
        name_en: "RPC Nodes"
      - name: "数据展示"
        name_en: "Data Visualization"
  - name: "行业动态"
    name_en: "Industry News"
    children:
      - name: "技术新闻"
        name_en: "Tech News"
      - name: "金融新闻"
        name_en: "Finance News"
      - name: "监管政策"
        name_en: "Regulations"
```

---

## LLM 路由策略

### 模型配置

```yaml
models:
  local:
    - name: "llama3:70b"
      memory: "40GB"
      purpose: "通用生成"
      priority: 1
    - name: "qwen2.5:32b"
      memory: "20GB"
      purpose: "中文优化"
      priority: 2
    - name: "mistral:7b"
      memory: "4GB"
      purpose: "快速任务"
      priority: 3

  cloud:
    - name: "claude-sonnet-4-20250514"
      provider: "anthropic"
      purpose: "复杂分析"
      cost_per_1k_input: 0.003
      cost_per_1k_output: 0.015
    - name: "claude-haiku"
      provider: "anthropic"
      purpose: "简单任务"
      cost_per_1k_input: 0.00025
      cost_per_1k_output: 0.00125
    - name: "gpt-4o"
      provider: "openai"
      purpose: "复杂分析"
    - name: "gpt-4o-mini"
      provider: "openai"
      purpose: "简单任务"

routing:
  content_generation_simple:
    primary: "llama3:70b"
    fallback: "claude-haiku"
  content_generation_complex:
    primary: "claude-sonnet-4-20250514"
    fallback: null
  summarization:
    primary: "qwen2.5:32b"
    fallback: "claude-haiku"
  classification:
    primary: "qwen2.5:32b"
    fallback: "claude-haiku"
  chat:
    primary: "llama3:70b"
    fallback: "claude-sonnet-4-20250514"
  translation:
    primary: "qwen2.5:32b"
    fallback: "claude-haiku"
```

### Prompt 模板

```yaml
prompts:
  knowledge_article:
    system: |
      你是一个 Web3 技术专家，正在为一位刚入职区块链公司的程序员撰写技术文档。

      要求：
      1. 使用中文撰写，保持专业性和准确性
      2. 专业术语格式：英文术语 (中文翻译)，如 "Rollup (卷叠)"
      3. 首次出现的缩写需要展开，如 "EVM (Ethereum Virtual Machine, 以太坊虚拟机)"
      4. 内容结构：概念定义 → 工作原理 → 技术细节 → 优劣势 → 实际应用 → 相关技术
      5. 深入浅出，既要有高层概述，也要有底层细节

    user: |
      请撰写关于「{topic}」的技术文章。

      {context}

      要求：
      - 文章长度：2000-4000 字
      - 包含代码示例（如适用）
      - 标注原始来源

  classification:
    system: |
      你是一个 Web3 内容分类专家。根据文章内容，推荐最合适的分类。

      可用分类：
      {categories}

    user: |
      请为以下内容推荐分类：

      标题：{title}
      内容摘要：{summary}

      返回 JSON 格式：
      {
        "primary_category": "分类路径",
        "secondary_categories": ["分类路径1", "分类路径2"],
        "suggested_tags": ["标签1", "标签2"],
        "new_category_suggestion": null 或 {"name": "新分类名", "parent": "父分类路径"}
      }

  chat_with_context:
    system: |
      你是一个 Web3 技术助手。用户正在阅读一篇关于「{article_title}」的文章，并对内容有疑问。

      文章内容：
      {article_content}

      请基于文章内容回答用户的问题。如果问题超出文章范围，可以补充相关知识。
      使用中文回答，保持专业术语的一致性。

  news_summary:
    system: |
      你是一个 Web3 新闻编辑。请将以下英文新闻翻译并总结为中文。

      要求：
      1. 保留原文的关键信息
      2. 专业术语使用「英文 (中文)」格式
      3. 总结长度：100-200 字
      4. 标注新闻类别：tech/finance/product/company/regulation
```

---

## 数据源配置

### 免费数据源

```yaml
free_sources:
  rss:
    - name: "Ethereum Blog"
      url: "https://blog.ethereum.org/feed.xml"
      interval: 3600
      category: "tech"
    - name: "Vitalik Buterin"
      url: "https://vitalik.eth.limo/feed.xml"
      interval: 21600
      category: "tech"
    - name: "zkSync Blog"
      url: "https://blog.zksync.io/rss"
      interval: 3600
      category: "tech"
    - name: "Paradigm Research"
      url: "https://www.paradigm.xyz/feed.xml"
      interval: 86400
      category: "research"
    - name: "a]16z Crypto"
      url: "https://a16zcrypto.com/feed/"
      interval: 86400
      category: "research"

  api:
    - name: "CoinGecko"
      base_url: "https://api.coingecko.com/api/v3"
      rate_limit: 30  # per minute
      endpoints:
        - "/coins/list"
        - "/coins/{id}"
    - name: "DeFiLlama"
      base_url: "https://api.llama.fi"
      rate_limit: null  # unlimited
      endpoints:
        - "/protocols"
        - "/tvl/{protocol}"

  github:
    repos:
      - "ethereum/go-ethereum"
      - "matter-labs/zksync-era"
      - "cosmos/cosmos-sdk"
      - "blockscout/blockscout"
    interval: 86400

crawl_sources:
  explorer_research:
    - name: "Etherscan"
      url: "https://etherscan.io"
      interval: 604800  # weekly
    - name: "Blockscout"
      url: "https://blockscout.com"
      interval: 604800
    - name: "Subscan"
      url: "https://www.subscan.io"
      interval: 604800

  politeness:
    delay_range: [30, 60]  # seconds
    concurrent: 2
    user_agent_rotation: true
    respect_robots_txt: true

paid_sources:
  tavily:
    enabled: false
    api_key_env: "TAVILY_API_KEY"
    cost_per_request: 0.01
  serpapi:
    enabled: false
    api_key_env: "SERPAPI_KEY"
    cost_per_request: 0.01
```

---

## UI 设计规范

### 配色方案

```css
:root {
  /* 主色调 */
  --color-text-primary: #000000;
  --color-text-secondary: #333333;
  --color-text-tertiary: #666666;
  --color-text-muted: #999999;
  --color-background: #FFFFFF;
  --color-background-secondary: #F9FAFB;
  --color-border: #E5E7EB;

  /* 强调色 */
  --color-accent: #0066FF;
  --color-accent-hover: #0052CC;

  /* 状态色 */
  --color-success: #22C55E;
  --color-warning: #F59E0B;
  --color-error: #EF4444;
  --color-info: #6366F1;

  /* 暗色模式 (可选，后期) */
  --color-dark-bg: #0A0A0A;
  --color-dark-text: #EDEDED;
}
```

### 字体

```css
:root {
  --font-sans: "Inter", "Source Han Sans SC", "Noto Sans SC", system-ui, sans-serif;
  --font-mono: "JetBrains Mono", "Fira Code", monospace;
}
```

### 间距规范

```css
:root {
  --spacing-1: 4px;
  --spacing-2: 8px;
  --spacing-3: 12px;
  --spacing-4: 16px;
  --spacing-6: 24px;
  --spacing-8: 32px;
  --spacing-12: 48px;
  --spacing-16: 64px;
}
```

---

## 实现阶段

### Phase 1: 基础 UI 和后台框架 (当前阶段)

**前端:**
- [ ] 项目初始化 (Next.js + TypeScript + Tailwind + shadcn/ui)
- [ ] 布局组件 (侧边栏、顶部导航、主内容区)
- [ ] 知识库浏览页面 (分类树 + 文章列表 + 文章详情)
- [ ] 悬浮问答窗口 (可拖拽、流式响应、保存对话)
- [ ] 即时研究页面 (搜索 + AI 生成 + 保存到分类)
- [ ] 后台管理 UI (配置、任务监控、成本、日志)
- [ ] 搜索功能 (全文 + 语义)

**后端:**
- [ ] 项目初始化 (Go + Gin + GORM)
- [ ] 数据库迁移和初始数据
- [ ] RESTful API (文章、分类、配置、任务)
- [ ] WebSocket 问答服务
- [ ] LLM 路由层 (Ollama + Claude/OpenAI 适配器)
- [ ] 后台 Worker 框架 (Asynq 任务调度)
- [ ] 配置热加载

**基础设施:**
- [ ] Docker Compose (PostgreSQL + Redis + Ollama)
- [ ] 开发环境配置

### Phase 2: 数据收集和内容生成

**爬虫模块:**
- [ ] RSS 订阅解析和同步
- [ ] 网页爬取 (Colly + 礼貌策略)
- [ ] GitHub 仓库监控
- [ ] API 数据源集成

**内容生成:**
- [ ] 知识文章自动生成
- [ ] 新闻摘要和翻译
- [ ] 分类自动推断
- [ ] 向量嵌入生成

### Phase 3: Explorer 调研和高级功能

**Explorer 调研:**
- [ ] Explorer 页面爬取和截图
- [ ] 功能对比分析
- [ ] 自动生成调研报告

**高级功能:**
- [ ] 知识图谱可视化
- [ ] 学习进度追踪
- [ ] 付费搜索 API 集成
- [ ] 暗色模式

---

## Phase 1 详细任务

### Task 1: 项目结构初始化

**Files:**
- Create: `web3-insight/` (项目根目录)
- Create: `web3-insight/frontend/` (Next.js 前端)
- Create: `web3-insight/backend/` (Go 后端)
- Create: `web3-insight/docker-compose.yml`
- Create: `web3-insight/.gitignore`
- Create: `web3-insight/README.md`

**Step 1: 创建项目根目录和基础文件**

```bash
mkdir -p web3-insight/{frontend,backend,docs}
cd web3-insight
```

**Step 2: 创建 docker-compose.yml**

```yaml
version: '3.8'

services:
  postgres:
    image: pgvector/pgvector:pg16
    container_name: web3-insight-db
    environment:
      POSTGRES_USER: web3insight
      POSTGRES_PASSWORD: web3insight_dev
      POSTGRES_DB: web3insight
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    container_name: web3-insight-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

**Step 3: 创建 .gitignore**

```gitignore
# Dependencies
node_modules/
vendor/

# Build
.next/
out/
dist/
bin/

# Environment
.env
.env.local
.env.*.local

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
*.log
logs/

# Database
*.db
*.sqlite

# Temp
tmp/
temp/
```

**Step 4: 初始化 Git**

```bash
git init
git add .
git commit -m "chore: initial project structure"
```

---

### Task 2: 前端项目初始化

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/tsconfig.json`
- Create: `frontend/tailwind.config.ts`
- Create: `frontend/next.config.js`
- Create: `frontend/app/layout.tsx`
- Create: `frontend/app/page.tsx`

**Step 1: 初始化 Next.js 项目**

```bash
cd frontend
pnpm create next-app@latest . --typescript --tailwind --eslint --app --src-dir=false --import-alias="@/*"
```

**Step 2: 安装依赖**

```bash
pnpm add @tanstack/react-query zustand recharts lucide-react class-variance-authority clsx tailwind-merge
pnpm add -D @types/node
```

**Step 3: 安装 shadcn/ui**

```bash
pnpm dlx shadcn@latest init
# 选择: New York style, Zinc color, CSS variables: yes
```

**Step 4: 添加 shadcn 组件**

```bash
pnpm dlx shadcn@latest add button card input textarea select dialog dropdown-menu scroll-area separator tabs toast tooltip
```

**Step 5: 创建基础布局**

`app/layout.tsx`:
```tsx
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { QueryProvider } from '@/components/providers/query-provider'
import { Toaster } from '@/components/ui/toaster'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Web3 Insight',
  description: 'Your personal Web3 knowledge base',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="zh-CN">
      <body className={inter.className}>
        <QueryProvider>
          {children}
          <Toaster />
        </QueryProvider>
      </body>
    </html>
  )
}
```

**Step 6: Commit**

```bash
git add .
git commit -m "feat(frontend): initialize Next.js with shadcn/ui"
```

---

### Task 3: 后端项目初始化

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/cmd/worker/main.go`
- Create: `backend/internal/api/router.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/config/config.yaml`

**Step 1: 初始化 Go 模块**

```bash
cd backend
go mod init github.com/user/web3-insight
```

**Step 2: 安装依赖**

```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/hibiken/asynq
go get github.com/go-redis/redis/v8
go get github.com/spf13/viper
go get github.com/gorilla/websocket
go get go.uber.org/zap
```

**Step 3: 创建配置文件**

`config/config.yaml`:
```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 5432
  user: "web3insight"
  password: "web3insight_dev"
  dbname: "web3insight"
  sslmode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

llm:
  default_local: "llama3:70b"
  ollama_host: "http://localhost:11434"

  claude:
    enabled: true
    api_key: "${ANTHROPIC_API_KEY}"
    default_model: "claude-sonnet-4-20250514"

  openai:
    enabled: false
    api_key: "${OPENAI_API_KEY}"
    default_model: "gpt-4o"

worker:
  concurrency: 5
  queues:
    critical: 6
    default: 3
    low: 1
```

**Step 4: 创建主入口**

`cmd/server/main.go`:
```go
package main

import (
    "log"

    "github.com/user/web3-insight/internal/api"
    "github.com/user/web3-insight/internal/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    router := api.NewRouter(cfg)

    addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
    log.Printf("Server starting on %s", addr)

    if err := router.Run(addr); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
```

**Step 5: Commit**

```bash
git add .
git commit -m "feat(backend): initialize Go project with Gin"
```

---

### Task 4: 数据库模型和迁移

**Files:**
- Create: `backend/internal/model/article.go`
- Create: `backend/internal/model/category.go`
- Create: `backend/internal/model/chat.go`
- Create: `backend/internal/model/task.go`
- Create: `backend/internal/model/config.go`
- Create: `backend/internal/database/migrate.go`
- Create: `backend/internal/database/seed.go`

**Step 1: 创建数据模型**

`internal/model/category.go`:
```go
package model

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Category struct {
    ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    Name         string     `gorm:"size:100;not null" json:"name"`
    NameEn       string     `gorm:"size:100" json:"nameEn"`
    Slug         string     `gorm:"size:100;uniqueIndex;not null" json:"slug"`
    ParentID     *uuid.UUID `gorm:"type:uuid" json:"parentId"`
    Parent       *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Children     []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
    Description  string     `gorm:"type:text" json:"description"`
    Icon         string     `gorm:"size:50" json:"icon"`
    SortOrder    int        `gorm:"default:0" json:"sortOrder"`
    AutoCreated  bool       `gorm:"default:false" json:"autoCreated"`
    ArticleCount int        `gorm:"default:0" json:"articleCount"`
    CreatedAt    time.Time  `json:"createdAt"`
    UpdatedAt    time.Time  `json:"updatedAt"`
}

func (Category) TableName() string {
    return "categories"
}
```

`internal/model/article.go`:
```go
package model

import (
    "time"

    "github.com/google/uuid"
    "github.com/lib/pq"
    "github.com/pgvector/pgvector-go"
)

type Article struct {
    ID               uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    Title            string          `gorm:"size:500;not null" json:"title"`
    Slug             string          `gorm:"size:500;uniqueIndex;not null" json:"slug"`
    Content          string          `gorm:"type:text;not null" json:"content"`
    ContentHTML      string          `gorm:"type:text" json:"contentHtml"`
    Summary          string          `gorm:"type:text" json:"summary"`
    CategoryID       *uuid.UUID      `gorm:"type:uuid" json:"categoryId"`
    Category         *Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
    Tags             pq.StringArray  `gorm:"type:text[]" json:"tags"`
    Status           string          `gorm:"size:20;default:'published'" json:"status"`
    SourceURLs       pq.StringArray  `gorm:"type:text[]" json:"sourceUrls"`
    SourceLanguage   string          `gorm:"size:10" json:"sourceLanguage"`
    ModelUsed        string          `gorm:"size:50" json:"modelUsed"`
    GenerationPrompt string          `gorm:"type:text" json:"generationPrompt"`
    ViewCount        int             `gorm:"default:0" json:"viewCount"`
    Embedding        pgvector.Vector `gorm:"type:vector(1536)" json:"-"`
    CreatedAt        time.Time       `json:"createdAt"`
    UpdatedAt        time.Time       `json:"updatedAt"`
}

func (Article) TableName() string {
    return "articles"
}
```

**Step 2: 创建迁移脚本**

`internal/database/migrate.go`:
```go
package database

import (
    "github.com/user/web3-insight/internal/model"
    "gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
    // Enable pgvector extension
    db.Exec("CREATE EXTENSION IF NOT EXISTS vector")

    return db.AutoMigrate(
        &model.Category{},
        &model.Article{},
        &model.ArticleVersion{},
        &model.ChatMessage{},
        &model.NewsItem{},
        &model.ExplorerResearch{},
        &model.Task{},
        &model.Config{},
        &model.DataSource{},
    )
}
```

**Step 3: 创建种子数据**

`internal/database/seed.go`:
```go
package database

import (
    "github.com/user/web3-insight/internal/model"
    "gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
    categories := []model.Category{
        {Name: "基础技术", NameEn: "Fundamentals", Slug: "fundamentals"},
        {Name: "扩容方案", NameEn: "Scaling Solutions", Slug: "scaling"},
        {Name: "跨链技术", NameEn: "Cross-chain", Slug: "cross-chain"},
        {Name: "生态系统", NameEn: "Ecosystems", Slug: "ecosystems"},
        {Name: "Explorer 技术", NameEn: "Explorer Tech", Slug: "explorer-tech"},
        {Name: "行业动态", NameEn: "Industry News", Slug: "news"},
    }

    for _, cat := range categories {
        db.FirstOrCreate(&cat, model.Category{Slug: cat.Slug})
    }

    return nil
}
```

**Step 4: Commit**

```bash
git add .
git commit -m "feat(backend): add database models and migrations"
```

---

### Task 5: RESTful API 实现

**Files:**
- Create: `backend/internal/api/router.go`
- Create: `backend/internal/api/article.go`
- Create: `backend/internal/api/category.go`
- Create: `backend/internal/api/config.go`
- Create: `backend/internal/api/task.go`
- Create: `backend/internal/api/search.go`
- Create: `backend/internal/repository/article.go`
- Create: `backend/internal/repository/category.go`

**Step 1: 创建路由**

`internal/api/router.go`:
```go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/user/web3-insight/internal/config"
)

func NewRouter(cfg *config.Config) *gin.Engine {
    r := gin.Default()

    // CORS
    r.Use(corsMiddleware())

    // API routes
    api := r.Group("/api")
    {
        // Articles
        articles := api.Group("/articles")
        {
            articles.GET("", listArticles)
            articles.GET("/:id", getArticle)
            articles.POST("", createArticle)
            articles.PUT("/:id", updateArticle)
            articles.DELETE("/:id", deleteArticle)
            articles.POST("/:id/regenerate", regenerateArticle)
        }

        // Categories
        categories := api.Group("/categories")
        {
            categories.GET("", listCategories)
            categories.GET("/tree", getCategoryTree)
            categories.POST("", createCategory)
            categories.PUT("/:id", updateCategory)
            categories.DELETE("/:id", deleteCategory)
        }

        // Search
        api.GET("/search", search)

        // Config
        api.GET("/config", getConfig)
        api.PUT("/config", updateConfig)

        // Tasks
        tasks := api.Group("/tasks")
        {
            tasks.GET("", listTasks)
            tasks.GET("/stats", getTaskStats)
            tasks.POST("/:id/cancel", cancelTask)
        }

        // Instant research
        api.POST("/research", instantResearch)
    }

    // WebSocket for chat
    r.GET("/ws/chat", handleChatWebSocket)

    return r
}
```

**Step 2-5: 实现各个 handler (详见完整代码)**

**Step 6: Commit**

```bash
git add .
git commit -m "feat(backend): implement RESTful API endpoints"
```

---

### Task 6: LLM 路由层

**Files:**
- Create: `backend/internal/llm/router.go`
- Create: `backend/internal/llm/ollama.go`
- Create: `backend/internal/llm/claude.go`
- Create: `backend/internal/llm/openai.go`
- Create: `backend/internal/llm/types.go`

**Step 1: 定义接口**

`internal/llm/types.go`:
```go
package llm

type LLMAdapter interface {
    Name() string
    Type() string // "local" or "cloud"
    Generate(prompt string, opts *GenerateOptions) (string, error)
    GenerateStream(prompt string, opts *GenerateOptions) (<-chan string, error)
    IsAvailable() bool
    EstimateCost(inputTokens, outputTokens int) float64
}

type GenerateOptions struct {
    SystemPrompt string
    MaxTokens    int
    Temperature  float64
    TopP         float64
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

**Step 2: 实现 Ollama 适配器**

`internal/llm/ollama.go`:
```go
package llm

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type OllamaAdapter struct {
    host  string
    model string
}

func NewOllamaAdapter(host, model string) *OllamaAdapter {
    return &OllamaAdapter{host: host, model: model}
}

func (o *OllamaAdapter) Name() string { return o.model }
func (o *OllamaAdapter) Type() string { return "local" }

func (o *OllamaAdapter) Generate(prompt string, opts *GenerateOptions) (string, error) {
    payload := map[string]interface{}{
        "model":  o.model,
        "prompt": prompt,
        "stream": false,
    }

    if opts != nil && opts.SystemPrompt != "" {
        payload["system"] = opts.SystemPrompt
    }

    body, _ := json.Marshal(payload)
    resp, err := http.Post(o.host+"/api/generate", "application/json", bytes.NewReader(body))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Response string `json:"response"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    return result.Response, nil
}

func (o *OllamaAdapter) GenerateStream(prompt string, opts *GenerateOptions) (<-chan string, error) {
    ch := make(chan string)

    go func() {
        defer close(ch)

        payload := map[string]interface{}{
            "model":  o.model,
            "prompt": prompt,
            "stream": true,
        }

        if opts != nil && opts.SystemPrompt != "" {
            payload["system"] = opts.SystemPrompt
        }

        body, _ := json.Marshal(payload)
        resp, err := http.Post(o.host+"/api/generate", "application/json", bytes.NewReader(body))
        if err != nil {
            return
        }
        defer resp.Body.Close()

        decoder := json.NewDecoder(resp.Body)
        for {
            var chunk struct {
                Response string `json:"response"`
                Done     bool   `json:"done"`
            }
            if err := decoder.Decode(&chunk); err == io.EOF {
                break
            } else if err != nil {
                break
            }
            ch <- chunk.Response
            if chunk.Done {
                break
            }
        }
    }()

    return ch, nil
}

func (o *OllamaAdapter) IsAvailable() bool {
    resp, err := http.Get(o.host + "/api/tags")
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200
}

func (o *OllamaAdapter) EstimateCost(inputTokens, outputTokens int) float64 {
    return 0 // Local model, no cost
}
```

**Step 3: 实现 Claude 适配器**

`internal/llm/claude.go`:
```go
package llm

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

type ClaudeAdapter struct {
    apiKey string
    model  string
}

func NewClaudeAdapter(model string) *ClaudeAdapter {
    return &ClaudeAdapter{
        apiKey: os.Getenv("ANTHROPIC_API_KEY"),
        model:  model,
    }
}

func (c *ClaudeAdapter) Name() string { return c.model }
func (c *ClaudeAdapter) Type() string { return "cloud" }

func (c *ClaudeAdapter) Generate(prompt string, opts *GenerateOptions) (string, error) {
    messages := []map[string]string{
        {"role": "user", "content": prompt},
    }

    payload := map[string]interface{}{
        "model":      c.model,
        "max_tokens": 4096,
        "messages":   messages,
    }

    if opts != nil && opts.SystemPrompt != "" {
        payload["system"] = opts.SystemPrompt
    }

    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-api-key", c.apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Content []struct {
            Text string `json:"text"`
        } `json:"content"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    if len(result.Content) > 0 {
        return result.Content[0].Text, nil
    }
    return "", fmt.Errorf("empty response")
}

// GenerateStream implementation with SSE...

func (c *ClaudeAdapter) IsAvailable() bool {
    return c.apiKey != ""
}

func (c *ClaudeAdapter) EstimateCost(inputTokens, outputTokens int) float64 {
    // Claude Sonnet pricing
    inputCost := float64(inputTokens) / 1000 * 0.003
    outputCost := float64(outputTokens) / 1000 * 0.015
    return inputCost + outputCost
}
```

**Step 4: 实现路由器**

`internal/llm/router.go`:
```go
package llm

import (
    "fmt"
    "sync"
)

type Router struct {
    adapters map[string]LLMAdapter
    routes   map[string][]string // task -> [primary, fallback...]
    mu       sync.RWMutex
}

func NewRouter() *Router {
    return &Router{
        adapters: make(map[string]LLMAdapter),
        routes:   make(map[string][]string),
    }
}

func (r *Router) RegisterAdapter(name string, adapter LLMAdapter) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.adapters[name] = adapter
}

func (r *Router) SetRoute(task string, models []string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.routes[task] = models
}

func (r *Router) Generate(task, prompt string, opts *GenerateOptions) (string, string, error) {
    r.mu.RLock()
    models := r.routes[task]
    r.mu.RUnlock()

    for _, modelName := range models {
        adapter, ok := r.adapters[modelName]
        if !ok {
            continue
        }

        if !adapter.IsAvailable() {
            continue
        }

        result, err := adapter.Generate(prompt, opts)
        if err == nil {
            return result, modelName, nil
        }
    }

    return "", "", fmt.Errorf("all models failed for task: %s", task)
}
```

**Step 5: Commit**

```bash
git add .
git commit -m "feat(backend): implement LLM router with Ollama and Claude adapters"
```

---

### Task 7: WebSocket 问答服务

**Files:**
- Create: `backend/internal/api/chat_ws.go`
- Create: `backend/internal/service/chat.go`

**Step 1: 实现 WebSocket handler**

`internal/api/chat_ws.go`:
```go
package api

import (
    "encoding/json"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/user/web3-insight/internal/service"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins in dev
    },
}

type ChatRequest struct {
    ArticleID    string    `json:"articleId"`
    Message      string    `json:"message"`
    SelectedText string    `json:"selectedText,omitempty"`
    SessionID    string    `json:"sessionId"`
}

type ChatResponse struct {
    Type    string `json:"type"` // "chunk", "done", "error"
    Content string `json:"content"`
    Model   string `json:"model,omitempty"`
}

func handleChatWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    chatService := service.NewChatService()

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            break
        }

        var req ChatRequest
        if err := json.Unmarshal(message, &req); err != nil {
            conn.WriteJSON(ChatResponse{Type: "error", Content: "Invalid request"})
            continue
        }

        // Stream response
        stream, model, err := chatService.Chat(req.ArticleID, req.Message, req.SelectedText)
        if err != nil {
            conn.WriteJSON(ChatResponse{Type: "error", Content: err.Error()})
            continue
        }

        for chunk := range stream {
            conn.WriteJSON(ChatResponse{Type: "chunk", Content: chunk})
        }

        conn.WriteJSON(ChatResponse{Type: "done", Model: model})
    }
}
```

**Step 2: 实现 Chat 服务**

`internal/service/chat.go`:
```go
package service

import (
    "github.com/user/web3-insight/internal/llm"
    "github.com/user/web3-insight/internal/repository"
)

type ChatService struct {
    llmRouter  *llm.Router
    articleRepo *repository.ArticleRepository
}

func NewChatService() *ChatService {
    return &ChatService{
        llmRouter:   llm.GetDefaultRouter(),
        articleRepo: repository.NewArticleRepository(),
    }
}

func (s *ChatService) Chat(articleID, message, selectedText string) (<-chan string, string, error) {
    // Get article context
    article, err := s.articleRepo.GetByID(articleID)
    if err != nil {
        return nil, "", err
    }

    // Build prompt
    systemPrompt := buildChatSystemPrompt(article.Title, article.Content)
    userPrompt := message
    if selectedText != "" {
        userPrompt = fmt.Sprintf("关于「%s」这部分内容：%s", selectedText, message)
    }

    // Stream from LLM
    opts := &llm.GenerateOptions{
        SystemPrompt: systemPrompt,
        MaxTokens:    2048,
        Temperature:  0.7,
    }

    adapter := s.llmRouter.GetAdapter("chat")
    stream, err := adapter.GenerateStream(userPrompt, opts)

    return stream, adapter.Name(), err
}

func buildChatSystemPrompt(title, content string) string {
    return fmt.Sprintf(`你是一个 Web3 技术助手。用户正在阅读一篇关于「%s」的文章，并对内容有疑问。

文章内容：
%s

请基于文章内容回答用户的问题。如果问题超出文章范围，可以补充相关知识。
使用中文回答，保持专业术语的一致性。`, title, content)
}
```

**Step 3: Commit**

```bash
git add .
git commit -m "feat(backend): implement WebSocket chat service with streaming"
```

---

### Task 8: 后台 Worker 框架

**Files:**
- Create: `backend/cmd/worker/main.go`
- Create: `backend/internal/worker/scheduler.go`
- Create: `backend/internal/worker/tasks.go`

**Step 1: 创建 Worker 入口**

`cmd/worker/main.go`:
```go
package main

import (
    "log"

    "github.com/hibiken/asynq"
    "github.com/user/web3-insight/internal/config"
    "github.com/user/web3-insight/internal/worker"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    redisOpt := asynq.RedisClientOpt{
        Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
    }

    srv := asynq.NewServer(redisOpt, asynq.Config{
        Concurrency: cfg.Worker.Concurrency,
        Queues: map[string]int{
            "critical": 6,
            "default":  3,
            "low":      1,
        },
    })

    mux := worker.NewTaskMux()

    log.Println("Worker starting...")
    if err := srv.Run(mux); err != nil {
        log.Fatalf("Worker failed: %v", err)
    }
}
```

**Step 2: 创建任务处理器**

`internal/worker/tasks.go`:
```go
package worker

import (
    "context"
    "encoding/json"

    "github.com/hibiken/asynq"
)

const (
    TaskTypeContentGenerate = "content:generate"
    TaskTypeRSSSync         = "rss:sync"
    TaskTypeWebCrawl        = "web:crawl"
    TaskTypeClassify        = "content:classify"
)

func NewTaskMux() *asynq.ServeMux {
    mux := asynq.NewServeMux()

    mux.HandleFunc(TaskTypeContentGenerate, handleContentGenerate)
    mux.HandleFunc(TaskTypeRSSSync, handleRSSSync)
    mux.HandleFunc(TaskTypeWebCrawl, handleWebCrawl)
    mux.HandleFunc(TaskTypeClassify, handleClassify)

    return mux
}

func handleContentGenerate(ctx context.Context, t *asynq.Task) error {
    var payload struct {
        Topic      string `json:"topic"`
        CategoryID string `json:"categoryId"`
    }
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return err
    }

    // TODO: Implement in Phase 2
    return nil
}

func handleRSSSync(ctx context.Context, t *asynq.Task) error {
    // TODO: Implement in Phase 2
    return nil
}

func handleWebCrawl(ctx context.Context, t *asynq.Task) error {
    // TODO: Implement in Phase 2
    return nil
}

func handleClassify(ctx context.Context, t *asynq.Task) error {
    // TODO: Implement in Phase 2
    return nil
}
```

**Step 3: 创建调度器**

`internal/worker/scheduler.go`:
```go
package worker

import (
    "time"

    "github.com/hibiken/asynq"
)

type Scheduler struct {
    scheduler *asynq.Scheduler
}

func NewScheduler(redisOpt asynq.RedisClientOpt) *Scheduler {
    return &Scheduler{
        scheduler: asynq.NewScheduler(redisOpt, nil),
    }
}

func (s *Scheduler) RegisterTasks() {
    // RSS sync every hour
    s.scheduler.Register("0 * * * *", asynq.NewTask(TaskTypeRSSSync, nil))

    // Content generation every 6 hours
    s.scheduler.Register("0 */6 * * *", asynq.NewTask(TaskTypeContentGenerate, nil))
}

func (s *Scheduler) Run() error {
    return s.scheduler.Run()
}
```

**Step 4: Commit**

```bash
git add .
git commit -m "feat(backend): implement background worker framework with Asynq"
```

---

### Task 9: 前端布局组件

**Files:**
- Create: `frontend/components/layout/sidebar.tsx`
- Create: `frontend/components/layout/header.tsx`
- Create: `frontend/components/layout/main-layout.tsx`
- Create: `frontend/components/knowledge/category-tree.tsx`

**Step 1: 创建侧边栏**

`components/layout/sidebar.tsx`:
```tsx
'use client'

import { useState } from 'react'
import Link from 'next/link'
import { cn } from '@/lib/utils'
import { ChevronRight, FileText, Newspaper, Search, Settings } from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { CategoryTree } from '@/components/knowledge/category-tree'

interface SidebarProps {
  className?: string
}

export function Sidebar({ className }: SidebarProps) {
  return (
    <aside className={cn(
      "w-64 border-r border-border bg-background flex flex-col",
      className
    )}>
      {/* Logo */}
      <div className="h-14 flex items-center px-4 border-b border-border">
        <Link href="/" className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-accent flex items-center justify-center">
            <span className="text-white font-bold text-sm">W3</span>
          </div>
          <span className="font-semibold">Web3 Insight</span>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="p-2">
        <Link href="/knowledge" className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-accent/10 text-sm">
          <FileText className="w-4 h-4" />
          知识库
        </Link>
        <Link href="/news" className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-accent/10 text-sm">
          <Newspaper className="w-4 h-4" />
          新闻
        </Link>
        <Link href="/research" className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-accent/10 text-sm">
          <Search className="w-4 h-4" />
          即时研究
        </Link>
        <Link href="/admin" className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-accent/10 text-sm">
          <Settings className="w-4 h-4" />
          后台管理
        </Link>
      </nav>

      {/* Category Tree */}
      <div className="flex-1 overflow-hidden">
        <div className="px-4 py-2 text-xs font-medium text-muted-foreground uppercase">
          分类
        </div>
        <ScrollArea className="h-full px-2">
          <CategoryTree />
        </ScrollArea>
      </div>

      {/* Recent */}
      <div className="border-t border-border p-4">
        <div className="text-xs font-medium text-muted-foreground uppercase mb-2">
          最近阅读
        </div>
        <div className="space-y-1 text-sm text-muted-foreground">
          <div className="truncate hover:text-foreground cursor-pointer">
            zkSync 工作原理
          </div>
          <div className="truncate hover:text-foreground cursor-pointer">
            Cosmos IBC 详解
          </div>
        </div>
      </div>
    </aside>
  )
}
```

**Step 2: 创建顶部导航**

`components/layout/header.tsx`:
```tsx
'use client'

import { useState } from 'react'
import { Search, Settings } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import Link from 'next/link'

export function Header() {
  const [searchQuery, setSearchQuery] = useState('')

  return (
    <header className="h-14 border-b border-border bg-background flex items-center justify-between px-6">
      {/* Breadcrumb or Title */}
      <div className="text-sm text-muted-foreground">
        知识库 / Layer 2 / ZK Rollup
      </div>

      {/* Search */}
      <div className="flex items-center gap-4">
        <div className="relative w-64">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="搜索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        <Link href="/admin">
          <Button variant="ghost" size="icon">
            <Settings className="w-4 h-4" />
          </Button>
        </Link>
      </div>
    </header>
  )
}
```

**Step 3: 创建主布局**

`components/layout/main-layout.tsx`:
```tsx
import { Sidebar } from './sidebar'
import { Header } from './header'

interface MainLayoutProps {
  children: React.ReactNode
}

export function MainLayout({ children }: MainLayoutProps) {
  return (
    <div className="h-screen flex">
      <Sidebar />
      <div className="flex-1 flex flex-col overflow-hidden">
        <Header />
        <main className="flex-1 overflow-auto bg-background-secondary">
          {children}
        </main>
      </div>
    </div>
  )
}
```

**Step 4: Commit**

```bash
git add .
git commit -m "feat(frontend): add layout components (sidebar, header, main-layout)"
```

---

### Task 10: 悬浮问答窗口

**Files:**
- Create: `frontend/components/chat/floating-chat.tsx`
- Create: `frontend/components/chat/chat-message.tsx`
- Create: `frontend/hooks/use-chat.ts`
- Create: `frontend/lib/websocket.ts`

**Step 1: 创建 WebSocket hook**

`lib/websocket.ts`:
```tsx
export function createChatWebSocket(onMessage: (data: any) => void) {
  const ws = new WebSocket('ws://localhost:8080/ws/chat')

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    onMessage(data)
  }

  return ws
}
```

`hooks/use-chat.ts`:
```tsx
'use client'

import { useState, useCallback, useRef, useEffect } from 'react'
import { createChatWebSocket } from '@/lib/websocket'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  model?: string
}

export function useChat(articleId: string) {
  const [messages, setMessages] = useState<Message[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [currentResponse, setCurrentResponse] = useState('')
  const wsRef = useRef<WebSocket | null>(null)
  const sessionId = useRef(crypto.randomUUID())

  useEffect(() => {
    wsRef.current = createChatWebSocket((data) => {
      if (data.type === 'chunk') {
        setCurrentResponse(prev => prev + data.content)
      } else if (data.type === 'done') {
        setMessages(prev => [...prev, {
          id: crypto.randomUUID(),
          role: 'assistant',
          content: currentResponse,
          model: data.model
        }])
        setCurrentResponse('')
        setIsLoading(false)
      } else if (data.type === 'error') {
        setIsLoading(false)
        // Handle error
      }
    })

    return () => {
      wsRef.current?.close()
    }
  }, [articleId])

  const sendMessage = useCallback((content: string, selectedText?: string) => {
    if (!wsRef.current || isLoading) return

    setMessages(prev => [...prev, {
      id: crypto.randomUUID(),
      role: 'user',
      content
    }])

    setIsLoading(true)

    wsRef.current.send(JSON.stringify({
      articleId,
      message: content,
      selectedText,
      sessionId: sessionId.current
    }))
  }, [articleId, isLoading])

  return {
    messages,
    isLoading,
    currentResponse,
    sendMessage,
    clearMessages: () => setMessages([])
  }
}
```

**Step 2: 创建悬浮窗口组件**

`components/chat/floating-chat.tsx`:
```tsx
'use client'

import { useState, useRef, useEffect } from 'react'
import { MessageCircle, X, Minus, Send, Save, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ChatMessage } from './chat-message'
import { useChat } from '@/hooks/use-chat'

interface FloatingChatProps {
  articleId: string
  articleTitle: string
}

export function FloatingChat({ articleId, articleTitle }: FloatingChatProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [isMinimized, setIsMinimized] = useState(false)
  const [input, setInput] = useState('')
  const [position, setPosition] = useState({ x: 20, y: 20 }) // from bottom-right
  const { messages, isLoading, currentResponse, sendMessage, clearMessages } = useChat(articleId)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Auto scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages, currentResponse])

  const handleSubmit = () => {
    if (!input.trim() || isLoading) return
    sendMessage(input)
    setInput('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleSubmit()
    }
  }

  // Keyboard shortcut to toggle
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === '/') {
        e.preventDefault()
        setIsOpen(prev => !prev)
        setIsMinimized(false)
      }
      if (e.key === 'Escape' && isOpen) {
        setIsMinimized(true)
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [isOpen])

  if (!isOpen) {
    return (
      <button
        onClick={() => setIsOpen(true)}
        className="fixed bottom-6 right-6 w-12 h-12 rounded-full bg-accent text-white shadow-lg flex items-center justify-center hover:bg-accent-hover transition-colors"
        title="打开问答 (⌘/)"
      >
        <MessageCircle className="w-5 h-5" />
      </button>
    )
  }

  if (isMinimized) {
    return (
      <button
        onClick={() => setIsMinimized(false)}
        className="fixed bottom-6 right-6 px-4 py-2 rounded-full bg-accent text-white shadow-lg flex items-center gap-2 hover:bg-accent-hover transition-colors"
      >
        <MessageCircle className="w-4 h-4" />
        <span className="text-sm">问答窗口</span>
      </button>
    )
  }

  return (
    <div
      className="fixed bg-background border border-border rounded-lg shadow-xl flex flex-col"
      style={{
        bottom: position.y,
        right: position.x,
        width: 380,
        height: 500,
        maxHeight: '70vh'
      }}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border">
        <div className="flex items-center gap-2">
          <MessageCircle className="w-4 h-4 text-accent" />
          <span className="text-sm font-medium">关于本文的问答</span>
        </div>
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="icon" className="w-7 h-7" onClick={() => setIsMinimized(true)}>
            <Minus className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="icon" className="w-7 h-7" onClick={() => setIsOpen(false)}>
            <X className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Messages */}
      <ScrollArea className="flex-1 p-4" ref={scrollRef}>
        {messages.length === 0 && !currentResponse && (
          <div className="text-center text-muted-foreground text-sm py-8">
            <p>对「{articleTitle}」有疑问？</p>
            <p className="mt-1">在下方输入你的问题</p>
          </div>
        )}

        {messages.map((msg) => (
          <ChatMessage key={msg.id} message={msg} />
        ))}

        {currentResponse && (
          <ChatMessage
            message={{
              id: 'current',
              role: 'assistant',
              content: currentResponse
            }}
            isStreaming
          />
        )}

        {isLoading && !currentResponse && (
          <div className="flex items-center gap-2 text-muted-foreground text-sm">
            <div className="w-2 h-2 rounded-full bg-accent animate-pulse" />
            <span>思考中...</span>
          </div>
        )}
      </ScrollArea>

      {/* Toolbar */}
      <div className="px-4 py-2 border-t border-border flex items-center gap-2">
        <Button variant="ghost" size="sm" className="text-xs">
          <Save className="w-3 h-3 mr-1" />
          保存对话
        </Button>
        <Button variant="ghost" size="sm" className="text-xs" onClick={clearMessages}>
          <Trash2 className="w-3 h-3 mr-1" />
          清空
        </Button>
      </div>

      {/* Input */}
      <div className="p-4 pt-0">
        <div className="relative">
          <Textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="输入你的问题... (⌘↵ 发送)"
            className="pr-10 resize-none"
            rows={2}
          />
          <Button
            size="icon"
            className="absolute bottom-2 right-2 w-7 h-7"
            onClick={handleSubmit}
            disabled={!input.trim() || isLoading}
          >
            <Send className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
```

**Step 3: 创建消息组件**

`components/chat/chat-message.tsx`:
```tsx
import { cn } from '@/lib/utils'
import { User, Bot } from 'lucide-react'

interface ChatMessageProps {
  message: {
    id: string
    role: 'user' | 'assistant'
    content: string
    model?: string
  }
  isStreaming?: boolean
}

export function ChatMessage({ message, isStreaming }: ChatMessageProps) {
  const isUser = message.role === 'user'

  return (
    <div className={cn(
      "flex gap-3 mb-4",
      isUser && "flex-row-reverse"
    )}>
      <div className={cn(
        "w-7 h-7 rounded-full flex items-center justify-center flex-shrink-0",
        isUser ? "bg-accent" : "bg-muted"
      )}>
        {isUser ? (
          <User className="w-4 h-4 text-white" />
        ) : (
          <Bot className="w-4 h-4" />
        )}
      </div>

      <div className={cn(
        "flex-1 text-sm",
        isUser && "text-right"
      )}>
        <div className={cn(
          "inline-block px-3 py-2 rounded-lg max-w-[85%]",
          isUser ? "bg-accent text-white" : "bg-muted"
        )}>
          {message.content}
          {isStreaming && (
            <span className="inline-block w-1.5 h-4 bg-current ml-0.5 animate-pulse" />
          )}
        </div>
        {message.model && (
          <div className="text-xs text-muted-foreground mt-1">
            {message.model}
          </div>
        )}
      </div>
    </div>
  )
}
```

**Step 4: Commit**

```bash
git add .
git commit -m "feat(frontend): add floating chat component with WebSocket streaming"
```

---

### Task 11: 后台管理 UI

**Files:**
- Create: `frontend/app/admin/page.tsx`
- Create: `frontend/app/admin/layout.tsx`
- Create: `frontend/components/admin/config-panel.tsx`
- Create: `frontend/components/admin/task-monitor.tsx`
- Create: `frontend/components/admin/cost-chart.tsx`
- Create: `frontend/components/admin/system-status.tsx`

**Step 1: 创建管理后台布局**

`app/admin/layout.tsx`:
```tsx
'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard,
  Settings,
  Activity,
  DollarSign,
  FileText,
  ArrowLeft
} from 'lucide-react'

const navItems = [
  { href: '/admin', label: '概览', icon: LayoutDashboard },
  { href: '/admin/config', label: '配置', icon: Settings },
  { href: '/admin/tasks', label: '任务', icon: Activity },
  { href: '/admin/costs', label: '成本', icon: DollarSign },
  { href: '/admin/content', label: '内容', icon: FileText },
]

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()

  return (
    <div className="h-screen flex">
      {/* Sidebar */}
      <aside className="w-56 border-r border-border bg-background flex flex-col">
        <div className="h-14 flex items-center px-4 border-b border-border">
          <Link href="/" className="flex items-center gap-2 text-muted-foreground hover:text-foreground">
            <ArrowLeft className="w-4 h-4" />
            <span className="text-sm">返回主站</span>
          </Link>
        </div>

        <div className="px-4 py-4">
          <h1 className="font-semibold">后台管理</h1>
        </div>

        <nav className="flex-1 px-2">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-2 px-3 py-2 rounded-md text-sm mb-1",
                pathname === item.href
                  ? "bg-accent/10 text-accent"
                  : "hover:bg-muted text-muted-foreground"
              )}
            >
              <item.icon className="w-4 h-4" />
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>

      {/* Main */}
      <main className="flex-1 overflow-auto bg-background-secondary">
        {children}
      </main>
    </div>
  )
}
```

**Step 2: 创建概览页面**

`app/admin/page.tsx`:
```tsx
import { SystemStatus } from '@/components/admin/system-status'
import { TaskMonitor } from '@/components/admin/task-monitor'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function AdminPage() {
  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-semibold">系统概览</h1>

      <SystemStatus />

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              今日新文章
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">12</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              API 调用次数
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">847</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              今日成本
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">$0.32</div>
          </CardContent>
        </Card>
      </div>

      {/* Task Queue */}
      <Card>
        <CardHeader>
          <CardTitle>当前任务队列</CardTitle>
        </CardHeader>
        <CardContent>
          <TaskMonitor />
        </CardContent>
      </Card>
    </div>
  )
}
```

**Step 3: 创建系统状态组件**

`components/admin/system-status.tsx`:
```tsx
'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { RefreshCw } from 'lucide-react'

interface Status {
  name: string
  status: 'online' | 'offline' | 'warning'
  detail?: string
}

export function SystemStatus() {
  const { data, refetch, isLoading } = useQuery({
    queryKey: ['system-status'],
    queryFn: async () => {
      // TODO: Fetch from API
      return [
        { name: '后台服务', status: 'online' as const },
        { name: 'PostgreSQL', status: 'online' as const, detail: '已连接' },
        { name: 'Ollama', status: 'online' as const, detail: 'llama3:70b' },
        { name: 'Claude API', status: 'warning' as const, detail: '余额 $42.50' },
      ]
    },
    refetchInterval: 30000
  })

  const statusColors = {
    online: 'bg-green-500',
    offline: 'bg-red-500',
    warning: 'bg-yellow-500'
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-medium">系统状态</h3>
          <Button variant="ghost" size="sm" onClick={() => refetch()} disabled={isLoading}>
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          </Button>
        </div>

        <div className="space-y-3">
          {data?.map((item) => (
            <div key={item.name} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${statusColors[item.status]}`} />
                <span className="text-sm">{item.name}</span>
              </div>
              {item.detail && (
                <span className="text-sm text-muted-foreground">{item.detail}</span>
              )}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
```

**Step 4: 创建任务监控组件**

`components/admin/task-monitor.tsx`:
```tsx
'use client'

import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Eye, XCircle, Clock, CheckCircle, Loader2 } from 'lucide-react'

interface Task {
  id: string
  type: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  description: string
  model?: string
  progress?: number
  startedAt?: string
  completedAt?: string
}

export function TaskMonitor() {
  const { data: tasks } = useQuery({
    queryKey: ['tasks'],
    queryFn: async () => {
      // TODO: Fetch from API
      return [
        {
          id: '1842',
          type: 'content:generate',
          status: 'running' as const,
          description: '生成文章: "EIP-4844详解"',
          model: 'llama3:70b',
          progress: 67,
          startedAt: new Date().toISOString()
        },
        {
          id: '1843',
          type: 'web:crawl',
          status: 'pending' as const,
          description: '爬取 etherscan.io',
        },
        {
          id: '1841',
          type: 'rss:sync',
          status: 'completed' as const,
          description: 'RSS 同步 (zkSync)',
          completedAt: new Date().toISOString()
        }
      ]
    },
    refetchInterval: 5000
  })

  const statusIcons = {
    pending: <Clock className="w-4 h-4 text-muted-foreground" />,
    running: <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />,
    completed: <CheckCircle className="w-4 h-4 text-green-500" />,
    failed: <XCircle className="w-4 h-4 text-red-500" />
  }

  return (
    <ScrollArea className="h-[300px]">
      <div className="space-y-3">
        {tasks?.map((task) => (
          <div
            key={task.id}
            className="flex items-start justify-between p-3 rounded-lg border border-border"
          >
            <div className="flex items-start gap-3">
              {statusIcons[task.status]}
              <div>
                <div className="text-sm font-medium">{task.description}</div>
                {task.model && (
                  <div className="text-xs text-muted-foreground mt-1">
                    模型: {task.model}
                    {task.progress && ` | 进度: ${task.progress}%`}
                  </div>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Button variant="ghost" size="icon" className="w-7 h-7">
                <Eye className="w-4 h-4" />
              </Button>
              {task.status === 'running' && (
                <Button variant="ghost" size="icon" className="w-7 h-7 text-red-500">
                  <XCircle className="w-4 h-4" />
                </Button>
              )}
            </div>
          </div>
        ))}
      </div>
    </ScrollArea>
  )
}
```

**Step 5: Commit**

```bash
git add .
git commit -m "feat(frontend): add admin dashboard with system status and task monitor"
```

---

### Task 12: 知识库浏览页面

**Files:**
- Create: `frontend/app/knowledge/page.tsx`
- Create: `frontend/app/knowledge/[slug]/page.tsx`
- Create: `frontend/components/knowledge/article-list.tsx`
- Create: `frontend/components/knowledge/article-view.tsx`
- Create: `frontend/components/knowledge/article-edit.tsx`

**Step 1: 创建知识库列表页**

`app/knowledge/page.tsx`:
```tsx
import { MainLayout } from '@/components/layout/main-layout'
import { ArticleList } from '@/components/knowledge/article-list'

export default function KnowledgePage() {
  return (
    <MainLayout>
      <div className="p-6">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-semibold">知识库</h1>
        </div>
        <ArticleList />
      </div>
    </MainLayout>
  )
}
```

**Step 2: 创建文章详情页**

`app/knowledge/[slug]/page.tsx`:
```tsx
'use client'

import { useParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/main-layout'
import { ArticleView } from '@/components/knowledge/article-view'
import { FloatingChat } from '@/components/chat/floating-chat'

export default function ArticlePage() {
  const params = useParams()
  const slug = params.slug as string

  const { data: article, isLoading } = useQuery({
    queryKey: ['article', slug],
    queryFn: async () => {
      const res = await fetch(`/api/articles/${slug}`)
      return res.json()
    }
  })

  if (isLoading) {
    return (
      <MainLayout>
        <div className="p-6">加载中...</div>
      </MainLayout>
    )
  }

  return (
    <MainLayout>
      <ArticleView article={article} />
      <FloatingChat
        articleId={article.id}
        articleTitle={article.title}
      />
    </MainLayout>
  )
}
```

**Step 3: 创建文章视图组件**

`components/knowledge/article-view.tsx`:
```tsx
'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Edit, RefreshCw, MoreHorizontal, ExternalLink, Clock } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

interface Article {
  id: string
  title: string
  content: string
  contentHtml: string
  summary: string
  category: { name: string; slug: string }
  tags: string[]
  sourceUrls: string[]
  modelUsed: string
  createdAt: string
  updatedAt: string
}

interface ArticleViewProps {
  article: Article
}

export function ArticleView({ article }: ArticleViewProps) {
  const [isEditing, setIsEditing] = useState(false)

  return (
    <div className="max-w-4xl mx-auto p-6">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-start justify-between mb-4">
          <div>
            <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
              <span>{article.category?.name}</span>
              <span>·</span>
              <span className="flex items-center gap-1">
                <Clock className="w-3 h-3" />
                {formatDistanceToNow(new Date(article.updatedAt), {
                  addSuffix: true,
                  locale: zhCN
                })}
              </span>
            </div>
            <h1 className="text-3xl font-bold">{article.title}</h1>
          </div>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => setIsEditing(true)}>
                <Edit className="w-4 h-4 mr-2" />
                编辑
              </DropdownMenuItem>
              <DropdownMenuItem>
                <RefreshCw className="w-4 h-4 mr-2" />
                重新生成
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Tags */}
        <div className="flex items-center gap-2">
          {article.tags?.map((tag) => (
            <Badge key={tag} variant="secondary">{tag}</Badge>
          ))}
        </div>
      </div>

      {/* Content */}
      <article
        className="prose prose-neutral max-w-none"
        dangerouslySetInnerHTML={{ __html: article.contentHtml }}
      />

      {/* Sources */}
      {article.sourceUrls?.length > 0 && (
        <div className="mt-12 pt-6 border-t border-border">
          <h3 className="text-sm font-medium text-muted-foreground mb-3">📚 原始来源</h3>
          <ul className="space-y-1">
            {article.sourceUrls.map((url) => (
              <li key={url}>
                <a
                  href={url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-accent hover:underline flex items-center gap-1"
                >
                  {url}
                  <ExternalLink className="w-3 h-3" />
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Model info */}
      <div className="mt-6 text-xs text-muted-foreground">
        生成模型: {article.modelUsed}
      </div>
    </div>
  )
}
```

**Step 4: Commit**

```bash
git add .
git commit -m "feat(frontend): add knowledge base article list and detail pages"
```

---

### Task 13: 即时研究页面

**Files:**
- Create: `frontend/app/research/page.tsx`
- Create: `frontend/components/research/research-panel.tsx`

**Step 1: 创建即时研究页面**

`app/research/page.tsx`:
```tsx
import { MainLayout } from '@/components/layout/main-layout'
import { ResearchPanel } from '@/components/research/research-panel'

export default function ResearchPage() {
  return (
    <MainLayout>
      <div className="max-w-4xl mx-auto p-6">
        <h1 className="text-2xl font-semibold mb-6">即时研究</h1>
        <p className="text-muted-foreground mb-8">
          输入任何 Web3 技术名词或问题，AI 将为你搜索、分析并生成详细解释。
        </p>
        <ResearchPanel />
      </div>
    </MainLayout>
  )
}
```

**Step 2: 创建研究面板组件**

`components/research/research-panel.tsx`:
```tsx
'use client'

import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'
import { Search, Save, Loader2 } from 'lucide-react'

export function ResearchPanel() {
  const [query, setQuery] = useState('')
  const [result, setResult] = useState<any>(null)
  const [selectedCategory, setSelectedCategory] = useState<string>('')

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      const res = await fetch('/api/categories/tree')
      return res.json()
    }
  })

  const researchMutation = useMutation({
    mutationFn: async (topic: string) => {
      const res = await fetch('/api/research', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ topic })
      })
      return res.json()
    },
    onSuccess: (data) => {
      setResult(data)
      if (data.suggestedCategory) {
        setSelectedCategory(data.suggestedCategory.id)
      }
    }
  })

  const saveMutation = useMutation({
    mutationFn: async () => {
      const res = await fetch('/api/articles', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: result.title,
          content: result.content,
          categoryId: selectedCategory,
          tags: result.suggestedTags,
          sourceUrls: result.sources
        })
      })
      return res.json()
    }
  })

  const handleSearch = () => {
    if (!query.trim()) return
    researchMutation.mutate(query)
  }

  return (
    <div className="space-y-6">
      {/* Search */}
      <div className="flex gap-2">
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="例如: zkPorter, EIP-4844, Cosmos IBC..."
          className="flex-1"
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
        />
        <Button onClick={handleSearch} disabled={researchMutation.isPending}>
          {researchMutation.isPending ? (
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
          ) : (
            <Search className="w-4 h-4 mr-2" />
          )}
          研究
        </Button>
      </div>

      {/* Result */}
      {result && (
        <Card>
          <CardHeader>
            <CardTitle>{result.title}</CardTitle>
            <div className="flex items-center gap-2 mt-2">
              {result.suggestedTags?.map((tag: string) => (
                <Badge key={tag} variant="secondary">{tag}</Badge>
              ))}
            </div>
          </CardHeader>
          <CardContent>
            <article
              className="prose prose-neutral max-w-none"
              dangerouslySetInnerHTML={{ __html: result.contentHtml }}
            />

            {/* Save */}
            <div className="mt-6 pt-6 border-t border-border">
              <div className="flex items-center gap-4">
                <span className="text-sm">保存到分类:</span>
                <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                  <SelectTrigger className="w-[200px]">
                    <SelectValue placeholder="选择分类" />
                  </SelectTrigger>
                  <SelectContent>
                    {categories?.map((cat: any) => (
                      <SelectItem key={cat.id} value={cat.id}>
                        {cat.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button
                  onClick={() => saveMutation.mutate()}
                  disabled={!selectedCategory || saveMutation.isPending}
                >
                  <Save className="w-4 h-4 mr-2" />
                  保存到知识库
                </Button>
              </div>

              {result.suggestedNewCategory && (
                <div className="mt-3 p-3 bg-muted rounded-lg text-sm">
                  💡 建议创建新分类: <strong>{result.suggestedNewCategory.name}</strong>
                  <Button variant="link" size="sm" className="ml-2">
                    创建并保存
                  </Button>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
```

**Step 3: Commit**

```bash
git add .
git commit -m "feat(frontend): add instant research page"
```

---

### Task 14: 配置管理页面

**Files:**
- Create: `frontend/app/admin/config/page.tsx`
- Create: `frontend/components/admin/model-config.tsx`
- Create: `frontend/components/admin/source-config.tsx`

**Step 1: 创建配置页面**

`app/admin/config/page.tsx`:
```tsx
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ModelConfig } from '@/components/admin/model-config'
import { SourceConfig } from '@/components/admin/source-config'

export default function ConfigPage() {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-semibold mb-6">配置管理</h1>

      <Tabs defaultValue="models">
        <TabsList>
          <TabsTrigger value="models">模型设置</TabsTrigger>
          <TabsTrigger value="sources">数据源</TabsTrigger>
          <TabsTrigger value="crawler">爬虫</TabsTrigger>
          <TabsTrigger value="prompts">Prompt 模板</TabsTrigger>
        </TabsList>

        <TabsContent value="models" className="mt-6">
          <ModelConfig />
        </TabsContent>

        <TabsContent value="sources" className="mt-6">
          <SourceConfig />
        </TabsContent>

        <TabsContent value="crawler" className="mt-6">
          {/* CrawlerConfig */}
        </TabsContent>

        <TabsContent value="prompts" className="mt-6">
          {/* PromptConfig */}
        </TabsContent>
      </Tabs>
    </div>
  )
}
```

**Step 2: 创建模型配置组件**

`components/admin/model-config.tsx`:
```tsx
'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'
import { Check, RefreshCw } from 'lucide-react'

export function ModelConfig() {
  const queryClient = useQueryClient()

  const { data: config, isLoading } = useQuery({
    queryKey: ['model-config'],
    queryFn: async () => {
      const res = await fetch('/api/config?key=models')
      return res.json()
    }
  })

  const saveMutation = useMutation({
    mutationFn: async (newConfig: any) => {
      await fetch('/api/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: 'models', value: newConfig })
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['model-config'] })
    }
  })

  if (isLoading) return <div>加载中...</div>

  return (
    <div className="space-y-6">
      {/* Local Models */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">本地模型 (Ollama)</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <span className="text-sm w-24">默认模型</span>
            <Select defaultValue="llama3:70b">
              <SelectTrigger className="w-[200px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="llama3:70b">llama3:70b</SelectItem>
                <SelectItem value="qwen2.5:32b">qwen2.5:32b</SelectItem>
                <SelectItem value="mistral:7b">mistral:7b</SelectItem>
              </SelectContent>
            </Select>
            <Button variant="outline" size="sm">
              <RefreshCw className="w-4 h-4 mr-2" />
              测试连接
            </Button>
          </div>

          <div className="border rounded-lg divide-y">
            {['llama3:70b', 'qwen2.5:32b', 'mistral:7b'].map((model) => (
              <div key={model} className="flex items-center justify-between p-3">
                <div className="flex items-center gap-3">
                  <Switch defaultChecked={model !== 'mistral:7b'} />
                  <span className="font-mono text-sm">{model}</span>
                </div>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span>{model === 'llama3:70b' ? '40GB' : model === 'qwen2.5:32b' ? '20GB' : '4GB'}</span>
                  <Badge variant={model === 'llama3:70b' ? 'default' : 'secondary'}>
                    {model === 'llama3:70b' ? '已加载' : '未加载'}
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Cloud Models */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">云端模型</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Claude */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="font-medium">Claude API</span>
              <Switch defaultChecked />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm text-muted-foreground">API Key</label>
                <Input type="password" defaultValue="sk-ant-••••••••" />
              </div>
              <div>
                <label className="text-sm text-muted-foreground">默认模型</label>
                <Select defaultValue="claude-sonnet-4-20250514">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="claude-sonnet-4-20250514">claude-sonnet-4-20250514</SelectItem>
                    <SelectItem value="claude-haiku">claude-haiku</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <span className="text-sm text-muted-foreground">月预算上限</span>
              <Input type="number" defaultValue={50} className="w-24" />
              <span className="text-sm text-muted-foreground">当前用量: $38.50 / $50.00</span>
            </div>
          </div>

          {/* OpenAI */}
          <div className="space-y-3 pt-4 border-t">
            <div className="flex items-center justify-between">
              <span className="font-medium">OpenAI API</span>
              <div className="flex items-center gap-2">
                <Badge variant="outline">未配置</Badge>
                <Switch />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Routing */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">模型路由策略</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg">
            <div className="grid grid-cols-3 gap-4 p-3 bg-muted text-sm font-medium">
              <span>任务类型</span>
              <span>首选模型</span>
              <span>Fallback</span>
            </div>
            {[
              { task: '内容生成 (简单)', primary: 'llama3:70b', fallback: 'claude-haiku' },
              { task: '内容生成 (复杂)', primary: 'claude-sonnet', fallback: '-' },
              { task: '摘要/分类', primary: 'qwen2.5:32b', fallback: 'claude-haiku' },
              { task: '问答对话', primary: 'llama3:70b', fallback: 'claude-sonnet' },
              { task: '翻译', primary: 'qwen2.5:32b', fallback: 'claude-haiku' },
            ].map((route) => (
              <div key={route.task} className="grid grid-cols-3 gap-4 p-3 border-t text-sm">
                <span>{route.task}</span>
                <Select defaultValue={route.primary}>
                  <SelectTrigger className="h-8">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="llama3:70b">llama3:70b</SelectItem>
                    <SelectItem value="qwen2.5:32b">qwen2.5:32b</SelectItem>
                    <SelectItem value="claude-sonnet">claude-sonnet</SelectItem>
                    <SelectItem value="claude-haiku">claude-haiku</SelectItem>
                  </SelectContent>
                </Select>
                <Select defaultValue={route.fallback}>
                  <SelectTrigger className="h-8">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="-">无</SelectItem>
                    <SelectItem value="claude-sonnet">claude-sonnet</SelectItem>
                    <SelectItem value="claude-haiku">claude-haiku</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end gap-2">
        <Button variant="outline">恢复默认</Button>
        <Button onClick={() => saveMutation.mutate(config)}>
          <Check className="w-4 h-4 mr-2" />
          保存配置
        </Button>
      </div>
    </div>
  )
}
```

**Step 3: Commit**

```bash
git add .
git commit -m "feat(frontend): add model and source configuration pages"
```

---

### Task 15: Docker Compose 和开发脚本

**Files:**
- Update: `docker-compose.yml`
- Create: `Makefile`
- Create: `scripts/dev.sh`

**Step 1: 完善 Docker Compose**

`docker-compose.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: pgvector/pgvector:pg16
    container_name: web3-insight-db
    environment:
      POSTGRES_USER: web3insight
      POSTGRES_PASSWORD: web3insight_dev
      POSTGRES_DB: web3insight
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U web3insight"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: web3-insight-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  redis_data:
```

**Step 2: 创建 Makefile**

`Makefile`:
```makefile
.PHONY: dev dev-backend dev-frontend db-up db-down migrate seed test build clean

# Development
dev: db-up
	@make -j2 dev-backend dev-frontend

dev-backend:
	cd backend && go run cmd/server/main.go

dev-frontend:
	cd frontend && pnpm dev

# Database
db-up:
	docker-compose up -d postgres redis
	@echo "Waiting for database..."
	@sleep 3

db-down:
	docker-compose down

migrate:
	cd backend && go run cmd/migrate/main.go

seed:
	cd backend && go run cmd/seed/main.go

# Worker
worker:
	cd backend && go run cmd/worker/main.go

# Build
build-backend:
	cd backend && go build -o bin/server cmd/server/main.go
	cd backend && go build -o bin/worker cmd/worker/main.go

build-frontend:
	cd frontend && pnpm build

build: build-backend build-frontend

# Test
test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && pnpm test

test: test-backend test-frontend

# Clean
clean:
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/node_modules
```

**Step 3: Commit**

```bash
git add .
git commit -m "chore: add Docker Compose and Makefile for development"
```

---

## Phase 2 任务概要 (后续实现)

### Task 16-20: 数据收集模块
- RSS 解析器和同步任务
- 网页爬虫 (Colly) 实现
- GitHub 仓库监控
- API 数据源集成
- 付费搜索 API (Tavily, SerpAPI)

### Task 21-25: 内容生成模块
- 知识文章自动生成流程
- 新闻摘要和翻译
- 分类自动推断
- 向量嵌入生成
- 质量评估和重试机制

### Task 26-30: Explorer 调研模块
- Explorer 页面爬取
- 功能提取和对比
- 自动截图
- 调研报告生成

---

## 文件清单

### 前端文件 (frontend/)
```
frontend/
├── app/
│   ├── layout.tsx
│   ├── page.tsx
│   ├── globals.css
│   ├── knowledge/
│   │   ├── page.tsx
│   │   └── [slug]/page.tsx
│   ├── news/
│   │   └── page.tsx
│   ├── research/
│   │   └── page.tsx
│   └── admin/
│       ├── layout.tsx
│       ├── page.tsx
│       ├── config/page.tsx
│       ├── tasks/page.tsx
│       └── costs/page.tsx
├── components/
│   ├── ui/ (shadcn components)
│   ├── layout/
│   │   ├── sidebar.tsx
│   │   ├── header.tsx
│   │   └── main-layout.tsx
│   ├── knowledge/
│   │   ├── category-tree.tsx
│   │   ├── article-list.tsx
│   │   ├── article-view.tsx
│   │   └── article-edit.tsx
│   ├── chat/
│   │   ├── floating-chat.tsx
│   │   └── chat-message.tsx
│   ├── research/
│   │   └── research-panel.tsx
│   ├── admin/
│   │   ├── system-status.tsx
│   │   ├── task-monitor.tsx
│   │   ├── cost-chart.tsx
│   │   ├── model-config.tsx
│   │   └── source-config.tsx
│   └── providers/
│       └── query-provider.tsx
├── hooks/
│   └── use-chat.ts
├── lib/
│   ├── utils.ts
│   └── websocket.ts
└── package.json
```

### 后端文件 (backend/)
```
backend/
├── cmd/
│   ├── server/main.go
│   ├── worker/main.go
│   ├── migrate/main.go
│   └── seed/main.go
├── internal/
│   ├── api/
│   │   ├── router.go
│   │   ├── article.go
│   │   ├── category.go
│   │   ├── config.go
│   │   ├── task.go
│   │   ├── search.go
│   │   ├── research.go
│   │   └── chat_ws.go
│   ├── model/
│   │   ├── article.go
│   │   ├── category.go
│   │   ├── chat.go
│   │   ├── task.go
│   │   ├── news.go
│   │   ├── explorer.go
│   │   ├── config.go
│   │   └── source.go
│   ├── repository/
│   │   ├── article.go
│   │   ├── category.go
│   │   └── ...
│   ├── service/
│   │   ├── generator.go
│   │   ├── classifier.go
│   │   ├── search.go
│   │   └── chat.go
│   ├── llm/
│   │   ├── types.go
│   │   ├── router.go
│   │   ├── ollama.go
│   │   ├── claude.go
│   │   └── openai.go
│   ├── database/
│   │   ├── connect.go
│   │   ├── migrate.go
│   │   └── seed.go
│   ├── worker/
│   │   ├── scheduler.go
│   │   └── tasks.go
│   └── config/
│       └── config.go
├── config/
│   └── config.yaml
└── go.mod
```

---

**计划完成，已保存到 `docs/plans/2025-02-02-web3-insight-full-design.md`。**

两种执行方式：

**1. Subagent-Driven (本会话)** - 我在本会话中逐个任务派发 subagent，每个任务完成后 review，快速迭代

**2. Parallel Session (新会话)** - 在 worktree 中开新会话，使用 executing-plans 批量执行带检查点

你选择哪种方式？
