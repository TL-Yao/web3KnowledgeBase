// components/knowledge/__tests__/article-list.test.tsx
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { http, HttpResponse } from 'msw'
import { server } from '@/__mocks__/server'
import { ArticleList } from '../article-list'
import { mockArticles, mockCategories } from '@/__mocks__/data'

// Add handlers for API endpoints
// Use beforeEach because vitest.setup.tsx resets handlers afterEach
beforeEach(() => {
  server.use(
    http.get('/api/articles', ({ request }) => {
      const url = new URL(request.url)
      // Note: API uses category_id (snake_case) as the query parameter
      const categoryId = url.searchParams.get('category_id')

      // Filter articles by category if categoryId is provided
      let articles = mockArticles
      if (categoryId) {
        articles = mockArticles.filter(a => a.categoryId === categoryId)
      }

      return HttpResponse.json({
        articles,
        total: articles.length,
        page: 1,
        limit: 20,
      })
    }),
    http.get('/api/categories/tree', () => {
      return HttpResponse.json(mockCategories)
    })
  )
})

// Create a fresh QueryClient for each test
function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
        staleTime: 0,
      },
    },
  })
}

// Wrapper component with QueryClientProvider
function renderWithQueryClient(ui: React.ReactElement) {
  const queryClient = createTestQueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>
  )
}

// Mock router functions
const mockSearchParams = new URLSearchParams()

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    back: vi.fn(),
    forward: vi.fn(),
    refresh: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => '/knowledge',
  useSearchParams: () => mockSearchParams,
  useParams: () => ({}),
}))

