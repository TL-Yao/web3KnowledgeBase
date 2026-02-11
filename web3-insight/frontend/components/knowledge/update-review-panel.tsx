'use client'

import { diffLines } from 'diff'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { X, RefreshCw, Check, Loader2, FileText } from 'lucide-react'
import { cn } from '@/lib/utils'

interface UpdateReviewPanelProps {
  articleTitle: string
  originalContent: string
  updatedContent: string
  changeSummary: string
  model: string
  onApply: () => void
  onRegenerate: () => void
  onCancel: () => void
  isApplying: boolean
  isRegenerating: boolean
  variant?: 'overlay' | 'inline'
}

export function UpdateReviewPanel({
  articleTitle,
  originalContent,
  updatedContent,
  changeSummary,
  model,
  onApply,
  onRegenerate,
  onCancel,
  isApplying,
  isRegenerating,
  variant = 'inline',
}: UpdateReviewPanelProps) {
  const changes = diffLines(originalContent, updatedContent)
  const hasChanges = changes.some(part => part.added || part.removed)

  const isOverlay = variant === 'overlay'

  const panel = (
    <div className={cn(
      "flex flex-col overflow-hidden",
      isOverlay
        ? "fixed inset-4 md:inset-8 bg-background border border-border rounded-lg shadow-2xl"
        : "h-full"
    )}>
      {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <div className="flex items-center gap-3">
            <FileText className="w-5 h-5 text-primary" />
            <div>
              <h2 className="text-lg font-semibold">文章更新预览</h2>
              <p className="text-sm text-muted-foreground truncate max-w-md">{articleTitle}</p>
            </div>
          </div>
          <Button variant="ghost" size="icon" onClick={onCancel}>
            <X className="w-5 h-5" />
          </Button>
        </div>

        {/* Change summary */}
        <div className="px-6 py-3 border-b border-border bg-muted/30">
          <div className="flex items-center gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">更新说明：</span>
              <span className="font-medium">{changeSummary}</span>
            </div>
            <div className="text-muted-foreground">
              模型：{model}
            </div>
          </div>
        </div>

        {/* Diff content */}
        <ScrollArea className="flex-1 min-h-0">
          <div className="p-6">
            {!hasChanges ? (
              <div className="text-center py-12 text-muted-foreground">
                <p className="text-lg font-medium">AI 未找到需要更新的内容</p>
                <p className="mt-1 text-sm">可以尝试重新生成或继续对话讨论</p>
              </div>
            ) : (
              <div className="font-mono text-sm border border-border rounded-md overflow-hidden">
                {changes.map((part, i) => {
                  if (!part.added && !part.removed) {
                    // Unchanged lines — show with collapse for large blocks
                    const lines = part.value.split('\n')
                    if (lines.length > 8) {
                      return (
                        <div key={i}>
                          {lines.slice(0, 3).map((line, j) => (
                            <div key={`${i}-start-${j}`} className="px-4 py-0.5 whitespace-pre-wrap text-muted-foreground">
                              {line || '\u00A0'}
                            </div>
                          ))}
                          <div className="px-4 py-1 text-center text-xs text-muted-foreground bg-muted/50 border-y border-border">
                            ··· {lines.length - 6} 行未变更 ···
                          </div>
                          {lines.slice(-3).map((line, j) => (
                            <div key={`${i}-end-${j}`} className="px-4 py-0.5 whitespace-pre-wrap text-muted-foreground">
                              {line || '\u00A0'}
                            </div>
                          ))}
                        </div>
                      )
                    }
                    return (
                      <div key={i}>
                        {lines.map((line, j) => (
                          <div key={`${i}-${j}`} className="px-4 py-0.5 whitespace-pre-wrap text-muted-foreground">
                            {line || '\u00A0'}
                          </div>
                        ))}
                      </div>
                    )
                  }

                  // Changed lines
                  const lines = part.value.split('\n')
                  return (
                    <div key={i}>
                      {lines.map((line, j) => (
                        // Skip trailing empty line from split
                        (j === lines.length - 1 && line === '') ? null : (
                          <div
                            key={`${i}-${j}`}
                            className={cn(
                              "px-4 py-0.5 whitespace-pre-wrap border-l-2",
                              part.added && "bg-green-50 text-green-900 border-l-green-500 dark:bg-green-950/30 dark:text-green-200",
                              part.removed && "bg-red-50 text-red-900 border-l-red-500 line-through dark:bg-red-950/30 dark:text-red-200"
                            )}
                          >
                            <span className="select-none mr-2 text-xs opacity-60">{part.added ? '+' : '-'}</span>
                            {line || '\u00A0'}
                          </div>
                        )
                      ))}
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </ScrollArea>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-border">
          <Button variant="outline" onClick={onCancel} disabled={isApplying}>
            取消
          </Button>
          <Button
            variant="outline"
            onClick={onRegenerate}
            disabled={isApplying || isRegenerating}
          >
            {isRegenerating ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <RefreshCw className="w-4 h-4 mr-2" />
            )}
            重新生成
          </Button>
          <Button
            onClick={onApply}
            disabled={isApplying || isRegenerating || !hasChanges}
          >
            {isApplying ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Check className="w-4 h-4 mr-2" />
            )}
            {isApplying ? '应用中...' : '应用更新'}
          </Button>
        </div>
    </div>
  )

  if (isOverlay) {
    return (
      <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
        {panel}
      </div>
    )
  }

  return panel
}
