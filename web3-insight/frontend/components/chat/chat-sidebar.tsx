'use client'

import { useState, useRef, useEffect } from 'react'
import { MessageCircle, X, Send, Trash2, Sparkles, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ChatMessage } from './chat-message'
import { useChat, Message } from '@/hooks/use-chat'

interface ChatSidebarProps {
  articleId: string
  articleTitle: string
  isOpen: boolean
  onToggle: () => void
  onGenerateUpdate?: (messages: Message[]) => void
  isGeneratingUpdate?: boolean
  width: number
  isDragging?: boolean
  clearTrigger?: number
}

export function ChatSidebar({
  articleId,
  articleTitle,
  isOpen,
  onToggle,
  onGenerateUpdate,
  isGeneratingUpdate,
  width,
  isDragging,
  clearTrigger,
}: ChatSidebarProps) {
  const [input, setInput] = useState('')
  const { messages, isLoading, currentResponse, sendMessage, clearMessages } = useChat(articleId)
  const bottomRef = useRef<HTMLDivElement>(null)

  // Clear messages when parent signals (e.g. after applying update)
  const clearTriggerRef = useRef(clearTrigger)
  useEffect(() => {
    if (clearTrigger !== undefined && clearTrigger !== clearTriggerRef.current) {
      clearTriggerRef.current = clearTrigger
      clearMessages()
    }
  }, [clearTrigger, clearMessages])

  // Auto scroll to bottom
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
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

  // Keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === '/') {
        e.preventDefault()
        onToggle()
      }
      if (e.key === 'Escape' && isOpen) {
        const tag = document.activeElement?.tagName
        if (tag === 'TEXTAREA' || tag === 'INPUT') {
          (document.activeElement as HTMLElement).blur()
        } else {
          onToggle()
        }
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [isOpen, onToggle])

  return (
    <div
      className={cn(
        "h-full flex flex-col border-l border-border bg-background overflow-hidden shrink-0",
        !isDragging && "transition-[width] duration-200 ease-out"
      )}
      style={{ width: isOpen ? width : 0 }}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div className="flex items-center gap-2">
          <MessageCircle className="w-4 h-4 text-primary" />
          <span className="text-sm font-medium whitespace-nowrap">关于本文的问答</span>
        </div>
        <Button variant="ghost" size="icon" className="w-7 h-7" onClick={onToggle}>
          <X className="w-4 h-4" />
        </Button>
      </div>

      {/* Messages */}
      <ScrollArea className="flex-1 min-h-0 p-4">
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
        <div ref={bottomRef} />
      </ScrollArea>

      {/* Toolbar */}
      <div className="px-4 py-2 border-t border-border flex items-center gap-2 shrink-0">
        {messages.length >= 2 && onGenerateUpdate && (
          <Button
            variant="default"
            size="sm"
            className="text-xs"
            onClick={() => onGenerateUpdate(messages)}
            disabled={isLoading || isGeneratingUpdate}
          >
            {isGeneratingUpdate ? (
              <Loader2 className="w-3 h-3 mr-1 animate-spin" />
            ) : (
              <Sparkles className="w-3 h-3 mr-1" />
            )}
            {isGeneratingUpdate ? '更新中...' : '更新文章（基于对话）'}
          </Button>
        )}
        <div className="flex-1" />
        <Button variant="ghost" size="sm" className="text-xs" onClick={clearMessages}>
          <Trash2 className="w-3 h-3 mr-1" />
          清空
        </Button>
      </div>

      {/* Input */}
      <div className="p-3 pt-0 shrink-0">
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
        <p className="text-xs text-muted-foreground mt-1">⌘↵ 发送 · ⌘/ 切换面板</p>
      </div>
    </div>
  )
}
