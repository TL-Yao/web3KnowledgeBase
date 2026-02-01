'use client'

import { useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'
import { Search, Save, Loader2 } from 'lucide-react'

export function ResearchPanel() {
  const [query, setQuery] = useState('')
  const [result, setResult] = useState<any>(null)
  const [selectedCategory, setSelectedCategory] = useState<string>('')

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: async () => {
      // TODO: Fetch from API
      return [
        { id: '1', name: 'Layer 1' },
        { id: '2', name: 'Layer 2' },
        { id: '3', name: 'DeFi' },
        { id: '4', name: 'NFT' },
        { id: '5', name: '共识机制' },
      ]
    }
  })

  const researchMutation = useMutation({
    mutationFn: async (topic: string) => {
      // TODO: Call actual API
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 2000))

      // Mock response
      return {
        title: `${topic} 技术详解`,
        content: `这是关于 ${topic} 的详细技术分析...`,
        contentHtml: `<h2>${topic} 简介</h2><p>这是关于 ${topic} 的详细技术分析，包括其核心原理、实现机制和应用场景。</p><h2>核心原理</h2><p>${topic} 的核心原理基于...</p><h2>应用场景</h2><p>${topic} 在 Web3 生态中有广泛的应用...</p>`,
        suggestedTags: [topic, 'Web3', '区块链'],
        suggestedCategory: { id: '2', name: 'Layer 2' },
        sources: ['https://ethereum.org/docs', 'https://docs.arbitrum.io'],
        suggestedNewCategory: null
      }
    },
    onSuccess: (data) => {
      setResult(data)
      if (data.suggestedCategory) {
        setSelectedCategory(data.suggestedCategory.id)
      }
    }
  })

  const saveMutation = useMutation({
    mutationFn: async () => {
      const res = await fetch('/api/articles', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: result.title,
          content: result.content,
          categoryId: selectedCategory,
          tags: result.suggestedTags,
          sourceUrls: result.sources
        })
      })
      return res.json()
    }
  })

  const handleSearch = () => {
    if (!query.trim()) return
    researchMutation.mutate(query)
  }

  return (
    <div className="space-y-6">
      {/* Search */}
      <div className="flex gap-2">
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="例如: zkPorter, EIP-4844, Cosmos IBC..."
          className="flex-1"
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
        />
        <Button onClick={handleSearch} disabled={researchMutation.isPending}>
          {researchMutation.isPending ? (
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
          ) : (
            <Search className="w-4 h-4 mr-2" />
          )}
          研究
        </Button>
      </div>

      {/* Loading State */}
      {researchMutation.isPending && (
        <Card>
          <CardContent className="py-12">
            <div className="flex flex-col items-center justify-center gap-4">
              <Loader2 className="w-8 h-8 animate-spin text-primary" />
              <p className="text-muted-foreground">正在研究 "{query}"...</p>
              <p className="text-sm text-muted-foreground">AI 正在搜索、分析并生成内容</p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Result */}
      {result && !researchMutation.isPending && (
        <Card>
          <CardHeader>
            <CardTitle>{result.title}</CardTitle>
            <div className="flex items-center gap-2 mt-2">
              {result.suggestedTags?.map((tag: string) => (
                <Badge key={tag} variant="secondary">{tag}</Badge>
              ))}
            </div>
          </CardHeader>
          <CardContent>
            <article
              className="prose prose-neutral max-w-none dark:prose-invert"
              dangerouslySetInnerHTML={{ __html: result.contentHtml }}
            />

            {/* Sources */}
            {result.sources?.length > 0 && (
              <div className="mt-6 pt-4 border-t border-border">
                <p className="text-sm text-muted-foreground mb-2">参考来源:</p>
                <ul className="text-sm space-y-1">
                  {result.sources.map((url: string) => (
                    <li key={url}>
                      <a href={url} target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
                        {url}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* Save */}
            <div className="mt-6 pt-6 border-t border-border">
              <div className="flex items-center gap-4">
                <span className="text-sm">保存到分类:</span>
                <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                  <SelectTrigger className="w-[200px]">
                    <SelectValue placeholder="选择分类" />
                  </SelectTrigger>
                  <SelectContent>
                    {categories?.map((cat: any) => (
                      <SelectItem key={cat.id} value={cat.id}>
                        {cat.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button
                  onClick={() => saveMutation.mutate()}
                  disabled={!selectedCategory || saveMutation.isPending}
                >
                  <Save className="w-4 h-4 mr-2" />
                  保存到知识库
                </Button>
              </div>

              {result.suggestedNewCategory && (
                <div className="mt-3 p-3 bg-muted rounded-lg text-sm">
                  💡 建议创建新分类: <strong>{result.suggestedNewCategory.name}</strong>
                  <Button variant="link" size="sm" className="ml-2">
                    创建并保存
                  </Button>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
