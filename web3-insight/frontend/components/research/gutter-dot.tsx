'use client'

import { HoverCard, HoverCardContent, HoverCardTrigger } from '@/components/ui/hover-card'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { PinPreviewCard } from './pin-preview-card'
import { useEffect, useState } from 'react'
import type { ResearchPinnedFinding } from '@/lib/api'

interface GutterDotProps {
  finding: ResearchPinnedFinding
  findingIndex: number
  onMove?: (findingIndex: number) => void
  onRemove?: (findingIndex: number) => void
}

export function GutterDot({
  finding,
  findingIndex,
  onMove,
  onRemove,
}: GutterDotProps) {
  const [isTouchDevice, setIsTouchDevice] = useState(false)

  useEffect(() => {
    setIsTouchDevice(window.matchMedia('(pointer: coarse)').matches)
  }, [])

  const trigger = (
    <button
      className="size-1.5 rounded-full bg-primary hover:scale-150 transition-transform cursor-pointer"
      aria-label="查看固定的发现"
      onClick={(e) => e.stopPropagation()}
    />
  )

  const previewCard = (
    <PinPreviewCard
      content={finding.messageContent}
      onMove={onMove ? () => onMove(findingIndex) : undefined}
      onRemove={onRemove ? () => onRemove(findingIndex) : undefined}
    />
  )

  if (isTouchDevice) {
    return (
      <Popover>
        <PopoverTrigger asChild>{trigger}</PopoverTrigger>
        <PopoverContent className="w-80" align="start">
          {previewCard}
        </PopoverContent>
      </Popover>
    )
  }

  return (
    <HoverCard openDelay={200} closeDelay={100}>
      <HoverCardTrigger asChild>{trigger}</HoverCardTrigger>
      <HoverCardContent className="w-80" align="start">
        {previewCard}
      </HoverCardContent>
    </HoverCard>
  )
}
