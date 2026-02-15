'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Sparkles, Loader2 } from 'lucide-react'

interface IntegrateButtonProps {
  pinCount: number
  onIntegrate: () => void
  isIntegrating?: boolean
  disabled?: boolean
}

export function IntegrateButton({
  pinCount,
  onIntegrate,
  isIntegrating,
  disabled,
}: IntegrateButtonProps) {
  const [showConfirm, setShowConfirm] = useState(false)

  if (pinCount === 0) return null

  const handleConfirm = () => {
    setShowConfirm(false)
    onIntegrate()
  }

  return (
    <>
      <Button
        size="sm"
        onClick={() => setShowConfirm(true)}
        disabled={isIntegrating || disabled}
        title={disabled ? '研究完成后可整合到报告' : undefined}
      >
        {isIntegrating ? (
          <Loader2 className="size-3.5 mr-1.5 animate-spin" />
        ) : (
          <Sparkles className="size-3.5 mr-1.5" />
        )}
        {isIntegrating ? '整合中...' : `整合 ${pinCount} 条发现`}
      </Button>

      <AlertDialog open={showConfirm} onOpenChange={setShowConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>整合发现到报告</AlertDialogTitle>
            <AlertDialogDescription>
              将 {pinCount} 条固定的发现整合到研究报告中。AI
              将重新分析并更新报告内容，这可能需要几分钟时间。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirm}>
              <Sparkles className="size-3.5 mr-1.5" />
              开始整合
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
