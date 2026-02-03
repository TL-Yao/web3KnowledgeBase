# Fetch Categories from Backend Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace hardcoded frontend category mock data with API calls to fetch categories from the backend.

**Architecture:** The backend already has a complete category API with tree structure support (`/api/categories/tree`). Frontend needs to replace the `mockCategories` array with React Query to fetch categories from the backend API, handle loading/error states, and display an empty state when no categories exist.

**Tech Stack:**
- Frontend: React, TypeScript, TanStack Query (React Query)
- Backend: Go, Gin, GORM (already implemented)
- API Endpoint: `GET /api/categories/tree`

---

## Issue Analysis

**Problem confirmed:**
- Frontend file: `web3-insight/frontend/components/knowledge/category-tree.tsx:14-49`
- Contains hardcoded `mockCategories` array with Layer 1, Layer 2, DeFi, NFT, etc.
- Comment on line 13: `// Mock data - will be replaced with API call`
- Component does not fetch from backend API

**Backend status:**
- ✅ API endpoint exists: `GET /api/categories/tree` (router.go:97)
- ✅ Handler implemented: `CategoryHandler.GetTree()` (category.go:46-54)
- ✅ Repository method: `CategoryRepository.GetTree()` (category.go:40-53)
- ✅ Returns hierarchical tree structure with children
- ✅ Model includes all necessary fields (id, name, nameEn, children, etc.)

**Frontend API layer:**
- ❌ Missing: No `categoryAPI.getTree()` method in `lib/api.ts`
- ⚠️ Exists: `categoryAPI.list()` but returns flat list, not tree

---

## Task 1: Add Category Tree Type Definitions

**Files:**
- Modify: `web3-insight/frontend/lib/api.ts:99-109`

**Step 1: Update Category interface to match backend model**

In `lib/api.ts`, replace the existing Category interface (lines 100-105) with:

```typescript
export interface Category {
  id: string
  name: string
  nameEn: string
  slug: string
  parentId?: string
  description?: string
  icon?: string
  sortOrder: number
  children?: Category[]
  articleCount: number
  createdAt: string
  updatedAt: string
}
```

**Step 2: Add getTree method to categoryAPI**

In `lib/api.ts`, replace the categoryAPI object (lines 107-109) with:

```typescript
export const categoryAPI = {
  list: () => fetchAPI<Category[]>('/api/categories'),
  getTree: () => fetchAPI<Category[]>('/api/categories/tree'),
}
```

