// components/ui/__tests__/input.test.tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Input } from '../input'

describe('Input', () => {
  describe('rendering', () => {
    it('renders with placeholder', () => {
      render(<Input placeholder="Enter your name" />)
      expect(screen.getByPlaceholderText('Enter your name')).toBeInTheDocument()
    })

    it('renders with data-slot attribute', () => {
      render(<Input placeholder="Test" />)
      expect(screen.getByRole('textbox')).toHaveAttribute('data-slot', 'input')
    })

    it('renders as textbox role by default', () => {
      render(<Input />)
      expect(screen.getByRole('textbox')).toBeInTheDocument()
    })
  })

  describe('user input', () => {
    it('accepts user input', async () => {
      const user = userEvent.setup()
      render(<Input placeholder="Type here" />)

      const input = screen.getByPlaceholderText('Type here')
      await user.type(input, 'Hello World')
      expect(input).toHaveValue('Hello World')
    })

    it('calls onChange when value changes', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<Input onChange={handleChange} placeholder="Test" />)

      const input = screen.getByPlaceholderText('Test')
      await user.type(input, 'a')
      expect(handleChange).toHaveBeenCalled()
    })

    it('updates value on each keystroke', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<Input onChange={handleChange} placeholder="Test" />)

      const input = screen.getByPlaceholderText('Test')
      await user.type(input, 'abc')
      expect(handleChange).toHaveBeenCalledTimes(3)
    })
  })

  describe('disabled state', () => {
    it('respects disabled state', () => {
      render(<Input disabled placeholder="Disabled input" />)
      const input = screen.getByPlaceholderText('Disabled input')
      expect(input).toBeDisabled()
    })

    it('applies disabled styles when disabled', () => {
      render(<Input disabled placeholder="Disabled" />)
      const input = screen.getByPlaceholderText('Disabled')
      expect(input).toHaveClass('disabled:opacity-50')
      expect(input).toHaveClass('disabled:pointer-events-none')
    })

    it('does not accept input when disabled', async () => {
      const user = userEvent.setup()
      const handleChange = vi.fn()
      render(<Input disabled onChange={handleChange} placeholder="Test" />)

      const input = screen.getByPlaceholderText('Test')
      await user.type(input, 'test')
      expect(handleChange).not.toHaveBeenCalled()
    })
  })

  describe('input types', () => {
    it('handles text type (default)', () => {
      render(<Input type="text" placeholder="Text input" />)
      const input = screen.getByPlaceholderText('Text input')
      expect(input).toHaveAttribute('type', 'text')
    })

    it('handles email type', () => {
      render(<Input type="email" placeholder="Email input" />)
      const input = screen.getByPlaceholderText('Email input')
      expect(input).toHaveAttribute('type', 'email')
    })

    it('handles password type', () => {
      render(<Input type="password" placeholder="Password input" />)
      const input = screen.getByPlaceholderText('Password input')
      expect(input).toHaveAttribute('type', 'password')
    })

    it('handles number type', () => {
      render(<Input type="number" placeholder="Number input" />)
      const input = screen.getByPlaceholderText('Number input')
      expect(input).toHaveAttribute('type', 'number')
    })

    it('handles search type', () => {
      render(<Input type="search" placeholder="Search input" />)
      const input = screen.getByPlaceholderText('Search input')
      expect(input).toHaveAttribute('type', 'search')
    })
  })

  describe('className prop', () => {
    it('applies custom className', () => {
      render(<Input className="custom-class" placeholder="Test" />)
      const input = screen.getByPlaceholderText('Test')
      expect(input).toHaveClass('custom-class')
    })

    it('merges custom className with default classes', () => {
      render(<Input className="custom-class" placeholder="Test" />)
      const input = screen.getByPlaceholderText('Test')
      expect(input).toHaveClass('custom-class')
      expect(input).toHaveClass('rounded-md')
      expect(input).toHaveClass('border')
    })
  })

  describe('additional props', () => {
    it('forwards additional props to input element', () => {
      render(<Input placeholder="Test" data-testid="custom-input" aria-label="Custom label" />)
      const input = screen.getByPlaceholderText('Test')
      expect(input).toHaveAttribute('data-testid', 'custom-input')
      expect(input).toHaveAttribute('aria-label', 'Custom label')
    })

    it('supports controlled value', () => {
      render(<Input value="controlled value" onChange={() => {}} placeholder="Test" />)
      const input = screen.getByPlaceholderText('Test')
      expect(input).toHaveValue('controlled value')
    })

    it('supports defaultValue', () => {
      render(<Input defaultValue="default value" placeholder="Test" />)
      const input = screen.getByPlaceholderText('Test')
      expect(input).toHaveValue('default value')
    })

    it('supports required attribute', () => {
      render(<Input required placeholder="Required input" />)
      const input = screen.getByPlaceholderText('Required input')
      expect(input).toBeRequired()
    })

    it('supports readOnly attribute', () => {
      render(<Input readOnly value="readonly" placeholder="Test" />)
      const input = screen.getByPlaceholderText('Test')
      expect(input).toHaveAttribute('readonly')
    })
  })
})
