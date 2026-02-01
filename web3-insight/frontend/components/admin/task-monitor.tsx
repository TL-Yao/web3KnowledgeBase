'use client'

import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Eye, XCircle, Clock, CheckCircle, Loader2 } from 'lucide-react'

interface Task {
  id: string
  type: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  description: string
  model?: string
  progress?: number
  startedAt?: string
  completedAt?: string
}

export function TaskMonitor() {
  const { data: tasks } = useQuery({
    queryKey: ['tasks'],
    queryFn: async () => {
      // TODO: Fetch from API
      return [
        {
          id: '1842',
          type: 'content:generate',
          status: 'running' as const,
          description: '生成文章: "EIP-4844详解"',
          model: 'llama3:70b',
          progress: 67,
          startedAt: new Date().toISOString()
        },
        {
          id: '1843',
          type: 'web:crawl',
          status: 'pending' as const,
          description: '爬取 etherscan.io',
        },
        {
          id: '1841',
          type: 'rss:sync',
          status: 'completed' as const,
          description: 'RSS 同步 (zkSync)',
          completedAt: new Date().toISOString()
        }
      ]
    },
    refetchInterval: 5000
  })

  const statusIcons = {
    pending: <Clock className="w-4 h-4 text-muted-foreground" />,
    running: <Loader2 className="w-4 h-4 text-blue-500 animate-spin" />,
    completed: <CheckCircle className="w-4 h-4 text-green-500" />,
    failed: <XCircle className="w-4 h-4 text-red-500" />
  }

  return (
    <ScrollArea className="h-[300px]">
      <div className="space-y-3">
        {tasks?.map((task) => (
          <div
            key={task.id}
            className="flex items-start justify-between p-3 rounded-lg border border-border"
          >
            <div className="flex items-start gap-3">
              {statusIcons[task.status]}
              <div>
                <div className="text-sm font-medium">{task.description}</div>
                {task.model && (
                  <div className="text-xs text-muted-foreground mt-1">
                    模型: {task.model}
                    {task.progress && ` | 进度: ${task.progress}%`}
                  </div>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Button variant="ghost" size="icon" className="w-7 h-7">
                <Eye className="w-4 h-4" />
              </Button>
              {task.status === 'running' && (
                <Button variant="ghost" size="icon" className="w-7 h-7 text-red-500">
                  <XCircle className="w-4 h-4" />
                </Button>
              )}
            </div>
          </div>
        ))}
      </div>
    </ScrollArea>
  )
}
