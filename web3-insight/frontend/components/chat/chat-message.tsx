import { cn } from '@/lib/utils'
import { User, Bot } from 'lucide-react'

interface ChatMessageProps {
  message: {
    id: string
    role: 'user' | 'assistant'
    content: string
    model?: string
  }
  isStreaming?: boolean
}

export function ChatMessage({ message, isStreaming }: ChatMessageProps) {
  const isUser = message.role === 'user'

  return (
    <div className={cn(
      "flex gap-3 mb-4",
      isUser && "flex-row-reverse"
    )}>
      <div className={cn(
        "w-7 h-7 rounded-full flex items-center justify-center flex-shrink-0",
        isUser ? "bg-primary" : "bg-muted"
      )}>
        {isUser ? (
          <User className="w-4 h-4 text-primary-foreground" />
        ) : (
          <Bot className="w-4 h-4" />
        )}
      </div>

      <div className={cn(
        "flex-1 text-sm",
        isUser && "text-right"
      )}>
        <div className={cn(
          "inline-block px-3 py-2 rounded-lg max-w-[85%]",
          isUser ? "bg-primary text-primary-foreground" : "bg-muted"
        )}>
          {message.content}
          {isStreaming && (
            <span className="inline-block w-1.5 h-4 bg-current ml-0.5 animate-pulse" />
          )}
        </div>
        {message.model && (
          <div className="text-xs text-muted-foreground mt-1">
            {message.model}
          </div>
        )}
      </div>
    </div>
  )
}
