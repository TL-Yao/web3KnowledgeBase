'use client'

import { useState, useCallback, useRef, useEffect } from 'react'
import { createChatWebSocket } from '@/lib/websocket'

export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  model?: string
}

function loadMessages(articleId: string): Message[] {
  if (typeof window === 'undefined') return []
  try {
    const saved = localStorage.getItem(`chat-${articleId}`)
    return saved ? JSON.parse(saved) : []
  } catch {
    return []
  }
}

function saveMessages(articleId: string, messages: Message[]) {
  try {
    if (messages.length > 0) {
      localStorage.setItem(`chat-${articleId}`, JSON.stringify(messages))
    } else {
      localStorage.removeItem(`chat-${articleId}`)
    }
  } catch {
    // localStorage quota exceeded — continue without persistence
  }
}

export function useChat(articleId: string) {
  const [messages, setMessages] = useState<Message[]>(() => loadMessages(articleId))
  const [isLoading, setIsLoading] = useState(false)
  const [currentResponse, setCurrentResponse] = useState('')
  const wsRef = useRef<WebSocket | null>(null)
  const sessionId = useRef(crypto.randomUUID())
  const currentResponseRef = useRef('')
  const messagesRef = useRef(messages)

  // Keep messagesRef in sync for use in sendMessage
  useEffect(() => {
    messagesRef.current = messages
  }, [messages])

  // Persist messages to localStorage
  useEffect(() => {
    saveMessages(articleId, messages)
  }, [articleId, messages])

  useEffect(() => {
    wsRef.current = createChatWebSocket((data) => {
      if (data.type === 'chunk') {
        currentResponseRef.current += data.content
        setCurrentResponse(currentResponseRef.current)
      } else if (data.type === 'done') {
        setMessages(prev => [...prev, {
          id: crypto.randomUUID(),
          role: 'assistant',
          content: currentResponseRef.current,
          model: data.model
        }])
        currentResponseRef.current = ''
        setCurrentResponse('')
        setIsLoading(false)
      } else if (data.type === 'error') {
        setIsLoading(false)
      }
    })

    return () => {
      wsRef.current?.close()
    }
  }, [articleId])

  const sendMessage = useCallback((content: string, selectedText?: string) => {
    if (!wsRef.current || isLoading) return

    const userMessage: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content
    }

    setMessages(prev => [...prev, userMessage])
    setIsLoading(true)
    currentResponseRef.current = ''

    // Send with full conversation history for multi-turn context
    const history = messagesRef.current.map(m => ({ role: m.role, content: m.content }))

    wsRef.current.send(JSON.stringify({
      articleId,
      message: content,
      selectedText,
      sessionId: sessionId.current,
      history,
    }))
  }, [articleId, isLoading])

  const clearMessages = useCallback(() => {
    setMessages([])
    localStorage.removeItem(`chat-${articleId}`)
  }, [articleId])

  return {
    messages,
    isLoading,
    currentResponse,
    sendMessage,
    clearMessages,
  }
}
