// components/ui/__tests__/tabs.test.tsx
import { describe, it, expect, vi, beforeAll } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../tabs'

// Mock pointer capture methods for Radix UI compatibility with JSDOM
beforeAll(() => {
  Element.prototype.hasPointerCapture = vi.fn(() => false)
  Element.prototype.setPointerCapture = vi.fn()
  Element.prototype.releasePointerCapture = vi.fn()
})

// Helper component for testing
function TestTabs({
  defaultValue = 'tab1',
  onValueChange,
  disabledTab,
}: {
  defaultValue?: string
  onValueChange?: (value: string) => void
  disabledTab?: string
}) {
  return (
    <Tabs defaultValue={defaultValue} onValueChange={onValueChange}>
      <TabsList>
        <TabsTrigger value="tab1" disabled={disabledTab === 'tab1'}>
          Tab 1
        </TabsTrigger>
        <TabsTrigger value="tab2" disabled={disabledTab === 'tab2'}>
          Tab 2
        </TabsTrigger>
        <TabsTrigger value="tab3" disabled={disabledTab === 'tab3'}>
          Tab 3
        </TabsTrigger>
      </TabsList>
      <TabsContent value="tab1">Content for Tab 1</TabsContent>
      <TabsContent value="tab2">Content for Tab 2</TabsContent>
      <TabsContent value="tab3">Content for Tab 3</TabsContent>
    </Tabs>
  )
}

