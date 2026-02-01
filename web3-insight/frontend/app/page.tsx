'use client'

import { MainLayout } from '@/components/layout/main-layout'
import { Card, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Search, BookOpen, Sparkles, Settings, ArrowRight } from 'lucide-react'
import Link from 'next/link'
import { useState } from 'react'
import { useRouter } from 'next/navigation'

export default function HomePage() {
  const [searchQuery, setSearchQuery] = useState('')
  const router = useRouter()

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (searchQuery.trim()) {
      router.push(`/knowledge?q=${encodeURIComponent(searchQuery)}`)
    }
  }

  const quickAccess = [
    {
      title: '知识库',
      description: '浏览和管理你的 Web3 知识文章',
      icon: BookOpen,
      href: '/knowledge',
      color: 'text-blue-500',
      bgColor: 'bg-blue-500/10',
    },
    {
      title: '即时研究',
      description: '输入问题，AI 帮你深入研究',
      icon: Sparkles,
      href: '/research',
      color: 'text-purple-500',
      bgColor: 'bg-purple-500/10',
    },
    {
      title: '系统设置',
      description: '配置模型和数据源',
      icon: Settings,
      href: '/admin/config',
      color: 'text-gray-500',
      bgColor: 'bg-gray-500/10',
    },
  ]

  return (
    <MainLayout>
      <div className="flex-1 flex flex-col">
        {/* Hero Section */}
        <div className="flex-1 flex flex-col items-center justify-center p-8 max-w-4xl mx-auto w-full">
          <div className="text-center mb-8">
            <h1 className="text-4xl font-bold tracking-tight mb-4">
              Web3 Insight
            </h1>
            <p className="text-lg text-muted-foreground max-w-2xl">
              你的私人 Web3 知识管理系统，支持 AI 驱动的研究和智能问答
            </p>
          </div>

          {/* Search Box */}
          <form onSubmit={handleSearch} className="w-full max-w-2xl mb-12">
            <div className="relative">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
              <Input
                type="text"
                placeholder="搜索知识库或输入问题..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-12 pr-24 h-14 text-lg rounded-full border-2 focus-visible:ring-2"
              />
              <Button
                type="submit"
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded-full"
                size="lg"
              >
                搜索
              </Button>
            </div>
          </form>

          {/* Quick Access Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full max-w-4xl">
            {quickAccess.map((item) => (
              <Link key={item.href} href={item.href}>
                <Card className="h-full hover:shadow-lg transition-shadow cursor-pointer group">
                  <CardHeader>
                    <div className={`w-12 h-12 rounded-lg ${item.bgColor} flex items-center justify-center mb-2`}>
                      <item.icon className={`h-6 w-6 ${item.color}`} />
                    </div>
                    <CardTitle className="flex items-center gap-2">
                      {item.title}
                      <ArrowRight className="h-4 w-4 opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0 transition-all" />
                    </CardTitle>
                    <CardDescription>{item.description}</CardDescription>
                  </CardHeader>
                </Card>
              </Link>
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="text-center py-6 text-sm text-muted-foreground border-t">
          <p>Web3 Insight - 本地部署的私人知识管理系统</p>
        </div>
      </div>
    </MainLayout>
  )
}
