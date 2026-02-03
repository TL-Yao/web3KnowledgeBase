// components/knowledge/__tests__/category-tree.test.tsx
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { http, HttpResponse } from 'msw'
import { server } from '@/__mocks__/server'
import { CategoryTree } from '../category-tree'
import { mockCategories } from '@/__mocks__/data'

// Add handler for categories tree endpoint (without base URL since NEXT_PUBLIC_API_URL is empty in tests)
// Use beforeEach because vitest.setup.tsx resets handlers afterEach
beforeEach(() => {
  server.use(
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
const mockPush = vi.fn()
const mockSearchParams = new URLSearchParams()

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
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

describe('CategoryTree', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockSearchParams.delete('category')
  })

  describe('loading state', () => {
    it('renders loading spinner while fetching categories', () => {
      renderWithQueryClient(<CategoryTree />)

      // Should show loading spinner initially
      const loadingContainer = document.querySelector('.animate-spin')
      expect(loadingContainer).toBeInTheDocument()
    })
  })

  describe('rendering categories from API', () => {
    it('renders categories from API (MSW mock data)', async () => {
      renderWithQueryClient(<CategoryTree />)

      // Wait for categories to load (from MSW mock handler)
      await waitFor(() => {
        expect(screen.getByText('Layer 1')).toBeInTheDocument()
      })

      expect(screen.getByText('DeFi')).toBeInTheDocument()
    })

    it('renders the "All" option at the top', async () => {
      renderWithQueryClient(<CategoryTree />)

      await waitFor(() => {
        expect(screen.getByText('全部')).toBeInTheDocument()
      })
    })
  })

  describe('expand/collapse behavior', () => {
    it('expands category to show children when clicked', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<CategoryTree />)

      // Wait for categories to load
      await waitFor(() => {
        expect(screen.getByText('Layer 1')).toBeInTheDocument()
      })

      // Initially, children should not be visible
      expect(screen.queryByText('Ethereum')).not.toBeInTheDocument()

      // Click on Layer 1 to expand it
      await user.click(screen.getByText('Layer 1'))

      // Now Ethereum child should be visible
      await waitFor(() => {
        expect(screen.getByText('Ethereum')).toBeInTheDocument()
      })
    })

    it('collapses expanded category when clicked again', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<CategoryTree />)

      // Wait for categories to load
      await waitFor(() => {
        expect(screen.getByText('Layer 1')).toBeInTheDocument()
      })

      // Click to expand
      await user.click(screen.getByText('Layer 1'))

      await waitFor(() => {
        expect(screen.getByText('Ethereum')).toBeInTheDocument()
      })

      // Click again to collapse
      await user.click(screen.getByText('Layer 1'))

      await waitFor(() => {
        expect(screen.queryByText('Ethereum')).not.toBeInTheDocument()
      })
    })
  })

  describe('selection behavior', () => {
    it('calls router.push when category is clicked', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<CategoryTree />)

      // Wait for categories to load
      await waitFor(() => {
        expect(screen.getByText('DeFi')).toBeInTheDocument()
      })

      // Click on DeFi category
      await user.click(screen.getByText('DeFi'))

      // Should call router.push with category id
      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/knowledge?category=cat-2')
      })
    })

    it('navigates to knowledge page without category when "All" is clicked', async () => {
      const user = userEvent.setup()
      // Set an initial category filter
      mockSearchParams.set('category', 'cat-1')

      renderWithQueryClient(<CategoryTree />)

      await waitFor(() => {
        expect(screen.getByText('全部')).toBeInTheDocument()
      })

      // Click on "All" option
      await user.click(screen.getByText('全部'))

      // Should navigate to /knowledge without category param
      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/knowledge')
      })
    })

    it('selects child category and navigates to correct URL', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<CategoryTree />)

      // Wait for categories to load
      await waitFor(() => {
        expect(screen.getByText('Layer 1')).toBeInTheDocument()
      })

      // Expand Layer 1
      await user.click(screen.getByText('Layer 1'))

      // Wait for children to appear
      await waitFor(() => {
        expect(screen.getByText('Ethereum')).toBeInTheDocument()
      })

      // Click on child category
      await user.click(screen.getByText('Ethereum'))

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/knowledge?category=cat-1-1')
      })
    })
  })

  describe('visual highlighting', () => {
    it('highlights "All" when no category is selected', async () => {
      renderWithQueryClient(<CategoryTree />)

      await waitFor(() => {
        expect(screen.getByText('全部')).toBeInTheDocument()
      })

      // The "All" option container should have the selected styling
      const allOption = screen.getByText('全部').closest('div')
      expect(allOption).toHaveClass('bg-primary/10')
      expect(allOption).toHaveClass('text-primary')
    })

    it('highlights selected category based on URL search params', async () => {
      // Set a category in search params
      mockSearchParams.set('category', 'cat-2')

      renderWithQueryClient(<CategoryTree />)

      await waitFor(() => {
        expect(screen.getByText('DeFi')).toBeInTheDocument()
      })

      // The DeFi category should have selected styling
      const defiOption = screen.getByText('DeFi').closest('div')
      expect(defiOption).toHaveClass('bg-primary/10')
      expect(defiOption).toHaveClass('text-primary')

      // The "All" option should NOT have selected styling
      const allOption = screen.getByText('全部').closest('div')
      expect(allOption).not.toHaveClass('bg-primary/10')
    })
  })

  describe('icons and visual elements', () => {
    it('shows folder icon for categories with children', async () => {
      renderWithQueryClient(<CategoryTree />)

      await waitFor(() => {
        expect(screen.getByText('Layer 1')).toBeInTheDocument()
      })

      // Layer 1 has children, should have a folder icon (SVG element)
      const layer1Row = screen.getByText('Layer 1').closest('div')
      const svgIcons = layer1Row?.querySelectorAll('svg')
      // Should have chevron and folder icons
      expect(svgIcons?.length).toBeGreaterThanOrEqual(2)
    })

    it('shows chevron icon that rotates when expanded', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<CategoryTree />)

      await waitFor(() => {
        expect(screen.getByText('Layer 1')).toBeInTheDocument()
      })

      // Find the chevron icon
      const layer1Row = screen.getByText('Layer 1').closest('div')
      const chevronIcon = layer1Row?.querySelector('svg')

      // Initially should not have rotate-90 class
      expect(chevronIcon).not.toHaveClass('rotate-90')

      // Click to expand
      await user.click(screen.getByText('Layer 1'))

      // After expansion, chevron should have rotate-90 class
      await waitFor(() => {
        const expandedChevron = screen.getByText('Layer 1').closest('div')?.querySelector('svg')
        expect(expandedChevron).toHaveClass('rotate-90')
      })
    })
  })
})
