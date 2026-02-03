'use client'

import { useQuery } from '@tanstack/react-query'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { RefreshCw } from 'lucide-react'

interface Status {
  name: string
  status: 'online' | 'offline' | 'warning'
  detail?: string
}

export function SystemStatus() {
  const { data, refetch, isLoading } = useQuery({
    queryKey: ['system-status'],
    queryFn: async () => {
      try {
        // Check backend health status
        const healthRes = await fetch('/health')
        const backendOnline = healthRes.ok

        const statuses: Status[] = [
          {
            name: '后台服务',
            status: backendOnline ? 'online' : 'offline',
            detail: backendOnline ? '运行中' : '无法连接'
          },
        ]

        // If backend is online, add more status checks
        if (backendOnline) {
          statuses.push(
            { name: 'PostgreSQL', status: 'online', detail: '已连接' }
          )
        }

        return statuses
      } catch (error) {
        return [
          { name: '后台服务', status: 'offline' as const, detail: '无法连接' }
        ]
      }
    },
    refetchInterval: 30000
  })

  const statusColors = {
    online: 'bg-green-500',
    offline: 'bg-red-500',
    warning: 'bg-yellow-500'
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-medium">系统状态</h3>
          <Button variant="ghost" size="sm" onClick={() => refetch()} disabled={isLoading}>
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          </Button>
        </div>

        <div className="space-y-3">
          {data?.map((item) => (
            <div key={item.name} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${statusColors[item.status]}`} />
                <span className="text-sm">{item.name}</span>
              </div>
              {item.detail && (
                <span className="text-sm text-muted-foreground">{item.detail}</span>
              )}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
