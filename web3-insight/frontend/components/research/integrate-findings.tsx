'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
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
import { Sparkles, X, Loader2, Crosshair, ArrowRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { ResearchPinnedFinding } from '@/lib/api'

export type FindingWithIndex = ResearchPinnedFinding & { originalIndex: number }

interface IntegrateFindingsProps {
  findings: FindingWithIndex[]
  onIntegrate: () => void
  onRemove?: (index: number) => void
  onPlacePin?: (index: number) => void
  onClearAll?: () => void
  isIntegrating?: boolean
  disableIntegrate?: boolean
  placingPinIndex?: number | null
}

export function IntegrateFindings({
  findings,
  onIntegrate,
  onRemove,
  onPlacePin,
  onClearAll,
  isIntegrating,
  disableIntegrate,
  placingPinIndex,
}: IntegrateFindingsProps) {
  const [showConfirm, setShowConfirm] = useState(false)
  const [isExpanded, setIsExpanded] = useState(false)

  const pinnedCount = findings.length
  if (pinnedCount === 0) return null

  const handleIntegrate = () => {
    setShowConfirm(false)
    onIntegrate()
  }

  return (
    <>
      <div className="sticky bottom-0 border-t border-border bg-background/95 backdrop-blur px-6 py-3">
        {/* Placement mode banner */}
        {placingPinIndex != null && (
          <div className="mb-2 flex items-center gap-2 text-sm text-primary bg-primary/10 rounded-md px-3 py-2">
            <Crosshair className="size-4 shrink-0 animate-pulse" />
            <span>点击文章中的章节标题来放置此发现</span>
          </div>
        )}

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

        {isExpanded && findings.length > 0 && (
          <div className="mt-3 space-y-2 max-h-40 overflow-y-auto">
            {findings.map((finding, idx) => {
              const preview = finding.messageContent.length > 120
                ? finding.messageContent.slice(0, 120) + '...'
                : finding.messageContent
              const isPlacing = placingPinIndex === finding.originalIndex

              return (
                <div
                  key={finding.originalIndex}
                  className={cn(
                    'flex items-start gap-2 text-sm bg-muted/50 rounded-md px-3 py-2',
                    'group',
                    isPlacing && 'ring-2 ring-primary bg-primary/5'
                  )}
                >
                  <span className="text-muted-foreground font-mono text-xs mt-0.5 shrink-0">
                    {idx + 1}.
                  </span>
                  <span className="flex-1 line-clamp-2 text-muted-foreground">
                    {preview}
                  </span>
                  {finding.targetPreview && (
                    <Badge variant="secondary" className="shrink-0 text-xs gap-1">
                      <ArrowRight className="size-2.5" />
                      {finding.targetPreview}
                    </Badge>
                  )}
                  <div className="flex items-center gap-0.5 shrink-0">
                    {onPlacePin && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className={cn(
                          'size-5',
                          !isPlacing && 'opacity-0 group-hover:opacity-100 transition-opacity'
                        )}
                        onClick={() => onPlacePin(finding.originalIndex)}
                        disabled={isIntegrating}
                        title="放置到文章中"
                      >
                        <Crosshair className="size-3" />
                      </Button>
                    )}
                    {onRemove && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-5 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={() => onRemove(finding.originalIndex)}
                        disabled={isIntegrating}
                      >
                        <X className="size-3" />
                      </Button>
                    )}
                  </div>
                </div>
              )
            })}
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
