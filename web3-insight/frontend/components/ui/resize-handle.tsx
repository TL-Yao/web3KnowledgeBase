'use client'

import { cn } from '@/lib/utils'
import { useResize } from '@/hooks/use-resize'

interface ResizeHandleProps {
  onResize: (delta: number) => void
  onResizeEnd?: () => void
  onReset?: () => void
  onDragChange?: (isDragging: boolean) => void
}

export function ResizeHandle({ onResize, onResizeEnd, onReset, onDragChange }: ResizeHandleProps) {
  const { isDragging, handleMouseDown } = useResize({
    onResize,
    onResizeEnd: () => {
      onDragChange?.(false)
      onResizeEnd?.()
    },
  })

  const handleStart = (e: React.MouseEvent | React.TouchEvent) => {
    onDragChange?.(true)
    handleMouseDown(e)
  }

  return (
    <div
      className={cn(
        'w-2 shrink-0 cursor-col-resize transition-colors',
        isDragging ? 'bg-primary/40' : 'bg-border hover:bg-primary/30'
      )}
      onMouseDown={handleStart}
      onTouchStart={handleStart}
      onDoubleClick={onReset}
    />
  )
}
