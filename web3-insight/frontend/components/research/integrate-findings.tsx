'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Sparkles, X, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface IntegrateFindingsProps {
  pinnedCount: number
  pinnedPreviews?: string[]
  onIntegrate: () => void
  onRemove?: (index: number) => void
  onClearAll?: () => void
  isIntegrating?: boolean
  disableIntegrate?: boolean
}

export function IntegrateFindings({
  pinnedCount,
  pinnedPreviews = [],
  onIntegrate,
  onRemove,
  onClearAll,
  isIntegrating,
  disableIntegrate,
}: IntegrateFindingsProps) {
  const [showConfirm, setShowConfirm] = useState(false)
  const [isExpanded, setIsExpanded] = useState(false)

  if (pinnedCount === 0) return null

  const handleIntegrate = () => {
    setShowConfirm(false)
    onIntegrate()
  }

  return (
    <>
      <div className="sticky bottom-0 border-t border-border bg-background/95 backdrop-blur px-6 py-3">
        <div className="flex items-center justify-between">
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="text-sm font-medium hover:underline underline-offset-2"
          >
            {pinnedCount} 条发现已固定
          </button>
          <div className="flex items-center gap-2">
            {onClearAll && (
              <Button
                variant="ghost"
                size="sm"
                className="text-xs"
                onClick={onClearAll}
                disabled={isIntegrating}
              >
                清除全部
              </Button>
            )}
            <Button
              size="sm"
              onClick={() => setShowConfirm(true)}
              disabled={isIntegrating || disableIntegrate}
              title={disableIntegrate ? '研究完成后可整合到报告' : undefined}
            >
              {isIntegrating ? (
                <Loader2 className="size-3.5 mr-1.5 animate-spin" />
              ) : (
                <Sparkles className="size-3.5 mr-1.5" />
              )}
              {isIntegrating ? '整合中...' : '整合到报告'}
            </Button>
          </div>
        </div>

        {isExpanded && pinnedPreviews.length > 0 && (
          <div className="mt-3 space-y-2 max-h-40 overflow-y-auto">
            {pinnedPreviews.map((preview, idx) => (
              <div
                key={idx}
                className={cn(
                  'flex items-start gap-2 text-sm bg-muted/50 rounded-md px-3 py-2',
                  'group'
                )}
              >
                <span className="text-muted-foreground font-mono text-xs mt-0.5 shrink-0">
                  {idx + 1}.
                </span>
                <span className="flex-1 line-clamp-2 text-muted-foreground">
                  {preview}
                </span>
                {onRemove && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-5 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
                    onClick={() => onRemove(idx)}
                    disabled={isIntegrating}
                  >
                    <X className="size-3" />
                  </Button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      <AlertDialog open={showConfirm} onOpenChange={setShowConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>整合发现到报告</AlertDialogTitle>
            <AlertDialogDescription>
              将 {pinnedCount} 条固定的发现整合到研究报告中。AI
              将重新分析并更新报告内容，这可能需要几分钟时间。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleIntegrate}>
              <Sparkles className="size-3.5 mr-1.5" />
              开始整合
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
