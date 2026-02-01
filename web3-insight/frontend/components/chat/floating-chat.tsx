'use client'

import { useState, useRef, useEffect } from 'react'
import { MessageCircle, X, Minus, Send, Save, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ChatMessage } from './chat-message'
import { useChat } from '@/hooks/use-chat'

interface FloatingChatProps {
  articleId: string
  articleTitle: string
}

export function FloatingChat({ articleId, articleTitle }: FloatingChatProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [isMinimized, setIsMinimized] = useState(false)
  const [input, setInput] = useState('')
  const [position, setPosition] = useState({ x: 20, y: 20 }) // from bottom-right
  const { messages, isLoading, currentResponse, sendMessage, clearMessages } = useChat(articleId)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Auto scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages, currentResponse])

  const handleSubmit = () => {
    if (!input.trim() || isLoading) return
    sendMessage(input)
    setInput('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleSubmit()
    }
  }

  // Keyboard shortcut to toggle
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === '/') {
        e.preventDefault()
        setIsOpen(prev => !prev)
        setIsMinimized(false)
      }
      if (e.key === 'Escape' && isOpen) {
        setIsMinimized(true)
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [isOpen])

  if (!isOpen) {
    return (
      <button
        onClick={() => setIsOpen(true)}
        className="fixed bottom-6 right-6 w-12 h-12 rounded-full bg-primary text-primary-foreground shadow-lg flex items-center justify-center hover:bg-primary/90 transition-colors"
        title="打开问答 (⌘/)"
      >
        <MessageCircle className="w-5 h-5" />
      </button>
    )
  }

  if (isMinimized) {
    return (
      <button
        onClick={() => setIsMinimized(false)}
        className="fixed bottom-6 right-6 px-4 py-2 rounded-full bg-primary text-primary-foreground shadow-lg flex items-center gap-2 hover:bg-primary/90 transition-colors"
      >
        <MessageCircle className="w-4 h-4" />
        <span className="text-sm">问答窗口</span>
      </button>
    )
  }

  return (
    <div
      className="fixed bg-background border border-border rounded-lg shadow-xl flex flex-col"
      style={{
        bottom: position.y,
        right: position.x,
        width: 380,
        height: 500,
        maxHeight: '70vh'
      }}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border">
        <div className="flex items-center gap-2">
          <MessageCircle className="w-4 h-4 text-primary" />
          <span className="text-sm font-medium">关于本文的问答</span>
        </div>
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="icon" className="w-7 h-7" onClick={() => setIsMinimized(true)}>
            <Minus className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="icon" className="w-7 h-7" onClick={() => setIsOpen(false)}>
            <X className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Messages */}
      <ScrollArea className="flex-1 p-4" ref={scrollRef}>
        {messages.length === 0 && !currentResponse && (
          <div className="text-center text-muted-foreground text-sm py-8">
            <p>对「{articleTitle}」有疑问？</p>
            <p className="mt-1">在下方输入你的问题</p>
          </div>
        )}

        {messages.map((msg) => (
          <ChatMessage key={msg.id} message={msg} />
        ))}

        {currentResponse && (
          <ChatMessage
            message={{
              id: 'current',
              role: 'assistant',
              content: currentResponse
            }}
            isStreaming
          />
        )}

        {isLoading && !currentResponse && (
          <div className="flex items-center gap-2 text-muted-foreground text-sm">
            <div className="w-2 h-2 rounded-full bg-primary animate-pulse" />
            <span>思考中...</span>
          </div>
        )}
      </ScrollArea>

      {/* Toolbar */}
      <div className="px-4 py-2 border-t border-border flex items-center gap-2">
        <Button variant="ghost" size="sm" className="text-xs">
          <Save className="w-3 h-3 mr-1" />
          保存对话
        </Button>
        <Button variant="ghost" size="sm" className="text-xs" onClick={clearMessages}>
          <Trash2 className="w-3 h-3 mr-1" />
          清空
        </Button>
      </div>

      {/* Input */}
      <div className="p-4 pt-0">
        <div className="relative">
          <Textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="输入你的问题... (⌘↵ 发送)"
            className="pr-10 resize-none"
            rows={2}
          />
          <Button
            size="icon"
            className="absolute bottom-2 right-2 w-7 h-7"
            onClick={handleSubmit}
            disabled={!input.trim() || isLoading}
          >
            <Send className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
