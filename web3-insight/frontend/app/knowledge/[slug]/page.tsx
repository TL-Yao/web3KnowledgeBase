'use client'

import { useState, useRef, useCallback, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/main-layout'
import { ArticleView } from '@/components/knowledge/article-view'
import { ArticleEditor } from '@/components/knowledge/article-editor'
import { ChatSidebar } from '@/components/chat/chat-sidebar'
import { SidebarToggle } from '@/components/chat/sidebar-toggle'
import { ResizeHandle } from '@/components/ui/resize-handle'
import { UpdateReviewPanel } from '@/components/knowledge/update-review-panel'
import { UpdateProgressOverlay } from '@/components/knowledge/update-progress-overlay'
import { Button } from '@/components/ui/button'
import { articleAPI } from '@/lib/api'
import { AlertCircle, ArrowLeft, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import Link from 'next/link'
import type { Message, ChatModel } from '@/hooks/use-chat'
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

const SIDEBAR_MIN = 320
const SIDEBAR_MAX = 600
const SIDEBAR_DEFAULT = 400

export default function ArticlePage() {
  const params = useParams()
  const slug = params.slug as string
  const queryClient = useQueryClient()

  const [isEditing, setIsEditing] = useState(false)
  const [isSavingEdit, setIsSavingEdit] = useState(false)
  const [isApplying, setIsApplying] = useState(false)
  const [clearTrigger, setClearTrigger] = useState(0)
  const lastMessagesRef = useRef<Message[]>([])
  const lastModelRef = useRef<ChatModel>('sonnet')
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const jobIdRef = useRef<string | null>(null)

  // Sidebar state with localStorage persistence (SSR-safe)
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)
  const [sidebarWidth, setSidebarWidth] = useState(SIDEBAR_DEFAULT)
  const [isDragging, setIsDragging] = useState(false)

  useEffect(() => {
    const savedOpen = localStorage.getItem('chat-sidebar-open')
    if (savedOpen === 'false') setIsSidebarOpen(false)
    const savedWidth = localStorage.getItem('chat-sidebar-width')
    if (savedWidth) setSidebarWidth(parseInt(savedWidth, 10))
  }, [])

  const toggleSidebar = useCallback(() => {
    setIsSidebarOpen(prev => {
      const next = !prev
      localStorage.setItem('chat-sidebar-open', String(next))
      return next
    })
  }, [])

  const handleResize = useCallback((delta: number) => {
    setSidebarWidth(prev => {
      // Dragging left (negative delta) = wider sidebar
      const next = Math.min(SIDEBAR_MAX, Math.max(SIDEBAR_MIN, prev - delta))
      return next
    })
  }, [])

  const handleResizeEnd = useCallback(() => {
    setSidebarWidth(prev => {
      localStorage.setItem('chat-sidebar-width', String(prev))
      return prev
    })
  }, [])

  const handleResizeReset = useCallback(() => {
    setSidebarWidth(SIDEBAR_DEFAULT)
    localStorage.setItem('chat-sidebar-width', String(SIDEBAR_DEFAULT))
  }, [])

  const [showCliFailedDialog, setShowCliFailedDialog] = useState(false)

  const [updateState, setUpdateState] = useState<{
    isGenerating: boolean
    isReviewOpen: boolean
    updatedContent: string | null
    changeSummary: string | null
    model: string | null
  }>({
    isGenerating: false,
    isReviewOpen: false,
    updatedContent: null,
    changeSummary: null,
    model: null,
  })

  const { data: article, isLoading, isError, refetch } = useQuery({
    queryKey: ['article', slug],
    queryFn: () => articleAPI.get(slug),
    retry: 1,
    staleTime: 30000,
  })

  const stopPolling = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current)
      pollingRef.current = null
    }
  }, [])

  // Cleanup polling on unmount
  useEffect(() => {
    return () => stopPolling()
  }, [stopPolling])

  const startPolling = useCallback((articleId: string, jobId: string) => {
    stopPolling()
    jobIdRef.current = jobId

    pollingRef.current = setInterval(async () => {
      try {
        const status = await articleAPI.getUpdateStatus(articleId, jobId)

        if (status.status === 'completed' && status.result) {
          stopPolling()
          jobIdRef.current = null
          if (status.result.noChange) {
            toast.info(status.result.noChangeReason || '对话内容未涉及文章的实质补充，无需更新')
            setUpdateState(prev => ({ ...prev, isGenerating: false }))
            return
          }
          setUpdateState({
            isGenerating: false,
            isReviewOpen: true,
            updatedContent: status.result.updatedContent,
            changeSummary: status.result.changeSummary,
            model: status.result.model,
          })
        } else if (status.status === 'failed') {
          stopPolling()
          jobIdRef.current = null
          if (status.errorType === 'cli_failed') {
            setUpdateState(prev => ({ ...prev, isGenerating: false }))
            setShowCliFailedDialog(true)
            return
          }
          toast.error(status.error || '生成更新失败')
          setUpdateState(prev => ({ ...prev, isGenerating: false }))
        }
      } catch {
        // Polling fetch failed — keep retrying silently
      }
    }, 3000)
  }, [stopPolling])

  // Transform API response to match ArticleView's expected format
  const displayArticle = article ? {
    id: article.id,
    slug: article.slug,
    title: article.title,
    content: article.content,
    contentHtml: article.contentHtml || '',
    summary: article.summary,
    category: {
      id: article.category?.id || 'uncategorized',
      name: article.category?.name || '未分类',
      slug: article.category?.slug || 'uncategorized'
    },
    tags: article.tags,
    sourceUrls: article.sourceUrls || [],
    modelUsed: article.modelUsed || '',
    archived: article.archived,
    createdAt: article.createdAt,
    updatedAt: article.updatedAt,
  } : null

  if (isLoading) {
    return (
      <MainLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">加载文章中...</p>
          </div>
        </div>
      </MainLayout>
    )
  }

  if (isError) {
    return (
      <MainLayout>
        <div className="flex flex-col items-center justify-center h-full p-6">
          <AlertCircle className="w-16 h-16 text-red-400 mb-4" />
          <h2 className="text-xl font-semibold text-red-600 mb-2">加载失败</h2>
          <p className="text-muted-foreground mb-6">
            无法加载文章，请检查后端服务
          </p>
          <div className="flex gap-4">
            <Link href="/knowledge">
              <Button variant="outline">
                <ArrowLeft className="w-4 h-4 mr-2" />
                返回知识库
              </Button>
            </Link>
            <Button onClick={() => refetch()}>
              <RefreshCw className="w-4 h-4 mr-2" />
              重试
            </Button>
          </div>
        </div>
      </MainLayout>
    )
  }

  if (!displayArticle) {
    return (
      <MainLayout>
        <div className="flex flex-col items-center justify-center h-full p-6">
          <AlertCircle className="w-16 h-16 text-muted-foreground mb-4" />
          <h2 className="text-xl font-semibold mb-2">文章未找到</h2>
          <p className="text-muted-foreground mb-6">
            无法找到 slug 为 &quot;{slug}&quot; 的文章
          </p>
          <div className="flex gap-4">
            <Link href="/knowledge">
              <Button variant="outline">
                <ArrowLeft className="w-4 h-4 mr-2" />
                返回知识库
              </Button>
            </Link>
            <Button onClick={() => refetch()}>
              <RefreshCw className="w-4 h-4 mr-2" />
              重试
            </Button>
          </div>
        </div>
      </MainLayout>
    )
  }

  const handleGenerateUpdate = async (messages: Message[], modelOrMethod?: ChatModel | 'cli' | 'api', method?: 'cli' | 'api') => {
    // Called from chat sidebar: (messages, chatModel)
    // Called from retry/API dialog: (messages, undefined, method)
    let chatModel: ChatModel | undefined
    let updateMethod: 'cli' | 'api' | undefined = method
    if (modelOrMethod === 'cli' || modelOrMethod === 'api') {
      updateMethod = modelOrMethod
    } else if (modelOrMethod) {
      chatModel = modelOrMethod
    }

    lastMessagesRef.current = messages
    if (chatModel) lastModelRef.current = chatModel
    stopPolling()

    setUpdateState(prev => ({ ...prev, isGenerating: true }))
    if (!isSidebarOpen) {
      setIsSidebarOpen(true)
      localStorage.setItem('chat-sidebar-open', 'true')
    }
    try {
      const { jobId } = await articleAPI.generateUpdate(
        displayArticle!.id,
        { conversationHistory: messages.map(m => ({ role: m.role, content: m.content })) },
        { method: updateMethod, model: lastModelRef.current },
      )
      startPolling(displayArticle!.id, jobId)
    } catch {
      toast.error('生成更新失败')
      setUpdateState(prev => ({ ...prev, isGenerating: false }))
    }
  }

  const handleCancelGeneration = () => {
    const jobId = jobIdRef.current
    const articleId = displayArticle?.id
    stopPolling()
    jobIdRef.current = null
    setUpdateState(prev => ({ ...prev, isGenerating: false }))
    toast.info('已取消更新生成')
    if (jobId && articleId) {
      articleAPI.cancelUpdate(articleId, jobId).catch(() => {})
    }
  }

  const handleRetryCli = () => {
    setShowCliFailedDialog(false)
    handleGenerateUpdate(lastMessagesRef.current, undefined, 'cli')
  }

  const handleUseApi = () => {
    setShowCliFailedDialog(false)
    handleGenerateUpdate(lastMessagesRef.current, undefined, 'api')
  }

  const handleRegenerate = () => {
    handleGenerateUpdate(lastMessagesRef.current)
  }

  const handleApplyUpdate = async () => {
    if (!updateState.updatedContent || !updateState.changeSummary) return
    setIsApplying(true)
    try {
      await articleAPI.applyUpdate(displayArticle!.id, {
        updatedContent: updateState.updatedContent,
        changeSummary: updateState.changeSummary,
      })
      toast.success('文章已更新')
      setUpdateState({ isGenerating: false, isReviewOpen: false, updatedContent: null, changeSummary: null, model: null })
      setClearTrigger(prev => prev + 1)
      queryClient.invalidateQueries({ queryKey: ['article', slug] })
      queryClient.invalidateQueries({ queryKey: ['article-versions', displayArticle!.id] })
    } catch {
      toast.error('应用更新失败')
    } finally {
      setIsApplying(false)
    }
  }

  const handleCancelReview = () => {
    setUpdateState({ isGenerating: false, isReviewOpen: false, updatedContent: null, changeSummary: null, model: null })
  }

  const handleSaveEdit = async (title: string, content: string) => {
    setIsSavingEdit(true)
    try {
      await articleAPI.saveEdit(displayArticle!.id, { title, content })
      toast.success('文章已保存')
      setIsEditing(false)
      queryClient.invalidateQueries({ queryKey: ['article', slug] })
      queryClient.invalidateQueries({ queryKey: ['article-versions', displayArticle!.id] })
    } catch {
      toast.error('保存失败')
    } finally {
      setIsSavingEdit(false)
    }
  }

  const handleCancelEdit = () => {
    setIsEditing(false)
  }

  const canEdit = !updateState.isGenerating && !updateState.isReviewOpen

  return (
    <MainLayout>
      <div className="flex h-full relative">
        {/* Article content or UpdateReviewPanel */}
        <div className="flex-1 min-w-0 overflow-auto relative">
          {isEditing ? (
            <ArticleEditor
              article={displayArticle}
              onSave={handleSaveEdit}
              onCancel={handleCancelEdit}
              isSaving={isSavingEdit}
            />
          ) : updateState.isReviewOpen && updateState.updatedContent ? (
            <UpdateReviewPanel
              variant="inline"
              articleTitle={displayArticle.title}
              originalContent={displayArticle.content}
              updatedContent={updateState.updatedContent}
              changeSummary={updateState.changeSummary!}
              model={updateState.model!}
              onApply={handleApplyUpdate}
              onRegenerate={handleRegenerate}
              onCancel={handleCancelReview}
              isApplying={isApplying}
              isRegenerating={updateState.isGenerating}
            />
          ) : (
            <ArticleView
              article={displayArticle}
              onEdit={() => setIsEditing(true)}
              canEdit={canEdit}
            />
          )}
          {updateState.isGenerating && <UpdateProgressOverlay onCancel={handleCancelGeneration} />}
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
        <ChatSidebar
          articleId={displayArticle.id}
          articleTitle={displayArticle.title}
          isOpen={isSidebarOpen}
          onToggle={toggleSidebar}
          onGenerateUpdate={handleGenerateUpdate}
          isGeneratingUpdate={updateState.isGenerating}
          isEditing={isEditing}
          width={sidebarWidth}
          isDragging={isDragging}
          clearTrigger={clearTrigger}
        />

        {/* Toggle button when sidebar is closed */}
        {!isSidebarOpen && <SidebarToggle onClick={toggleSidebar} />}
      </div>

      {/* CLI failure dialog */}
      <AlertDialog open={showCliFailedDialog} onOpenChange={setShowCliFailedDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>CLI 生成失败</AlertDialogTitle>
            <AlertDialogDescription>
              本地 CLI 生成更新时出错。你可以重试，或切换到 API 生成（会产生少量费用）。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              className="border border-input bg-background text-foreground shadow-xs hover:bg-accent hover:text-accent-foreground"
              onClick={handleRetryCli}
            >
              重试
            </AlertDialogAction>
            <AlertDialogAction onClick={handleUseApi}>使用 API 生成（付费）</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </MainLayout>
  )
}
