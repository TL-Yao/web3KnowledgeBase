'use client'

import { useState, useEffect } from 'react'
import { Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'

function getProgress(elapsed: number): number {
  if (elapsed <= 7) return (elapsed / 7) * 35
  if (elapsed <= 19) return 35 + ((elapsed - 7) / 12) * 45
  return Math.min(95, 80 + (elapsed - 19) * 0.67)
}

function getStageText(elapsed: number): string {
  if (elapsed < 7) return '正在启动 AI...'
  if (elapsed < 14) return '正在分析对话...'
  return '正在生成更新...'
}

interface UpdateProgressOverlayProps {
  onCancel: () => void
}

export function UpdateProgressOverlay({ onCancel }: UpdateProgressOverlayProps) {
  const [elapsed, setElapsed] = useState(0)

  useEffect(() => {
    const id = setInterval(() => setElapsed(s => s + 1), 1000)
    return () => clearInterval(id)
  }, [])

  const progress = getProgress(elapsed)

  return (
    <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/80 backdrop-blur-sm">
      <div className="flex flex-col items-center gap-4 w-72">
        <Sparkles className="w-8 h-8 text-primary animate-pulse" />
        <p className="text-base font-medium">正在生成文章更新</p>
        <div className="w-full h-2 rounded-full bg-muted overflow-hidden">
          <div
            className="h-full rounded-full bg-primary"
            style={{ width: `${progress}%`, transition: 'width 1s ease' }}
          />
        </div>
        <div className="flex items-center justify-between w-full text-xs text-muted-foreground">
          <span>{getStageText(elapsed)}</span>
          <span>已用时 {elapsed}s</span>
        </div>
        <Button variant="ghost" size="sm" className="text-xs text-muted-foreground" onClick={onCancel}>
          取消
        </Button>
      </div>
    </div>
  )
}
