'use client'

import { useParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/main-layout'
import { ArticleView } from '@/components/knowledge/article-view'
import { FloatingChat } from '@/components/chat/floating-chat'

export default function ArticlePage() {
  const params = useParams()
  const slug = params.slug as string

  const { data: article, isLoading } = useQuery({
    queryKey: ['article', slug],
    queryFn: async () => {
      const res = await fetch(`/api/articles/${slug}`)
      return res.json()
    }
  })

  if (isLoading) {
    return (
      <MainLayout>
        <div className="p-6">加载中...</div>
      </MainLayout>
    )
  }

  return (
    <MainLayout>
      <ArticleView article={article} />
      <FloatingChat
        articleId={article?.id || slug}
        articleTitle={article?.title || '文章'}
      />
    </MainLayout>
  )
}
