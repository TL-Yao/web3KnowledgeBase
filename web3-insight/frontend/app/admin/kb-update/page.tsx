'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { kbUpdateAPI, KBUpdateJob, Theme } from '@/lib/api'
import {
  AlertCircle,
  CheckCircle2,
  Clock,
  Play,
  RefreshCw,
  Loader2,
  AlertTriangle,
  BookOpen,
  Zap,
  History,
  Building2,
  Package,
  Minus,
  Plus,
} from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { toast } from 'sonner'

function formatDuration(startedAt?: string, completedAt?: string): string {
  if (!startedAt || !completedAt) return '-'
  const seconds = Math.round((new Date(completedAt).getTime() - new Date(startedAt).getTime()) / 1000)
  if (seconds < 0) return '-'
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${minutes}分${secs}秒`
}

function getStatusBadge(status: KBUpdateJob['status']) {
  switch (status) {
    case 'running':
      return <Badge className="bg-blue-500"><Loader2 className="w-3 h-3 mr-1 animate-spin" />运行中</Badge>
    case 'completed':
      return <Badge className="bg-green-500"><CheckCircle2 className="w-3 h-3 mr-1" />已完成</Badge>
    case 'failed':
      return <Badge variant="destructive"><AlertCircle className="w-3 h-3 mr-1" />失败</Badge>
    case 'pending':
      return <Badge variant="secondary"><Clock className="w-3 h-3 mr-1" />等待中</Badge>
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}

const categoryIcons: Record<string, React.ReactNode> = {
  '基础知识': <BookOpen className="w-4 h-4" />,
  '进阶': <Zap className="w-4 h-4" />,
  '历史事件': <History className="w-4 h-4" />,
  '产品介绍': <Package className="w-4 h-4" />,
  '公司介绍': <Building2 className="w-4 h-4" />,
}

function ThemeCard({ theme, isActive, onActivate, isActivating }: {
  theme: Theme
  isActive: boolean
  onActivate: () => void
  isActivating: boolean
}) {
  const isEmpty = (theme.keywordsTotal ?? 0) === 0
  return (
    <div
      className={`relative rounded-lg border p-4 transition-colors ${
        isActive
          ? 'border-primary bg-primary/5 ring-1 ring-primary'
          : isEmpty
            ? 'border-border opacity-60 hover:opacity-80 hover:border-primary/40 cursor-pointer'
            : 'border-border hover:border-primary/40 cursor-pointer'
      }`}
      onClick={() => !isActive && !isActivating && onActivate()}
    >
      <div className="flex items-start justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className="font-medium text-sm">{theme.name}</span>
          {isActive && (
            <Badge className="bg-primary text-primary-foreground text-xs px-1.5 py-0">
              当前
            </Badge>
          )}
          {isEmpty && (
            <Badge variant="outline" className="text-xs px-1.5 py-0 text-muted-foreground">
              未初始化
            </Badge>
          )}
        </div>
      </div>
      <p className="text-xs text-muted-foreground mb-3 line-clamp-2">{theme.description}</p>
      <div className="flex gap-3 text-xs text-muted-foreground">
        <span>待用 <span className="font-medium text-foreground">{theme.keywordsPending ?? 0}</span></span>
        <span>已用 <span className="font-medium text-foreground">{theme.keywordsUsed ?? 0}</span></span>
        <span>总计 <span className="font-medium text-foreground">{theme.keywordsTotal ?? 0}</span></span>
      </div>
      {isActivating && !isActive && (
        <div className="absolute inset-0 bg-background/60 rounded-lg flex items-center justify-center">
          <Loader2 className="w-4 h-4 animate-spin" />
        </div>
      )}
    </div>
  )
}

export default function KBUpdatePage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [batchSizeInput, setBatchSizeInput] = useState<number | null>(null)
  const pageSize = 20

  // Get themes
  const { data: themesData } = useQuery({
    queryKey: ['kb-themes'],
    queryFn: kbUpdateAPI.getThemes,
    refetchInterval: 30000,
  })

  // Get KB config
  const { data: kbConfig } = useQuery({
    queryKey: ['kb-config'],
    queryFn: kbUpdateAPI.getConfig,
  })

  // Get keyword stats
  const { data: keywordStats } = useQuery({
    queryKey: ['kb-keyword-stats'],
    queryFn: kbUpdateAPI.getKeywordStats,
    refetchInterval: 30000,
  })

  // Get job history
  const { data: jobsData, isLoading: jobsLoading } = useQuery({
    queryKey: ['kb-jobs', page],
    queryFn: () => kbUpdateAPI.getJobs(page, pageSize),
    refetchInterval: 5000,
  })

  const hasRunningJob = jobsData?.items?.some(job => job.status === 'running') ?? false
  const runningJob = jobsData?.items?.find(job => job.status === 'running')
  const isRunning = hasRunningJob

  // Activate theme mutation
  const activateThemeMutation = useMutation({
    mutationFn: kbUpdateAPI.setActiveTheme,
    onSuccess: () => {
      toast.success('主题已切换')
      queryClient.invalidateQueries({ queryKey: ['kb-themes'] })
      queryClient.invalidateQueries({ queryKey: ['kb-keyword-stats'] })
    },
    onError: (error: Error) => {
      toast.error('切换失败', { description: error.message })
    },
  })

  // Trigger update mutation
  const triggerMutation = useMutation({
    mutationFn: kbUpdateAPI.trigger,
    onSuccess: (data) => {
      toast.success('更新已触发', { description: data.message })
      queryClient.invalidateQueries({ queryKey: ['kb-jobs'] })
    },
    onError: (error: Error) => {
      toast.error('触发失败', { description: error.message })
    },
  })

  // Initialize keywords mutation
  const initKeywordsMutation = useMutation({
    mutationFn: kbUpdateAPI.initKeywords,
    onSuccess: (data) => {
      toast.success('关键词池初始化成功', { description: `已生成 ${data.count} 个关键词` })
      queryClient.invalidateQueries({ queryKey: ['kb-keyword-stats'] })
      queryClient.invalidateQueries({ queryKey: ['kb-themes'] })
    },
    onError: (error: Error) => {
      toast.error('初始化失败', { description: error.message })
    },
  })

  // Update batch size mutation
  const updateBatchSizeMutation = useMutation({
    mutationFn: kbUpdateAPI.updateBatchSize,
    onSuccess: () => {
      toast.success('批量大小已更新')
      queryClient.invalidateQueries({ queryKey: ['kb-config'] })
      setBatchSizeInput(null)
    },
    onError: (error: Error) => {
      toast.error('更新失败', { description: error.message })
    },
  })

  const handleTriggerUpdate = () => {
    triggerMutation.mutate({ triggerType: 'manual' })
  }

  const handleInitKeywords = () => {
    initKeywordsMutation.mutate({ count: 400 })
  }

  const currentBatchSize = batchSizeInput ?? kbConfig?.batchSize ?? 3
  const maxBatchSize = kbConfig?.maxBatchSize ?? 10

  const handleBatchSizeChange = (delta: number) => {
    const newSize = Math.max(1, Math.min(maxBatchSize, currentBatchSize + delta))
    setBatchSizeInput(newSize)
  }

  const handleSaveBatchSize = () => {
    if (batchSizeInput !== null) {
      updateBatchSizeMutation.mutate(batchSizeInput)
    }
  }

  // Group themes by category
  const themesByCategory = (themesData?.themes ?? []).reduce<Record<string, Theme[]>>((acc, theme) => {
    const cat = theme.category
    if (!acc[cat]) acc[cat] = []
    acc[cat].push(theme)
    return acc
  }, {})

  const activeThemeId = themesData?.activeThemeId
  const activeTheme = themesData?.themes?.find(t => t.id === activeThemeId)

  const showInitButton = keywordStats && keywordStats.pending === 0

  return (
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold">知识库自动更新</h1>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              queryClient.invalidateQueries({ queryKey: ['kb-jobs'] })
              queryClient.invalidateQueries({ queryKey: ['kb-keyword-stats'] })
              queryClient.invalidateQueries({ queryKey: ['kb-themes'] })
              queryClient.invalidateQueries({ queryKey: ['kb-config'] })
            }}
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
        </div>

        {/* Theme Selector */}
        {Object.keys(themesByCategory).length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">内容主题</CardTitle>
            </CardHeader>
            <CardContent className="space-y-5">
              {Object.entries(themesByCategory).map(([category, themes]) => (
                <div key={category}>
                  <div className="flex items-center gap-2 mb-3 text-sm font-medium text-muted-foreground">
                    {categoryIcons[category] || <BookOpen className="w-4 h-4" />}
                    <span>{category}</span>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                    {themes.map((theme) => (
                      <ThemeCard
                        key={theme.id}
                        theme={theme}
                        isActive={theme.id === activeThemeId}
                        onActivate={() => activateThemeMutation.mutate(theme.id)}
                        isActivating={activateThemeMutation.isPending}
                      />
                    ))}
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        )}

        {/* Action & Config */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Action Buttons */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">操作</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Button
                onClick={handleTriggerUpdate}
                disabled={isRunning || triggerMutation.isPending || (themesData && !activeThemeId)}
                className="w-full"
              >
                {triggerMutation.isPending ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    触发中...
                  </>
                ) : isRunning ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    任务运行中 - 请等待
                  </>
                ) : (
                  <>
                    <Play className="w-4 h-4 mr-2" />
                    立即更新知识库
                  </>
                )}
              </Button>
              {activeTheme ? (
                <p className="text-xs text-muted-foreground text-center">
                  当前主题: {activeTheme.name}
                </p>
              ) : themesData && !activeThemeId ? (
                <div className="flex items-center gap-2 text-xs text-amber-600 justify-center">
                  <AlertTriangle className="w-3 h-3" />
                  <span>请先在上方选择一个内容主题</span>
                </div>
              ) : null}

              {showInitButton && (
                <Button
                  variant="outline"
                  onClick={handleInitKeywords}
                  disabled={initKeywordsMutation.isPending}
                  className="w-full"
                >
                  {initKeywordsMutation.isPending ? (
                    <>
                      <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                      初始化中...
                    </>
                  ) : (
                    '初始化关键词池'
                  )}
                </Button>
              )}
            </CardContent>
          </Card>

          {/* Batch Size Config + Keyword Stats */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">配置</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Batch Size */}
              <div className="space-y-2">
                <Label className="text-sm">每次生成文章数</Label>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => handleBatchSizeChange(-1)}
                    disabled={currentBatchSize <= 1}
                  >
                    <Minus className="w-3 h-3" />
                  </Button>
                  <Input
                    type="number"
                    value={currentBatchSize}
                    onChange={(e) => {
                      const val = parseInt(e.target.value)
                      if (!isNaN(val) && val >= 1 && val <= maxBatchSize) {
                        setBatchSizeInput(val)
                      }
                    }}
                    className="w-16 h-8 text-center"
                    min={1}
                    max={maxBatchSize}
                  />
                  <Button
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => handleBatchSizeChange(1)}
                    disabled={currentBatchSize >= maxBatchSize}
                  >
                    <Plus className="w-3 h-3" />
                  </Button>
                  {batchSizeInput !== null && batchSizeInput !== (kbConfig?.batchSize ?? 3) && (
                    <Button
                      size="sm"
                      className="h-8"
                      onClick={handleSaveBatchSize}
                      disabled={updateBatchSizeMutation.isPending}
                    >
                      {updateBatchSizeMutation.isPending ? (
                        <Loader2 className="w-3 h-3 animate-spin" />
                      ) : '保存'}
                    </Button>
                  )}
                </div>
                <p className="text-xs text-muted-foreground">范围: 1-{maxBatchSize} 篇</p>
              </div>

              {/* Keyword Stats */}
              {keywordStats && (
                <div className="pt-3 border-t">
                  <Label className="text-sm mb-2 block">关键词池</Label>
                  <div className="grid grid-cols-3 gap-3">
                    <div>
                      <div className="text-xs text-muted-foreground">待用</div>
                      <div className="text-lg font-bold">{keywordStats.pending ?? 0}</div>
                    </div>
                    <div>
                      <div className="text-xs text-muted-foreground">已用</div>
                      <div className="text-lg font-bold">{keywordStats.used ?? 0}</div>
                    </div>
                    <div>
                      <div className="text-xs text-muted-foreground">总计</div>
                      <div className="text-lg font-bold">{keywordStats.total ?? 0}</div>
                    </div>
                  </div>
                  {keywordStats.pending < 30 && keywordStats.pending > 0 && (
                    <div className="mt-2 flex items-center gap-2 text-xs text-amber-600">
                      <AlertTriangle className="w-3 h-3" />
                      <span>待用关键词不足 30 个，建议及时补充</span>
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Current Job Status */}
        {isRunning && runningJob && (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">当前任务</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">状态</span>
                {getStatusBadge(runningJob.status)}
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">已生成文章</span>
                <span className="font-medium">{runningJob.articlesGenerated ?? 0} / {currentBatchSize}</span>
              </div>
              {runningJob.articlesPublished > 0 && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">已发布文章</span>
                  <span className="font-medium">{runningJob.articlesPublished ?? 0}</span>
                </div>
              )}
              {runningJob.startedAt && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">开始时间</span>
                  <span className="text-sm">
                    {formatDistanceToNow(new Date(runningJob.startedAt), {
                      addSuffix: true,
                      locale: zhCN,
                    })}
                  </span>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Job History */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">更新历史</CardTitle>
          </CardHeader>
          <CardContent>
            {jobsLoading ? (
              <div className="flex items-center justify-center py-8 text-muted-foreground">
                <Loader2 className="w-6 h-6 animate-spin mr-2" />
                加载中...
              </div>
            ) : jobsData?.items && jobsData.items.length > 0 ? (
              <>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>状态</TableHead>
                      <TableHead>触发方式</TableHead>
                      <TableHead>文章数</TableHead>
                      <TableHead>开始时间</TableHead>
                      <TableHead>耗时</TableHead>
                      <TableHead>错误</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {jobsData.items.map((job) => (
                      <TableRow key={job.id}>
                        <TableCell>{getStatusBadge(job.status)}</TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {job.triggerType === 'manual' ? '手动' : '定时'}
                          </Badge>
                        </TableCell>
                        <TableCell>{job.articlesGenerated ?? 0}</TableCell>
                        <TableCell>
                          {job.startedAt
                            ? new Date(job.startedAt).toLocaleString('zh-CN')
                            : '-'}
                        </TableCell>
                        <TableCell>{formatDuration(job.startedAt, job.completedAt)}</TableCell>
                        <TableCell>
                          {job.errorMessage ? (
                            <TooltipProvider>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <span className="text-xs text-destructive cursor-help truncate max-w-[120px] inline-block">
                                    {job.errorMessage.slice(0, 30)}...
                                  </span>
                                </TooltipTrigger>
                                <TooltipContent side="left" className="max-w-sm">
                                  <p className="text-xs whitespace-pre-wrap">{job.errorMessage}</p>
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                          ) : (
                            <span className="text-xs text-muted-foreground">-</span>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>

                {/* Pagination */}
                {jobsData.total > pageSize && (
                  <div className="flex items-center justify-between mt-4">
                    <div className="text-sm text-muted-foreground">
                      共 {jobsData.total} 条记录
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={page === 1}
                        onClick={() => setPage(page - 1)}
                      >
                        上一页
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={page * pageSize >= jobsData.total}
                        onClick={() => setPage(page + 1)}
                      >
                        下一页
                      </Button>
                    </div>
                  </div>
                )}
              </>
            ) : (
              <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                <Clock className="w-12 h-12 mb-4 opacity-50" />
                <p className="text-lg font-medium">暂无更新记录</p>
                <p className="text-sm">点击上方按钮开始第一次更新</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
  )
}
