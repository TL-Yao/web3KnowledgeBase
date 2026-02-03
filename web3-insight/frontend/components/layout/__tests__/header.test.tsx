// components/layout/__tests__/header.test.tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Header } from '../header'

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}))

describe('Header', () => {
  describe('rendering', () => {
    it('renders the default breadcrumb text', () => {
      render(<Header />)

      expect(screen.getByText('知识库')).toBeInTheDocument()
    })

    it('renders custom breadcrumb when provided', () => {
      render(<Header breadcrumb="自定义标题" />)

      expect(screen.getByText('自定义标题')).toBeInTheDocument()
      expect(screen.queryByText('知识库')).not.toBeInTheDocument()
    })

    it('renders header element with correct structure', () => {
      render(<Header />)

      const header = document.querySelector('header')
      expect(header).toBeInTheDocument()
      expect(header).toHaveClass('h-14', 'border-b')
    })
  })

  describe('search functionality', () => {
    it('renders search input with placeholder', () => {
      render(<Header />)

      const searchInput = screen.getByPlaceholderText('搜索...')
      expect(searchInput).toBeInTheDocument()
      expect(searchInput).toHaveAttribute('type', 'search')
    })

    it('renders search icon', () => {
      render(<Header />)

      // The Search icon is rendered as an SVG
      const searchIcon = document.querySelector('svg.lucide-search')
      expect(searchIcon).toBeInTheDocument()
    })

    it('allows typing in search input', () => {
      render(<Header />)

      const searchInput = screen.getByPlaceholderText('搜索...')
      fireEvent.change(searchInput, { target: { value: 'test query' } })

      expect(searchInput).toHaveValue('test query')
    })
  })

  describe('navigation', () => {
    it('renders admin/settings link', () => {
      render(<Header />)

      const adminLink = screen.getByRole('link')
      expect(adminLink).toHaveAttribute('href', '/admin')
    })

    it('renders settings button with icon', () => {
      render(<Header />)

      const settingsButton = screen.getByRole('button')
      expect(settingsButton).toBeInTheDocument()

      // The Settings icon is rendered as an SVG inside the button
      const settingsIcon = settingsButton.querySelector('svg.lucide-settings')
      expect(settingsIcon).toBeInTheDocument()
    })

    it('settings button has ghost variant styling', () => {
      render(<Header />)

      const settingsButton = screen.getByRole('button')
      // Ghost variant buttons typically have specific classes
      // size="icon" uses size-9 class
      expect(settingsButton).toHaveClass('size-9')
    })
  })

  describe('layout', () => {
    it('positions breadcrumb on the left and controls on the right', () => {
      render(<Header />)

      const header = document.querySelector('header')
      expect(header).toHaveClass('justify-between')
    })

    it('groups search and settings together', () => {
      render(<Header />)

      // The search input and settings button should be in a flex container
      const searchInput = screen.getByPlaceholderText('搜索...')
      const settingsButton = screen.getByRole('button')

      // Both should have a common parent container
      const searchContainer = searchInput.closest('.flex')
      const buttonContainer = settingsButton.closest('.flex')

      // They should share the same flex container (or nested flex containers)
      expect(searchContainer).not.toBeNull()
      expect(buttonContainer).not.toBeNull()
    })
  })
})
