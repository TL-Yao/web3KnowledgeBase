'use client'

import { useState } from 'react'
import { Search, Settings } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import Link from 'next/link'

interface HeaderProps {
  breadcrumb?: string
}

export function Header({ breadcrumb = '知识库' }: HeaderProps) {
  const [searchQuery, setSearchQuery] = useState('')

  return (
    <header className="h-14 border-b border-border bg-background flex items-center justify-between px-6">
      {/* Breadcrumb or Title */}
      <div className="text-sm text-muted-foreground">
        {breadcrumb}
      </div>

      {/* Search */}
      <div className="flex items-center gap-4">
        <div className="relative w-64">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="搜索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        <Link href="/admin">
          <Button variant="ghost" size="icon">
            <Settings className="w-4 h-4" />
          </Button>
        </Link>
      </div>
    </header>
  )
}
