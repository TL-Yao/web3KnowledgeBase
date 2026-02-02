'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Clock, FileText, AlertCircle, Inbox } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { articleAPI, Article as APIArticle, APIError } from '@/lib/api'

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

// Mock data for fallback when API is unavailable
const mockArticles: Article[] = [
  {
    id: '1',
    slug: 'eip-4844-proto-danksharding',
    title: 'EIP-4844: Proto-Danksharding 详解',
    summary: 'EIP-4844 引入了一种新的交易类型，允许在以太坊上发布 blob 数据，为 Rollup 提供更便宜的数据可用性层。',
    category: { name: 'Layer 2', slug: 'layer-2' },
    tags: ['EIP', 'Danksharding', 'Rollup'],
    modelUsed: 'llama3:70b',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  },
  {
    id: '2',
    slug: 'zk-rollup-technology',
    title: 'ZK Rollup 技术原理与应用',
    summary: '深入理解零知识证明在 Rollup 扩容方案中的应用，包括 zkSync、StarkNet 等主流项目的技术架构。',
    category: { name: 'Layer 2', slug: 'layer-2' },
    tags: ['ZK Proof', 'zkSync', 'StarkNet'],
    modelUsed: 'claude-sonnet-4-20250514',
    createdAt: new Date(Date.now() - 86400000).toISOString(),
    updatedAt: new Date(Date.now() - 86400000).toISOString()
  },
  {
    id: '3',
    slug: 'defi-lending-protocols',
    title: 'DeFi 借贷协议深度分析',
    summary: '分析 Aave、Compound 等主流借贷协议的机制设计、风险模型和治理架构。',
    category: { name: 'DeFi', slug: 'defi' },
    tags: ['Aave', 'Compound', 'Lending'],
    modelUsed: 'llama3:70b',
    createdAt: new Date(Date.now() - 172800000).toISOString(),
    updatedAt: new Date(Date.now() - 172800000).toISOString()
  }
]

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

export function ArticleList() {
  const { data, isLoading, error, isError } = useQuery({
    queryKey: ['articles'],
    queryFn: async () => {
      const response = await articleAPI.list({ limit: 20 })
      return response.data.map(transformArticle)
    },
    retry: 1,
    staleTime: 30000,
  })

  // Use mock data when API fails or returns empty
  const articles = (data && data.length > 0) ? data : (isError ? mockArticles : data)
  const showingMockData = isError || (data && data.length === 0)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12 text-muted-foreground">
        <div className="animate-pulse">加载中...</div>
      </div>
    )
  }

  if (!articles || articles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <Inbox className="w-12 h-12 mb-4 opacity-50" />
        <p className="text-lg font-medium">暂无文章</p>
        <p className="text-sm">开始研究以添加新的知识文章</p>
      </div>
    )
  }

  const getCategoryName = (category: string | { name: string; slug: string }) => {
    if (typeof category === 'string') return category
    return category?.name || '未分类'
  }

  return (
    <div>
      {showingMockData && (
        <div className="flex items-center gap-2 p-3 mb-4 text-sm text-amber-600 bg-amber-50 rounded-lg border border-amber-200">
          <AlertCircle className="w-4 h-4" />
          <span>后端服务未连接，显示示例数据</span>
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
                        <Badge key={tag} variant="secondary" className="text-xs">
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
