// components/ui/__tests__/button.test.tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button } from '../button'

describe('Button', () => {
  describe('rendering', () => {
    it('renders with text', () => {
      render(<Button>Click me</Button>)
      expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument()
    })

    it('renders with data-slot attribute', () => {
      render(<Button>Click me</Button>)
      expect(screen.getByRole('button')).toHaveAttribute('data-slot', 'button')
    })
  })

  describe('click handling', () => {
    it('calls onClick when clicked', async () => {
      const user = userEvent.setup()
      const handleClick = vi.fn()
      render(<Button onClick={handleClick}>Click me</Button>)

      await user.click(screen.getByRole('button'))
      expect(handleClick).toHaveBeenCalledTimes(1)
    })

    it('does not call onClick when disabled', async () => {
      const user = userEvent.setup()
      const handleClick = vi.fn()
      render(<Button disabled onClick={handleClick}>Click me</Button>)

      await user.click(screen.getByRole('button'))
      expect(handleClick).not.toHaveBeenCalled()
    })
  })

  describe('variant prop', () => {
    it('applies default variant by default', () => {
      render(<Button>Default</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-variant', 'default')
      expect(button).toHaveClass('bg-primary')
    })

    it('applies destructive variant classes', () => {
      render(<Button variant="destructive">Destructive</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-variant', 'destructive')
      expect(button).toHaveClass('bg-destructive')
    })

    it('applies outline variant classes', () => {
      render(<Button variant="outline">Outline</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-variant', 'outline')
      expect(button).toHaveClass('border')
      expect(button).toHaveClass('bg-background')
    })

    it('applies secondary variant classes', () => {
      render(<Button variant="secondary">Secondary</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-variant', 'secondary')
      expect(button).toHaveClass('bg-secondary')
    })

    it('applies ghost variant classes', () => {
      render(<Button variant="ghost">Ghost</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-variant', 'ghost')
      expect(button).toHaveClass('hover:bg-accent')
    })

    it('applies link variant classes', () => {
      render(<Button variant="link">Link</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-variant', 'link')
      expect(button).toHaveClass('underline-offset-4')
    })
  })

  describe('size prop', () => {
    it('applies default size by default', () => {
      render(<Button>Default</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'default')
      expect(button).toHaveClass('h-9')
    })

    it('applies xs size classes', () => {
      render(<Button size="xs">Extra Small</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'xs')
      expect(button).toHaveClass('h-6')
      expect(button).toHaveClass('text-xs')
    })

    it('applies sm size classes', () => {
      render(<Button size="sm">Small</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'sm')
      expect(button).toHaveClass('h-8')
    })

    it('applies lg size classes', () => {
      render(<Button size="lg">Large</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'lg')
      expect(button).toHaveClass('h-10')
    })

    it('applies icon size classes', () => {
      render(<Button size="icon">I</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'icon')
      expect(button).toHaveClass('size-9')
    })

    it('applies icon-xs size classes', () => {
      render(<Button size="icon-xs">I</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'icon-xs')
      expect(button).toHaveClass('size-6')
    })

    it('applies icon-sm size classes', () => {
      render(<Button size="icon-sm">I</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'icon-sm')
      expect(button).toHaveClass('size-8')
    })

    it('applies icon-lg size classes', () => {
      render(<Button size="icon-lg">I</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveAttribute('data-size', 'icon-lg')
      expect(button).toHaveClass('size-10')
    })
  })

  describe('asChild prop', () => {
    it('renders as button element by default', () => {
      render(<Button>Click me</Button>)
      expect(screen.getByRole('button').tagName).toBe('BUTTON')
    })

    it('renders as child element when asChild is true', () => {
      render(
        <Button asChild>
          <a href="/test">Link Button</a>
        </Button>
      )
      const link = screen.getByRole('link', { name: 'Link Button' })
      expect(link).toBeInTheDocument()
      expect(link.tagName).toBe('A')
      expect(link).toHaveAttribute('href', '/test')
      expect(link).toHaveAttribute('data-slot', 'button')
    })
  })

  describe('className prop', () => {
    it('merges custom className with default classes', () => {
      render(<Button className="custom-class">Custom</Button>)
      const button = screen.getByRole('button')
      expect(button).toHaveClass('custom-class')
      expect(button).toHaveClass('inline-flex')
    })
  })

  describe('disabled state', () => {
    it('applies disabled styles when disabled', () => {
      render(<Button disabled>Disabled</Button>)
      const button = screen.getByRole('button')
      expect(button).toBeDisabled()
      expect(button).toHaveClass('disabled:opacity-50')
    })
  })
})
