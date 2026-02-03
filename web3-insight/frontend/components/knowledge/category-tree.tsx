'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronRight, Folder, FolderOpen } from 'lucide-react'
import { cn } from '@/lib/utils'
import { categoryAPI, Category } from '@/lib/api'

// Constants for tree indentation
const INDENT_PER_LEVEL = 12
const BASE_PADDING = 8

interface CategoryNodeProps {
  category: Category
  level: number
}

function CategoryNode({ category, level }: CategoryNodeProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const hasChildren = category.children && category.children.length > 0

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-1 px-2 py-1.5 rounded-md hover:bg-accent/10 cursor-pointer text-sm",
          "transition-colors"
        )}
        style={{ paddingLeft: `${level * INDENT_PER_LEVEL + BASE_PADDING}px` }}
        onClick={() => {
          if (hasChildren) {
            setIsExpanded(!isExpanded)
          }
        }}
      >
        {hasChildren ? (
          <ChevronRight
            className={cn(
              "w-4 h-4 text-muted-foreground transition-transform",
              isExpanded && "rotate-90"
            )}
          />
        ) : (
          <span className="w-4" />
        )}
        {hasChildren ? (
          isExpanded ? (
            <FolderOpen className="w-4 h-4 text-accent" />
          ) : (
            <Folder className="w-4 h-4 text-muted-foreground" />
          )
        ) : (
          <span className="w-4 h-4 flex items-center justify-center text-muted-foreground">
            &bull;
          </span>
        )}
        <span className="truncate">{category.name}</span>
      </div>
      {hasChildren && isExpanded && (
        <div>
          {(category.children || []).map((child) => (
            <CategoryNode key={child.id} category={child} level={level + 1} />
          ))}
        </div>
      )}
    </div>
  )
}

export function CategoryTree() {
  const { data: categories, isLoading, error } = useQuery({
    queryKey: ['categories', 'tree'],
    queryFn: categoryAPI.getTree,
    retry: 1,
    staleTime: 30000,
  })

  if (isLoading) {
    return (
      <div className="py-4 px-2 text-sm text-muted-foreground">
        加载分类中...
      </div>
    )
  }

  if (error) {
    return (
      <div className="py-4 px-2 text-sm text-destructive">
        加载分类失败，请稍后重试
      </div>
    )
  }

  if (!categories || categories.length === 0) {
    return (
      <div className="py-4 px-2 text-sm text-muted-foreground">
        暂无分类
      </div>
    )
  }

  return (
    <div className="py-1">
      {categories.map((category) => (
        <CategoryNode key={category.id} category={category} level={0} />
      ))}
    </div>
  )
}
