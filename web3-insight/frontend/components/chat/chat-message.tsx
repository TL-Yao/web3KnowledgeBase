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
      "flex gap-2 mb-2.5",
      isUser && "flex-row-reverse"
    )}>
      <div className={cn(
        "w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0",
        isUser ? "bg-primary" : "bg-muted"
      )}>
        {isUser ? (
          <User className="w-3.5 h-3.5 text-primary-foreground" />
        ) : (
          <Bot className="w-3.5 h-3.5" />
        )}
      </div>

      <div className={cn(
        "flex-1 text-[13px] leading-normal",
        isUser && "text-right"
      )}>
        <div className={cn(
          "inline-block px-2.5 py-1.5 rounded-md max-w-[90%]",
          isUser ? "bg-primary text-primary-foreground" : "bg-muted"
        )}>
          {isUser ? (
            <span className="text-[13px] leading-normal whitespace-pre-wrap break-words">
              {message.content}
            </span>
          ) : (
            <div className="text-[13px] leading-normal space-y-1
              [&_p]:my-0.5
              [&_ul]:my-1 [&_ul]:pl-4 [&_ol]:my-1 [&_ol]:pl-4
              [&_li]:my-0
              [&_h3]:font-semibold [&_h3]:text-[13px] [&_h3]:mt-2 [&_h3]:mb-0.5
              [&_h2]:font-semibold [&_h2]:text-sm [&_h2]:mt-2 [&_h2]:mb-0.5
              [&_h1]:font-bold [&_h1]:text-sm [&_h1]:mt-2 [&_h1]:mb-0.5
              [&_pre]:my-1.5 [&_pre]:p-2 [&_pre]:text-xs [&_pre]:rounded [&_pre]:bg-muted [&_pre]:overflow-x-auto
              [&_code]:text-xs [&_code]:bg-muted [&_code]:px-1 [&_code]:py-0.5 [&_code]:rounded-sm
              [&_pre_code]:bg-transparent [&_pre_code]:p-0
              [&_blockquote]:my-1 [&_blockquote]:pl-2.5 [&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:text-muted-foreground
              [&_a]:text-primary [&_a]:underline [&_a]:underline-offset-2
              [&_strong]:font-semibold
              [&_hr]:my-2 [&_hr]:border-border
              [&_table]:text-xs [&_table]:my-1">
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
          <div className="text-[11px] text-muted-foreground mt-0.5">
            {message.model}
          </div>
        )}
      </div>
    </div>
  )
}
