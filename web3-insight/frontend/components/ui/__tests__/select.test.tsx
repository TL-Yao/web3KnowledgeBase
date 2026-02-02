// components/ui/__tests__/select.test.tsx
import { describe, it, expect, vi, beforeAll } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from '../select'

// Mock pointer capture methods for Radix UI compatibility with JSDOM
beforeAll(() => {
  Element.prototype.hasPointerCapture = vi.fn(() => false)
  Element.prototype.setPointerCapture = vi.fn()
  Element.prototype.releasePointerCapture = vi.fn()
})

// Helper component for testing
function TestSelect({
  placeholder = 'Select an option',
  disabled = false,
  defaultValue,
  onValueChange,
}: {
  placeholder?: string
  disabled?: boolean
  defaultValue?: string
  onValueChange?: (value: string) => void
}) {
  return (
    <Select defaultValue={defaultValue} onValueChange={onValueChange} disabled={disabled}>
      <SelectTrigger disabled={disabled}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>Fruits</SelectLabel>
          <SelectItem value="apple">Apple</SelectItem>
          <SelectItem value="banana">Banana</SelectItem>
          <SelectItem value="orange">Orange</SelectItem>
        </SelectGroup>
        <SelectSeparator />
        <SelectGroup>
          <SelectLabel>Vegetables</SelectLabel>
          <SelectItem value="carrot">Carrot</SelectItem>
          <SelectItem value="potato" disabled>Potato (disabled)</SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  )
}

