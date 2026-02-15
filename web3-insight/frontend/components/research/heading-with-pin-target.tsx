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

function normalizeHeading(text: string): string {
  return text.trim().toLowerCase().replace(/\s+/g, ' ')
}

interface HeadingWithPinTargetProps {
  level: 2 | 3
  children: React.ReactNode
  isPlacementMode: boolean
  onSelectSection?: (headingText: string) => void
  pinnedFindings: ResearchPinnedFinding[]
  onRemovePosition?: (index: number) => void
  onRemovePin?: (index: number) => void
}

export function HeadingWithPinTarget({
  level,
  children,
  isPlacementMode,
  onSelectSection,
  pinnedFindings,
  onRemovePosition,
  onRemovePin,
}: HeadingWithPinTargetProps) {
  const Tag = `h${level}` as const
  const headingText = extractText(children)
  const normalizedHeading = normalizeHeading(headingText)

  // Find pins targeting this heading
  const matchedPins = pinnedFindings
    .map((f, idx) => ({ finding: f, index: idx }))
    .filter(
      ({ finding }) =>
        finding.targetSection != null &&
        normalizeHeading(finding.targetSection) === normalizedHeading
    )

  const handleClick = () => {
    if (isPlacementMode && onSelectSection) {
      onSelectSection(headingText)
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
