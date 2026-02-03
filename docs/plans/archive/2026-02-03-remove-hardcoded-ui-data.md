# Remove Hardcoded UI Data Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove all hardcoded/mock data from frontend UI so startup shows empty state with proper defaults

**Architecture:** Replace hardcoded values in admin page stats and task monitor component with API calls that return empty data, ensuring all components gracefully handle empty states

**Tech Stack:** Next.js, React, TypeScript, TanStack Query

---

## Task 1: Remove hardcoded admin stats

**Files:**
- Modify: `web3-insight/frontend/app/admin/page.tsx:13-46`

**Problem:**
Lines 21, 32, 43 show hardcoded numbers: "12", "847", "$0.32"

**Step 1: Replace hardcoded stats with API calls**

Remove the hardcoded stat cards and replace with data-driven components:

```tsx
// Before (lines 13-46):
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
  // ... more hardcoded cards
</div>

// After:
'use client'

import { useQuery } from '@tanstack/react-query'
// ... existing imports

export default function AdminPage() {
  const { data: stats } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: async () => {
      // Return empty stats for now - backend can implement later
      return {
        todayArticles: 0,
        apiCalls: 0,
        todayCost: 0
      }
    }
  })

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
            <div className="text-3xl font-bold">{stats?.todayArticles ?? 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              API 调用次数
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.apiCalls ?? 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              今日成本
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              ${(stats?.todayCost ?? 0).toFixed(2)}
            </div>
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

**Step 2: Verify admin page shows zeros**

Run: `curl http://localhost:3000/admin` or check browser

Expected: All stat cards show "0" or "$0.00"

**Step 3: Commit**

```bash
git add web3-insight/frontend/app/admin/page.tsx
git commit -m "fix: remove hardcoded admin stats, use API query with empty defaults"
```

---

## Task 2: Remove hardcoded task monitor data

**Files:**
- Modify: `web3-insight/frontend/components/admin/task-monitor.tsx:19-50`

**Problem:**
Lines 24-47 return hardcoded mock tasks array instead of actual API data

**Step 1: Replace hardcoded tasks with empty default**

```tsx
// Before (lines 20-50):
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
      // ... more mock tasks
    ]
  },
  refetchInterval: 5000
})

// After:
const { data: tasks, isLoading } = useQuery({
  queryKey: ['tasks'],
  queryFn: async () => {
    // Return empty array - backend endpoint not yet implemented
    return []
  },
  refetchInterval: 5000
})
```

**Step 2: Add empty state UI**

Add empty state display after the ScrollArea wrapper:

```tsx
// After line 60:
return (
  <ScrollArea className="h-[300px]">
    <div className="space-y-3">
      {isLoading ? (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          <div className="animate-pulse">加载中...</div>
        </div>
      ) : tasks && tasks.length > 0 ? (
        tasks.map((task) => (
          // ... existing task card JSX
        ))
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
          <Clock className="w-12 h-12 mb-4 opacity-50" />
          <p className="text-lg font-medium">暂无运行中的任务</p>
          <p className="text-sm">当前队列为空</p>
        </div>
      )}
    </div>
  </ScrollArea>
)
```

**Step 3: Verify task monitor shows empty state**

Run: Check http://localhost:3000/admin in browser

Expected: Task queue section shows "暂无运行中的任务" message instead of mock tasks

**Step 4: Commit**

```bash
git add web3-insight/frontend/components/admin/task-monitor.tsx
git commit -m "fix: remove hardcoded task monitor data, show empty state by default"
```

---

## Task 3: Verify full empty state

**Step 1: Stop all services**

```bash
./web3-insight/scripts/stop-all.sh
```

**Step 2: Clear database (optional - to ensure truly empty)**

```bash
docker exec -it web3-insight-db psql -U web3insight -d web3insight -c "TRUNCATE categories, articles, data_sources, content_tasks CASCADE;"
```

**Step 3: Start all services**

```bash
./web3-insight/scripts/start-all.sh
```

**Step 4: Check each page manually**

1. Home page (http://localhost:3000): Should show search box and quick access cards only - no data
2. Knowledge page (http://localhost:3000/knowledge): Should show "暂无文章" empty state
3. Admin page (http://localhost:3000/admin): Should show all zeros in stats, empty task queue

Expected: All pages show appropriate empty states, no hardcoded/mock data visible

**Step 5: Document verification**

No commit needed - this is verification only

---

## Task 4: Update CLAUDE.md with findings

**Files:**
- Modify: `CLAUDE.md` - Implementation History section

**Step 1: Add implementation history entry**

Add after the 2026-02-03 Mock Data Removal entry:

```markdown
### 2026-02-03 - Remove Hardcoded UI Data (Follow-up)

**What was completed:**
- Removed hardcoded admin page stats (12 articles, 847 API calls, $0.32 cost)
- Removed hardcoded task monitor mock data (3 sample tasks)
- Replaced with proper API queries returning empty defaults
- Added empty state UI to task monitor component

**Important takeaways:**
- **Frontend mock data locations**: Mock data can hide in multiple places:
  1. `__mocks__/` directory (MSW handlers) - covered in previous cleanup
  2. Component-level hardcoded values (admin page stats)
  3. Query function return values (task-monitor.tsx)
- Always check both page files and component files for hardcoded data
- Empty states should be explicit and user-friendly, not just blank screens
- Use `?? 0` pattern for safe default values in stats display

**Related commits:** [commit range]

**Files modified:**
- app/admin/page.tsx (converted to 'use client', added useQuery for stats)
- components/admin/task-monitor.tsx (removed mock tasks, added empty state)
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add hardcoded UI data removal to implementation history"
```

---

## Summary

This plan removes ALL remaining hardcoded/mock data from the frontend:

1. Admin page stats (12, 847, $0.32) → API query with 0 defaults
2. Task monitor (3 mock tasks) → Empty array with friendly empty state

After completion, the entire system will show truly empty state on startup.
