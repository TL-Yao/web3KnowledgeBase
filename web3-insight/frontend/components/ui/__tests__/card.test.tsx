// components/ui/__tests__/card.test.tsx
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
  CardAction,
} from '../card'

describe('Card', () => {
  describe('rendering', () => {
    it('renders card with all parts', () => {
      render(
        <Card>
          <CardHeader>
            <CardTitle>Test Title</CardTitle>
            <CardDescription>Test Description</CardDescription>
            <CardAction>Action</CardAction>
          </CardHeader>
          <CardContent>Test Content</CardContent>
          <CardFooter>Test Footer</CardFooter>
        </Card>
      )

      expect(screen.getByText('Test Title')).toBeInTheDocument()
      expect(screen.getByText('Test Description')).toBeInTheDocument()
      expect(screen.getByText('Action')).toBeInTheDocument()
      expect(screen.getByText('Test Content')).toBeInTheDocument()
      expect(screen.getByText('Test Footer')).toBeInTheDocument()
    })

    it('applies custom className to Card', () => {
      render(<Card className="custom-card-class">Card content</Card>)
      const card = screen.getByText('Card content').closest('[data-slot="card"]')
      expect(card).toHaveClass('custom-card-class')
      expect(card).toHaveClass('rounded-xl')
    })

    it('renders CardTitle correctly', () => {
      render(
        <Card>
          <CardHeader>
            <CardTitle>Important Title</CardTitle>
          </CardHeader>
        </Card>
      )

      const title = screen.getByText('Important Title')
      expect(title).toBeInTheDocument()
      expect(title).toHaveClass('font-semibold')
      expect(title).toHaveAttribute('data-slot', 'card-title')
    })

    it('renders only content without header/footer', () => {
      render(
        <Card>
          <CardContent>Standalone content</CardContent>
        </Card>
      )

      expect(screen.getByText('Standalone content')).toBeInTheDocument()
      expect(screen.queryByText('header')).not.toBeInTheDocument()
      expect(screen.queryByText('footer')).not.toBeInTheDocument()
    })
  })

  describe('data-slot attributes', () => {
    it('Card applies correct data-slot attribute', () => {
      render(<Card data-testid="card">Content</Card>)
      expect(screen.getByTestId('card')).toHaveAttribute('data-slot', 'card')
    })

    it('CardHeader applies correct data-slot attribute', () => {
      render(<CardHeader data-testid="card-header">Header</CardHeader>)
      expect(screen.getByTestId('card-header')).toHaveAttribute('data-slot', 'card-header')
    })

    it('CardTitle applies correct data-slot attribute', () => {
      render(<CardTitle data-testid="card-title">Title</CardTitle>)
      expect(screen.getByTestId('card-title')).toHaveAttribute('data-slot', 'card-title')
    })

    it('CardDescription applies correct data-slot attribute', () => {
      render(<CardDescription data-testid="card-description">Description</CardDescription>)
      expect(screen.getByTestId('card-description')).toHaveAttribute('data-slot', 'card-description')
    })

    it('CardContent applies correct data-slot attribute', () => {
      render(<CardContent data-testid="card-content">Content</CardContent>)
      expect(screen.getByTestId('card-content')).toHaveAttribute('data-slot', 'card-content')
    })

    it('CardFooter applies correct data-slot attribute', () => {
      render(<CardFooter data-testid="card-footer">Footer</CardFooter>)
      expect(screen.getByTestId('card-footer')).toHaveAttribute('data-slot', 'card-footer')
    })

    it('CardAction applies correct data-slot attribute', () => {
      render(<CardAction data-testid="card-action">Action</CardAction>)
      expect(screen.getByTestId('card-action')).toHaveAttribute('data-slot', 'card-action')
    })
  })

  describe('className prop', () => {
    it('merges custom className with CardHeader', () => {
      render(<CardHeader className="custom-header" data-testid="header">Header</CardHeader>)
      expect(screen.getByTestId('header')).toHaveClass('custom-header')
      expect(screen.getByTestId('header')).toHaveClass('px-6')
    })

    it('merges custom className with CardContent', () => {
      render(<CardContent className="custom-content" data-testid="content">Content</CardContent>)
      expect(screen.getByTestId('content')).toHaveClass('custom-content')
      expect(screen.getByTestId('content')).toHaveClass('px-6')
    })

    it('merges custom className with CardFooter', () => {
      render(<CardFooter className="custom-footer" data-testid="footer">Footer</CardFooter>)
      expect(screen.getByTestId('footer')).toHaveClass('custom-footer')
      expect(screen.getByTestId('footer')).toHaveClass('px-6')
    })

    it('merges custom className with CardDescription', () => {
      render(<CardDescription className="custom-desc" data-testid="desc">Desc</CardDescription>)
      expect(screen.getByTestId('desc')).toHaveClass('custom-desc')
      expect(screen.getByTestId('desc')).toHaveClass('text-sm')
    })

    it('merges custom className with CardAction', () => {
      render(<CardAction className="custom-action" data-testid="action">Action</CardAction>)
      expect(screen.getByTestId('action')).toHaveClass('custom-action')
      expect(screen.getByTestId('action')).toHaveClass('col-start-2')
    })
  })
})
