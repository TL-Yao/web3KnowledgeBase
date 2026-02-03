// components/chat/__tests__/floating-chat.test.tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FloatingChat } from '../floating-chat'

// Mock the useChat hook
const mockSendMessage = vi.fn()
const mockClearMessages = vi.fn()

vi.mock('@/hooks/use-chat', () => ({
  useChat: vi.fn(() => ({
    messages: [],
    isLoading: false,
    currentResponse: '',
    sendMessage: mockSendMessage,
    clearMessages: mockClearMessages
  }))
}))

// Import the mocked module
import { useChat } from '@/hooks/use-chat'

describe('FloatingChat', () => {
  const defaultProps = {
    articleId: 'article-123',
    articleTitle: 'Test Article Title'
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Reset mock to default values
    vi.mocked(useChat).mockReturnValue({
      messages: [],
      isLoading: false,
      currentResponse: '',
      sendMessage: mockSendMessage,
      clearMessages: mockClearMessages
    })
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('renders chat toggle button initially', () => {
      render(<FloatingChat {...defaultProps} />)

      const toggleButton = screen.getByRole('button', { name: /打开问答/i })
      expect(toggleButton).toBeInTheDocument()
    })

    it('renders MessageCircle icon in toggle button', () => {
      render(<FloatingChat {...defaultProps} />)

      const toggleButton = screen.getByRole('button')
      expect(toggleButton.querySelector('svg.lucide-message-circle')).toBeInTheDocument()
    })
  })

  describe('opening and closing chat panel', () => {
    it('opens chat panel when toggle button is clicked', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      const toggleButton = screen.getByRole('button', { name: /打开问答/i })
      await user.click(toggleButton)

      expect(screen.getByText('关于本文的问答')).toBeInTheDocument()
    })

    it('shows article title in empty state', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      expect(screen.getByText(/Test Article Title/)).toBeInTheDocument()
    })

    it('closes chat panel when X button is clicked', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      // Open the chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))
      expect(screen.getByText('关于本文的问答')).toBeInTheDocument()

      // Find and click the close button (X icon button)
      const closeButtons = screen.getAllByRole('button')
      const closeButton = closeButtons.find(btn => btn.querySelector('svg.lucide-x'))
      expect(closeButton).toBeDefined()
      await user.click(closeButton!)

      // Chat panel should be closed, toggle button should be back
      expect(screen.queryByText('关于本文的问答')).not.toBeInTheDocument()
      expect(screen.getByRole('button', { name: /打开问答/i })).toBeInTheDocument()
    })

    it('minimizes chat panel when minimize button is clicked', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      // Open the chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Find and click the minimize button (Minus icon button)
      const allButtons = screen.getAllByRole('button')
      const minimizeButton = allButtons.find(btn => btn.querySelector('svg.lucide-minus'))
      expect(minimizeButton).toBeDefined()
      await user.click(minimizeButton!)

      // Should show minimized state
      expect(screen.getByText('问答窗口')).toBeInTheDocument()
      expect(screen.queryByText('关于本文的问答')).not.toBeInTheDocument()
    })

    it('restores from minimized state when clicked', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      // Open the chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Minimize
      const allButtons = screen.getAllByRole('button')
      const minimizeButton = allButtons.find(btn => btn.querySelector('svg.lucide-minus'))
      await user.click(minimizeButton!)

      // Click the minimized button to restore
      const minimizedButton = screen.getByText('问答窗口').closest('button')
      expect(minimizedButton).toBeDefined()
      await user.click(minimizedButton!)

      // Should show full chat panel again
      expect(screen.getByText('关于本文的问答')).toBeInTheDocument()
    })
  })

  describe('sending messages', () => {
    it('sends message when send button is clicked', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Type a message
      const textarea = screen.getByPlaceholderText(/输入你的问题/i)
      await user.type(textarea, 'Hello, AI!')

      // Click send button
      const sendButton = screen.getAllByRole('button').find(btn =>
        btn.querySelector('svg.lucide-send')
      )
      expect(sendButton).toBeDefined()
      await user.click(sendButton!)

      // Check sendMessage was called
      expect(mockSendMessage).toHaveBeenCalledWith('Hello, AI!')
    })

    it('clears input after sending message', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Type and send message
      const textarea = screen.getByPlaceholderText(/输入你的问题/i) as HTMLTextAreaElement
      await user.type(textarea, 'Test message')
      expect(textarea.value).toBe('Test message')

      const sendButton = screen.getAllByRole('button').find(btn =>
        btn.querySelector('svg.lucide-send')
      )
      await user.click(sendButton!)

      // Input should be cleared
      await waitFor(() => {
        expect(textarea.value).toBe('')
      })
    })

    it('does not send empty messages', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Try to click send without typing
      const sendButton = screen.getAllByRole('button').find(btn =>
        btn.querySelector('svg.lucide-send')
      )

      // Button should be disabled
      expect(sendButton).toBeDisabled()
    })

    it('disables send button while loading', async () => {
      const user = userEvent.setup()

      // Set loading state
      vi.mocked(useChat).mockReturnValue({
        messages: [],
        isLoading: true,
        currentResponse: '',
        sendMessage: mockSendMessage,
        clearMessages: mockClearMessages
      })

      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Type a message
      const textarea = screen.getByPlaceholderText(/输入你的问题/i)
      await user.type(textarea, 'Test')

      const sendButton = screen.getAllByRole('button').find(btn =>
        btn.querySelector('svg.lucide-send')
      )

      // Button should be disabled because isLoading is true
      expect(sendButton).toBeDisabled()
    })
  })

  describe('displaying messages', () => {
    it('displays user messages', async () => {
      const user = userEvent.setup()

      vi.mocked(useChat).mockReturnValue({
        messages: [
          { id: '1', role: 'user', content: 'Hello there!' }
        ],
        isLoading: false,
        currentResponse: '',
        sendMessage: mockSendMessage,
        clearMessages: mockClearMessages
      })

      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // User message should appear
      expect(screen.getByText('Hello there!')).toBeInTheDocument()
    })

    it('displays assistant messages', async () => {
      const user = userEvent.setup()

      vi.mocked(useChat).mockReturnValue({
        messages: [
          { id: '1', role: 'user', content: 'Hello' },
          { id: '2', role: 'assistant', content: 'Hi, how can I help?' }
        ],
        isLoading: false,
        currentResponse: '',
        sendMessage: mockSendMessage,
        clearMessages: mockClearMessages
      })

      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Both messages should appear
      expect(screen.getByText('Hello')).toBeInTheDocument()
      expect(screen.getByText('Hi, how can I help?')).toBeInTheDocument()
    })

    it('displays streaming response', async () => {
      const user = userEvent.setup()

      vi.mocked(useChat).mockReturnValue({
        messages: [
          { id: '1', role: 'user', content: 'Question' }
        ],
        isLoading: true,
        currentResponse: 'Streaming response...',
        sendMessage: mockSendMessage,
        clearMessages: mockClearMessages
      })

      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Streaming response should appear
      expect(screen.getByText('Streaming response...')).toBeInTheDocument()
    })

    it('displays loading indicator when waiting for response', async () => {
      const user = userEvent.setup()

      vi.mocked(useChat).mockReturnValue({
        messages: [
          { id: '1', role: 'user', content: 'Question' }
        ],
        isLoading: true,
        currentResponse: '',
        sendMessage: mockSendMessage,
        clearMessages: mockClearMessages
      })

      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Loading indicator should appear
      expect(screen.getByText('思考中...')).toBeInTheDocument()
    })
  })

  describe('clear messages', () => {
    it('calls clearMessages when clear button is clicked', async () => {
      const user = userEvent.setup()

      vi.mocked(useChat).mockReturnValue({
        messages: [
          { id: '1', role: 'user', content: 'Test message' }
        ],
        isLoading: false,
        currentResponse: '',
        sendMessage: mockSendMessage,
        clearMessages: mockClearMessages
      })

      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Click clear button
      const clearButton = screen.getByText('清空').closest('button')
      expect(clearButton).toBeDefined()
      await user.click(clearButton!)

      // clearMessages should be called
      expect(mockClearMessages).toHaveBeenCalled()
    })
  })

  describe('keyboard shortcuts', () => {
    it('sends message with Cmd+Enter', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Type a message and press Cmd+Enter
      const textarea = screen.getByPlaceholderText(/输入你的问题/i)
      await user.type(textarea, 'Shortcut test')
      await user.keyboard('{Meta>}{Enter}{/Meta}')

      // sendMessage should be called
      expect(mockSendMessage).toHaveBeenCalledWith('Shortcut test')
    })

    it('sends message with Ctrl+Enter', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      // Open chat
      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      // Type a message and press Ctrl+Enter
      const textarea = screen.getByPlaceholderText(/输入你的问题/i)
      await user.type(textarea, 'Ctrl shortcut test')
      await user.keyboard('{Control>}{Enter}{/Control}')

      // sendMessage should be called
      expect(mockSendMessage).toHaveBeenCalledWith('Ctrl shortcut test')
    })
  })

  describe('toolbar', () => {
    it('renders save conversation button', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      expect(screen.getByText('保存对话')).toBeInTheDocument()
    })

    it('renders clear button', async () => {
      const user = userEvent.setup()
      render(<FloatingChat {...defaultProps} />)

      await user.click(screen.getByRole('button', { name: /打开问答/i }))

      expect(screen.getByText('清空')).toBeInTheDocument()
    })
  })
})
