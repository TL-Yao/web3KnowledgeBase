'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Search, Loader2, Tag as TagIcon } from 'lucide-react'
import { tagAPI } from '@/lib/api'

interface TagSidebarProps {
  activeTag: string
  onTagClick: (tag: string) => void
}

export function TagSidebar({ activeTag, onTagClick }: TagSidebarProps) {
  const [searchQuery, setSearchQuery] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['tags', 'in-use'],
    queryFn: tagAPI.getInUse,
    staleTime: 30000,
  })

  const allTags = data?.tags ?? []

  const filteredTags = searchQuery
    ? allTags.filter((t) => t.name.includes(searchQuery))
    : allTags

  const popularTags = filteredTags.filter((t) => t.articleCount >= 3)
  const restTags = filteredTags.filter((t) => t.articleCount < 3)

  const handleClick = (tagName: string) => {
    onTagClick(tagName === activeTag ? '' : tagName)
  }

  return (
    <div className="flex flex-col h-full">
      <div className="p-4 border-b border-border flex items-center justify-between">
        <h2 className="font-semibold text-lg">标签筛选</h2>
        {activeTag && (
          <Button
            variant="ghost"
            size="sm"
            className="text-xs h-7"
            onClick={() => onTagClick('')}
          >
            清除
          </Button>
        )}
      </div>

      <div className="overflow-y-auto flex-1 p-4 space-y-4">
        {allTags.length >= 20 && (
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
            <Input
              placeholder="搜索标签..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8 h-8 text-sm"
            />
          </div>
        )}

        {isLoading ? (
          <div className="flex items-center justify-center py-8 text-muted-foreground">
            <Loader2 className="w-5 h-5 animate-spin mr-2" />
            <span className="text-sm">加载中...</span>
          </div>
        ) : filteredTags.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
            <TagIcon className="w-8 h-8 mb-2 opacity-50" />
            <p className="text-sm">暂无标签</p>
          </div>
        ) : (
          <>
            {popularTags.length > 0 && (
              <div>
                <p className="text-xs text-muted-foreground mb-2">热门</p>
                <div className="flex flex-wrap gap-1.5">
                  {popularTags.map((tag) => (
                    <Badge
                      key={tag.name}
                      variant={tag.name === activeTag ? 'default' : 'secondary'}
                      className="text-xs cursor-pointer hover:bg-primary/20 transition-colors"
                      onClick={() => handleClick(tag.name)}
                    >
                      {tag.name}
                      <span className="ml-1 opacity-60">{tag.articleCount}</span>
                    </Badge>
                  ))}
                </div>
              </div>
            )}

            {restTags.length > 0 && (
              <div>
                {popularTags.length > 0 && (
                  <p className="text-xs text-muted-foreground mb-2">全部</p>
                )}
                <div className="flex flex-wrap gap-1.5">
                  {restTags.map((tag) => (
                    <Badge
                      key={tag.name}
                      variant={tag.name === activeTag ? 'default' : 'secondary'}
                      className="text-xs cursor-pointer hover:bg-primary/20 transition-colors"
                      onClick={() => handleClick(tag.name)}
                    >
                      {tag.name}
                      <span className="ml-1 opacity-60">{tag.articleCount}</span>
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