**Step 3: Verify TypeScript compilation**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm run type-check`
Expected: No type errors

**Step 4: Commit**

```bash
git add web3-insight/frontend/lib/api.ts
git commit -m "feat(api): add category tree interface and API method"
```

---

## Task 2: Update CategoryTree Component to Fetch from API

**Files:**
- Modify: `web3-insight/frontend/components/knowledge/category-tree.tsx`

**Step 1: Write test for CategoryTree with loading state**

Create: `web3-insight/frontend/components/knowledge/__tests__/category-tree.test.tsx`

```typescript
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { CategoryTree } from '../category-tree'

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('CategoryTree', () => {
  it('shows loading state initially', () => {
    render(<CategoryTree />, { wrapper: createWrapper() })
    expect(screen.getByText(/加载中/i)).toBeInTheDocument()
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm test -- category-tree.test.tsx`
Expected: FAIL - loading state not implemented

**Step 3: Replace mock data with useQuery**

In `category-tree.tsx`, replace lines 1-116 with:

```typescript
'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronRight, Folder, FolderOpen } from 'lucide-react'
import { cn } from '@/lib/utils'
import { categoryAPI, Category } from '@/lib/api'

interface CategoryNodeProps {
  category: Category
  level: number
}

function CategoryNode({ category, level }: CategoryNodeProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const hasChildren = category.children && category.children.length > 0

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-1 px-2 py-1.5 rounded-md hover:bg-accent/10 cursor-pointer text-sm",
          "transition-colors"
        )}
        style={{ paddingLeft: `${level * 12 + 8}px` }}
        onClick={() => {
          if (hasChildren) {
            setIsExpanded(!isExpanded)
          }
        }}
      >
        {hasChildren ? (
          <ChevronRight
            className={cn(
              "w-4 h-4 text-muted-foreground transition-transform",
              isExpanded && "rotate-90"
            )}
          />
        ) : (
          <span className="w-4" />
        )}
        {hasChildren ? (
          isExpanded ? (
            <FolderOpen className="w-4 h-4 text-accent" />
          ) : (
            <Folder className="w-4 h-4 text-muted-foreground" />
          )
        ) : (
          <span className="w-4 h-4 flex items-center justify-center text-muted-foreground">
            &bull;
          </span>
        )}
        <span className="truncate">{category.name}</span>
      </div>
      {hasChildren && isExpanded && (
        <div>
          {category.children!.map((child) => (
            <CategoryNode key={child.id} category={child} level={level + 1} />
          ))}
        </div>
      )}
    </div>
  )
}

export function CategoryTree() {
  const { data: categories, isLoading, error } = useQuery({
    queryKey: ['categories', 'tree'],
    queryFn: categoryAPI.getTree,
  })

  if (isLoading) {
    return (
      <div className="py-4 px-2 text-sm text-muted-foreground">
        加载分类中...
      </div>
    )
  }

  if (error) {
    return (
      <div className="py-4 px-2 text-sm text-destructive">
        加载分类失败，请稍后重试
      </div>
    )
  }

  if (!categories || categories.length === 0) {
    return (
      <div className="py-4 px-2 text-sm text-muted-foreground">
        暂无分类
      </div>
    )
  }

  return (
    <div className="py-1">
      {categories.map((category) => (
        <CategoryNode key={category.id} category={category} level={0} />
      ))}
    </div>
  )
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm test -- category-tree.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add web3-insight/frontend/components/knowledge/category-tree.tsx
git add web3-insight/frontend/components/knowledge/__tests__/category-tree.test.tsx
git commit -m "feat(categories): fetch category tree from backend API

- Replace mockCategories with useQuery
- Add loading, error, and empty states
- Add test for loading state"
```

---

## Task 3: Verify Sidebar Integration

**Files:**
- Check: `web3-insight/frontend/components/layout/sidebar.tsx`

**Step 1: Verify CategoryTree is wrapped in QueryClientProvider**

Run: `grep -r "QueryClientProvider" /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend/app`
Expected: Should find QueryClientProvider in app layout

**Step 2: Check if not found, add QueryClientProvider to root layout**

If QueryClientProvider is not found, modify `app/layout.tsx` to wrap children with QueryClientProvider.

Read: `web3-insight/frontend/app/layout.tsx`

If missing, wrap children with:

```typescript
'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState } from 'react'

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const [queryClient] = useState(() => new QueryClient())

  return (
    <html lang="zh-CN">
      <body>
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      </body>
    </html>
  )
}
```

**Step 3: Verify by running frontend**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm run dev`
Expected: Frontend starts without errors

**Step 4: Commit if changes made**

```bash
git add web3-insight/frontend/app/layout.tsx
git commit -m "feat: add QueryClientProvider to root layout"
```

---

## Task 4: Test End-to-End with Backend

**Files:**
- Test: Full stack integration

**Step 1: Start backend services**

Run: `/Users/tongleyao/claudeProjects/explorerResearch/web3-insight/scripts/start-all.sh`
Expected: All services start successfully

**Step 2: Verify backend API returns empty array**

Run: `curl http://localhost:8080/api/categories/tree`
Expected: `[]` (empty array, since no categories exist)

**Step 3: Open frontend in Chrome using automation**

Use Chrome automation to:
- Navigate to `http://localhost:3000`
- Take screenshot of sidebar
- Verify "暂无分类" (no categories) message is displayed

**Step 4: Create test categories via API**

Run:
```bash
curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Layer 1",
    "nameEn": "Layer 1",
    "slug": "layer-1",
    "sortOrder": 1
  }'

curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Layer 2",
    "nameEn": "Layer 2",
    "slug": "layer-2",
    "sortOrder": 2
  }'
```

Expected: 201 Created responses with category objects

**Step 5: Verify frontend displays new categories**

Use Chrome automation to:
- Refresh page at `http://localhost:3000`
- Take screenshot of sidebar
- Verify "Layer 1" and "Layer 2" categories are displayed

**Step 6: Check console for errors**

Use Chrome automation to read console messages with pattern: `error|Error|ERROR`
Expected: No API errors

**Step 7: Document test results**

Test completed successfully on 2026-02-04:

✅ **Empty State Test:**
- Backend API returned `[]` (empty array)
- Frontend displayed "暂无分类" message in sidebar
- Screenshot captured: Empty state working correctly

✅ **Populated State Test:**
- Created 2 test categories via API:
  - Layer 1 (id: d1b93e4f-4ccf-41fd-b34b-5f951364eef7)
  - Layer 2 (id: b0c7eb30-09b2-4eda-be6c-b89f13064ae3)
- Backend API returned both categories in tree format
- Frontend refreshed and displayed both categories correctly
- Screenshot captured: Categories visible in sidebar as bullet list

✅ **Network Requests:**
- Multiple successful GET requests to `/api/categories/tree`
- All requests returned 200 OK status
- CORS preflight OPTIONS requests returned 204 No Content

✅ **Console Errors:**
- No errors found in browser console
- No JavaScript errors
- No API errors

**Test Evidence:**
- Screenshots: ss_7846syajs (empty state), ss_2756vmz5n (populated state)
- Network requests: 6 successful API calls tracked
- Categories displayed correctly in sidebar with proper styling

**Step 8: Commit test documentation**

```bash
git add docs/plans/2026-02-04-fetch-categories-from-backend.md
git commit -m "docs: add end-to-end test results to plan"
```

---

## Task 5: Update Documentation

**Files:**
- Modify: `web3-insight/CLAUDE.md`

**Step 1: Add entry to Implementation History**

Append to CLAUDE.md under Implementation History section:

```markdown
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
```

**Step 2: Verify CLAUDE.md formatting**

Run: `head -50 /Users/tongleyao/claudeProjects/explorerResearch/CLAUDE.md`
Expected: Valid markdown structure

**Step 3: Commit documentation**

```bash
git add CLAUDE.md
git commit -m "docs: document category API integration"
```

---

## Task 6: Clean Up and Final Verification

**Files:**
- Check: All modified files

**Step 1: Run frontend tests**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm test`
Expected: All tests pass

**Step 2: Run TypeScript type check**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm run type-check`
Expected: No type errors

**Step 3: Run frontend linter**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm run lint`
Expected: No linting errors

**Step 4: Build frontend to verify production build**

Run: `cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend && npm run build`
Expected: Build succeeds without errors

**Step 5: Final Chrome verification**

Use Chrome automation to:
- Navigate to `http://localhost:3000`
- Take screenshot showing working category tree
- Verify no console errors
- Test category expand/collapse interaction

**Step 6: Stop services**

Run: `/Users/tongleyao/claudeProjects/explorerResearch/web3-insight/scripts/stop-all.sh`

**Step 7: Final commit**

```bash
git add -A
git commit -m "chore: final verification and cleanup for category API integration"
```

---

## Success Criteria

- ✅ No hardcoded mockCategories in frontend code
- ✅ CategoryTree component fetches from `/api/categories/tree`
- ✅ Loading state displays "加载分类中..."
- ✅ Error state displays user-friendly error message
- ✅ Empty state displays "暂无分类"
- ✅ Categories display correctly when data exists
- ✅ Category expand/collapse works
- ✅ No console errors
- ✅ All tests pass
- ✅ TypeScript compiles without errors
- ✅ Documentation updated in CLAUDE.md

## Notes

- Backend API already fully implemented, no backend changes needed
- Frontend uses TanStack Query (React Query) for all API calls
- Component is already styled correctly, only data source needs to change
- Sidebar component (sidebar.tsx) also has hardcoded "最近阅读" data (lines 60-72) - this should be addressed in a separate task
