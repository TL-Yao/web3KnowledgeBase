'use client'

import { Button } from '@/components/ui/button'
import { Crosshair } from 'lucide-react'

interface PlacementHintProps {
  onSkip: () => void
}

export function PlacementHint({ onSkip }: PlacementHintProps) {
  return (
    <div className="sticky top-0 z-10 flex items-center justify-between gap-2 bg-primary/10 border-b border-primary/20 px-6 py-2 text-sm text-primary">
      <div className="flex items-center gap-2">
        <Crosshair className="size-4 shrink-0 animate-pulse" />
        <span>点击段落来放置此发现 · ESC 跳过</span>
      </div>
      <Button
        variant="ghost"
        size="sm"
        className="h-7 text-xs"
        onClick={onSkip}
      >
        跳过
      </Button>
    </div>
  )
}
