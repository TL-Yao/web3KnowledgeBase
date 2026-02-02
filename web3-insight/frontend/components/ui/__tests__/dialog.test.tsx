// components/ui/__tests__/dialog.test.tsx
import { describe, it, expect, vi, beforeAll } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../dialog'

// Mock pointer capture methods for Radix UI compatibility with JSDOM
beforeAll(() => {
  Element.prototype.hasPointerCapture = vi.fn(() => false)
  Element.prototype.setPointerCapture = vi.fn()
  Element.prototype.releasePointerCapture = vi.fn()
})

// Helper component for testing uncontrolled dialog
function TestDialog({
  onOpenChange,
  showCloseButton = true,
}: {
  onOpenChange?: (open: boolean) => void
  showCloseButton?: boolean
}) {
  return (
    <Dialog onOpenChange={onOpenChange}>
      <DialogTrigger>Open Dialog</DialogTrigger>
      <DialogContent showCloseButton={showCloseButton}>
        <DialogHeader>
          <DialogTitle>Test Dialog Title</DialogTitle>
          <DialogDescription>This is a test dialog description.</DialogDescription>
        </DialogHeader>
        <div>Dialog body content</div>
        <DialogFooter>
          <DialogClose>Cancel</DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// Helper component for testing controlled dialog
function ControlledDialog({
  initialOpen = false,
  onOpenChange,
}: {
  initialOpen?: boolean
  onOpenChange?: (open: boolean) => void
}) {
  const [open, setOpen] = useState(initialOpen)

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen)
    onOpenChange?.(newOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger>Open Controlled Dialog</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Controlled Dialog</DialogTitle>
          <DialogDescription>This dialog is controlled.</DialogDescription>
        </DialogHeader>
      </DialogContent>
    </Dialog>
  )
}

describe('Dialog', () => {
  describe('rendering', () => {
    it('is hidden by default', () => {
      render(<TestDialog />)

      // Dialog content should not be in the document when closed
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      expect(screen.queryByText('Test Dialog Title')).not.toBeInTheDocument()
    })

    it('renders trigger button', () => {
      render(<TestDialog />)

      expect(screen.getByRole('button', { name: 'Open Dialog' })).toBeInTheDocument()
    })

    it('renders with data-slot attribute on trigger', () => {
      render(<TestDialog />)

      const trigger = screen.getByRole('button', { name: 'Open Dialog' })
      expect(trigger).toHaveAttribute('data-slot', 'dialog-trigger')
    })
  })

  describe('opening behavior', () => {
    it('opens when trigger is clicked', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      const trigger = screen.getByRole('button', { name: 'Open Dialog' })
      await user.click(trigger)

      // Wait for dialog to open (Radix uses portals)
      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })
    })

    it('contains title and description when open', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(screen.getByText('Test Dialog Title')).toBeInTheDocument()
        expect(screen.getByText('This is a test dialog description.')).toBeInTheDocument()
      })
    })

    it('renders dialog body content when open', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(screen.getByText('Dialog body content')).toBeInTheDocument()
      })
    })
  })

  describe('closing behavior', () => {
    it('closes when close button (X) is clicked', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      // Open dialog
      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Click the X close button (sr-only text "Close")
      const closeButton = screen.getByRole('button', { name: 'Close' })
      await user.click(closeButton)

      // Dialog should close
      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })

    it('closes when DialogClose button in footer is clicked', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      // Open dialog
      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Click the Cancel button (DialogClose in footer)
      const cancelButton = screen.getByRole('button', { name: 'Cancel' })
      await user.click(cancelButton)

      // Dialog should close
      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })

    it('closes when Escape key is pressed', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      // Open dialog
      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Press Escape
      await user.keyboard('{Escape}')

      // Dialog should close
      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
      })
    })
  })

  describe('onOpenChange callback', () => {
    it('calls onOpenChange with true when dialog opens', async () => {
      const user = userEvent.setup()
      const handleOpenChange = vi.fn()
      render(<TestDialog onOpenChange={handleOpenChange} />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(handleOpenChange).toHaveBeenCalledWith(true)
      })
    })

    it('calls onOpenChange with false when dialog closes via X button', async () => {
      const user = userEvent.setup()
      const handleOpenChange = vi.fn()
      render(<TestDialog onOpenChange={handleOpenChange} />)

      // Open dialog
      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Reset mock to check only the close call
      handleOpenChange.mockClear()

      // Close via X button
      await user.click(screen.getByRole('button', { name: 'Close' }))

      await waitFor(() => {
        expect(handleOpenChange).toHaveBeenCalledWith(false)
      })
    })

    it('calls onOpenChange with false when dialog closes via Escape', async () => {
      const user = userEvent.setup()
      const handleOpenChange = vi.fn()
      render(<TestDialog onOpenChange={handleOpenChange} />)

      // Open dialog
      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // Reset mock
      handleOpenChange.mockClear()

      // Close via Escape
      await user.keyboard('{Escape}')

      await waitFor(() => {
        expect(handleOpenChange).toHaveBeenCalledWith(false)
      })
    })
  })

  describe('controlled dialog', () => {
    it('renders open dialog when open prop is true', async () => {
      render(<ControlledDialog initialOpen={true} />)

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
        expect(screen.getByText('Controlled Dialog')).toBeInTheDocument()
      })
    })

    it('renders closed dialog when open prop is false', () => {
      render(<ControlledDialog initialOpen={false} />)

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('updates state when controlled dialog is opened', async () => {
      const user = userEvent.setup()
      const handleOpenChange = vi.fn()
      render(<ControlledDialog initialOpen={false} onOpenChange={handleOpenChange} />)

      await user.click(screen.getByRole('button', { name: 'Open Controlled Dialog' }))

      await waitFor(() => {
        expect(handleOpenChange).toHaveBeenCalledWith(true)
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })
    })
  })

  describe('showCloseButton prop', () => {
    it('shows close button by default', async () => {
      const user = userEvent.setup()
      render(<TestDialog showCloseButton={true} />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Close' })).toBeInTheDocument()
      })
    })

    it('hides close button when showCloseButton is false', async () => {
      const user = userEvent.setup()
      render(<TestDialog showCloseButton={false} />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })

      // The X close button should not be present (but Cancel button from footer should still work)
      // There should be no button with just "Close" as the name (the X button)
      const closeButtons = screen.queryAllByRole('button', { name: 'Close' })
      expect(closeButtons.length).toBe(0)
    })
  })

  describe('DialogHeader', () => {
    it('renders with data-slot attribute', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        const header = screen.getByRole('dialog').querySelector('[data-slot="dialog-header"]')
        expect(header).toBeInTheDocument()
      })
    })
  })

  describe('DialogTitle', () => {
    it('renders with data-slot attribute', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        const title = screen.getByText('Test Dialog Title')
        expect(title).toHaveAttribute('data-slot', 'dialog-title')
      })
    })
  })

  describe('DialogDescription', () => {
    it('renders with data-slot attribute', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        const description = screen.getByText('This is a test dialog description.')
        expect(description).toHaveAttribute('data-slot', 'dialog-description')
      })
    })
  })

  describe('DialogFooter', () => {
    it('renders with data-slot attribute', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        const footer = screen.getByRole('dialog').querySelector('[data-slot="dialog-footer"]')
        expect(footer).toBeInTheDocument()
      })
    })
  })

  describe('DialogContent', () => {
    it('renders with data-slot attribute', async () => {
      const user = userEvent.setup()
      render(<TestDialog />)

      await user.click(screen.getByRole('button', { name: 'Open Dialog' }))

      await waitFor(() => {
        const content = screen.getByRole('dialog')
        expect(content).toHaveAttribute('data-slot', 'dialog-content')
      })
    })
  })
})
