'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { MainLayout } from '@/components/layout/main-layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { kbUpdateAPI, KBUpdateJob } from '@/lib/api'
import { AlertCircle, CheckCircle2, Clock, Play, RefreshCw, Loader2, AlertTriangle } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { toast } from 'sonner'

function formatDuration(seconds?: number): string {
  if (!seconds) return '-'
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

export default function KBUpdatePage() {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const pageSize = 20

  // Get keyword stats
  const { data: keywordStats } = useQuery({
    queryKey: ['kb-keyword-stats'],
    queryFn: kbUpdateAPI.getKeywordStats,
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  // Get job history
  const { data: jobsData, isLoading: jobsLoading } = useQuery({
    queryKey: ['kb-jobs', page],
    queryFn: () => kbUpdateAPI.getJobs(page, pageSize),
    refetchInterval: 5000, // Refresh every 5 seconds
  })

  // Check if any job is currently running (more reliable than just checking first item)
  const hasRunningJob = jobsData?.items?.some(job => job.status === 'running') ?? false
  const runningJob = jobsData?.items?.find(job => job.status === 'running')

  // For backward compatibility, keep isRunning
  const isRunning = hasRunningJob

  // Trigger update mutation
  const triggerMutation = useMutation({
    mutationFn: kbUpdateAPI.trigger,
    onSuccess: (data) => {
      toast.success('更新已触发', {
        description: data.message,
      })
      queryClient.invalidateQueries({ queryKey: ['kb-jobs'] })
    },
    onError: (error: Error) => {
      toast.error('触发失败', {
        description: error.message,
      })
    },
  })

  // Initialize keywords mutation
  const initKeywordsMutation = useMutation({
    mutationFn: kbUpdateAPI.initKeywords,
    onSuccess: (data) => {
      toast.success('关键词池初始化成功', {
        description: `已生成 ${data.count} 个关键词`,
      })
      queryClient.invalidateQueries({ queryKey: ['kb-keyword-stats'] })
    },
    onError: (error: Error) => {
      toast.error('初始化失败', {
        description: error.message,
      })
    },
  })

  const handleTriggerUpdate = () => {
    triggerMutation.mutate({ triggerType: 'manual' })
  }

  const handleInitKeywords = () => {
    initKeywordsMutation.mutate({ count: 200 })
  }

  const showInitButton = keywordStats && keywordStats.pending === 0

  return (
    <MainLayout>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-semibold">知识库自动更新</h1>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              queryClient.invalidateQueries({ queryKey: ['kb-jobs'] })
              queryClient.invalidateQueries({ queryKey: ['kb-keyword-stats'] })
            }}
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            刷新
          </Button>
        </div>

        {/* Action Buttons */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">操作</CardTitle>
          </CardHeader>
          <CardContent className="flex gap-4">
            <Button
              onClick={handleTriggerUpdate}
              disabled={isRunning || triggerMutation.isPending}
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

            {showInitButton && (
              <Button
                variant="outline"
                onClick={handleInitKeywords}
                disabled={initKeywordsMutation.isPending}
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

        {/* Keyword Stats */}
        {keywordStats && (
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">关键词池统计</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <div className="text-sm text-muted-foreground">待用</div>
                  <div className="text-2xl font-bold">{keywordStats.pending ?? 0}</div>
                </div>
                <div>
                  <div className="text-sm text-muted-foreground">已用</div>
                  <div className="text-2xl font-bold">{keywordStats.used ?? 0}</div>
                </div>
                <div>
                  <div className="text-sm text-muted-foreground">总计</div>
                  <div className="text-2xl font-bold">{keywordStats.total ?? 0}</div>
                </div>
              </div>
              {keywordStats.pending < 30 && keywordStats.pending > 0 && (
                <div className="mt-4 flex items-center gap-2 text-sm text-amber-600">
                  <AlertTriangle className="w-4 h-4" />
                  <span>待用关键词不足 30 个，建议及时补充</span>
                </div>
              )}
            </CardContent>
          </Card>
        )}

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
                <span className="font-medium">{runningJob.articlesGenerated ?? 0} / 20</span>
              </div>
              {runningJob.articlesPublished > 0 && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">已发布文章</span>
                  <span className="font-medium">{runningJob.articlesPublished ?? 0}</span>
                </div>
              )}
              {runningJob.startTime && (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">开始时间</span>
                  <span className="text-sm">
                    {formatDistanceToNow(new Date(runningJob.startTime), {
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
                          {job.startTime
                            ? new Date(job.startTime).toLocaleString('zh-CN')
                            : '-'}
                        </TableCell>
                        <TableCell>{formatDuration(job.duration)}</TableCell>
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
    </MainLayout>
  )
}
