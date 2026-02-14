'use client'

import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { cn } from '@/lib/utils'
import { ExternalLink, Loader2, Search, FileText, PenLine } from 'lucide-react'

export interface Citation {
  title: string
  url: string
  snippet?: string
}

type ResearchStage = 'planning' | 'researching' | 'writing' | 'completed' | 'failed'

interface ReportViewerProps {
  content?: string
  citations?: Citation[]
  stage?: ResearchStage
  stageDetail?: string
  error?: string
}

const STAGE_STEPS: { key: ResearchStage; label: string; icon: React.ElementType }[] = [
  { key: 'planning', label: '分析问题', icon: FileText },
  { key: 'researching', label: '搜索资料', icon: Search },
  { key: 'writing', label: '撰写报告', icon: PenLine },
]

function StageIndicator({ stage, stageDetail }: { stage: ResearchStage; stageDetail?: string }) {
  const activeIdx = STAGE_STEPS.findIndex((s) => s.key === stage)

  return (
    <div className="mb-8">
      <div className="flex items-center gap-3 mb-3">
        {STAGE_STEPS.map((step, idx) => {
          const isActive = idx === activeIdx
          const isDone = idx < activeIdx || stage === 'completed'
          const Icon = step.icon
          return (
            <div key={step.key} className="flex items-center gap-2">
              {idx > 0 && (
                <div
                  className={cn(
                    'w-8 h-px',
                    isDone ? 'bg-primary' : 'bg-border'
                  )}
                />
              )}
              <div
                className={cn(
                  'flex items-center gap-1.5 text-sm px-2.5 py-1 rounded-full border transition-colors',
                  isActive && 'border-primary bg-primary/10 text-primary font-medium',
                  isDone && !isActive && 'border-primary/50 text-primary/70',
                  !isActive && !isDone && 'border-border text-muted-foreground'
                )}
              >
                {isActive ? (
                  <Loader2 className="size-3.5 animate-spin" />
                ) : (
                  <Icon className="size-3.5" />
                )}
                {step.label}
              </div>
            </div>
          )
        })}
      </div>
      {stageDetail && (
        <p className="text-sm text-muted-foreground animate-pulse">{stageDetail}</p>
      )}
    </div>
  )
}

function LoadingSkeleton() {
  return (
    <div className="space-y-4 animate-pulse">
      <div className="h-6 bg-muted rounded w-3/4" />
      <div className="space-y-2">
        <div className="h-4 bg-muted rounded w-full" />
        <div className="h-4 bg-muted rounded w-5/6" />
        <div className="h-4 bg-muted rounded w-4/6" />
      </div>
      <div className="space-y-2 mt-6">
        <div className="h-4 bg-muted rounded w-full" />
        <div className="h-4 bg-muted rounded w-3/4" />
      </div>
    </div>
  )
}

function CitationList({ citations }: { citations: Citation[] }) {
  if (citations.length === 0) return null

  return (
    <div className="mt-8 pt-6 border-t border-border">
      <h3 className="text-sm font-semibold mb-3">参考来源</h3>
      <ol className="space-y-2">
        {citations.map((cite, idx) => (
          <li key={idx} className="flex items-start gap-2 text-sm">
            <span className="text-muted-foreground font-mono text-xs mt-0.5 shrink-0">
              [{idx + 1}]
            </span>
            <div className="min-w-0">
              <a
                href={cite.url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:underline underline-offset-2 inline-flex items-center gap-1"
              >
                {cite.title}
                <ExternalLink className="size-3 shrink-0" />
              </a>
              {cite.snippet && (
                <p className="text-muted-foreground text-xs mt-0.5 line-clamp-2">
                  {cite.snippet}
                </p>
              )}
            </div>
          </li>
        ))}
      </ol>
    </div>
  )
}

export function ReportViewer({ content, citations, stage, stageDetail, error }: ReportViewerProps) {
  const isGenerating = stage === 'planning' || stage === 'researching' || stage === 'writing'

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <div className="text-red-500 text-lg font-medium mb-2">生成失败</div>
        <p className="text-sm">{error}</p>
      </div>
    )
  }

  return (
    <div className="px-6 py-4">
      {isGenerating && <StageIndicator stage={stage!} stageDetail={stageDetail} />}

      {isGenerating && !content && <LoadingSkeleton />}

      {content && (
        <div
          className="prose prose-sm dark:prose-invert max-w-none
            prose-headings:font-semibold
            prose-h1:text-xl prose-h1:mb-4 prose-h1:mt-6
            prose-h2:text-lg prose-h2:mb-3 prose-h2:mt-5
            prose-h3:text-base prose-h3:mb-2 prose-h3:mt-4
            prose-p:text-sm prose-p:leading-relaxed prose-p:my-2
            prose-li:text-sm prose-li:my-0.5
            prose-a:text-primary prose-a:underline prose-a:underline-offset-2
            prose-blockquote:border-l-2 prose-blockquote:border-primary/30 prose-blockquote:text-muted-foreground
            prose-code:text-xs prose-code:bg-muted prose-code:px-1 prose-code:py-0.5 prose-code:rounded-sm
            prose-pre:bg-muted prose-pre:text-xs"
        >
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
              a: ({ href, children }) => (
                <a href={href} target="_blank" rel="noopener noreferrer">
                  {children}
                </a>
              ),
            }}
          >
            {content}
          </ReactMarkdown>
        </div>
      )}

      {citations && citations.length > 0 && <CitationList citations={citations} />}
    </div>
  )
}
