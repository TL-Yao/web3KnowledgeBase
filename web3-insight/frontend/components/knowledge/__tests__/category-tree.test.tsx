import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { CategoryTree } from '../category-tree'

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('CategoryTree', () => {
  it('shows loading state initially', () => {
    render(<CategoryTree />, { wrapper: createWrapper() })
    expect(screen.getByText(/加载分类中/i)).toBeInTheDocument()
  })
})
