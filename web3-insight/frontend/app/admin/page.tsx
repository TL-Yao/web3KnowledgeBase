'use client'

import { useQuery } from '@tanstack/react-query'
import { SystemStatus } from '@/components/admin/system-status'
import { TaskMonitor } from '@/components/admin/task-monitor'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { AlertCircle } from 'lucide-react'

interface AdminStats {
  todayArticles: number
  apiCalls: number
  todayCost: number
}

export default function AdminPage() {
  const { data: stats, isLoading, isError } = useQuery<AdminStats>({
    queryKey: ['admin-stats'],
    queryFn: async (): Promise<AdminStats> => {
      // TODO: Replace with real API endpoint when backend is ready
      // return fetchAPI<AdminStats>('/api/admin/stats')
      return {
        todayArticles: 0,
        apiCalls: 0,
        todayCost: 0
      }
    },
    refetchInterval: 30000,
    staleTime: 30000,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12 text-muted-foreground">
        <div className="animate-pulse">加载中...</div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <AlertCircle className="w-12 h-12 mb-4 text-red-400" />
        <p className="text-lg font-medium text-red-600">无法加载统计数据</p>
        <p className="text-sm">请检查服务状态后重试</p>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-semibold">系统概览</h1>

      <SystemStatus />

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              今日新文章
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.todayArticles ?? 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              API 调用次数
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats?.apiCalls ?? 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              今日成本
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">${(stats?.todayCost ?? 0).toFixed(2)}</div>
          </CardContent>
        </Card>
      </div>

      {/* Task Queue */}
      <Card>
        <CardHeader>
          <CardTitle>当前任务队列</CardTitle>
        </CardHeader>
        <CardContent>
          <TaskMonitor />
        </CardContent>
      </Card>
    </div>
  )
}
