// components/layout/__tests__/sidebar.test.tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Sidebar } from '../sidebar'

// Mock next/link
vi.mock('next/link', () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}))

// Mock the CategoryTree component since it has complex dependencies
vi.mock('@/components/knowledge/category-tree', () => ({
  CategoryTree: () => <div data-testid="category-tree">Category Tree Mock</div>,
}))

// Mock the ScrollArea component
vi.mock('@/components/ui/scroll-area', () => ({
  ScrollArea: ({ children, className }: { children: React.ReactNode; className?: string }) => (
    <div className={className} data-testid="scroll-area">{children}</div>
  ),
}))

describe('Sidebar', () => {
  describe('rendering', () => {
    it('renders sidebar element with correct structure', () => {
      render(<Sidebar />)

      const sidebar = document.querySelector('aside')
      expect(sidebar).toBeInTheDocument()
      expect(sidebar).toHaveClass('w-64', 'border-r')
    })

    it('applies custom className when provided', () => {
      render(<Sidebar className="custom-class" />)

      const sidebar = document.querySelector('aside')
      expect(sidebar).toHaveClass('custom-class')
    })

    it('renders the logo section', () => {
      render(<Sidebar />)

      expect(screen.getByText('W3')).toBeInTheDocument()
      expect(screen.getByText('Web3 Insight')).toBeInTheDocument()
    })

    it('logo links to homepage', () => {
      render(<Sidebar />)

      const logoLink = screen.getByRole('link', { name: /W3 Web3 Insight/i })
      expect(logoLink).toHaveAttribute('href', '/')
    })
  })

  describe('navigation items', () => {
    it('renders all main navigation items', () => {
      render(<Sidebar />)

      expect(screen.getByText('知识库')).toBeInTheDocument()
      expect(screen.getByText('新闻')).toBeInTheDocument()
      expect(screen.getByText('即时研究')).toBeInTheDocument()
      expect(screen.getByText('后台管理')).toBeInTheDocument()
    })

    it('renders navigation links with correct hrefs', () => {
      render(<Sidebar />)

      const knowledgeLink = screen.getByRole('link', { name: /知识库/ })
      expect(knowledgeLink).toHaveAttribute('href', '/knowledge')

      const newsLink = screen.getByRole('link', { name: /新闻/ })
      expect(newsLink).toHaveAttribute('href', '/news')

      const researchLink = screen.getByRole('link', { name: /即时研究/ })
      expect(researchLink).toHaveAttribute('href', '/research')

      const adminLink = screen.getByRole('link', { name: /后台管理/ })
      expect(adminLink).toHaveAttribute('href', '/admin')
    })

    it('renders navigation icons', () => {
      render(<Sidebar />)

      // Check for the presence of lucide icons
      const fileTextIcon = document.querySelector('svg.lucide-file-text')
      expect(fileTextIcon).toBeInTheDocument()

      const newspaperIcon = document.querySelector('svg.lucide-newspaper')
      expect(newspaperIcon).toBeInTheDocument()

      const searchIcon = document.querySelector('svg.lucide-search')
      expect(searchIcon).toBeInTheDocument()

      const settingsIcon = document.querySelector('svg.lucide-settings')
      expect(settingsIcon).toBeInTheDocument()
    })
  })

  describe('category section', () => {
    it('renders category section header', () => {
      render(<Sidebar />)

      expect(screen.getByText('分类')).toBeInTheDocument()
    })

    it('renders CategoryTree component', () => {
      render(<Sidebar />)

      expect(screen.getByTestId('category-tree')).toBeInTheDocument()
    })

    it('renders ScrollArea for category tree', () => {
      render(<Sidebar />)

      expect(screen.getByTestId('scroll-area')).toBeInTheDocument()
    })
  })

  describe('recent section', () => {
    it('renders recent section header', () => {
      render(<Sidebar />)

      expect(screen.getByText('最近阅读')).toBeInTheDocument()
    })

    it('renders recent reading items', () => {
      render(<Sidebar />)

      expect(screen.getByText('zkSync 工作原理')).toBeInTheDocument()
      expect(screen.getByText('Cosmos IBC 详解')).toBeInTheDocument()
    })
  })

  describe('layout structure', () => {
    it('has proper flex layout', () => {
      render(<Sidebar />)

      const sidebar = document.querySelector('aside')
      expect(sidebar).toHaveClass('flex', 'flex-col')
    })

    it('has border between logo and navigation', () => {
      render(<Sidebar />)

      const logoSection = document.querySelector('.h-14.border-b')
      expect(logoSection).toBeInTheDocument()
    })

    it('has border above recent section', () => {
      render(<Sidebar />)

      const recentSection = document.querySelector('.border-t')
      expect(recentSection).toBeInTheDocument()
    })
  })
})
