'use client'

import { HoverCard, HoverCardContent, HoverCardTrigger } from '@/components/ui/hover-card'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { PinPreviewCard } from './pin-preview-card'
import { MapPin } from 'lucide-react'
import { useEffect, useState } from 'react'

interface PinMarkerProps {
  content: string
  onRemovePosition: () => void
  onRemovePin: () => void
}

export function PinMarker({
  content,
  onRemovePosition,
  onRemovePin,
}: PinMarkerProps) {
  const [isTouchDevice, setIsTouchDevice] = useState(false)

  useEffect(() => {
    setIsTouchDevice(window.matchMedia('(pointer: coarse)').matches)
  }, [])

  const trigger = (
    <button
      className="inline-flex items-center justify-center size-5 rounded-full bg-orange-100 dark:bg-orange-900/40 text-orange-600 dark:text-orange-400 hover:bg-orange-200 dark:hover:bg-orange-900/60 transition-colors ml-1.5 align-middle cursor-pointer"
      aria-label="查看固定的发现"
      onClick={(e) => e.stopPropagation()}
    >
      <MapPin className="size-3" />
    </button>
  )

  const previewCard = (
    <PinPreviewCard
      content={content}
      onRemovePosition={onRemovePosition}
      onRemovePin={onRemovePin}
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
