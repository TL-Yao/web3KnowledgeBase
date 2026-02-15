'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { GutterDot } from './gutter-dot'
import type { ResearchPinnedFinding } from '@/lib/api'

function extractText(children: React.ReactNode): string {
  if (typeof children === 'string') return children
  if (typeof children === 'number') return String(children)
  if (Array.isArray(children)) return children.map(extractText).join('')
  if (React.isValidElement(children)) {
    const props = children.props as { children?: React.ReactNode }
    if (props.children) return extractText(props.children)
  }
  return ''
}

interface PinnableBlockProps {
  tag: 'p' | 'h2' | 'h3' | 'h4' | 'li' | 'blockquote' | 'pre'
  blockIndex: number
  children: React.ReactNode
  isPlacementMode: boolean
  pins: Array<{ finding: ResearchPinnedFinding; findingIndex: number }>
  onSelectBlock?: (blockIndex: number, previewText: string) => void
  onMovePin?: (findingIndex: number) => void
  onRemovePin?: (findingIndex: number) => void
}

export function PinnableBlock({
  tag: Tag,
  blockIndex,
  children,
  isPlacementMode,
  pins,
  onSelectBlock,
  onMovePin,
  onRemovePin,
}: PinnableBlockProps) {
  const handleClick = () => {
    if (isPlacementMode && onSelectBlock) {
      const text = extractText(children)
      onSelectBlock(blockIndex, text.slice(0, 60))
    }
  }

  return (
    <Tag
      data-block-index={blockIndex}
      className={cn(
        'relative group/block',
        isPlacementMode && [
          'cursor-crosshair',
          'border-l-2 border-transparent hover:border-primary',
          'hover:bg-primary/5 transition-colors',
        ]
      )}
      onClick={isPlacementMode ? handleClick : undefined}
    >
      {pins.length > 0 && (
        <div
          className="absolute -left-4 top-1 flex flex-col gap-0.5"
          onClick={(e) => e.stopPropagation()}
        >
          {pins.map((pin) => (
            <GutterDot
              key={pin.findingIndex}
              finding={pin.finding}
              findingIndex={pin.findingIndex}
              onMove={onMovePin}
              onRemove={onRemovePin}
            />
          ))}
        </div>
      )}
      {children}
    </Tag>
  )
}
