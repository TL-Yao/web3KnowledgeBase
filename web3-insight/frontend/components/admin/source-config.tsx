'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useFeatureFlag } from '@/hooks/use-feature-flag'
import { DisabledFeature } from '@/components/ui/disabled-feature'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Plus, Trash2, RefreshCw, ExternalLink } from 'lucide-react'

interface Source {
  id: string
  name: string
  type: 'rss' | 'api' | 'crawler'
  url: string
  enabled: boolean
  lastSync?: string
  status: 'active' | 'error' | 'pending'
}

export function SourceConfig() {
  const { isDisabled } = useFeatureFlag('dataSourceManagement')
  const queryClient = useQueryClient()

  // All hooks must be called before any conditional returns
  const { data: sources, isLoading } = useQuery<Source[]>({
    queryKey: ['sources'],
    queryFn: async () => {
      const response = await fetch('/api/datasources')
      if (!response.ok) throw new Error('Failed to fetch')
      return response.json()
    },
    enabled: !isDisabled,
  })

  const syncMutation = useMutation({
    mutationFn: async (sourceId: string) => {
      // TODO: Call sync API
      await new Promise(resolve => setTimeout(resolve, 1000))
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sources'] })
    }
  })

  const deleteMutation = useMutation({
    mutationFn: async (sourceId: string) => {
      // TODO: Call delete API
      await new Promise(resolve => setTimeout(resolve, 500))
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sources'] })
    }
  })

  // Feature flag check AFTER all hooks
  if (isDisabled) {
    return (
      <DisabledFeature
        featureName="数据源管理"
        description="后端 API 正在开发中"
        variant="card"
      />
    )
  }

  if (isLoading) return <div>加载中...</div>

  const typeLabels = {
    rss: 'RSS',
    api: 'API',
    crawler: '爬虫'
  }

  const statusColors = {
    active: 'bg-green-500',
    error: 'bg-red-500',
    pending: 'bg-yellow-500'
  }

  return (
    <div className="space-y-6">
      {/* Add New Source */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">添加数据源</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <Input placeholder="数据源名称" className="w-48" />
            <Input placeholder="URL" className="flex-1" />
            <Button>
              <Plus className="w-4 h-4 mr-2" />
              添加
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Source List */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">已配置数据源</CardTitle>
            <Button variant="outline" size="sm">
              <RefreshCw className="w-4 h-4 mr-2" />
              全部同步
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="border rounded-lg divide-y">
            {sources?.map((source) => (
              <div key={source.id} className="flex items-center justify-between p-4">
                <div className="flex items-center gap-4">
                  <Switch defaultChecked={source.enabled} />
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{source.name}</span>
                      <Badge variant="outline">{typeLabels[source.type]}</Badge>
                      <div className={`w-2 h-2 rounded-full ${statusColors[source.status]}`} />
                    </div>
                    <a
                      href={source.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-muted-foreground hover:text-primary flex items-center gap-1"
                    >
                      {source.url}
                      <ExternalLink className="w-3 h-3" />
                    </a>
                  </div>
                </div>

                <div className="flex items-center gap-4">
                  {source.lastSync && (
                    <span className="text-sm text-muted-foreground">
                      上次同步: {source.lastSync}
                    </span>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => syncMutation.mutate(source.id)}
                    disabled={syncMutation.isPending}
                  >
                    <RefreshCw className={`w-4 h-4 ${syncMutation.isPending ? 'animate-spin' : ''}`} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-red-500 hover:text-red-600"
                    onClick={() => deleteMutation.mutate(source.id)}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Sync Schedule */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">同步计划</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <span className="text-sm w-32">RSS 同步间隔</span>
            <Input type="number" defaultValue={60} className="w-24" />
            <span className="text-sm text-muted-foreground">分钟</span>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm w-32">爬虫运行时间</span>
            <Input type="text" defaultValue="02:00" className="w-24" />
            <span className="text-sm text-muted-foreground">每日 (UTC+8)</span>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm w-32">最大并发数</span>
            <Input type="number" defaultValue={5} className="w-24" />
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
