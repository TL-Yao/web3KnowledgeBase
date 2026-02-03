'use client'

import { useQuery } from '@tanstack/react-query'
import { SystemStatus } from '@/components/admin/system-status'
import { TaskMonitor } from '@/components/admin/task-monitor'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function AdminPage() {
  const { data: stats } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: async () => {
      return {
        todayArticles: 0,
        apiCalls: 0,
        todayCost: 0
      }
    }
  })

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
            <div className="text-3xl font-bold">${stats?.todayCost ?? 0}</div>
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
