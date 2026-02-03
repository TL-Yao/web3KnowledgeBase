# Fix Frontend-Backend API Response Mismatch

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix the "无法连接后端服务" error by aligning frontend API expectations with backend response format.

**Root Cause:** Backend returns `{"articles": [], ...}` but frontend expects `{"data": [], ...}`. The frontend code tries to access `response.data` which is undefined, causing the query to fail and show the connection error.

**Architecture:** Update frontend API client (`lib/api.ts`) to match backend response structure. Backend returns `articles` field, not `data` field. Also rename `pageSize` to `limit` for consistency.

**Tech Stack:** TypeScript, Next.js, TanStack Query

---

## Task 1: Fix Article List Response Interface

**Files:**
- Modify: `web3-insight/frontend/lib/api.ts:52-57`
- Modify: `web3-insight/frontend/components/knowledge/article-list.tsx:44-45`

**Step 1: Update ArticleListResponse interface to match backend**

In `lib/api.ts`, change the interface to match what backend actually returns:

```typescript
export interface ArticleListResponse {
  articles: Article[]  // Changed from 'data' to 'articles'
  total: number
  page: number
  pageSize: number    // Backend returns 'pageSize', not 'limit'
}
```

**Step 2: Update article-list.tsx to use correct field name**

In `components/knowledge/article-list.tsx`, change line 45:

```typescript
// Before:
return response.data.map(transformArticle)

// After:
return response.articles.map(transformArticle)
```

**Step 3: Verify the fix**

Run: Open http://localhost:3000/knowledge in browser
Expected: Should show "暂无文章" (no articles) instead of "无法连接后端服务" (connection error)

**Step 4: Test with backend response**

```bash
curl -s http://localhost:3000/api/articles | jq '.'
```

Expected output:
```json
{
  "articles": [],
  "total": 0,
  "page": 1,
  "pageSize": 20
}
```

**Step 5: Commit**

```bash
git add web3-insight/frontend/lib/api.ts web3-insight/frontend/components/knowledge/article-list.tsx
git commit -m "fix: align frontend ArticleListResponse with backend response format

- Change 'data' field to 'articles' to match backend
- Change 'limit' field to 'pageSize' to match backend
- Fixes 'unable to connect to backend' error on knowledge page

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Verify System Status Component

**Files:**
- Test: `web3-insight/frontend/components/admin/system-status.tsx:20`

**Step 1: Test system status health check**

The component fetches `/health` (without `/api` prefix). Verify this works:

```bash
curl -s http://localhost:3000/health | jq '.'
```

Expected output:
```json
{
  "status": "ok"
}
```

**Step 2: Open admin page in browser**

```bash
open -a "Google Chrome" http://localhost:3000/admin
```

Expected: System status should show "后台服务: 运行中" (Backend Service: Running) with green indicator

**Step 3: No code changes needed**

If the health check works, no changes are needed. The system status component already uses the correct endpoint.

**Step 4: Document the verification**

Add verification note to commit message in final task.

---

## Task 3: Test All API Endpoints

**Files:**
- Test: All API client functions in `web3-insight/frontend/lib/api.ts`

**Step 1: Test categories endpoint**

```bash
curl -s http://localhost:3000/api/categories | jq '.'
```

Expected: `[]` (empty array, backend returns array directly, not wrapped in object)

**Step 2: Check if categoryAPI needs updates**

Read `lib/api.ts` line 108:

```typescript
list: () => fetchAPI<Category[]>('/api/categories'),
```

This is correct - it expects an array, and backend returns an array. No changes needed.

**Step 3: Test explorer endpoints**

```bash
curl -s http://localhost:3000/api/explorers | jq '.'
```

Expected output structure:
```json
{
  "data": [...],
  "count": 0
}
```

**Step 4: Verify explorerAPI interface matches**

Read `lib/api.ts` line 333. It should expect `{data: ExplorerResearch[], count: number}`.

If it matches, no changes needed. If not, update the interface.

**Step 5: Test data sources endpoint**

```bash
curl -s http://localhost:3000/api/sources | jq '.'
```

Check if response matches `DataSource[]` interface (array of objects).

**Step 6: Document findings**

Create a list of any other endpoints that need fixing. If all are correct, note that in commit message.

---

## Task 4: Update Documentation

**Files:**
- Modify: `CLAUDE.md` (Implementation History section)

**Step 1: Add entry to Implementation History**

Add this entry after the last implementation history entry:

```markdown
### 2026-02-04 - Fix Frontend-Backend API Response Mismatch

