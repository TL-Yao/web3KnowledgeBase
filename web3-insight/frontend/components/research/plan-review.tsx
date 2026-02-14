'use client'

import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Check, PenLine, X, Loader2 } from 'lucide-react'

interface PlanReviewProps {
  plan: string
  onApprove: (editedPlan?: string) => void
  onCancel: () => void
  isApproving?: boolean
  isCancelling?: boolean
}

export function PlanReview({
  plan,
  onApprove,
  onCancel,
  isApproving,
  isCancelling,
}: PlanReviewProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editedPlan, setEditedPlan] = useState(plan)

  const handleApprove = () => {
    const changed = editedPlan.trim() !== plan.trim() ? editedPlan : undefined
    onApprove(changed)
  }

  return (
    <div className="px-6 py-4">
      <div className="mb-4">
        <h2 className="text-lg font-semibold mb-1">研究计划</h2>
        <p className="text-sm text-muted-foreground">
          请审核以下研究大纲，确认后将开始生成报告
        </p>
      </div>

      {isEditing ? (
        <div className="space-y-3">
          <Textarea
            value={editedPlan}
            onChange={(e) => setEditedPlan(e.target.value)}
            className="min-h-[300px] font-mono text-sm"
            placeholder="编辑研究计划..."
          />
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setIsEditing(false)
                setEditedPlan(plan)
              }}
            >
              <X className="size-3.5 mr-1" />
              取消编辑
            </Button>
          </div>
        </div>
      ) : (
        <div
          className="rounded-lg border border-border bg-muted/30 p-4 prose prose-sm dark:prose-invert max-w-none
            prose-headings:font-semibold
            prose-h2:text-base prose-h2:mb-2 prose-h2:mt-4
            prose-h3:text-sm prose-h3:mb-1.5 prose-h3:mt-3
            prose-p:text-sm prose-p:my-1.5
            prose-li:text-sm prose-li:my-0.5
            prose-ul:my-1.5 prose-ol:my-1.5
            prose-a:text-primary prose-a:underline prose-a:underline-offset-2"
        >
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {editedPlan}
          </ReactMarkdown>
        </div>
      )}

      <div className="flex items-center gap-3 mt-6">
        <Button onClick={handleApprove} disabled={isApproving || isCancelling}>
          {isApproving ? (
            <Loader2 className="size-4 mr-1.5 animate-spin" />
          ) : (
            <Check className="size-4 mr-1.5" />
          )}
          {isApproving ? '确认中...' : '确认并开始研究'}
        </Button>

        {!isEditing && (
          <Button
            variant="outline"
            onClick={() => setIsEditing(true)}
            disabled={isApproving || isCancelling}
          >
            <PenLine className="size-4 mr-1.5" />
            编辑计划
          </Button>
        )}

        <Button
          variant="ghost"
          className="text-destructive hover:text-destructive"
          onClick={onCancel}
          disabled={isApproving || isCancelling}
        >
          {isCancelling ? (
            <Loader2 className="size-4 mr-1.5 animate-spin" />
          ) : (
            <X className="size-4 mr-1.5" />
          )}
          取消研究
        </Button>
      </div>
    </div>
  )
}
