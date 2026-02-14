'use client'

import { useState } from 'react'
import { MainLayout } from '@/components/layout/main-layout'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import { DomainSelector } from '@/components/research/domain-selector'
import { SessionList } from '@/components/research/session-list'
import { useResearch, useResearchSessions } from '@/hooks/use-research'
import { Telescope, Loader2, Send } from 'lucide-react'

export default function ResearchPage() {
  const [question, setQuestion] = useState('')
  const [domain, setDomain] = useState('auto')
  const [reviewPlan, setReviewPlan] = useState(false)
  const [page, setPage] = useState(1)

  const { startResearch, isStarting } = useResearch()
  const { sessions, total, isLoading, deleteSession, isDeleting } = useResearchSessions(page)

  const handleSubmit = async () => {
    if (!question.trim() || isStarting) return
    await startResearch({ question: question.trim(), domain, reviewPlan })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleSubmit()
    }
  }

  const limit = 20
  const hasMore = total > page * limit

  return (
    <MainLayout>
      <div className="flex-1 flex flex-col">
        {/* Hero Section */}
        <div className="flex flex-col items-center pt-16 pb-10 px-8">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <Telescope className="size-5 text-primary" />
            </div>
            <h1 className="text-3xl font-bold tracking-tight">即时研究</h1>
          </div>
          <p className="text-muted-foreground mb-8">
            输入任何主题，AI 为你生成深度研究报告
          </p>

          {/* Research Input */}
          <div className="w-full max-w-2xl space-y-3">
            <div className="relative">
              <Textarea
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="你想研究什么？例如：RISC-V 架构的发展现状和未来趋势..."
                className="min-h-[80px] pr-14 text-base resize-none"
                rows={3}
              />
              <Button
                size="icon"
                className="absolute bottom-3 right-3"
                onClick={handleSubmit}
                disabled={!question.trim() || isStarting}
              >
                {isStarting ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Send className="size-4" />
                )}
              </Button>
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <DomainSelector value={domain} onChange={setDomain} />
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="review-plan"
                    checked={reviewPlan}
                    onCheckedChange={(checked) => setReviewPlan(checked === true)}
                  />
                  <label
                    htmlFor="review-plan"
                    className="text-sm text-muted-foreground cursor-pointer select-none"
                  >
                    先审核研究计划
                  </label>
                </div>
              </div>
              <p className="text-xs text-muted-foreground hidden sm:block">
                ⌘↵ 开始研究
              </p>
            </div>
          </div>
        </div>

        {/* History Section */}
        <div className="flex-1 px-8 pb-8 max-w-4xl mx-auto w-full">
          <h2 className="text-lg font-semibold mb-4">研究记录</h2>
          <SessionList
            sessions={sessions}
            isLoading={isLoading}
            hasMore={hasMore}
            onLoadMore={() => setPage((p) => p + 1)}
            onDelete={deleteSession}
            isDeleting={isDeleting}
          />
        </div>
      </div>
    </MainLayout>
  )
}
