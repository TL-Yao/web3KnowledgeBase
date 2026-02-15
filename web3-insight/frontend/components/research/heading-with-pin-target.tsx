'use client'

import React from 'react'
import { PinMarker } from './pin-marker'
import { cn } from '@/lib/utils'
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

interface HeadingWithPinTargetProps {
  level: 2 | 3
  blockIndex: number
  children: React.ReactNode
  isPlacementMode: boolean
  onSelectBlock?: (blockIndex: number, preview: string) => void
  pinnedFindings: ResearchPinnedFinding[]
  onRemovePosition?: (index: number) => void
  onRemovePin?: (index: number) => void
}

export function HeadingWithPinTarget({
  level,
  blockIndex,
  children,
  isPlacementMode,
  onSelectBlock,
  pinnedFindings,
  onRemovePosition,
  onRemovePin,
}: HeadingWithPinTargetProps) {
  const Tag = `h${level}` as const
  const headingText = extractText(children)

  // Find pins targeting this block
  const matchedPins = pinnedFindings
    .map((f, idx) => ({ finding: f, index: idx }))
    .filter(
      ({ finding }) =>
        finding.targetBlockIndex != null &&
        finding.targetBlockIndex === blockIndex
    )

  const handleClick = () => {
    if (isPlacementMode && onSelectBlock) {
      onSelectBlock(blockIndex, headingText.slice(0, 60))
    }
  }

  return (
    <Tag
      className={cn(
        isPlacementMode && [
          'border-2 border-dashed border-primary/40 rounded-md px-2 -mx-2',
          'cursor-pointer hover:border-primary hover:bg-primary/5 transition-colors',
        ]
      )}
      onClick={isPlacementMode ? handleClick : undefined}
    >
      {children}
      {matchedPins.map(({ finding, index }) => (
        <PinMarker
          key={index}
          content={finding.messageContent}
          onRemovePosition={() => onRemovePosition?.(index)}
          onRemovePin={() => onRemovePin?.(index)}
        />
      ))}
    </Tag>
  )
}
