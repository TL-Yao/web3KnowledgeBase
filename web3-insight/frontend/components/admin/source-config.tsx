'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useFeatureFlag } from '@/hooks/use-feature-flag'
import { DisabledFeature } from '@/components/ui/disabled-feature'
import { Plus, Trash2, RefreshCw, CheckCircle, XCircle, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { dataSourceAPI, DataSource, CreateDataSourceRequest } from '@/lib/api'
import { toast } from 'sonner'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

export function SourceConfig() {
  const { isDisabled } = useFeatureFlag('dataSourceManagement')
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [newSource, setNewSource] = useState<CreateDataSourceRequest>({
    name: '',
    type: 'rss',
    url: '',
    fetchInterval: 3600,
  })
  const [validationResult, setValidationResult] = useState<{
    valid?: boolean
    title?: string
    error?: string
  } | null>(null)
  const [isValidating, setIsValidating] = useState(false)

  const queryClient = useQueryClient()

  // All hooks must be called before any conditional returns
  const { data: sources, isLoading } = useQuery({
    queryKey: ['dataSources'],
    queryFn: dataSourceAPI.list,
    enabled: !isDisabled,
  })

  const createMutation = useMutation({
    mutationFn: dataSourceAPI.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dataSources'] })
      setIsDialogOpen(false)
      setNewSource({ name: '', type: 'rss', url: '', fetchInterval: 3600 })
      setValidationResult(null)
      toast.success('数据源创建成功')
    },
    onError: (error: Error) => {
      toast.error('创建失败: ' + error.message)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: dataSourceAPI.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dataSources'] })
      toast.success('数据源已删除')
    },
    onError: (error: Error) => {
      toast.error('删除失败: ' + error.message)
    },
  })

  const syncMutation = useMutation({
    mutationFn: dataSourceAPI.sync,
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['dataSources'] })
      toast.success(`同步完成: 发现 ${result.itemsFound} 条，新增 ${result.itemsNew} 条`)
    },
    onError: (error: Error) => {
      toast.error('同步失败: ' + error.message)
    },
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) => {
      const source = sources?.find((s) => s.id === id)
      if (!source) throw new Error('Source not found')
      return dataSourceAPI.update(id, {
        name: source.name,
        type: source.type,
        url: source.url,
        fetchInterval: source.fetchInterval,
        enabled,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dataSources'] })
    },
  })

  const handleValidateURL = async () => {
    if (!newSource.url) return

    setIsValidating(true)
    try {
      const result = await dataSourceAPI.validate(newSource.url, newSource.type)
      setValidationResult(result)
      if (result.valid && result.title && !newSource.name) {
        setNewSource((prev) => ({ ...prev, name: result.title! }))
      }
    } catch (error) {
      setValidationResult({ valid: false, error: (error as Error).message })
    } finally {
      setIsValidating(false)
    }
  }

  const handleCreate = () => {
    if (!newSource.name || !newSource.url) {
      toast.error('请填写名称和 URL')
      return
    }
    createMutation.mutate(newSource)
  }

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'rss':
        return 'RSS'
      case 'api':
        return 'API'
      case 'crawl':
        return '爬虫'
      default:
        return type
    }
  }

  const formatInterval = (seconds: number) => {
    if (seconds < 3600) return `${Math.round(seconds / 60)} 分钟`
    if (seconds < 86400) return `${Math.round(seconds / 3600)} 小时`
    return `${Math.round(seconds / 86400)} 天`
  }

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

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <Loader2 className="h-6 w-6 animate-spin" />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>数据源配置</CardTitle>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus className="h-4 w-4 mr-2" />
              添加数据源
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>添加数据源</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">类型</label>
                <Select
                  value={newSource.type}
                  onValueChange={(value: 'rss' | 'api' | 'crawl') =>
                    setNewSource((prev) => ({ ...prev, type: value }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="rss">RSS 订阅</SelectItem>
                    <SelectItem value="api">API</SelectItem>
                    <SelectItem value="crawl">网页爬虫</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">URL</label>
                <div className="flex gap-2">
                  <Input
                    value={newSource.url}
                    onChange={(e) => {
                      setNewSource((prev) => ({ ...prev, url: e.target.value }))
                      setValidationResult(null)
                    }}
                    placeholder="https://example.com/feed.xml"
                  />
                  <Button
                    variant="outline"
                    onClick={handleValidateURL}
                    disabled={isValidating || !newSource.url}
                  >
                    {isValidating ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      '验证'
                    )}
                  </Button>
                </div>
                {validationResult && (
                  <div
                    className={`text-sm ${
                      validationResult.valid ? 'text-green-600' : 'text-red-600'
                    }`}
                  >
                    {validationResult.valid ? (
                      <span className="flex items-center gap-1">
                        <CheckCircle className="h-4 w-4" />
                        有效: {validationResult.title}
                      </span>
                    ) : (
                      <span className="flex items-center gap-1">
                        <XCircle className="h-4 w-4" />
                        {validationResult.error}
                      </span>
                    )}
                  </div>
                )}
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">名称</label>
                <Input
                  value={newSource.name}
                  onChange={(e) =>
                    setNewSource((prev) => ({ ...prev, name: e.target.value }))
                  }
                  placeholder="数据源名称"
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">抓取间隔（秒）</label>
                <Input
                  type="number"
                  value={newSource.fetchInterval}
                  onChange={(e) =>
                    setNewSource((prev) => ({
                      ...prev,
                      fetchInterval: parseInt(e.target.value) || 3600,
                    }))
                  }
                  min={60}
                />
                <p className="text-xs text-muted-foreground">
                  当前设置: {formatInterval(newSource.fetchInterval || 3600)}
                </p>
              </div>

              <div className="flex justify-end gap-2 pt-4">
                <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                  取消
                </Button>
                <Button
                  onClick={handleCreate}
                  disabled={createMutation.isPending}
                >
                  {createMutation.isPending ? (
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  ) : null}
                  创建
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </CardHeader>
      <CardContent>
        {!sources?.length ? (
          <p className="text-center text-muted-foreground py-8">
            暂无数据源，点击上方按钮添加
          </p>
        ) : (
          <div className="space-y-3">
            {sources.map((source) => (
              <div
                key={source.id}
                className="flex items-center justify-between p-3 border rounded-lg"
              >
                <div className="flex items-center gap-3">
                  <Switch
                    checked={source.enabled}
                    onCheckedChange={(enabled) =>
                      toggleMutation.mutate({ id: source.id, enabled })
                    }
                  />
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{source.name}</span>
                      <Badge variant="secondary">{getTypeLabel(source.type)}</Badge>
                      {source.lastError && (
                        <Badge variant="destructive">错误</Badge>
                      )}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {source.url.length > 50
                        ? source.url.substring(0, 50) + '...'
                        : source.url}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      间隔: {formatInterval(source.fetchInterval)}
                      {source.lastFetchedAt && (
                        <span>
                          {' · '}上次同步:{' '}
                          {formatDistanceToNow(new Date(source.lastFetchedAt), {
                            addSuffix: true,
                            locale: zhCN,
                          })}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => syncMutation.mutate(source.id)}
                    disabled={syncMutation.isPending}
                  >
                    <RefreshCw
                      className={`h-4 w-4 ${
                        syncMutation.isPending ? 'animate-spin' : ''
                      }`}
                    />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => {
                      if (confirm('确定删除此数据源？')) {
                        deleteMutation.mutate(source.id)
                      }
                    }}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
