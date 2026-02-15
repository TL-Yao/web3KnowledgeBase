'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Button } from '@/components/ui/button'
import { GripVertical, Move, X } from 'lucide-react'
import type { ResearchPinnedFinding } from '@/lib/api'

interface PinnedCardProps {
  finding: ResearchPinnedFinding
  findingIndex: number
  initialPosition: { x: number; y: number }
  onClose: () => void
  onMove?: (findingIndex: number) => void
  onRemove?: (findingIndex: number) => void
}

export function PinnedCard({
  finding,
  findingIndex,
  initialPosition,
  onClose,
  onMove,
  onRemove,
}: PinnedCardProps) {
  const [position, setPosition] = useState(initialPosition)
  const dragRef = useRef<{ startX: number; startY: number; originX: number; originY: number } | null>(null)
  const cardRef = useRef<HTMLDivElement>(null)

  // Clamp position to viewport on mount
  useEffect(() => {
    if (!cardRef.current) return
    const rect = cardRef.current.getBoundingClientRect()
    const maxX = window.innerWidth - rect.width - 8
    const maxY = window.innerHeight - rect.height - 8
    setPosition(prev => ({
      x: Math.max(8, Math.min(prev.x, maxX)),
      y: Math.max(8, Math.min(prev.y, maxY)),
    }))
  }, [])

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    dragRef.current = {
      startX: e.clientX,
      startY: e.clientY,
      originX: position.x,
      originY: position.y,
    }

    const handleMouseMove = (e: MouseEvent) => {
      if (!dragRef.current) return
      const dx = e.clientX - dragRef.current.startX
      const dy = e.clientY - dragRef.current.startY
      setPosition({
        x: dragRef.current.originX + dx,
        y: dragRef.current.originY + dy,
      })
    }

    const handleMouseUp = () => {
      dragRef.current = null
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
  }, [position])

  const card = (
    <div
      ref={cardRef}
      className="fixed z-50 w-96 rounded-lg border border-border bg-popover shadow-xl"
      style={{ left: position.x, top: position.y }}
    >
      {/* Header */}
      <div
        className="flex items-center justify-between px-3 py-2 border-b border-border cursor-grab active:cursor-grabbing select-none"
        onMouseDown={handleMouseDown}
      >
        <div className="flex items-center gap-1.5 text-sm font-medium">
          <GripVertical className="size-4 text-muted-foreground" />
          <span>发现 #{findingIndex + 1}</span>
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="size-6"
          onClick={onClose}
        >
          <X className="size-3.5" />
        </Button>
      </div>

      {/* Content */}
      <div className="px-3 py-2 max-h-[400px] overflow-y-auto">
        <div className="prose prose-sm dark:prose-invert max-w-none prose-p:text-sm prose-p:my-1 prose-li:text-sm prose-li:my-0.5">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {finding.messageContent}
          </ReactMarkdown>
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2 px-3 py-2 border-t border-border">
        {onMove && (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 text-xs"
            onClick={() => {
              onMove(findingIndex)
              onClose()
            }}
          >
            <Move className="size-3 mr-1" />
            移动
          </Button>
        )}
        {onRemove && (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 text-xs text-destructive hover:text-destructive"
            onClick={() => {
              onRemove(findingIndex)
              onClose()
            }}
          >
            <X className="size-3 mr-1" />
            取消固定
          </Button>
        )}
      </div>
    </div>
  )

  return createPortal(card, document.body)
}
