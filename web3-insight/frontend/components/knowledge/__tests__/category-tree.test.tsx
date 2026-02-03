import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, beforeEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/__mocks__/server'
import { CategoryTree } from '../category-tree'

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

describe('CategoryTree', () => {
  beforeEach(() => {
    // Reset handlers after each test
    server.resetHandlers()
  })

  it('shows loading state initially', () => {
    // Set up handler that never resolves to keep loading state
    server.use(
      http.get('/api/categories/tree', () => {
        return new Promise(() => {})
      })
    )

    renderWithQueryClient(<CategoryTree />)
    expect(screen.getByText(/加载分类中/i)).toBeInTheDocument()
  })

  it('shows error state when API call fails', async () => {
    // Set up handler that returns error
    server.use(
      http.get('/api/categories/tree', () => {
        return new HttpResponse(null, { status: 500 })
      })
    )

    renderWithQueryClient(<CategoryTree />)

    await waitFor(
      () => {
        expect(screen.getByText(/加载分类失败，请稍后重试/i)).toBeInTheDocument()
      },
      { timeout: 3000 }
    )
  })

  it('shows empty state when no categories exist', async () => {
    // Set up handler that returns empty array
    server.use(
      http.get('/api/categories/tree', () => {
        return HttpResponse.json([])
      })
    )

    renderWithQueryClient(<CategoryTree />)

    await waitFor(() => {
      expect(screen.getByText(/暂无分类/i)).toBeInTheDocument()
    })
  })

  it('displays categories when data is loaded successfully', async () => {
    const mockCategories = [
      {
        id: '1',
        name: 'Blockchain',
        slug: 'blockchain',
        children: [
          {
            id: '2',
            name: 'Ethereum',
            slug: 'ethereum',
            children: [],
          },
        ],
      },
      {
        id: '3',
        name: 'DeFi',
        slug: 'defi',
        children: [],
      },
    ]

    // Set up handler that returns mock categories
    server.use(
      http.get('/api/categories/tree', () => {
        return HttpResponse.json(mockCategories)
      })
    )

    renderWithQueryClient(<CategoryTree />)

    await waitFor(() => {
      expect(screen.getByText('Blockchain')).toBeInTheDocument()
      expect(screen.getByText('DeFi')).toBeInTheDocument()
    })
  })
})
