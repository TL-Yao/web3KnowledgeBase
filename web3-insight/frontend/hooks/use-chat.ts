'use client'

import { useState, useCallback, useRef, useEffect } from 'react'
import { createChatWebSocket } from '@/lib/websocket'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  model?: string
}

export function useChat(articleId: string) {
  const [messages, setMessages] = useState<Message[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [currentResponse, setCurrentResponse] = useState('')
  const wsRef = useRef<WebSocket | null>(null)
  const sessionId = useRef(crypto.randomUUID())
  const currentResponseRef = useRef('')

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
        // Handle error
      }
    })

    return () => {
      wsRef.current?.close()
    }
  }, [articleId])

  const sendMessage = useCallback((content: string, selectedText?: string) => {
    if (!wsRef.current || isLoading) return

    setMessages(prev => [...prev, {
      id: crypto.randomUUID(),
      role: 'user',
      content
    }])

    setIsLoading(true)
    currentResponseRef.current = ''

    wsRef.current.send(JSON.stringify({
      articleId,
      message: content,
      selectedText,
      sessionId: sessionId.current
    }))
  }, [articleId, isLoading])

  return {
    messages,
    isLoading,
    currentResponse,
    sendMessage,
    clearMessages: () => setMessages([])
  }
}
