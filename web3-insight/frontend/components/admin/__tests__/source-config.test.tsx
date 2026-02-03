// components/admin/__tests__/source-config.test.tsx
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { http, HttpResponse } from 'msw'
import { server } from '@/__mocks__/server'
import { SourceConfig } from '../source-config'
import { mockDataSources } from '@/__mocks__/data'

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

// Add handlers for API endpoints
beforeEach(() => {
  vi.clearAllMocks()

  server.use(
    http.get('/api/sources', () => {
      return HttpResponse.json(mockDataSources)
    }),
    http.post('/api/sources', async ({ request }) => {
      const body = await request.json() as { name: string; type: string; url: string }
      return HttpResponse.json({
        id: 'new-ds-1',
        name: body.name,
        type: body.type,
        url: body.url,
        enabled: true,
        fetchInterval: 3600,
        createdAt: new Date().toISOString(),
      })
    }),
    http.delete('/api/sources/:id', () => {
      return HttpResponse.json({ message: 'Deleted' })
    }),
    http.post('/api/sources/:id/sync', () => {
      return HttpResponse.json({
        message: 'Sync completed',
        itemsFound: 10,
        itemsNew: 3,
      })
    }),
    http.put('/api/sources/:id', async ({ request }) => {
      const body = await request.json() as { name: string; type: string; url: string; enabled: boolean }
      return HttpResponse.json({
        id: 'ds-1',
        ...body,
        createdAt: mockDataSources[0].createdAt,
      })
    }),
    http.post('/api/sources/validate', async ({ request }) => {
      const body = await request.json() as { url: string; type: string }
      if (body.url.includes('invalid')) {
        return HttpResponse.json({
          valid: false,
          error: 'Invalid URL format',
        })
      }
      return HttpResponse.json({
        valid: true,
        title: 'Valid Feed Title',
        description: 'A valid RSS feed',
        itemCount: 25,
      })
    })
  )
})

