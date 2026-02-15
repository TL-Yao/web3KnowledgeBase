'use client'

import { Pin } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

interface PinButtonProps {
  isPinned?: boolean
  onClick: (shiftKey: boolean) => void
  className?: string
}

export function PinButton({ isPinned, onClick, className }: PinButtonProps) {
  return (
    <Button
      variant="ghost"
      size="icon"
      className={cn(
        'size-6 opacity-0 group-hover:opacity-100 transition-opacity',
        isPinned && 'opacity-100 text-primary',
        className
      )}
      onClick={(e) => onClick(e.shiftKey)}
      title={isPinned ? '已固定到报告 (Shift+点击: 不放置)' : '固定到报告 (Shift+点击: 不放置)'}
    >
      <Pin
        className={cn(
          'size-3.5',
          isPinned && 'fill-current'
        )}
      />
    </Button>
  )
}
