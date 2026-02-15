'use client'

import { Button } from '@/components/ui/button'
import { MapPinOff, X } from 'lucide-react'

interface PinPreviewCardProps {
  content: string
  onRemovePosition: () => void
  onRemovePin: () => void
}

export function PinPreviewCard({
  content,
  onRemovePosition,
  onRemovePin,
}: PinPreviewCardProps) {
  const truncated = content.length > 200 ? content.slice(0, 200) + '...' : content

  return (
    <div className="space-y-3">
      <div className="text-sm text-muted-foreground leading-relaxed max-h-[120px] overflow-hidden relative">
        {truncated}
        {content.length > 200 && (
          <div className="absolute bottom-0 left-0 right-0 h-6 bg-gradient-to-t from-popover to-transparent" />
        )}
      </div>
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          className="h-7 text-xs"
          onClick={onRemovePosition}
        >
          <MapPinOff className="size-3 mr-1" />
          移除定位
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 text-xs text-destructive hover:text-destructive"
          onClick={onRemovePin}
        >
          <X className="size-3 mr-1" />
          取消固定
        </Button>
      </div>
    </div>
  )
}
