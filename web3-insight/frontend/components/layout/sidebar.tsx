'use client'

import Link from 'next/link'
import { cn } from '@/lib/utils'
import { FileText, Newspaper, Search, Settings } from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { CategoryTree } from '@/components/knowledge/category-tree'

interface SidebarProps {
  className?: string
}

export function Sidebar({ className }: SidebarProps) {
  return (
    <aside className={cn(
      "w-64 border-r border-border bg-background flex flex-col",
      className
    )}>
      {/* Logo */}
      <div className="h-14 flex items-center px-4 border-b border-border">
        <Link href="/" className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-accent flex items-center justify-center">
            <span className="text-white font-bold text-sm">W3</span>
          </div>
          <span className="font-semibold">Web3 Insight</span>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="p-2">
        <Link href="/knowledge" className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-accent/10 text-sm">
          <FileText className="w-4 h-4" />
          知识库
        </Link>
        <Link href="/news" className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-accent/10 text-sm">
          <Newspaper className="w-4 h-4" />
          新闻
        </Link>
        <Link href="/research" className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-accent/10 text-sm">
          <Search className="w-4 h-4" />
          即时研究
        </Link>
        <Link href="/admin" className="flex items-center gap-2 px-3 py-2 rounded-md hover:bg-accent/10 text-sm">
          <Settings className="w-4 h-4" />
          后台管理
        </Link>
      </nav>

      {/* Category Tree */}
      <div className="flex-1 overflow-hidden">
        <div className="px-4 py-2 text-xs font-medium text-muted-foreground uppercase">
          分类
        </div>
        <ScrollArea className="h-full px-2">
          <CategoryTree />
        </ScrollArea>
      </div>

      {/* Recent */}
      <div className="border-t border-border p-4">
        <div className="text-xs font-medium text-muted-foreground uppercase mb-2">
          最近阅读
        </div>
        <div className="space-y-1 text-sm text-muted-foreground">
          <div className="truncate hover:text-foreground cursor-pointer">
            zkSync 工作原理
          </div>
          <div className="truncate hover:text-foreground cursor-pointer">
            Cosmos IBC 详解
          </div>
        </div>
      </div>
    </aside>
  )
}
