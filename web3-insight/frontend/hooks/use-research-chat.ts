'use client'

import { useState, useCallback, useRef, useEffect } from 'react'
import { createResearchChatWebSocket } from '@/lib/websocket'
import type { Message, ChatModel } from './use-chat'

const STORAGE_PREFIX = 'research-chat-'
const MODEL_STORAGE_KEY = 'research-chat-model'
const DEFAULT_MODEL: ChatModel = 'sonnet'

function loadMessages(sessionId: string): Message[] {
  if (typeof window === 'undefined') return []
  try {
    const saved = localStorage.getItem(`${STORAGE_PREFIX}${sessionId}`)
    return saved ? JSON.parse(saved) : []
  } catch {
    return []
  }
}

function saveMessages(sessionId: string, messages: Message[]) {
  try {
    if (messages.length > 0) {
      localStorage.setItem(`${STORAGE_PREFIX}${sessionId}`, JSON.stringify(messages))
    } else {
      localStorage.removeItem(`${STORAGE_PREFIX}${sessionId}`)
    }
  } catch {
    // localStorage quota exceeded — continue without persistence
  }
}

function loadModel(): ChatModel {
  if (typeof window === 'undefined') return DEFAULT_MODEL
  try {
    const saved = localStorage.getItem(MODEL_STORAGE_KEY)
    if (saved === 'haiku' || saved === 'sonnet' || saved === 'opus') return saved
  } catch {}
  return DEFAULT_MODEL
}

export function useResearchChat(sessionId: string) {
  const [messages, setMessages] = useState<Message[]>(() => loadMessages(sessionId))
  const [isLoading, setIsLoading] = useState(false)
  const [currentResponse, setCurrentResponse] = useState('')
  const [model, setModelState] = useState<ChatModel>(loadModel)
  const wsRef = useRef<WebSocket | null>(null)
  const wsSessionId = useRef(crypto.randomUUID())
  const currentResponseRef = useRef('')
  const messagesRef = useRef(messages)
  const modelRef = useRef(model)

  const setModel = useCallback((m: ChatModel) => {
    setModelState(m)
    modelRef.current = m
    try { localStorage.setItem(MODEL_STORAGE_KEY, m) } catch {}
  }, [])

  useEffect(() => {
    messagesRef.current = messages
  }, [messages])

  useEffect(() => {
    saveMessages(sessionId, messages)
  }, [sessionId, messages])

  useEffect(() => {
    wsRef.current = createResearchChatWebSocket((data) => {
      if (data.type === 'chunk') {
        currentResponseRef.current += data.content
        setCurrentResponse(currentResponseRef.current)
      } else if (data.type === 'done') {
        setMessages(prev => [...prev, {
          id: crypto.randomUUID(),
          role: 'assistant',
          content: currentResponseRef.current,
          model: data.model,
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
  }, [sessionId])

  const sendMessage = useCallback((content: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN || isLoading) return

    const userMessage: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content,
    }

    setMessages(prev => [...prev, userMessage])
    setIsLoading(true)
    currentResponseRef.current = ''

    const history = messagesRef.current.map(m => ({ role: m.role, content: m.content }))

    wsRef.current.send(JSON.stringify({
      sessionId,
      message: content,
      sessionWsId: wsSessionId.current,
      history,
      model: modelRef.current,
    }))
  }, [sessionId, isLoading])

  const clearMessages = useCallback(() => {
    setMessages([])
    localStorage.removeItem(`${STORAGE_PREFIX}${sessionId}`)
  }, [sessionId])

  return {
    messages,
    isLoading,
    currentResponse,
    sendMessage,
    clearMessages,
    model,
    setModel,
  }
}
