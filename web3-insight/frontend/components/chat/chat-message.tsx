import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
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
        "flex-1 text-xs",
        isUser && "text-right"
      )}>
        <div className={cn(
          "inline-block px-3 py-2 rounded-lg max-w-[85%]",
          isUser ? "bg-primary text-primary-foreground" : "bg-muted"
        )}>
          {isUser ? (
            message.content
          ) : (
            <div className="prose prose-sm prose-neutral dark:prose-invert max-w-none
                            prose-p:my-1 prose-ul:my-1 prose-ol:my-1
                            prose-li:my-0.5 prose-headings:my-2
                            prose-pre:my-2 prose-code:text-xs">
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                disallowedElements={['img']}
                components={{
                  a: ({ href, children }) => (
                    <a href={href} target="_blank" rel="noopener noreferrer">{children}</a>
                  ),
                }}
              >
                {message.content}
              </ReactMarkdown>
            </div>
          )}
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
