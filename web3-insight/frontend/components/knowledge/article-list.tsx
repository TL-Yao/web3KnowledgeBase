'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Clock, FileText } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

interface Article {
  id: string
  slug: string
  title: string
  summary: string
  category: { name: string; slug: string }
  tags: string[]
  modelUsed: string
  createdAt: string
  updatedAt: string
}

export function ArticleList() {
  const { data: articles, isLoading } = useQuery({
    queryKey: ['articles'],
    queryFn: async () => {
      // TODO: Fetch from API
      return [
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
      ] as Article[]
    }
  })

  if (isLoading) {
    return (
      <div className="text-muted-foreground">加载中...</div>
    )
  }

  return (
    <ScrollArea className="h-[calc(100vh-200px)]">
      <div className="space-y-4">
        {articles?.map((article) => (
          <Link key={article.id} href={`/knowledge/${article.slug}`}>
            <Card className="hover:bg-muted/50 transition-colors cursor-pointer">
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <Badge variant="outline">{article.category?.name}</Badge>
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
                  <span className="text-xs text-muted-foreground flex items-center gap-1">
                    <FileText className="w-3 h-3" />
                    {article.modelUsed}
                  </span>
                </div>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </ScrollArea>
  )
}