describe('ArticleList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockSearchParams.delete('category')
  })

  describe('loading state', () => {
    it('renders loading indicator while fetching articles', () => {
      renderWithQueryClient(<ArticleList />)

      // Should show loading text initially
      expect(screen.getByText('加载中...')).toBeInTheDocument()
    })
  })

  describe('rendering articles from API', () => {
    it('renders articles from API (MSW mock data)', async () => {
      renderWithQueryClient(<ArticleList />)

      // Wait for articles to load (from MSW mock handler)
      await waitFor(() => {
        expect(screen.getByText('以太坊 2.0 升级详解')).toBeInTheDocument()
      })

      expect(screen.getByText('DeFi 借贷协议对比')).toBeInTheDocument()
    })

    it('shows article summary', async () => {
      renderWithQueryClient(<ArticleList />)

      await waitFor(() => {
        expect(screen.getByText('本文详细介绍以太坊 2.0 的技术架构和升级路线图')).toBeInTheDocument()
      })

      expect(screen.getByText('深入对比 Aave、Compound 和 MakerDAO 等借贷协议')).toBeInTheDocument()
    })

    it('displays category badges', async () => {
      renderWithQueryClient(<ArticleList />)

      await waitFor(() => {
        expect(screen.getByText('以太坊 2.0 升级详解')).toBeInTheDocument()
      })

      // Should show category badges (may appear multiple times as both category and tag)
      // Use getAllByText since "Ethereum" and "DeFi" appear as both category badge and tag
      expect(screen.getAllByText('Ethereum').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('DeFi').length).toBeGreaterThanOrEqual(1)
    })

    it('displays article tags', async () => {
      renderWithQueryClient(<ArticleList />)

      await waitFor(() => {
        expect(screen.getByText('以太坊 2.0 升级详解')).toBeInTheDocument()
      })

      // Should show tags from mock articles
      expect(screen.getByText('PoS')).toBeInTheDocument()
      expect(screen.getByText('Sharding')).toBeInTheDocument()
      expect(screen.getByText('Lending')).toBeInTheDocument()
      expect(screen.getByText('Aave')).toBeInTheDocument()
    })
  })

  describe('navigation', () => {
    it('navigates to article detail on click (check link href)', async () => {
      renderWithQueryClient(<ArticleList />)

      await waitFor(() => {
        expect(screen.getByText('以太坊 2.0 升级详解')).toBeInTheDocument()
      })

      // Check that article titles are wrapped in links with correct hrefs
      const ethereumLink = screen.getByText('以太坊 2.0 升级详解').closest('a')
      expect(ethereumLink).toHaveAttribute('href', '/knowledge/ethereum-2-upgrade')

      const defiLink = screen.getByText('DeFi 借贷协议对比').closest('a')
      expect(defiLink).toHaveAttribute('href', '/knowledge/defi-lending-comparison')
    })
  })

  describe('category filtering', () => {
    it('filters articles by category when categoryId is in URL', async () => {
      // Set category filter in search params BEFORE setting up the handler
      mockSearchParams.set('category', 'cat-1-1')

      // Override handler specifically for this test to filter by category
      server.use(
        http.get('/api/articles', ({ request }) => {
          const url = new URL(request.url)
          // Note: API uses category_id (snake_case) as the query parameter
          const categoryId = url.searchParams.get('category_id')

          // Filter articles by category if categoryId is provided
          let articles = mockArticles
          if (categoryId) {
            articles = mockArticles.filter(a => a.categoryId === categoryId)
          }

          return HttpResponse.json({
            articles,
            total: articles.length,
            page: 1,
            limit: 20,
          })
        })
      )

      renderWithQueryClient(<ArticleList />)

      // Wait for filtered articles to load
      await waitFor(() => {
        expect(screen.getByText('以太坊 2.0 升级详解')).toBeInTheDocument()
      })

      // DeFi article should not be shown (different category)
      expect(screen.queryByText('DeFi 借贷协议对比')).not.toBeInTheDocument()
    })

    it('shows category filter indicator when filtering by category', async () => {
      // Set category filter in search params
      mockSearchParams.set('category', 'cat-1-1')

      renderWithQueryClient(<ArticleList />)

      // Wait for category name to be fetched and displayed
      await waitFor(() => {
        expect(screen.getByText('当前分类:')).toBeInTheDocument()
      })

      // Should show article count
      expect(screen.getByText(/篇文章/)).toBeInTheDocument()
    })

    it('shows all articles when no category filter is set', async () => {
      renderWithQueryClient(<ArticleList />)

      await waitFor(() => {
        expect(screen.getByText('以太坊 2.0 升级详解')).toBeInTheDocument()
      })

      // Both articles should be visible
      expect(screen.getByText('DeFi 借贷协议对比')).toBeInTheDocument()
    })
  })

  describe('empty state', () => {
    it('shows empty state when no articles match filter', async () => {
      // Set category filter to one with no articles
      mockSearchParams.set('category', 'cat-nonexistent')

      // Override handler to return empty articles
      server.use(
        http.get('/api/articles', () => {
          return HttpResponse.json({
            articles: [],
            total: 0,
            page: 1,
            limit: 20,
          })
        })
      )

      renderWithQueryClient(<ArticleList />)

      await waitFor(() => {
        expect(screen.getByText('暂无文章')).toBeInTheDocument()
      })
    })
  })

  describe('article cards', () => {
    it('renders articles in card format with proper structure', async () => {
      renderWithQueryClient(<ArticleList />)

      await waitFor(() => {
        expect(screen.getByText('以太坊 2.0 升级详解')).toBeInTheDocument()
      })

      // Articles should be in cards (check for card-related elements)
      const cards = document.querySelectorAll('[class*="card"]')
      expect(cards.length).toBeGreaterThanOrEqual(2)
    })

    it('shows relative time for article updates', async () => {
      renderWithQueryClient(<ArticleList />)

      await waitFor(() => {
        expect(screen.getByText('以太坊 2.0 升级详解')).toBeInTheDocument()
      })

      // Should show some time indicator (the exact text depends on date-fns locale)
      // The component uses formatDistanceToNow which adds "前" suffix in Chinese
      const timeElements = document.querySelectorAll('.text-muted-foreground')
      expect(timeElements.length).toBeGreaterThan(0)
    })
  })

  describe('error handling', () => {
    it('shows mock data warning when API fails', async () => {
      // Override handler to return error
      server.use(
        http.get('/api/articles', () => {
          return HttpResponse.error()
        })
      )

      renderWithQueryClient(<ArticleList />)

      // Wait for the error state and mock data fallback
      // The component shows mock data with warning banner when API fails
      await waitFor(
        () => {
          expect(screen.getByText('后端服务未连接，显示示例数据')).toBeInTheDocument()
        },
        { timeout: 3000 }
      )
    })
  })
})
