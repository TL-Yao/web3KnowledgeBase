'use client'

import { useState, useRef, useEffect } from 'react'
import { MessageCircle, X, Send, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ChatMessage } from '@/components/chat/chat-message'
import { PinButton } from './pin-button'
import type { Message, ChatModel } from '@/hooks/use-chat'

interface ResearchChatProps {
  isOpen: boolean
  onToggle: () => void
  width: number
  isDragging?: boolean
  messages?: Message[]
  currentResponse?: string
  isLoading?: boolean
  model?: ChatModel
  onModelChange?: (model: ChatModel) => void
  onSendMessage?: (content: string) => void
  onClearMessages?: () => void
  onPinFinding?: (content: string) => void
  pinnedContents?: Set<string>
}

export function ResearchChat({
  isOpen,
  onToggle,
  width,
  isDragging,
  messages = [],
  currentResponse,
  isLoading = false,
  model = 'sonnet',
  onModelChange,
  onSendMessage,
  onClearMessages,
  onPinFinding,
  pinnedContents,
}: ResearchChatProps) {
  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, currentResponse])

  const handleSubmit = () => {
    if (!input.trim() || isLoading) return
    onSendMessage?.(input)
    setInput('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleSubmit()
    }
  }

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

  const isPinned = (content: string) => pinnedContents?.has(content) ?? false

  return (
    <div
      className={cn(
        'h-full flex flex-col border-l border-border bg-background overflow-hidden shrink-0',
        !isDragging && 'transition-[width] duration-200 ease-out'
      )}
      style={{ width: isOpen ? width : 0 }}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div className="flex items-center gap-2">
          <MessageCircle className="w-4 h-4 text-primary" />
          <span className="text-sm font-medium whitespace-nowrap">研究对话</span>
        </div>
        <div className="flex items-center gap-1.5">
          <Select value={model} onValueChange={(v) => onModelChange?.(v as ChatModel)}>
            <SelectTrigger size="sm" className="h-7 text-xs gap-1 px-2 min-w-0 border-none shadow-none">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="haiku">Haiku · 快速</SelectItem>
              <SelectItem value="sonnet">Sonnet · 推荐</SelectItem>
              <SelectItem value="opus">Opus · 最强</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="ghost" size="icon" className="w-7 h-7" onClick={onToggle}>
            <X className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Messages */}
      <ScrollArea className="flex-1 min-h-0 p-4">
        {messages.length === 0 && !currentResponse && (
          <div className="text-center text-muted-foreground text-sm py-8">
            <p>针对本次研究进行提问</p>
            <p className="mt-1">AI 将基于研究上下文回答</p>
          </div>
        )}

        {messages.map((msg) => (
          <div key={msg.id} className="group relative">
            <ChatMessage message={msg} />
            {msg.role === 'assistant' && onPinFinding && (
              <div className="absolute top-1 right-1">
                <PinButton
                  isPinned={isPinned(msg.content)}
                  onClick={() => onPinFinding(msg.content)}
                />
              </div>
            )}
          </div>
        ))}

        {currentResponse && (
          <div className="group relative">
            <ChatMessage
              message={{
                id: 'current',
                role: 'assistant',
                content: currentResponse,
              }}
              isStreaming
            />
          </div>
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
      <div className="px-3 py-1.5 border-t border-border flex items-center gap-2 shrink-0">
        <div className="flex-1" />
        <Button variant="ghost" size="sm" className="text-xs" onClick={onClearMessages}>
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
            className="absolute bottom-2 right-2 w-6 h-6"
            onClick={handleSubmit}
            disabled={!input.trim() || isLoading}
          >
            <Send className="w-4 h-4" />
          </Button>
        </div>
        <p className="text-[11px] text-muted-foreground mt-1">⌘↵ 发送 · ⌘/ 切换面板</p>
      </div>
    </div>
  )
}
