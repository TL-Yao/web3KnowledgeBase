'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Clock, FileText, AlertCircle, Inbox, X } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { articleAPI, Article as APIArticle } from '@/lib/api'

interface Article {
  id: string
  slug: string
  title: string
  summary: string
  category: string | { name: string; slug: string }
  tags: string[]
  modelUsed?: string
  createdAt: string
  updatedAt: string
}

// Transform API article to component article format
function transformArticle(article: APIArticle): Article {
  return {
    id: article.id,
    slug: article.slug,
    title: article.title,
    summary: article.summary,
    category: article.category || 'Uncategorized',
    tags: article.tags || [],
    modelUsed: article.modelUsed,
    createdAt: article.createdAt,
    updatedAt: article.updatedAt,
  }
}

interface ArticleListProps {
  categoryId?: string
  searchQuery?: string
  activeTag?: string
  onTagClick?: (tag: string) => void
}

export function ArticleList({ categoryId, searchQuery, activeTag, onTagClick }: ArticleListProps) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['articles', categoryId, searchQuery, activeTag],
    queryFn: async () => {
      const response = await articleAPI.list({
        category: categoryId,
        tag: activeTag,
        q: searchQuery,
        limit: 20,
      })
      return response.articles.map(transformArticle)
    },
    retry: 1,
    staleTime: 30000,
  })

  const articles = data || []

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12 text-muted-foreground">
        <div className="animate-pulse">加载中...</div>
      </div>
    )
  }

  // Show error state when no cached data available
  if (isError && articles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <AlertCircle className="w-12 h-12 mb-4 text-red-400" />
        <p className="text-lg font-medium text-red-600">无法连接后端服务</p>
        <p className="text-sm">请检查服务状态后重试</p>
      </div>
    )
  }

  if (articles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <Inbox className="w-12 h-12 mb-4 opacity-50" />
        <p className="text-lg font-medium">暂无文章</p>
        <p className="text-sm">
          {activeTag ? `没有标签为「${activeTag}」的文章` : '开始研究以添加新的知识文章'}
        </p>
      </div>
    )
  }

  const getCategoryName = (category: string | { name: string; slug: string }) => {
    if (typeof category === 'string') return category
    return category?.name || '未分类'
  }

  const handleTagClick = (e: React.MouseEvent, tag: string) => {
    e.preventDefault()
    e.stopPropagation()
    onTagClick?.(tag)
  }

  return (
    <div>
      {isError && (
        <div className="flex items-center gap-2 p-3 mb-4 text-sm text-red-600 bg-red-50 rounded-lg border border-red-200">
          <AlertCircle className="w-4 h-4" />
          <span>部分数据可能已过时，后端服务连接异常</span>
        </div>
      )}

      {activeTag && (
        <div className="flex items-center gap-2 mb-4">
          <span className="text-sm text-muted-foreground">标签筛选:</span>
          <Badge
            variant="default"
            className="cursor-pointer gap-1"
            onClick={() => onTagClick?.('')}
          >
            {activeTag}
            <X className="w-3 h-3" />
          </Badge>
        </div>
      )}

      <ScrollArea className="h-[calc(100vh-200px)]">
        <div className="space-y-4">
          {articles.map((article) => (
            <Link key={article.id} href={`/knowledge/${article.slug}`}>
              <Card className="hover:bg-muted/50 transition-colors cursor-pointer">
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between">
                    <Badge variant="outline">{getCategoryName(article.category)}</Badge>
                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {formatDistanceToNow(new Date(article.updatedAt), {
                        addSuffix: true,
                        locale: zhCN
                      })}
                    </span>
                  </div>
                  <CardTitle className="text-lg mt-2">{article.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground line-clamp-2 mb-3">
                    {article.summary}
                  </p>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {article.tags?.slice(0, 3).map((tag) => (
                        <Badge
                          key={tag}
                          variant={tag === activeTag ? 'default' : 'secondary'}
                          className="text-xs cursor-pointer hover:bg-primary/20 transition-colors"
                          onClick={(e) => handleTagClick(e, tag)}
                        >
                          {tag}
                        </Badge>
                      ))}
                    </div>
                    {article.modelUsed && (
                      <span className="text-xs text-muted-foreground flex items-center gap-1">
                        <FileText className="w-3 h-3" />
                        {article.modelUsed}
                      </span>
                    )}
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      </ScrollArea>
    </div>
  )
}