describe('Tabs', () => {
  describe('rendering', () => {
    it('renders all tab triggers', () => {
      render(<TestTabs />)

      expect(screen.getByRole('tab', { name: 'Tab 1' })).toBeInTheDocument()
      expect(screen.getByRole('tab', { name: 'Tab 2' })).toBeInTheDocument()
      expect(screen.getByRole('tab', { name: 'Tab 3' })).toBeInTheDocument()
    })

    it('renders tablist with correct role', () => {
      render(<TestTabs />)

      expect(screen.getByRole('tablist')).toBeInTheDocument()
    })

    it('renders with data-slot attributes', () => {
      render(<TestTabs />)

      expect(screen.getByRole('tablist')).toHaveAttribute('data-slot', 'tabs-list')
      expect(screen.getByRole('tab', { name: 'Tab 1' })).toHaveAttribute('data-slot', 'tabs-trigger')
    })

    it('renders with data-orientation attribute', () => {
      render(<TestTabs />)

      const tabs = screen.getByRole('tablist').closest('[data-slot="tabs"]')
      expect(tabs).toHaveAttribute('data-orientation', 'horizontal')
    })
  })

  describe('default tab content', () => {
    it('shows default tab content', () => {
      render(<TestTabs defaultValue="tab1" />)

      expect(screen.getByText('Content for Tab 1')).toBeInTheDocument()
    })

    it('shows correct default content for different default values', () => {
      render(<TestTabs defaultValue="tab2" />)

      expect(screen.getByText('Content for Tab 2')).toBeInTheDocument()
    })

    it('hides non-selected tab content', () => {
      render(<TestTabs defaultValue="tab1" />)

      // Tab 2 and Tab 3 content should not be visible
      expect(screen.queryByText('Content for Tab 2')).not.toBeInTheDocument()
      expect(screen.queryByText('Content for Tab 3')).not.toBeInTheDocument()
    })
  })

  describe('tab switching', () => {
    it('switches content when tab is clicked', async () => {
      const user = userEvent.setup()
      render(<TestTabs defaultValue="tab1" />)

      // Initially shows Tab 1 content
      expect(screen.getByText('Content for Tab 1')).toBeInTheDocument()

      // Click on Tab 2
      await user.click(screen.getByRole('tab', { name: 'Tab 2' }))

      // Now shows Tab 2 content
      await waitFor(() => {
        expect(screen.getByText('Content for Tab 2')).toBeInTheDocument()
      })

      // Tab 1 content should be hidden
      expect(screen.queryByText('Content for Tab 1')).not.toBeInTheDocument()
    })

    it('switches between multiple tabs correctly', async () => {
      const user = userEvent.setup()
      render(<TestTabs defaultValue="tab1" />)

      // Click Tab 3
      await user.click(screen.getByRole('tab', { name: 'Tab 3' }))

      await waitFor(() => {
        expect(screen.getByText('Content for Tab 3')).toBeInTheDocument()
      })

      // Click Tab 1
      await user.click(screen.getByRole('tab', { name: 'Tab 1' }))

      await waitFor(() => {
        expect(screen.getByText('Content for Tab 1')).toBeInTheDocument()
      })
    })
  })

  describe('onValueChange callback', () => {
    it('calls onValueChange when tab changes', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestTabs defaultValue="tab1" onValueChange={handleChange} />)

      await user.click(screen.getByRole('tab', { name: 'Tab 2' }))

      expect(handleChange).toHaveBeenCalledWith('tab2')
    })

    it('calls onValueChange with correct value for each tab', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestTabs defaultValue="tab1" onValueChange={handleChange} />)

      await user.click(screen.getByRole('tab', { name: 'Tab 3' }))
      expect(handleChange).toHaveBeenCalledWith('tab3')

      await user.click(screen.getByRole('tab', { name: 'Tab 1' }))
      expect(handleChange).toHaveBeenCalledWith('tab1')
    })

    it('does not call onValueChange when clicking already active tab', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestTabs defaultValue="tab1" onValueChange={handleChange} />)

      // Click on the already active tab
      await user.click(screen.getByRole('tab', { name: 'Tab 1' }))

      expect(handleChange).not.toHaveBeenCalled()
    })
  })

  describe('aria-selected attribute', () => {
    it('marks active tab as selected (aria-selected)', () => {
      render(<TestTabs defaultValue="tab1" />)

      expect(screen.getByRole('tab', { name: 'Tab 1' })).toHaveAttribute('aria-selected', 'true')
      expect(screen.getByRole('tab', { name: 'Tab 2' })).toHaveAttribute('aria-selected', 'false')
      expect(screen.getByRole('tab', { name: 'Tab 3' })).toHaveAttribute('aria-selected', 'false')
    })

    it('updates aria-selected when tab changes', async () => {
      const user = userEvent.setup()
      render(<TestTabs defaultValue="tab1" />)

      await user.click(screen.getByRole('tab', { name: 'Tab 2' }))

      await waitFor(() => {
        expect(screen.getByRole('tab', { name: 'Tab 1' })).toHaveAttribute('aria-selected', 'false')
        expect(screen.getByRole('tab', { name: 'Tab 2' })).toHaveAttribute('aria-selected', 'true')
        expect(screen.getByRole('tab', { name: 'Tab 3' })).toHaveAttribute('aria-selected', 'false')
      })
    })

    it('correctly reflects aria-selected for different default values', () => {
      render(<TestTabs defaultValue="tab3" />)

      expect(screen.getByRole('tab', { name: 'Tab 1' })).toHaveAttribute('aria-selected', 'false')
      expect(screen.getByRole('tab', { name: 'Tab 2' })).toHaveAttribute('aria-selected', 'false')
      expect(screen.getByRole('tab', { name: 'Tab 3' })).toHaveAttribute('aria-selected', 'true')
    })
  })

  describe('disabled state', () => {
    it('respects disabled state on trigger', () => {
      render(<TestTabs disabledTab="tab2" />)

      const disabledTab = screen.getByRole('tab', { name: 'Tab 2' })
      expect(disabledTab).toBeDisabled()
    })

    it('does not switch to disabled tab when clicked', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestTabs defaultValue="tab1" disabledTab="tab2" onValueChange={handleChange} />)

      // Try to click the disabled tab
      await user.click(screen.getByRole('tab', { name: 'Tab 2' }))

      // onValueChange should not be called
      expect(handleChange).not.toHaveBeenCalled()

      // Content should still be Tab 1
      expect(screen.getByText('Content for Tab 1')).toBeInTheDocument()
    })

    it('allows clicking non-disabled tabs when one is disabled', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestTabs defaultValue="tab1" disabledTab="tab2" onValueChange={handleChange} />)

      // Click Tab 3 (not disabled)
      await user.click(screen.getByRole('tab', { name: 'Tab 3' }))

      expect(handleChange).toHaveBeenCalledWith('tab3')
      await waitFor(() => {
        expect(screen.getByText('Content for Tab 3')).toBeInTheDocument()
      })
    })
  })

  describe('keyboard navigation', () => {
    it('navigates with arrow keys', async () => {
      const user = userEvent.setup()
      render(<TestTabs defaultValue="tab1" />)

      // Focus on the first tab
      screen.getByRole('tab', { name: 'Tab 1' }).focus()

      // Press ArrowRight to move to next tab
      await user.keyboard('{ArrowRight}')

      await waitFor(() => {
        expect(screen.getByRole('tab', { name: 'Tab 2' })).toHaveFocus()
      })
    })

    it('activates tab with Enter key', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestTabs defaultValue="tab1" onValueChange={handleChange} />)

      // Focus on Tab 2
      screen.getByRole('tab', { name: 'Tab 2' }).focus()

      // Press Enter to activate
      await user.keyboard('{Enter}')

      expect(handleChange).toHaveBeenCalledWith('tab2')
    })

    it('activates tab with Space key', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestTabs defaultValue="tab1" onValueChange={handleChange} />)

      // Focus on Tab 2
      screen.getByRole('tab', { name: 'Tab 2' }).focus()

      // Press Space to activate
      await user.keyboard(' ')

      expect(handleChange).toHaveBeenCalledWith('tab2')
    })
  })

  describe('TabsList variants', () => {
    it('applies default variant styling', () => {
      render(<TestTabs />)

      const tabsList = screen.getByRole('tablist')
      expect(tabsList).toHaveAttribute('data-variant', 'default')
    })

    it('applies line variant when specified', () => {
      render(
        <Tabs defaultValue="tab1">
          <TabsList variant="line">
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Content</TabsContent>
        </Tabs>
      )

      const tabsList = screen.getByRole('tablist')
      expect(tabsList).toHaveAttribute('data-variant', 'line')
    })
  })

  describe('className prop', () => {
    it('merges custom className on Tabs', () => {
      render(
        <Tabs defaultValue="tab1" className="custom-tabs-class">
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Content</TabsContent>
        </Tabs>
      )

      const tabs = screen.getByRole('tablist').closest('[data-slot="tabs"]')
      expect(tabs).toHaveClass('custom-tabs-class')
    })

    it('merges custom className on TabsList', () => {
      render(
        <Tabs defaultValue="tab1">
          <TabsList className="custom-list-class">
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Content</TabsContent>
        </Tabs>
      )

      const tabsList = screen.getByRole('tablist')
      expect(tabsList).toHaveClass('custom-list-class')
    })

    it('merges custom className on TabsTrigger', () => {
      render(
        <Tabs defaultValue="tab1">
          <TabsList>
            <TabsTrigger value="tab1" className="custom-trigger-class">
              Tab 1
            </TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Content</TabsContent>
        </Tabs>
      )

      const trigger = screen.getByRole('tab', { name: 'Tab 1' })
      expect(trigger).toHaveClass('custom-trigger-class')
    })

    it('merges custom className on TabsContent', () => {
      render(
        <Tabs defaultValue="tab1">
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1" className="custom-content-class">
            Content
          </TabsContent>
        </Tabs>
      )

      const content = screen.getByText('Content').closest('[data-slot="tabs-content"]')
      expect(content).toHaveClass('custom-content-class')
    })
  })

  describe('orientation', () => {
    it('defaults to horizontal orientation', () => {
      render(<TestTabs />)

      const tabs = screen.getByRole('tablist').closest('[data-slot="tabs"]')
      expect(tabs).toHaveAttribute('data-orientation', 'horizontal')
    })

    it('supports vertical orientation', () => {
      render(
        <Tabs defaultValue="tab1" orientation="vertical">
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Content</TabsContent>
        </Tabs>
      )

      const tabs = screen.getByRole('tablist').closest('[data-slot="tabs"]')
      expect(tabs).toHaveAttribute('data-orientation', 'vertical')
    })
  })
})
