'use client'

import Link from 'next/link'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Clock, Inbox, Trash2, Loader2 } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { cn } from '@/lib/utils'
import { useState } from 'react'
import { DOMAINS } from './domain-selector'

export type SessionStatus = 'pending' | 'planning' | 'plan_review' | 'researching' | 'writing' | 'completed' | 'failed'

export interface ResearchSession {
  id: string
  question: string
  domain: string
  status: SessionStatus
  stage?: string
  stageDetail?: string
  articleSlug?: string
  error?: string
  createdAt: string
  updatedAt: string
}

const STATUS_CONFIG: Record<SessionStatus, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; className?: string }> = {
  pending: { label: '等待中', variant: 'secondary' },
  planning: { label: '规划中', variant: 'outline', className: 'border-blue-300 text-blue-600 animate-pulse' },
  plan_review: { label: '待审核', variant: 'outline', className: 'border-amber-300 text-amber-600' },
  researching: { label: '研究中', variant: 'outline', className: 'border-blue-300 text-blue-600 animate-pulse' },
  writing: { label: '撰写中', variant: 'outline', className: 'border-blue-300 text-blue-600 animate-pulse' },
  completed: { label: '已完成', variant: 'default' },
  failed: { label: '失败', variant: 'destructive' },
}

interface SessionListProps {
  sessions?: ResearchSession[]
  isLoading?: boolean
  hasMore?: boolean
  onLoadMore?: () => void
  onDelete?: (id: string) => void
  isDeleting?: boolean
}

export function SessionList({
  sessions = [],
  isLoading,
  hasMore,
  onLoadMore,
  onDelete,
  isDeleting,
}: SessionListProps) {
  const [deleteId, setDeleteId] = useState<string | null>(null)

  if (isLoading && sessions.length === 0) {
    return (
      <div className="flex items-center justify-center py-12 text-muted-foreground">
        <div className="animate-pulse">加载中...</div>
      </div>
    )
  }

  if (sessions.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <Inbox className="w-12 h-12 mb-4 opacity-50" />
        <p className="text-lg font-medium">暂无研究记录</p>
        <p className="text-sm">在上方输入问题，开始你的第一次研究</p>
      </div>
    )
  }

  const getDomainName = (domainId: string) => {
    return DOMAINS.find((d) => d.id === domainId)?.name ?? domainId
  }

  const handleConfirmDelete = () => {
    if (deleteId) {
      onDelete?.(deleteId)
      setDeleteId(null)
    }
  }

  return (
    <>
      <ScrollArea className="h-[calc(100vh-400px)]">
        <div className="space-y-3">
          {sessions.map((session) => {
            const statusCfg = STATUS_CONFIG[session.status] ?? STATUS_CONFIG.pending
            const DomainIcon = DOMAINS.find((d) => d.id === session.domain)?.icon

            return (
              <Link key={session.id} href={`/research/${session.id}`}>
                <Card className="hover:bg-muted/50 transition-colors cursor-pointer group">
                  <CardHeader className="pb-2">
                    <div className="flex items-center justify-between gap-2">
                      <CardTitle className="text-base line-clamp-1 flex-1">
                        {session.question}
                      </CardTitle>
                      <div className="flex items-center gap-2 shrink-0">
                        <Badge
                          variant={statusCfg.variant}
                          className={cn('text-xs', statusCfg.className)}
                        >
                          {statusCfg.label}
                        </Badge>
                        {onDelete && (
                          <Button
                            variant="ghost"
                            size="icon"
                            className="size-7 opacity-0 group-hover:opacity-100 transition-opacity"
                            onClick={(e) => {
                              e.preventDefault()
                              e.stopPropagation()
                              setDeleteId(session.id)
                            }}
                          >
                            <Trash2 className="size-3.5 text-muted-foreground" />
                          </Button>
                        )}
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <div className="flex items-center justify-between text-xs text-muted-foreground">
                      <div className="flex items-center gap-2">
                        {DomainIcon && <DomainIcon className="size-3.5" />}
                        <span>{getDomainName(session.domain)}</span>
                      </div>
                      <span className="flex items-center gap-1">
                        <Clock className="size-3" />
                        {formatDistanceToNow(new Date(session.createdAt), {
                          addSuffix: true,
                          locale: zhCN,
                        })}
                      </span>
                    </div>
                  </CardContent>
                </Card>
              </Link>
            )
          })}
        </div>

        {hasMore && (
          <div className="flex justify-center mt-4">
            <Button variant="outline" size="sm" onClick={onLoadMore} disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="size-3.5 mr-1.5 animate-spin" />
                  加载中...
                </>
              ) : (
                '加载更多'
              )}
            </Button>
          </div>
        )}
      </ScrollArea>

      <AlertDialog open={!!deleteId} onOpenChange={(open) => !open && setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              删除后将无法恢复此研究记录及其关联的报告。确定要继续吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmDelete}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? <Loader2 className="size-3.5 mr-1.5 animate-spin" /> : null}
              删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
