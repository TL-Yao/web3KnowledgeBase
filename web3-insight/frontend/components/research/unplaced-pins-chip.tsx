'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { MapPin, Crosshair, X } from 'lucide-react'
import type { ResearchPinnedFinding } from '@/lib/api'

interface UnplacedPinsChipProps {
  findings: Array<{ finding: ResearchPinnedFinding; findingIndex: number }>
  onPlace?: (findingIndex: number) => void
  onRemove?: (findingIndex: number) => void
}

export function UnplacedPinsChip({
  findings,
  onPlace,
  onRemove,
}: UnplacedPinsChipProps) {
  const [open, setOpen] = useState(false)

  if (findings.length === 0) return null

  return (
    <div className="sticky bottom-4 flex justify-end pr-4 pointer-events-none">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="secondary"
            size="sm"
            className="shadow-md pointer-events-auto"
          >
            <MapPin className="size-3.5 mr-1.5" />
            {findings.length} 条未放置
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-80" align="end">
          <div className="space-y-2 max-h-60 overflow-y-auto">
            {findings.map(({ finding, findingIndex }) => {
              const preview = finding.messageContent.length > 100
                ? finding.messageContent.slice(0, 100) + '...'
                : finding.messageContent
              return (
                <div
                  key={findingIndex}
                  className="flex items-start gap-2 text-sm bg-muted/50 rounded-md px-3 py-2 group"
                >
                  <span className="flex-1 line-clamp-2 text-muted-foreground">
                    {preview}
                  </span>
                  <div className="flex items-center gap-0.5 shrink-0">
                    {onPlace && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-6"
                        onClick={() => {
                          onPlace(findingIndex)
                          setOpen(false)
                        }}
                        title="放置到文章中"
                      >
                        <Crosshair className="size-3.5" />
                      </Button>
                    )}
                    {onRemove && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-6 text-destructive hover:text-destructive"
                        onClick={() => onRemove(findingIndex)}
                        title="取消固定"
                      >
                        <X className="size-3.5" />
                      </Button>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}