**What was completed:**
- Fixed ArticleListResponse interface to match backend response format
- Changed 'data' field to 'articles' to match backend's actual response
- Changed 'limit' field to 'pageSize' to match backend's actual response
- Verified system status health check endpoint works correctly
- Tested all API endpoints for consistency

**Important takeaways:**
- **API contract alignment**: Frontend TypeScript interfaces must exactly match backend JSON responses
- **Response field naming**: Backend returns `{"articles": [...]}` not `{"data": [...]}`
- **Pagination field naming**: Backend uses `pageSize` not `limit` in response
- **Health check endpoint**: Uses `/health` (no `/api` prefix) and works through Next.js rewrites
- **Error manifestation**: Mismatched field names cause "无法连接后端服务" error even when backend is running fine
- **Testing approach**: Use curl to verify actual backend responses before writing frontend interfaces

**Related commits:** [commit hash from Task 1]

**Files modified:**
- web3-insight/frontend/lib/api.ts (ArticleListResponse interface)
- web3-insight/frontend/components/knowledge/article-list.tsx (response field access)
- CLAUDE.md (this documentation)
```

**Step 2: Commit documentation update**

```bash
git add CLAUDE.md
git commit -m "docs: add API response mismatch fix to implementation history

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Final Verification

**Files:**
- Test: Full application in browser

**Step 1: Restart frontend to ensure changes are loaded**

```bash
# Stop frontend (find PID on port 3000)
lsof -ti :3000 | xargs kill

# Start frontend
cd /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/frontend
npm run dev > /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/logs/frontend.log 2>&1 &
echo $! > /Users/tongleyao/claudeProjects/explorerResearch/web3-insight/logs/frontend.pid
```

**Step 2: Open all pages in browser**

```bash
open -a "Google Chrome" http://localhost:3000
open -a "Google Chrome" http://localhost:3000/knowledge
open -a "Google Chrome" http://localhost:3000/admin
open -a "Google Chrome" http://localhost:3000/research
```

**Step 3: Verify each page**

- **Home page**: Should load without errors
- **Knowledge page**: Should show "暂无文章" (no articles) instead of connection error
- **Admin page**:
  - System status should show "后台服务: 运行中" (green)
  - Stats should show 0 values (not connection error)
  - Task monitor should show "暂无运行中的任务" (no tasks)
- **Research page**: Should load without errors

**Step 4: Check browser console for errors**

Open DevTools (Cmd+Option+I) and check Console tab. Should see no red errors about failed API calls.

**Step 5: Success criteria**

✅ No "无法连接后端服务" errors anywhere
✅ Empty states display correctly
✅ System status shows backend as online
✅ No console errors for API calls
✅ All pages load successfully

---

## Notes

**Why this happened:**
- During previous cleanup of mock data, backend response structure wasn't verified against frontend expectations
- TypeScript interfaces in `lib/api.ts` were based on assumptions rather than actual backend responses
- The error message "无法连接后端服务" was misleading - the backend WAS connected and working, but the response parsing failed

**Prevention:**
- Always curl test the actual backend endpoint before writing frontend interfaces
- Use browser DevTools Network tab to see actual API responses
- TypeScript interfaces should be generated from or verified against actual API responses
- Add API contract tests to catch these mismatches early