describe('Select', () => {
  describe('rendering', () => {
    it('renders with placeholder', () => {
      render(<TestSelect placeholder="Choose a fruit" />)

      const trigger = screen.getByRole('combobox')
      expect(trigger).toBeInTheDocument()
      expect(trigger).toHaveTextContent('Choose a fruit')
    })

    it('renders with data-slot attribute on trigger', () => {
      render(<TestSelect />)

      const trigger = screen.getByRole('combobox')
      expect(trigger).toHaveAttribute('data-slot', 'select-trigger')
    })

    it('renders with default size attribute', () => {
      render(<TestSelect />)

      const trigger = screen.getByRole('combobox')
      expect(trigger).toHaveAttribute('data-size', 'default')
    })
  })

  describe('dropdown behavior', () => {
    it('opens dropdown on click', async () => {
      const user = userEvent.setup()
      render(<TestSelect />)

      const trigger = screen.getByRole('combobox')
      await user.click(trigger)

      // Wait for the dropdown to open (Radix uses portals)
      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
      })

      // Check that options are visible
      expect(screen.getByRole('option', { name: 'Apple' })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: 'Banana' })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: 'Orange' })).toBeInTheDocument()
    })

    it('displays group labels when open', async () => {
      const user = userEvent.setup()
      render(<TestSelect />)

      await user.click(screen.getByRole('combobox'))

      await waitFor(() => {
        expect(screen.getByText('Fruits')).toBeInTheDocument()
        expect(screen.getByText('Vegetables')).toBeInTheDocument()
      })
    })
  })

  describe('selection behavior', () => {
    it('selects an option on click', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestSelect onValueChange={handleChange} />)

      // Open dropdown
      await user.click(screen.getByRole('combobox'))

      // Wait for dropdown to be visible
      await waitFor(() => {
        expect(screen.getByRole('option', { name: 'Banana' })).toBeInTheDocument()
      })

      // Click option
      await user.click(screen.getByRole('option', { name: 'Banana' }))

      // Verify callback was called with correct value
      expect(handleChange).toHaveBeenCalledWith('banana')
    })

    it('displays selected value in trigger', async () => {
      const user = userEvent.setup()
      render(<TestSelect />)

      const trigger = screen.getByRole('combobox')

      // Open dropdown
      await user.click(trigger)

      // Wait for options to be visible
      await waitFor(() => {
        expect(screen.getByRole('option', { name: 'Apple' })).toBeInTheDocument()
      })

      // Select option
      await user.click(screen.getByRole('option', { name: 'Apple' }))

      // Wait for dropdown to close and check trigger shows selected value
      await waitFor(() => {
        expect(trigger).toHaveTextContent('Apple')
      })
    })

    it('displays default value when provided', () => {
      render(<TestSelect defaultValue="orange" />)

      const trigger = screen.getByRole('combobox')
      expect(trigger).toHaveTextContent('Orange')
    })

    it('closes dropdown after selection', async () => {
      const user = userEvent.setup()
      render(<TestSelect />)

      // Open dropdown
      await user.click(screen.getByRole('combobox'))

      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
      })

      // Select option
      await user.click(screen.getByRole('option', { name: 'Carrot' }))

      // Dropdown should close
      await waitFor(() => {
        expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
      })
    })
  })

  describe('disabled state', () => {
    it('respects disabled state on trigger', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestSelect disabled onValueChange={handleChange} />)

      const trigger = screen.getByRole('combobox')
      expect(trigger).toBeDisabled()

      // Try to click
      await user.click(trigger)

      // Dropdown should not open
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    })

    it('shows disabled item with correct styling', async () => {
      const user = userEvent.setup()
      render(<TestSelect />)

      await user.click(screen.getByRole('combobox'))

      await waitFor(() => {
        const disabledOption = screen.getByRole('option', { name: 'Potato (disabled)' })
        expect(disabledOption).toHaveAttribute('data-disabled')
      })
    })

    it('does not select disabled item', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<TestSelect onValueChange={handleChange} />)

      await user.click(screen.getByRole('combobox'))

      await waitFor(() => {
        expect(screen.getByRole('option', { name: 'Potato (disabled)' })).toBeInTheDocument()
      })

      // Try to click disabled option
      await user.click(screen.getByRole('option', { name: 'Potato (disabled)' }))

      // Callback should not be called
      expect(handleChange).not.toHaveBeenCalled()
    })
  })

  describe('keyboard navigation', () => {
    it('opens dropdown with Enter key', async () => {
      const user = userEvent.setup()
      render(<TestSelect />)

      const trigger = screen.getByRole('combobox')
      trigger.focus()

      await user.keyboard('{Enter}')

      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
      })
    })

    it('opens dropdown with Space key', async () => {
      const user = userEvent.setup()
      render(<TestSelect />)

      const trigger = screen.getByRole('combobox')
      trigger.focus()

      await user.keyboard(' ')

      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
      })
    })

    it('opens dropdown with ArrowDown key', async () => {
      const user = userEvent.setup()
      render(<TestSelect />)

      const trigger = screen.getByRole('combobox')
      trigger.focus()

      await user.keyboard('{ArrowDown}')

      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
      })
    })

    it('closes dropdown with Escape key', async () => {
      const user = userEvent.setup()
      render(<TestSelect />)

      // Open dropdown using keyboard
      const trigger = screen.getByRole('combobox')
      trigger.focus()
      await user.keyboard('{Enter}')

      await waitFor(() => {
        expect(screen.getByRole('listbox')).toBeInTheDocument()
      })

      // Press Escape
      await user.keyboard('{Escape}')

      await waitFor(() => {
        expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
      })
    })
  })

  describe('SelectTrigger size variants', () => {
    it('applies sm size when specified', () => {
      render(
        <Select>
          <SelectTrigger size="sm">
            <SelectValue placeholder="Small" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="test">Test</SelectItem>
          </SelectContent>
        </Select>
      )

      const trigger = screen.getByRole('combobox')
      expect(trigger).toHaveAttribute('data-size', 'sm')
    })
  })

  describe('className prop', () => {
    it('merges custom className on trigger', () => {
      render(
        <Select>
          <SelectTrigger className="custom-trigger-class">
            <SelectValue placeholder="Custom" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="test">Test</SelectItem>
          </SelectContent>
        </Select>
      )

      const trigger = screen.getByRole('combobox')
      expect(trigger).toHaveClass('custom-trigger-class')
    })
  })
})
