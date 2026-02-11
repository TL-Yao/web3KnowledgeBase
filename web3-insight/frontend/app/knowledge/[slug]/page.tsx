'use client'

import { useState, useRef } from 'react'
import { useParams } from 'next/navigation'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/main-layout'
import { ArticleView } from '@/components/knowledge/article-view'
import { FloatingChat } from '@/components/chat/floating-chat'
import { UpdateReviewPanel } from '@/components/knowledge/update-review-panel'
import { Button } from '@/components/ui/button'
import { articleAPI } from '@/lib/api'
import { AlertCircle, ArrowLeft, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import Link from 'next/link'
import type { Message } from '@/hooks/use-chat'

export default function ArticlePage() {
  const params = useParams()
  const slug = params.slug as string
  const queryClient = useQueryClient()

  const [isApplying, setIsApplying] = useState(false)
  const lastMessagesRef = useRef<Message[]>([])

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

  // Transform API response to match ArticleView's expected format
  const displayArticle = article ? {
    id: article.id,
    slug: article.slug,
    title: article.title,
    content: article.content,
    contentHtml: article.contentHtml || '', // API doesn't provide HTML, component will fall back to content
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

  // Handle error state first - show error message when backend fails
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

  // Handle not found state - article doesn't exist
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

  const handleGenerateUpdate = async (messages: Message[]) => {
    lastMessagesRef.current = messages
    setUpdateState(prev => ({ ...prev, isGenerating: true }))
    try {
      const result = await articleAPI.generateUpdate(displayArticle!.id, {
        conversationHistory: messages.map(m => ({ role: m.role, content: m.content }))
      })
      setUpdateState({
        isGenerating: false,
        isReviewOpen: true,
        updatedContent: result.updatedContent,
        changeSummary: result.changeSummary,
        model: result.model,
      })
    } catch {
      toast.error('生成更新失败')
      setUpdateState(prev => ({ ...prev, isGenerating: false }))
    }
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

  return (
    <MainLayout>
      <ArticleView article={displayArticle} />
      <FloatingChat
        articleId={displayArticle.id}
        articleTitle={displayArticle.title}
        onGenerateUpdate={handleGenerateUpdate}
        isGeneratingUpdate={updateState.isGenerating}
      />

      {updateState.isReviewOpen && updateState.updatedContent && (
        <UpdateReviewPanel
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
      )}
    </MainLayout>
  )
}