describe('SourceConfig', () => {
  describe('initial rendering', () => {
    it('renders data sources list', async () => {
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByText('CoinDesk RSS')).toBeInTheDocument()
      })
    })

    it('shows source type badge (RSS)', async () => {
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByText('RSS')).toBeInTheDocument()
      })
    })

    it('shows enabled/disabled status via switch', async () => {
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        // The switch should be present and checked (enabled) for the mock source
        const switchElement = screen.getByRole('switch')
        expect(switchElement).toBeInTheDocument()
        expect(switchElement).toHaveAttribute('data-state', 'checked')
      })
    })

    it('shows add source button', async () => {
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })
    })

    it('shows last sync time', async () => {
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        // The mock data source has lastFetchedAt set, so it should show relative time
        expect(screen.getByText(/上次同步/i)).toBeInTheDocument()
      })
    })

    it('renders card title', async () => {
      renderWithQueryClient(<SourceConfig />)

      // The card should have the title "数据源配置"
      await waitFor(() => {
        expect(screen.getByText('数据源配置')).toBeInTheDocument()
      })
    })

    it('shows empty state when no sources exist', async () => {
      server.use(
        http.get('/api/sources', () => {
          return HttpResponse.json([])
        })
      )

      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByText(/暂无数据源/i)).toBeInTheDocument()
      })
    })
  })

  describe('add source dialog', () => {
    it('opens add source dialog when clicking button', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
        expect(screen.getByText('添加数据源', { selector: 'h2' })).toBeInTheDocument()
      })
    })

    it('shows form fields in add source dialog', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        // Check for labels by their text content
        expect(screen.getByText('类型')).toBeInTheDocument()
        expect(screen.getByText('URL')).toBeInTheDocument()
        expect(screen.getByText('名称')).toBeInTheDocument()
        expect(screen.getByText(/抓取间隔/i)).toBeInTheDocument()
      })
    })

    it('closes dialog when clicking cancel', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /取消/i }))

      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })
  })

  describe('URL validation', () => {
    it('validates URL input successfully', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Find URL input and type a valid URL
      const urlInput = screen.getByPlaceholderText('https://example.com/feed.xml')
      await user.type(urlInput, 'https://example.com/feed.xml')

      // Click validate button
      const validateButton = screen.getByRole('button', { name: /验证/i })
      await user.click(validateButton)

      await waitFor(() => {
        expect(screen.getByText(/有效/i)).toBeInTheDocument()
        expect(screen.getByText(/Valid Feed Title/i)).toBeInTheDocument()
      })
    })

    it('shows validation error for invalid URL', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Type invalid URL
      const urlInput = screen.getByPlaceholderText('https://example.com/feed.xml')
      await user.type(urlInput, 'https://invalid-url.com/feed')

      // Click validate button
      const validateButton = screen.getByRole('button', { name: /验证/i })
      await user.click(validateButton)

      await waitFor(() => {
        expect(screen.getByText(/Invalid URL format/i)).toBeInTheDocument()
      })
    })

    it('auto-fills name from validation result', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      const urlInput = screen.getByPlaceholderText('https://example.com/feed.xml')
      await user.type(urlInput, 'https://example.com/feed.xml')

      const validateButton = screen.getByRole('button', { name: /验证/i })
      await user.click(validateButton)

      await waitFor(() => {
        const nameInput = screen.getByPlaceholderText('数据源名称') as HTMLInputElement
        expect(nameInput.value).toBe('Valid Feed Title')
      })
    })

    it('disables validate button when URL is empty', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      const validateButton = screen.getByRole('button', { name: /验证/i })
      expect(validateButton).toBeDisabled()
    })
  })

  describe('manual sync', () => {
    it('triggers manual sync when clicking sync button', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByText('CoinDesk RSS')).toBeInTheDocument()
      })

      // Find all buttons and filter to get icon-only buttons in the source row
      const allButtons = screen.getAllByRole('button')
      // Icon buttons have no text content (only contain SVG icons)
      const iconButtons = allButtons.filter(btn => {
        const textContent = btn.textContent?.trim()
        return !textContent || textContent === ''
      })

      // There should be at least 2 icon buttons (sync and delete)
      expect(iconButtons.length).toBeGreaterThanOrEqual(2)

      // The sync button is the first one in the actions area
      // Click it and verify no errors occur
      await user.click(iconButtons[0])

      // After sync, the source list should still be visible
      await waitFor(() => {
        expect(screen.getByText('CoinDesk RSS')).toBeInTheDocument()
      })
    })
  })

  describe('delete source', () => {
    it('can delete a source with confirmation', async () => {
      const user = userEvent.setup()
      const confirmMock = vi.fn().mockReturnValue(true)
      vi.stubGlobal('confirm', confirmMock)

      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByText('CoinDesk RSS')).toBeInTheDocument()
      })

      // Find all buttons and filter to get icon-only buttons
      const allButtons = screen.getAllByRole('button')
      const iconButtons = allButtons.filter(btn => {
        const textContent = btn.textContent?.trim()
        return !textContent || textContent === ''
      })

      // Second icon button should be delete (Trash2 icon)
      expect(iconButtons.length).toBeGreaterThanOrEqual(2)
      await user.click(iconButtons[1])

      expect(confirmMock).toHaveBeenCalledWith('确定删除此数据源？')

      vi.unstubAllGlobals()
    })

    it('does not delete when confirmation is cancelled', async () => {
      const user = userEvent.setup()
      const confirmMock = vi.fn().mockReturnValue(false)
      vi.stubGlobal('confirm', confirmMock)

      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByText('CoinDesk RSS')).toBeInTheDocument()
      })

      const allButtons = screen.getAllByRole('button')
      const iconButtons = allButtons.filter(btn => {
        const textContent = btn.textContent?.trim()
        return !textContent || textContent === ''
      })

      await user.click(iconButtons[1])

      expect(confirmMock).toHaveBeenCalled()

      // Source should still be visible
      expect(screen.getByText('CoinDesk RSS')).toBeInTheDocument()

      vi.unstubAllGlobals()
    })
  })

  describe('toggle source enabled state', () => {
    it('toggles enabled state when clicking switch', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByText('CoinDesk RSS')).toBeInTheDocument()
      })

      const switchElement = screen.getByRole('switch')
      expect(switchElement).toHaveAttribute('data-state', 'checked')

      await user.click(switchElement)

      // The toggle mutation should be triggered
      await waitFor(() => {
        // Component re-renders after mutation
        expect(screen.getByText('CoinDesk RSS')).toBeInTheDocument()
      })
    })
  })

  describe('create new source', () => {
    it('creates a new source successfully', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Fill in the form
      const urlInput = screen.getByPlaceholderText('https://example.com/feed.xml')
      await user.type(urlInput, 'https://example.com/new-feed.xml')

      const nameInput = screen.getByPlaceholderText('数据源名称')
      await user.type(nameInput, 'New Test Feed')

      // Click create button
      const createButton = screen.getByRole('button', { name: /创建/i })
      await user.click(createButton)

      // Dialog should close after successful creation
      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })

    it('shows error when name or URL is missing', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Try to create without filling required fields
      const createButton = screen.getByRole('button', { name: /创建/i })
      await user.click(createButton)

      // Form should stay open (toast error shown but dialog not closed)
      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })
    })
  })

  describe('source type selection', () => {
    it('shows type select in dialog', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Check that the type select is present (default is RSS)
      const typeSelect = screen.getByRole('combobox')
      expect(typeSelect).toBeInTheDocument()
      expect(screen.getByText('RSS 订阅')).toBeInTheDocument()
    })
  })

  describe('fetch interval display', () => {
    it('displays formatted fetch interval', async () => {
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        // Mock source has fetchInterval: 3600 (1 hour)
        expect(screen.getByText(/间隔: 1 小时/i)).toBeInTheDocument()
      })
    })

    it('shows interval description in dialog', async () => {
      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
        // Default interval is 3600 seconds = 1 hour
        expect(screen.getByText(/当前设置: 1 小时/i)).toBeInTheDocument()
      })
    })
  })

  describe('loading states', () => {
    it('shows loading state during validation', async () => {
      server.use(
        http.post('/api/sources/validate', async () => {
          await new Promise(resolve => setTimeout(resolve, 100))
          return HttpResponse.json({
            valid: true,
            title: 'Test Feed',
          })
        })
      )

      const user = userEvent.setup()
      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /添加数据源/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /添加数据源/i }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      const urlInput = screen.getByPlaceholderText('https://example.com/feed.xml')
      await user.type(urlInput, 'https://example.com/feed.xml')

      const validateButton = screen.getByRole('button', { name: /验证/i })
      await user.click(validateButton)

      // Button should be disabled during loading
      expect(validateButton).toBeDisabled()

      await waitFor(() => {
        expect(screen.getByText(/有效/i)).toBeInTheDocument()
      })
    })
  })

  describe('error badge', () => {
    it('shows error badge when source has lastError', async () => {
      server.use(
        http.get('/api/sources', () => {
          return HttpResponse.json([{
            ...mockDataSources[0],
            lastError: 'Connection timeout',
          }])
        })
      )

      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByText(/错误/i)).toBeInTheDocument()
      })
    })
  })

  describe('URL truncation', () => {
    it('truncates long URLs in the list', async () => {
      server.use(
        http.get('/api/sources', () => {
          return HttpResponse.json([{
            ...mockDataSources[0],
            url: 'https://www.example.com/very/long/path/to/a/rss/feed/that/exceeds/fifty/characters/in/total.xml',
          }])
        })
      )

      renderWithQueryClient(<SourceConfig />)

      await waitFor(() => {
        expect(screen.getByText(/\.\.\.$/)).toBeInTheDocument()
      })
    })
  })
})
