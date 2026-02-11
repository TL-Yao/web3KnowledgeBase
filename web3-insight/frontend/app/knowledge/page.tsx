'use client'

import { Suspense, useState } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { MainLayout } from '@/components/layout/main-layout'
import { ArticleList } from '@/components/knowledge/article-list'
import { TagSidebar } from '@/components/knowledge/tag-sidebar'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Search, X } from 'lucide-react'

function KnowledgeContent() {
  const router = useRouter()
  const searchParams = useSearchParams()

  const activeTag = searchParams.get('tag') || ''
  const [searchQuery, setSearchQuery] = useState(searchParams.get('q') || '')

  const handleTagClick = (tag: string) => {
    const params = new URLSearchParams(searchParams.toString())
    if (tag) {
      params.set('tag', tag)
    } else {
      params.delete('tag')
    }
    router.push(`/knowledge?${params.toString()}`)
  }

  const handleSearch = () => {
    const params = new URLSearchParams(searchParams.toString())
    if (searchQuery.trim()) {
      params.set('q', searchQuery.trim())
    } else {
      params.delete('q')
    }
    router.push(`/knowledge?${params.toString()}`)
  }

  const handleClearSearch = () => {
    setSearchQuery('')
    const params = new URLSearchParams(searchParams.toString())
    params.delete('q')
    router.push(`/knowledge?${params.toString()}`)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  return (
    <div className="flex h-[calc(100vh-64px)]">
      {/* Left Sidebar - Tag Navigation - Hidden on mobile */}
      <div className="hidden md:block w-64 border-r border-border bg-card">
        <TagSidebar activeTag={activeTag} onTagClick={handleTagClick} />
      </div>

      {/* Right Content - Articles */}
      <div className="flex-1 flex flex-col w-full md:w-auto">
        {/* Search Bar */}
        <div className="p-4 border-b border-border bg-card">
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                placeholder="搜索文章标题、内容..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={handleKeyDown}
                className="pl-9 pr-9"
              />
              {searchQuery && (
                <button
                  onClick={handleClearSearch}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>
            <Button onClick={handleSearch}>搜索</Button>
          </div>
        </div>

        {/* Article List */}
        <div className="flex-1 overflow-hidden p-6">
          <ArticleList
            searchQuery={searchParams.get('q') || ''}
            activeTag={activeTag}
            onTagClick={handleTagClick}
          />
        </div>
      </div>
    </div>
  )
}

export default function KnowledgePage() {
  return (
    <MainLayout>
      <Suspense fallback={
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
            <p className="text-muted-foreground">加载中...</p>
          </div>
        </div>
      }>
        <KnowledgeContent />
      </Suspense>
    </MainLayout>
  )
}
