'use client'

import { useParams } from 'next/navigation'
import { useQuery } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/main-layout'
import { ArticleView } from '@/components/knowledge/article-view'
import { FloatingChat } from '@/components/chat/floating-chat'
import { Button } from '@/components/ui/button'
import { articleAPI, Article } from '@/lib/api'
import { AlertCircle, ArrowLeft, RefreshCw } from 'lucide-react'
import Link from 'next/link'

// Mock article for fallback
const mockArticle: Article = {
  id: 'mock-1',
  title: 'EIP-4844: Proto-Danksharding 详解',
  slug: 'eip-4844-proto-danksharding',
  content: `# EIP-4844: Proto-Danksharding

## 概述

EIP-4844，也称为 Proto-Danksharding，是以太坊的一个重要升级提案。它引入了一种新的交易类型——携带 blob 的交易（blob-carrying transactions）。

## 主要特点

### 1. Blob 数据

- Blob 是一种大型数据块，最大可达 128 KB
- Blob 数据不会被 EVM 执行
- Blob 数据只在信标链上保留约 18 天

### 2. 数据可用性

Proto-Danksharding 为 Layer 2 Rollup 提供了更便宜的数据可用性层：

- 降低 Rollup 的数据发布成本
- 提高以太坊的整体吞吐量
- 为完整的 Danksharding 做准备

### 3. 费用市场

引入了独立的 blob 费用市场：

- Blob gas 与普通 gas 分离
- 动态定价机制
- 目标是每个区块 3 个 blob

## 对 Layer 2 的影响

EIP-4844 将显著降低 Optimistic Rollups 和 ZK Rollups 的运营成本，预计可降低 10-100 倍的数据发布费用。

## 结论

Proto-Danksharding 是以太坊扩容路线图上的关键一步，为未来的完整分片做好准备。
`,
  summary: 'EIP-4844 引入了一种新的交易类型，允许在以太坊上发布 blob 数据，为 Rollup 提供更便宜的数据可用性层。',
  category: { id: 'mock-cat', name: 'Layer 2', slug: 'layer-2' },
  tags: ['EIP', 'Danksharding', 'Rollup', 'Ethereum'],
  sourceUrls: ['https://eips.ethereum.org/EIPS/eip-4844'],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
}

export default function ArticlePage() {
  const params = useParams()
  const slug = params.slug as string

  const { data: article, isLoading, error, isError, refetch } = useQuery({
    queryKey: ['article', slug],
    queryFn: () => articleAPI.get(slug),
    retry: 1,
    staleTime: 30000,
  })

  // Use mock article when API fails
  const displayArticle = isError ? mockArticle : article
  const showingMockData = isError

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
      {showingMockData && (
        <div className="flex items-center gap-2 p-3 mx-6 mt-4 text-sm text-amber-600 bg-amber-50 rounded-lg border border-amber-200">
          <AlertCircle className="w-4 h-4" />
          <span>后端服务未连接，显示示例文章</span>
        </div>
      )}
      <ArticleView article={displayArticle} />
      <FloatingChat
        articleId={displayArticle.id}
        articleTitle={displayArticle.title}
      />
    </MainLayout>
  )
}
