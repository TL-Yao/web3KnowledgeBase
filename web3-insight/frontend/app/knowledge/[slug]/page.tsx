'use client'

import { useParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/main-layout'
import { ArticleView } from '@/components/knowledge/article-view'
import { FloatingChat } from '@/components/chat/floating-chat'
import { Button } from '@/components/ui/button'
import { articleAPI } from '@/lib/api'
import { AlertCircle, ArrowLeft, RefreshCw } from 'lucide-react'
import Link from 'next/link'

export default function ArticlePage() {
  const params = useParams()
  const slug = params.slug as string

  const { data: article, isLoading, isError, refetch } = useQuery({
    queryKey: ['article', slug],
    queryFn: () => articleAPI.get(slug),
    retry: 1,
    staleTime: 30000,
  })

  // Transform API response to match ArticleView's expected format
  const displayArticle = article ? {
    id: article.id,
    title: article.title,
    content: article.content,
    contentHtml: '', // API doesn't provide HTML, component will fall back to content
    summary: article.summary,
    category: { name: article.category || '未分类', slug: (article.category || 'uncategorized').toLowerCase() },
    tags: article.tags,
    sourceUrls: article.source_url ? [article.source_url] : [],
    modelUsed: '',
    createdAt: article.created_at,
    updatedAt: article.updated_at,
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

  return (
    <MainLayout>
      <ArticleView article={displayArticle} />
      <FloatingChat
        articleId={displayArticle.id}
        articleTitle={displayArticle.title}
      />
    </MainLayout>
  )
}
