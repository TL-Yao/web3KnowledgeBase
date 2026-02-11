'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Clock, AlertCircle, Inbox, X, Tag, Key } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { articleAPI, Article as APIArticle } from '@/lib/api'
import { cn } from '@/lib/utils'

interface Article {
  id: string
  slug: string
  title: string
  summary: string
  tags: string[]
  sourceKeyword?: string
  archived?: boolean
  createdAt: string
  updatedAt: string
}

function transformArticle(article: APIArticle): Article {
  return {
    id: article.id,
    slug: article.slug,
    title: article.title,
    summary: article.summary,
    tags: article.tags || [],
    sourceKeyword: article.sourceKeyword,
    archived: article.archived,
    createdAt: article.createdAt,
    updatedAt: article.updatedAt,
  }
}

interface ArticleListProps {
  searchQuery?: string
  activeTag?: string
  archived?: 'true' | 'false' | 'all'
  onTagClick?: (tag: string) => void
}

export function ArticleList({ searchQuery, activeTag, archived, onTagClick }: ArticleListProps) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['articles', searchQuery, activeTag, archived],
    queryFn: async () => {
      const response = await articleAPI.list({
        tag: activeTag,
        q: searchQuery,
        limit: 20,
        archived,
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
              <Card className={cn(
                "hover:bg-muted/50 transition-colors cursor-pointer",
                article.archived && "opacity-60 bg-muted/50"
              )}>
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <CardTitle className="text-lg">{article.title}</CardTitle>
                      {article.archived && (
                        <Badge variant="secondary" className="bg-amber-100 text-amber-800 text-xs">
                          已归档
                        </Badge>
                      )}
                    </div>
                    <span className="text-xs text-muted-foreground flex items-center gap-1 shrink-0 ml-4">
                      <Clock className="w-3 h-3" />
                      {formatDistanceToNow(new Date(article.updatedAt), {
                        addSuffix: true,
                        locale: zhCN
                      })}
                    </span>
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground line-clamp-2 mb-3">
                    {article.summary}
                  </p>
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-1.5 flex-wrap">
                      <Tag className="w-3 h-3 text-muted-foreground shrink-0" />
                      {article.tags?.map((tag) => (
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
                    {article.sourceKeyword && (
                      <span className="text-xs text-muted-foreground flex items-center gap-1 shrink-0">
                        <Key className="w-3 h-3" />
                        {article.sourceKeyword}
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
