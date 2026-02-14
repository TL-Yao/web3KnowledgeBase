'use client'

import { useState, useCallback, useEffect, useMemo } from 'react'
import { useParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/main-layout'
import { ReportViewer } from '@/components/research/report-viewer'
import { PlanReview } from '@/components/research/plan-review'
import { ResearchChat } from '@/components/research/research-chat'
import { IntegrateFindings } from '@/components/research/integrate-findings'
import { SidebarToggle } from '@/components/chat/sidebar-toggle'
import { ResizeHandle } from '@/components/ui/resize-handle'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useResearch } from '@/hooks/use-research'
import { useResearchChat } from '@/hooks/use-research-chat'
import { articleAPI } from '@/lib/api'
import { AlertCircle, ArrowLeft, RefreshCw } from 'lucide-react'
import Link from 'next/link'

const SIDEBAR_MIN = 320
const SIDEBAR_MAX = 600
const SIDEBAR_DEFAULT = 400

export default function ResearchSessionPage() {
  const params = useParams()
  const sessionId = params.id as string

  const {
    session,
    status,
    isLoadingSession,
    approvePlan,
    isApproving,
    cancelResearch,
    isCancelling,
    pinFinding,
    unpinFinding,
    integrateFindings,
    isIntegrating,
  } = useResearch(sessionId)

  const {
    messages,
    isLoading: isChatLoading,
    currentResponse,
    sendMessage,
    clearMessages,
    model,
    setModel,
  } = useResearchChat(sessionId)

  // Sidebar state with localStorage persistence
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)
  const [sidebarWidth, setSidebarWidth] = useState(SIDEBAR_DEFAULT)
  const [isDragging, setIsDragging] = useState(false)

  useEffect(() => {
    const savedOpen = localStorage.getItem('research-sidebar-open')
    if (savedOpen === 'false') setIsSidebarOpen(false)
    const savedWidth = localStorage.getItem('research-sidebar-width')
    if (savedWidth) setSidebarWidth(parseInt(savedWidth, 10))
  }, [])

  const toggleSidebar = useCallback(() => {
    setIsSidebarOpen(prev => {
      const next = !prev
      localStorage.setItem('research-sidebar-open', String(next))
      return next
    })
  }, [])

  const handleResize = useCallback((delta: number) => {
    setSidebarWidth(prev =>
      Math.min(SIDEBAR_MAX, Math.max(SIDEBAR_MIN, prev - delta))
    )
  }, [])

  const handleResizeEnd = useCallback(() => {
    setSidebarWidth(prev => {
      localStorage.setItem('research-sidebar-width', String(prev))
      return prev
    })
  }, [])

  const handleResizeReset = useCallback(() => {
    setSidebarWidth(SIDEBAR_DEFAULT)
    localStorage.setItem('research-sidebar-width', String(SIDEBAR_DEFAULT))
  }, [])

  // Track pinned message contents for visual feedback in chat
  const pinnedContents = useMemo(() => {
    const set = new Set<string>()
    session?.pinnedFindings?.forEach(f => set.add(f.messageContent))
    return set
  }, [session?.pinnedFindings])

  // Fetch linked article when session is completed
  const articleSlug = status?.articleSlug ?? session?.articleSlug
  const { data: article } = useQuery({
    queryKey: ['article', articleSlug],
    queryFn: () => articleAPI.get(articleSlug!),
    enabled: !!articleSlug,
    staleTime: 30000,
  })

  const currentStatus = status?.status ?? session?.status
  const currentStage = status?.stage ?? session?.stage
  const currentStageDetail = status?.stageDetail ?? session?.stageDetail
  const currentPlan = status?.researchPlan ?? session?.researchPlan
  const currentError = status?.error ?? session?.error

  // Loading state
  if (isLoadingSession && !session) {
    return (
      <MainLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4" />
            <p className="text-muted-foreground">加载研究会话...</p>
          </div>
        </div>
      </MainLayout>
    )
  }

  // Error state — session not found
  if (!session && !isLoadingSession) {
    return (
      <MainLayout>
        <div className="flex flex-col items-center justify-center h-full p-6">
          <AlertCircle className="w-16 h-16 text-red-400 mb-4" />
          <h2 className="text-xl font-semibold text-red-600 mb-2">会话未找到</h2>
          <p className="text-muted-foreground mb-6">
            无法找到此研究会话，可能已被删除
          </p>
          <Link href="/research">
            <Button variant="outline">
              <ArrowLeft className="w-4 h-4 mr-2" />
              返回研究首页
            </Button>
          </Link>
        </div>
      </MainLayout>
    )
  }

  const isActive = currentStatus === 'planning' || currentStatus === 'researching' || currentStatus === 'writing'
  const isPlanReview = currentStatus === 'plan_review'
  const isCompleted = currentStatus === 'completed'
  const isFailed = currentStatus === 'failed'

  // Map status to ReportViewer stage
  const reportStage = currentStatus === 'planning' ? 'planning' as const
    : currentStatus === 'researching' ? 'researching' as const
    : currentStatus === 'writing' ? 'writing' as const
    : isCompleted ? 'completed' as const
    : isFailed ? 'failed' as const
    : undefined

  const pinnedPreviews = session?.pinnedFindings?.map(f =>
    f.messageContent.length > 120 ? f.messageContent.slice(0, 120) + '...' : f.messageContent
  ) ?? []

  return (
    <MainLayout>
      <div className="flex h-full relative">
        {/* Left panel — Report / Plan / Status */}
        <div className="flex-1 min-w-0 overflow-auto flex flex-col">
          {/* Header bar */}
          <div className="px-6 py-3 border-b border-border flex items-center justify-between shrink-0">
            <div className="flex items-center gap-3 min-w-0">
              <Link href="/research">
                <Button variant="ghost" size="icon" className="size-8 shrink-0">
                  <ArrowLeft className="size-4" />
                </Button>
              </Link>
              <h1 className="text-sm font-medium truncate">
                {session?.question}
              </h1>
            </div>
            {currentStatus && (
              <Badge
                variant={isCompleted ? 'default' : isFailed ? 'destructive' : 'outline'}
                className="shrink-0 ml-2"
              >
                {isActive && <span className="inline-block size-1.5 rounded-full bg-current mr-1.5 animate-pulse" />}
                {currentStatus === 'planning' ? '规划中'
                  : currentStatus === 'plan_review' ? '待审核'
                  : currentStatus === 'researching' ? '研究中'
                  : currentStatus === 'writing' ? '撰写中'
                  : isCompleted ? '已完成'
                  : isFailed ? '失败'
                  : '等待中'}
              </Badge>
            )}
          </div>

          {/* Main content area */}
          <div className="flex-1 overflow-auto">
            {isPlanReview && currentPlan && (
              <PlanReview
                plan={currentPlan}
                onApprove={(edited) => approvePlan(edited)}
                onCancel={() => cancelResearch()}
                isApproving={isApproving}
                isCancelling={isCancelling}
              />
            )}

            {isActive && (
              <ReportViewer
                stage={reportStage}
                stageDetail={currentStageDetail}
                content={article?.content}
              />
            )}

            {isCompleted && article && (
              <ReportViewer
                stage="completed"
                content={article.content}
                citations={session?.citations}
              />
            )}

            {isFailed && (
              <ReportViewer
                error={currentError || '研究过程中发生未知错误'}
              />
            )}

            {/* Fallback — pending with no specific state yet */}
            {currentStatus === 'pending' && (
              <div className="flex items-center justify-center py-16 text-muted-foreground">
                <div className="text-center">
                  <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mx-auto mb-3" />
                  <p className="text-sm">正在启动研究...</p>
                </div>
              </div>
            )}
          </div>

          {/* Pinned findings bar */}
          {isCompleted && (session?.pinnedFindings?.length ?? 0) > 0 && (
            <IntegrateFindings
              pinnedCount={session?.pinnedFindings?.length ?? 0}
              pinnedPreviews={pinnedPreviews}
              onIntegrate={() => integrateFindings()}
              onRemove={(idx) => unpinFinding(idx)}
              isIntegrating={isIntegrating}
            />
          )}
        </div>

        {/* Resize handle */}
        {isSidebarOpen && (
          <ResizeHandle
            onResize={handleResize}
            onResizeEnd={handleResizeEnd}
            onReset={handleResizeReset}
            onDragChange={setIsDragging}
          />
        )}

        {/* Chat sidebar */}
        <ResearchChat
          isOpen={isSidebarOpen}
          onToggle={toggleSidebar}
          width={sidebarWidth}
          isDragging={isDragging}
          messages={messages}
          currentResponse={currentResponse}
          isLoading={isChatLoading}
          model={model}
          onModelChange={setModel}
          onSendMessage={sendMessage}
          onClearMessages={clearMessages}
          onPinFinding={isCompleted ? (content) => pinFinding(content) : undefined}
          pinnedContents={pinnedContents}
        />

        {/* Toggle button when sidebar is closed */}
        {!isSidebarOpen && <SidebarToggle onClick={toggleSidebar} />}
      </div>
    </MainLayout>
  )
}
